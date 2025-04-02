package server

import (
	"log"
	"net/http"
)

func Serve() {
	log.Println("Serving current directory at http://localhost:8080")

	err := http.ListenAndServe(":8080", http.FileServer(http.Dir("./out")))
	if err != nil {
		log.Printf("Server error: %v", err)
	}
}
