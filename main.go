package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/rehoy/audioplayer/server"
)

func main() {
	s := server.NewServer()
	s.SetupServer("templates")

	http.Handle("/style.css", http.FileServer(http.Dir(".")))

	fmt.Println("listening on http://localhost:8080/")

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}();

	c := make(chan os.Signal, 1)
	stop := make(chan bool)

	go func() {
		<-c
		stop <- true
	}()
	
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Shutting down...")
	s.DB.Close()

}
