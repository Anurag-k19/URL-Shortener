package auth

import (
	"fmt"
	"net/http"
)

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Hello")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {}
