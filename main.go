package main

import (
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	"github.com/rehoy/audio/handler"
)



func main() {
	fmt.Println("Hello, World!")

	db, err := handler.GetDB("./my.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	episode, err := handler.QueryRowById(10, db)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(episode.Title, episode.Series_id, episode.Episode_id, len(episode.Audio))


	// insertFolder("beef", db)



}
