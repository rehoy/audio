package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"

// 	"github.com/rehoy/audioplayer/server"
// )

// func main() {
// 	s := server.NewServer()
// 	s.SetupServer("templates")

// 	http.Handle("/style.css", http.FileServer(http.Dir(".")))

// 	fmt.Println("listening on http://localhost:8080/")

// 	go func() {
// 		err := http.ListenAndServe(":8080", nil)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}();

// 	c := make(chan os.Signal, 1)
// 	stop := make(chan bool)

// 	go func() {
// 		<-c
// 		stop <- true
// 	}()

// 	signal.Notify(c, os.Interrupt)
// 	<-c
// 	log.Println("Shutting down...")
// 	s.DB.Close()

// }

import (
	"fmt"
	"github.com/rehoy/audioplayer/db"
)

func main() {

	db := podb.NewDB()

	defer db.Close()
	fmt.Println("Database opened")

	for i := 0; i < 10; i++ {
		series_name, err := db.SeriesNameFromID(i)
		if err != nil {
			continue
		}

		fmt.Println("Series Name: ", series_name)
	}

	podcast_title := "Not Another D&D Podcast"
	db.UpdatePodcast(podcast_title)





	




}
