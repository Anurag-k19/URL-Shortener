package main

import (
	"fmt"
	"net/http"
	"url_shortner/handlers"
	"url_shortner/storage"
)

func newRouter(shortenHandler, redirectHandler http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", shortenHandler)
	mux.HandleFunc("/redirect", redirectHandler)
	mux.HandleFunc("/", redirectHandler)
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Vary", "Origin")
			if origin == "http://localhost:5500" || origin == "http://127.0.0.1:5500" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// main starts the HTTP server and registers route handlers
func main() {
	storage.InitDB()
	storage.InitRedis()

	router := newRouter(handlers.Shortenhandler, handlers.RedirectHandler)
	fmt.Printf("Server is running on port 8000......    \n\n ")
	http.ListenAndServe(":8000", router)
}
