package main

import (
	"fmt"
	"net/http"
	"url_shortner/handlers"
	"url_shortner/storage"
)

// main starts the HTTP server and registers route handlers
func main() {
	storage.InitDB()
	// Register POST /shorten endpoint for creating short URLs
	http.HandleFunc("/shorten", handlers.Shortenhandler)
	// Register GET / endpoint for redirecting to original URLs
	http.HandleFunc("/", handlers.RedirectHandler)
	fmt.Printf("Server is running on port 8000......    \n\n ")
	http.ListenAndServe(":8000", nil)
}
