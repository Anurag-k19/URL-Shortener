package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"url_shortner/auth"
	"url_shortner/handlers"
	"url_shortner/storage"
)

func newRouter(shortenHandler, redirectHandler, googleLoginHandler, googleCallbackHandler, meHandler, logoutHandler http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/auth/google/login", googleLoginHandler)
	mux.Handle("/auth/google/callback", googleCallbackHandler)
	mux.Handle("/auth/me", meHandler)
	mux.Handle("/auth/logout", logoutHandler)
	mux.Handle("/shorten", shortenHandler)
	mux.Handle("/redirect", redirectHandler)
	mux.Handle("/", redirectHandler)
	return corsMiddleware(auth.Middleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Vary", "Origin")
			if origin == "http://localhost:5500" || origin == "http://127.0.0.1:5500" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
	if err := auth.InitGoogle(context.Background()); err != nil {
		log.Fatal("could not set up Google login: ", err)
	}
	if err := storage.EnsurePartitions(); err != nil {
		log.Fatal("could not create monthly partitions: ", err)
	}
	storage.StartPartitionMaintenance(24 * time.Hour)

	router := newRouter(handlers.ShortenHandler, handlers.RedirectHandler, auth.GoogleLoginHandler, auth.GoogleCallbackHandler, auth.MeHandler, auth.LogoutHandler)
	fmt.Printf("Server is running on port 8000......    \n\n ")
	http.ListenAndServe(":8000", router)
}
