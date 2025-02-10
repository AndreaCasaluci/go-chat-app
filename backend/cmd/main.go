package main

import (
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/handlers"
	"log"
	"net/http"
	"os"

	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/gorilla/mux"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Service is running")
	}).Methods("GET")

	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Starting Go Chat App...")
	fmt.Println(`
 ________  ________          ________  ___  ___  ________  _________   
|\   ____\|\   __  \        |\   ____\|\  \|\  \|\   __  \|\___   ___\ 
\ \  \___|\ \  \|\  \       \ \  \___|\ \  \\\  \ \  \|\  \|___ \  \_| 
 \ \  \  __\ \  \\\  \       \ \  \    \ \   __  \ \   __  \   \ \  \  
  \ \  \|\  \ \  \\\  \       \ \  \____\ \  \ \  \ \  \ \  \   \ \  \ 
   \ \_______\ \_______\       \ \_______\ \__\ \__\ \__\ \__\   \ \__\
    \|_______|\|_______|        \|_______|\|__|\|__|\|__|\|__|    \|__|` + "\n")

	log.Printf("Server started on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
