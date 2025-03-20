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
	fmt.Println("Hello, World!")

	portnumber := "8080"

	db, err := handler.NewDB("./my.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	
	http.HandleFunc("/", slashHandler)
	http.ListenAndServe(":" + portnumber, nil)
}