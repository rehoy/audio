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
	Episode_id  int    `json:"episode_id"`
	Title       string `json:"title"`
	Pubdate     string `json:"pubdate"`
	Description string `json:"description"`
	AudioURL    string `json:"audioURL"`
	ImageURL    string `json:"imageURL"`
}
type Podcast struct {
	Title       string
	Description string
	Episodes    map[string]Episode
}

func (db *DB) Check() {
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

func (db *DB) AddPodcastToDB(jsonPath, name, database string) error {

	var podcasts map[string]Podcast

	file, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}

	err = json.NewDecoder(file).Decode(&podcasts)
	if err != nil {
		return fmt.Errorf("Error decoding JSON: %v", err)
	}

	podcast, ok := podcasts[strings.Trim(name, " ")]
	if !ok {
		return fmt.Errorf("Podcast not found: %v", name)
	}

	fmt.Println("podcast title:", podcast.Title, ", number of episodes:", len(podcast.Episodes))

	id, err := db.insertSeries(podcast.Title, podcast.Description)
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
func (db *DB) insertSeries(title, description string) (int, error) {
	res, err := db.conn.Exec("INSERT INTO series (title, description) VALUES (?, ?)", title, description)
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

func (db *DB) seriesNameFromID(seriesID int) (string, error) {
	var title string
	query := "SELECT title FROM series WHERE id = ?"
	err := db.conn.QueryRow(query, seriesID).Scan(&title)
	if err != nil {
		return "", fmt.Errorf("Error querying series name: %v", err)
	}
	return title, nil
}

func (db *DB) GetEpisodesFromSeries(args ...any) ([]Episode, error) {
	var seriesID int
	var err error

	switch args[0].(type) {
	case string:
		seriesID, err = db.getSeriesIDByName(args[0].(string))
		if err != nil {
			return nil, fmt.Errorf("Error getting series ID by name: %v", err)
		}
	case int:
		seriesID = args[0].(int)
	default:
		return nil, fmt.Errorf("Invalid argument type: %T", args[0])
	}

	query := "SELECT episode_id, title, pubdate, description, audiourl, imageurl FROM episodes WHERE series_id = ?"

	fmt.Println("seriesID:", seriesID, "arg:", args[0].(string))
	rows, err := db.conn.Query(query, seriesID)
	if err != nil {
		return nil, fmt.Errorf("Error querying episodes: %v", err)
	}
	defer rows.Close()

	episodes := make([]Episode, 0)
	for rows.Next() {
		var episode Episode
		var pubdate string
		err := rows.Scan(&episode.Episode_id, &episode.Title, &pubdate, &episode.Description, &episode.AudioURL, &episode.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("Error scanning episode: %v", err)
		}
		episode.Pubdate = pubdate
		episodes = append(episodes, episode)
	}

	return episodes, nil
}

func (db *DB) getSeriesIDByName(seriesName string) (int, error) {
	var seriesID int
	query := "SELECT series_id FROM series WHERE title = ?"
	err := db.conn.QueryRow(query, seriesName).Scan(&seriesID)
	if err != nil {
		return 0, fmt.Errorf("Error querying series ID by name: %v", err)
	}
	return seriesID, nil
}

func (db *DB) GetEpisode(id int) Episode {
	query := "SELECT episode_id, title, pubdate, description, audiourl, imageurl FROM episodes WHERE episode_id = ?"
	row := db.conn.QueryRow(query, id)

	var episode Episode
	err := row.Scan(&episode.Episode_id, &episode.Title, &episode.Pubdate, &episode.Description, &episode.AudioURL, &episode.ImageURL)
	if err != nil {
		fmt.Println("Error getting episode:", err)
		return Episode{}
	}
	return episode
}

func (db *DB) GetSeries() ([]string, error) {
	query := "SELECT title FROM series"
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Error querying series: %v", err)
	}
	defer rows.Close()

	var series []string 

	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			return nil, fmt.Errorf("Error scanning series title: %v", err)
		}

		series = append(series, title)
	}

	return series, nil
}

//package dbhandler
// package main

// import (
// 	"github.com/rehoy/audioplayer/db"
// 	"flag"
// 	"fmt"
// )

// func main(){

// 	db := podb.NewDB()
// 	defer db.Close()

// 	db.Check()

// 	name := flag.String("name", "", "Name of podcast to add to database")
// 	rss := flag.String("rss", "", "RSS feed URL")

// 	flag.Parse()

// 	if *name != "" {
// 		err := db.AddPodcastToDB("podcasts.json", *name, "pod.db")
// 		if err != nil {
// 			fmt.Println("Error adding podcast to database:", err)
// 			return
// 		}
// 	}

// 	if *rss != "" {

// 		err := podb.WritePodcastsToJSON(*rss, "podcasts.json")
// 		if err != nil {
// 			fmt.Println("Error writing podcasts to JSON:", err)
// 			return
// 		}
// 		fmt.Println("Added rss to podcasts.json", )
// 	}
//}
