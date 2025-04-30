// Provides a simple HTTP server to serve the HTML files we have generated.
package server

import (
	"log"
	"net/http"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/joelseq/sqliteadmin-go"
	"github.com/rs/cors"
)

func ServeBacked() {
	db, err := sql.Open("sqlite3", "airtable-export.db")
	if err != nil {
		log.Printf("Server error: %v", err)
	}

	logger := slog.Default()

	config := sqliteadmin.Config{
		DB:       db,
		Username: "user",
		Password: "password",
		Logger:   logger,
	}
	admin := sqliteadmin.New(config)

	mux := http.NewServeMux()

	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			admin.HandlePost(w, r)
		}
	})

	// Configure CORS to allow all origins, methods, and headers
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := corsHandler.Handler(mux)

	s := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(&s, done)

	log.Println("Backend listening on port 8080")

	err = s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}
