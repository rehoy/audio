package main

import (
	"fmt"
	"net/http"
	"github.com/rehoy/audioplayer/server"
	"time"
	"log"
	

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
	
	index := 0
	for {
		fmt.Println(index)
		index++
		time.Sleep(time.Second * 1)
	}

}
