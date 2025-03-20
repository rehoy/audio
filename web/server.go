package main

import (
	"fmt"
	"net/http"
	"github.com/rehoy/audio/handler"
)

func slashHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}
func main() {
	portnumber := "8080"

	db, err := handler.NewDB("./my.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()



	// http.HandleFunc("/", slashHandler)
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)
	http.HandleFunc("/episode", db.HandleEpisode)
	fmt.Println("Server listening on port", portnumber)
	http.ListenAndServe(":" + portnumber, nil)
}