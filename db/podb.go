package podb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	// "sync"
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
	Series_id int
	Title       string
	Description string
	Episodes    []Episode
	RssFeed     string
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

func (db *DB) AddPodcastToDB(rssFeed, database string) error {

	podcast, err := db.PodcastFromFeed(rssFeed)
	if err != nil {
		return fmt.Errorf("failed to parse podcast feed: %w", err)
	}
	// Check if the podcast already exists in the database
	existingPodcasts, err := db.GetSeries()
	if err != nil {
		return fmt.Errorf("failed to get existing podcasts: %w", err)
	}
	// Check if the podcast title already exists in the database
	for _, existingPodcast := range existingPodcasts {
		if strings.EqualFold(existingPodcast, podcast.Title) {
			fmt.Printf("Podcast '%s' already exists in the database.\n", podcast.Title)
			return nil
		}
	}

	fmt.Println("podcast title:", podcast.Title, ", number of episodes:", len(podcast.Episodes))

	id, err := db.insertSeries(podcast.Title, podcast.Description, podcast.RssFeed)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	go func(){
		for _, episode := range podcast.Episodes {
			id, err := db.insertEpisode(episode, id)
			if err != nil {
				fmt.Println("Error inserting episode:", err)
			} else {
				fmt.Printf("Inserted episode: %s into series: %s with id: %d\n", episode.Title, podcast.Title, id)
			}
		}
	}();



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
func (db *DB) insertSeries(title, description, rssFeed string) (int, error) {
	res, err := db.conn.Exec("INSERT INTO series (title, description, feedurl) VALUES (?, ?, ?)", title, description, rssFeed)
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
	layouts := []string{
		time.RFC3339,                     // "2006-01-02T15:04:05Z07:00"
		time.RFC1123Z,                    // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                     // "Mon, 02 Jan 2006 15:04:05 MST"
		"Mon, 2 Jan 2006 15:04:05 -0700", // non-padded day with numeric tz
		"Mon, 2 Jan 2006 15:04:05 MST",   // non-padded day with abbreviated tz
	}
	var parsedTime time.Time
	var err error
	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, pubdate)
		if err == nil {
			return parsedTime, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse pubdate %q: %v", pubdate, err)
}

func (db *DB) PodcastFromFeed(feedURL string) (Podcast, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return Podcast{}, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	// Use the itunes:image tag from the channel as the default image
	defaultImageURL := ""
	if feed.ITunesExt != nil && feed.ITunesExt.Image != "" {
		defaultImageURL = feed.ITunesExt.Image
	}

	podcast := Podcast{
		Title:       feed.Title,
		Description: feed.Description,
		Episodes:    make([]Episode, 0),
		RssFeed:     feedURL,
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
		podcast.Episodes = append(podcast.Episodes, episode)
	}

	return podcast, nil

}

func (db *DB) WritePodcastsToJSON(feedURL, jsonPath string) error {

	podcast, err := db.PodcastFromFeed(feedURL)
	if err != nil {
		return fmt.Errorf("failed to parse podcast feed: %w", err)
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

func (db *DB) SeriesNameFromID(seriesID int) (string, error) {
	var title string
	query := "SELECT title FROM series WHERE series_id = ?"
	err := db.conn.QueryRow(query, seriesID).Scan(&title)
	if err != nil {
		return "", fmt.Errorf("Error querying series name: %v, seriesID: %d", err, seriesID)
	}
	return title, nil
}

func (db *DB) GetEpisodesFromSeries(args ...any) ([]Episode, error) {
	var seriesID int
	var err error

	// Extract sorting parameters
	sortColumn := "pubdate"
	sortOrder := "ASC"
	if len(args) > 1 {
		if col, ok := args[1].(string); ok && col != "" {
			sortColumn = col
		}
	}
	if len(args) > 2 {
		if order, ok := args[2].(string); ok && (strings.ToUpper(order) == "ASC" || strings.ToUpper(order) == "DESC") {
			sortOrder = strings.ToUpper(order)
		}
	}

	// Determine series ID based on the first argument
	switch args[0].(type) {
	case string:
		seriesID, err = db.GetSeriesIDByName(args[0].(string))
		if err != nil {
			return nil, fmt.Errorf("Error getting series ID by name: %v", err)
		}
	case int:
		seriesID = args[0].(int)
	default:
		return nil, fmt.Errorf("Invalid argument type: %T", args[0])
	}

	// Construct query with sorting
	query := fmt.Sprintf("SELECT episode_id, title, pubdate, description, audiourl, imageurl FROM episodes WHERE series_id = ? ORDER BY %s %s", sortColumn, sortOrder)

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

func (db *DB) sortChronologically(episodes []Episode, reverse ...bool) {
	sort.Slice(episodes, func(i, j int) bool {
		ti, err1 := db.pubdateToTimeStamp(episodes[i].Pubdate)
		tj, err2 := db.pubdateToTimeStamp(episodes[j].Pubdate)

		if err1 != nil || err2 != nil {
			fmt.Println(err1, err2)
			if len(reverse) > 0 && reverse[0] {
				return episodes[i].Pubdate > episodes[j].Pubdate
			}
			return episodes[i].Pubdate < episodes[j].Pubdate
		}

		if len(reverse) > 0 && reverse[0] {
			return ti.After(tj)
		}
		return ti.Before(tj)
	})
}

func sortAlphabetically(episodes []Episode) []Episode {
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Title < episodes[j].Title
	})
	return episodes
}

func (db *DB) GetSeriesIDByName(seriesName string) (int, error) {
	var seriesID int
	query := "SELECT series_id FROM series WHERE title = ?"
	err := db.conn.QueryRow(query, seriesName).Scan(&seriesID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no series found with name %q", seriesName)
		}
		return 0, fmt.Errorf("error querying series ID by name: %v", err)
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

func (db *DB) DeleteSeries(series_id int) error {
	query := "DELETE FROM series WHERE series_id = ?"
	_, err := db.conn.Exec(query, series_id)
	if err != nil {
		fmt.Println("Error deleting series:", err)
		return fmt.Errorf("Error deleting series: %v", err)
	}
	fmt.Println("Deleted series with ID:", series_id)
	query = "DELETE FROM episodes WHERE series_id = ?"
	_, err = db.conn.Exec(query, series_id)
	if err != nil {
		fmt.Println("COuld not delete episodes:", err)
		return fmt.Errorf("Error deleting episodes: %v", err)
	}
	return nil
}

func (db *DB) compareEpisodeList(newRssFeed string) ([]Episode, error) {
    newPodcast, err := db.PodcastFromFeed(newRssFeed)
    if err != nil {
        return nil, fmt.Errorf("failed to parse podcast feed: %w", err)
    }

    newEpisodes := make([]Episode, 0)
    oldEpisodes, _ := db.GetEpisodesFromSeries(newPodcast.Title)

    for _, newEpisode := range newPodcast.Episodes {
        found := false
        for _, oldEpisode := range oldEpisodes {
            if strings.EqualFold(newEpisode.Title, oldEpisode.Title) {
                found = true
                break
            }
        }
        if !found {
            newEpisodes = append(newEpisodes, newEpisode)
        }
    }

    return newEpisodes, nil
}

func (db *DB) FindNewEpisodes(rssFeed string) ([]Episode, error) {
	newPodcast, err := db.PodcastFromFeed(rssFeed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse podcast feed: %w", err)
	}

	seriesID, err := db.GetSeriesIDByName(newPodcast.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to get series ID: %w", err)
	}

	existingEpisodes, err := db.GetEpisodesFromSeries(seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing episodes: %w", err)
	}

	newEpisodes := make([]Episode, 0)
	for _, newEpisode := range newPodcast.Episodes {
        found := false
        for _, existingEpisode := range existingEpisodes {
            if strings.EqualFold(newEpisode.Title, existingEpisode.Title) {
                found = true
                break
            }
        }
        if !found {
            newEpisodes = append(newEpisodes, newEpisode)
        }
	}

	return newEpisodes, nil
}

func (db *DB) AddNewEpisodes(rssFeed string) ([]Episode, error) {
	newEpisodes, err := db.FindNewEpisodes(rssFeed)

	if err != nil {
		return nil, err
	}

	podcast, _ := db.PodcastFromFeed(rssFeed)
	seriesID, err := db.GetSeriesIDByName(podcast.Title)

	    // 5. Insert the new episodes into the database
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, episode := range newEpisodes {
		_, err := tx.Exec("INSERT INTO episodes (title, pubdate, description, audiourl, imageurl, series_id) VALUES (?, ?, ?, ?, ?, ?)", episode.Title, episode.Pubdate, episode.Description, episode.AudioURL, episode.ImageURL, seriesID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert episode: %w", err)
		}
	}

	 

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	for i, episode := range newEpisodes {
		var id int
		err := db.conn.QueryRow("SELECT episode_id FROM episodes WHERE title = ? AND series_id = ?", episode.Title, seriesID).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch episode ID: %w", err)
		}
		newEpisodes[i].Episode_id = id
	}

	return newEpisodes, nil
}

func (db *DB) getSeriesMap() (map[string]Podcast, error) {

	seriesMap := make(map[string]Podcast)

	series, err := db.GetSeries()

	if err != nil {
		fmt.Println("Error getting series:", err)
		return nil, err
	}

	for _, seriesName := range series {
		seriesID, err := db.GetSeriesIDByName(seriesName)
		if err != nil {
			return nil, err
		}

		var podcast Podcast
		query := "select series_id, title, description, feedurl from series where series_id = ?"
		err = db.conn.QueryRow(query, seriesID).Scan(&podcast.Series_id, &podcast.Title, &podcast.Description, &podcast.RssFeed)

		if err != nil {
			return nil, err
		}

		seriesMap[podcast.Title] = podcast
	}

	return seriesMap, nil

}

func (db *DB) GetPodcast(id int) (Podcast, error) {
	query := "SELECT series_id, title, description, feedurl FROM series WHERE series_id = ?"
	row := db.conn.QueryRow(query, id)

	var podcast Podcast
	err := row.Scan(&podcast.Series_id, &podcast.Title, &podcast.Description, &podcast.RssFeed)
	if err != nil {
		fmt.Println("Error getting series:", err)
		return Podcast{}, err
	}

	episodes, err := db.GetEpisodesFromSeries(id)
	if err != nil {
		fmt.Println("Error getting episodes:", err)
		return Podcast{}, err
	}
	podcast.Episodes = episodes

	return podcast, nil
}