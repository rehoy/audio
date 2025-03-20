package handler

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/rehoy/audio/processor"
)

type Episode struct {
	Episode_id int
	Title      string
	Audio      []byte
	Series_id  int
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
	series_id, err := db.getSeriesIDFromName("beef", "series")
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

func (db *DB) getSeriesIDFromName(series, table string) (int, error) {
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
