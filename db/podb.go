package podb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"


	_ "github.com/glebarez/go-sqlite"
	"github.com/mmcdole/gofeed"
)

type Episode struct {
	Title string `json:"title"`
	Pubdate string `json:"pubdate"`
	Description string `json:"description"`
	AudioURL string `json:"audioURL"`
	ImageURL string `json:"imageURL"`
}
type Podcast struct {
	Title       string
	Description string
	Episodes    map[string]Episode
}

func (db *DB)Check() {
	var version string


	err := db.conn.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		fmt.Println("Error querying database version", err)
		return
	} 
	fmt.Println("SQLite version:", version)
} 

type DB struct {
	conn *sql.DB
}
func (db *DB) Close() {
	db.conn.Close()
}

func NewDB() *DB {
	conn, err := sql.Open("sqlite", "pod.db")
	if err != nil {
		fmt.Println("Error opening database", err)
		log.Fatal()
	}
	return &DB{conn: conn}
}


func (db *DB)AddPodcastToDB(jsonPath, name, database string) error {

	var podcasts map[string]Podcast

	file, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}

	err = json.NewDecoder(file).Decode(&podcasts)
	if err != nil {
		return fmt.Errorf("Error decoding JSON: %v", err)
	}

	for key:=range podcasts{
		fmt.Println(key)
	}



	podcast, ok := podcasts[strings.Trim(name , " ")]
	if !ok {
		return fmt.Errorf("Podcast not found: %v", name)
	}

	fmt.Println("podcast title:", podcast.Title, ", number of episodes:", len(podcast.Episodes))

	id, err := db.insertSeries(podcast.Title)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	for _, episode := range podcast.Episodes {
		id, err := db.insertEpisode(episode, id)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		fmt.Printf("Inserted episode: %s into series: %s with id: %d\n", episode.Title, podcast.Title, id)
	}

	return nil
}

func (db *DB) insertEpisode(episode Episode, seriesID int) (int, error) {
	query := "INSERT INTO episodes (title, pubdate, description, audiourl, imageurl, series_id) VALUES (?, ?, ?, ?, ?, ?)"

	timestamp, err := db.pubdateToTimeStamp(episode.Pubdate)
	if err != nil {
		return 0, fmt.Errorf("Error converting pubdate to timestamp: %v", err)
	}
	// timestamp.Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec(query, episode.Title, timestamp.Format("2006-01-02 15:04:05"), episode.Description, episode.AudioURL, episode.ImageURL, seriesID)
	if err != nil {
		return 0, fmt.Errorf("Error inserting episode: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Error getting last insert ID: %v", err)
	}

	return int(id), nil

}
func (db *DB) insertSeries(title string) (int, error) {
	res, err := db.conn.Exec("INSERT INTO series (title) VALUES (?)", title)
	if err != nil {
		return 0, fmt.Errorf("Error inserting series: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Error getting last insert ID: %v", err)
	}

	return int(id), nil
}

func (db *DB) pubdateToTimeStamp(pubdate string) (time.Time, error) {
	layout := time.RFC1123Z
	var parsedTime time.Time
	parsedTime, err := time.Parse(layout, pubdate)

	if err != nil {
		parsedTime, err = time.Parse(time.RFC1123, pubdate)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return time.Time{}, err
		}
	}

	return parsedTime, nil

}

func WritePodcastsToJSON(feedURL, jsonPath string) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	// Use the itunes:image tag from the channel as the default image
	defaultImageURL := ""
	if feed.ITunesExt != nil && feed.ITunesExt.Image != "" {
		defaultImageURL = feed.ITunesExt.Image
	}

	podcast := Podcast{
		Title:       feed.Title,
		Description: feed.Description,
		Episodes:    make(map[string]Episode),
	}

	for _, item := range feed.Items {
		imageURL := defaultImageURL
		if item.ITunesExt != nil && item.ITunesExt.Image != "" {
			imageURL = item.ITunesExt.Image
		}

		episode := Episode{
			Title:       item.Title,
			Pubdate:     item.Published,
			Description: item.Description,
			AudioURL:    item.Enclosures[0].URL,
			ImageURL:    imageURL,
		}
		podcast.Episodes[episode.Title] = episode
	}

	// Read existing data from podcasts.json
	existingData := make(map[string]Podcast)
	file, err := os.Open(jsonPath)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&existingData); err != nil {
			return fmt.Errorf("failed to decode existing JSON: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to open JSON file: %w", err)
	}

	// Append the new podcast data
	existingData[podcast.Title] = podcast

	// Write the updated data back to podcasts.json
	file, err = os.Create("podcasts.json")
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(existingData); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}





