package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dustin-ward/CYH2021-Backend/data"
	"github.com/gorilla/mux"
)

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}

func main() {
	data.Setup()
	defer data.Takedown()

	r := mux.NewRouter()
	r.HandleFunc("/", greet)
	r.HandleFunc("/users", data.GetAllUsers)
	r.HandleFunc("/users/{id}", data.GetUser)
	fmt.Println("Now serving on 8080...")
	http.ListenAndServe(":8080", r)
}
