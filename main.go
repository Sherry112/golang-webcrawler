package main

import (
	"log"
	"net/http"

	"github.com/Sherry112/go-webcrawler/handlers" // Adjust the import path accordingly
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})
	http.HandleFunc("/analyze", handlers.AnalyzeHandler)
	http.HandleFunc("/sse", handlers.SSE.SSEHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
