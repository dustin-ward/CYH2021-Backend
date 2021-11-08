package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dustin-ward/CYH2021-Backend/auth"
	"github.com/dustin-ward/CYH2021-Backend/data"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}

func main() {
	// Setup DB connections
	data.Setup()
	defer data.Takedown()

	// Setup token expiration system
	go auth.CleanTokens()

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/", greet)
	r.HandleFunc("/users", data.GetAllUsers).Methods("GET")
	r.HandleFunc("/users/{id}", data.GetUser).Methods("GET")
	r.HandleFunc("/register", auth.CreateUser).Methods("POST")
	r.HandleFunc("/login", auth.Login).Methods("POST")
	r.HandleFunc("/logout", auth.Logout)
	fmt.Println("Now serving on 8080...")

	corsWrapper := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})

	http.ListenAndServe(":8080", corsWrapper.Handler(r))
}
