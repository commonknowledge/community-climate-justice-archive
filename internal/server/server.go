// Provides a simple HTTP server to serve the HTML files we have generated.
package server

import (
	"log"
	"net/http"
)

// Serve starts a simple HTTP server to serve the HTML files we have generated on port 8080.
func Serve() {
	log.Println("Serving current directory at http://localhost:8080, use Ctrl+C to stop")

	err := http.ListenAndServe(":8080", http.FileServer(http.Dir("./out")))
	if err != nil {
		log.Printf("Server error: %v", err)
	}
}
