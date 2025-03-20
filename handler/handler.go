package handler

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/rehoy/audio/processor"

	"net/http"
	"encoding/json"
)

type Episode struct {
	Episode_id int    `json:"episode_id"`
	Title     string `json:"title"`
	Audio     []byte `json:"audio"`
	Series_id  int    `json:"series_id"`
}

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func (db *DB) InsertSeries(series string) (sql.Result, error) {
	return db.conn.Exec("INSERT INTO series (name) VALUES (?)", series)
}

func (db *DB) insertEpisode(episode Episode) (sql.Result, error) {
	return db.conn.Exec("INSERT INTO episodes (name, audio, series_id) VALUES (?, ?, ?)", episode.Title, episode.Audio, episode.Series_id)
}

func (db *DB) insertFolder(folder string) error {
	series_id, err := db.getIDFromName("beef", "series")
	if err != nil {
		fmt.Println(err)
		return err
	}

	mp3Files, mp3Blobs, err := processor.ReadMP3Files(folder)
	if err != nil {
		log.Println("Failed to read MP3 files", err)
		return fmt.Errorf("Failed to read MP3 files %v", err)
	}

	count_sucessful := 0

	for _, mp3File := range mp3Files {
		episode := Episode{
			Title:     mp3File,
			Audio:     mp3Blobs[mp3File],
			Series_id: series_id,
		}
		_, err := db.insertEpisode(episode)
		if err != nil {
			log.Printf("Failed to insert episode %s: %v\n", episode.Title, err)
		}
		count_sucessful++
		fmt.Printf("Inserted episode %s to table episodes\n", episode.Title)
	}

	fmt.Println("Inserted ", count_sucessful, " episodes out of total", len(mp3Files))
	return nil
}

func (db *DB) getIDFromName(series, table string) (int, error) {
	var series_id int
	err := db.conn.QueryRow(fmt.Sprintf("SELECT series_id FROM %s WHERE name = ?", table), series).Scan(&series_id)
	if err != nil {
		return 0, err
	}

	return series_id, nil
}

func (db *DB) QueryRowById(id int) (*Episode, error) {
	episode := &Episode{}

	query := `SELECT episode_id, name, series_id, audio from episodes WHERE episode_id = ?`

	row := db.conn.QueryRow(query, id)
	err := row.Scan(&episode.Episode_id, &episode.Title, &episode.Series_id, &episode.Audio)
	if err != nil {
		return nil, fmt.Errorf("No episode found with id %d", id)
	}

	return episode, nil

}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) HandleEpisode(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	// id := query.Get("id")
	name := query.Get("name")

	log.Println("Request for episode", name)

	episode_id, err := db.getIDFromName(name, "episodes")
	log.Println("episode_id", episode_id)

	if err != nil {
		log.Println("Failed to get episode id", err, name)
		http.Error(w, "Failed to get episode id", http.StatusInternalServerError)
		return
	}

	episode, err := db.QueryRowById(episode_id)
	if err != nil {
		log.Println("Failed to query episode", err, episode_id)
		http.Error(w, "Failed to query episode", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Write(episode.Audio)

}

func (db *DB) HandlePodcast(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	title := query.Get("title")

	if title == "" {
		http.Error(w, "Missing title", http.StatusBadRequest)
		log.Println("Missing title for podcast request")
		return
	}

	series_id, err := db.getIDFromName(title, "series")
	if err != nil {
		log.Println("Failed to get series id", err, title)
		http.Error(w, "Failed to get series id", http.StatusInternalServerError)
		return
	}

	rows, err := db.conn.Query("SELECT episode_id, name, series_id, audio from episodes WHERE series_id = ?", series_id)
	if err != nil {
		log.Println("Failed to query episodes", err)
		http.Error(w, "Failed to query episodes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	episodes := make(map[string]*Episode)

	for rows.Next() {
		episode := &Episode{}
		err := rows.Scan(&episode.Episode_id, &episode.Title, &episode.Series_id, &episode.Audio)
		if err != nil {
			log.Println("Failed to scan episode", err)
			http.Error(w, "Failed to scan episode", http.StatusInternalServerError)
			return
		}

		episodes[episode.Title] = episode

	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)

}
