package main

import (
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	"github.com/rehoy/audio/handler"
	"github.com/rehoy/audio/processor"
)



func main() {
	fmt.Println("Hello, World!")

	db, err := handler.NewDB("./my.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	episode, err := db.QueryRowById(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(episode.Title, episode.Series_id, episode.Episode_id, len(episode.Audio))

	err = processor.EncodeToMP3(episode.Audio, "output/output.mp3")
	if err != nil {
		fmt.Println("failed to encode mp3", err)
		return
	}
	fmt.Println("encoded MP3")
	

}
