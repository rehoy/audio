package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/rehoy/audio/processor"
)

type Episode struct {
	Episode_id int
	Title string
	Audio []byte
	Series_id int
}

func main() {

	// db, err := sql.Open("sqlite", "./my.db")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// insertFolder("beef", db)



}

func insertEpisode(episode Episode, db *sql.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO episodes (name, audio, series_id) VALUES (?, ?, ?)", episode.Title, episode.Audio, episode.Series_id)
}

func insertFolder(folder string, db *sql.DB) error {

	series_id, err := getSeriesIDFromName("beef", "series", db)
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
			Title: mp3File,
			Audio: mp3Blobs[mp3File],
			Series_id: series_id,
		}
		_, err := insertEpisode(episode, db)
		if err != nil {
			log.Printf("Failed to insert episode %s: %v\n", episode.Title, err)
		}
		count_sucessful++
		fmt.Printf("Inserted episode %s to table episodes\n", episode.Title)
	}

	fmt.Println("Inserted ", count_sucessful, " episodes out of total", len(mp3Files))
	return nil
}

func getSeriesIDFromName(series, table string, db *sql.DB) (int, error) {
	var series_id int
	err := db.QueryRow(fmt.Sprintf("SELECT series_id FROM %s WHERE name = ?", table), series).Scan(&series_id)
	if err != nil {
		return 0, err
	}

	return series_id, nil
}