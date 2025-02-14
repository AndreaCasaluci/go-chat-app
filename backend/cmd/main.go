package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndreaCasaluci/go-chat-app/db"
	"github.com/AndreaCasaluci/go-chat-app/handlers"
	"github.com/AndreaCasaluci/go-chat-app/middleware"
	"github.com/gorilla/mux"
)

func main() {
	db, err := database.GetDb()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := RunServer(port)
	GracefulShutdown(server, db, 10*time.Second)
}

func GracefulShutdown(server *http.Server, db *sql.DB, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	if db != nil {
		log.Println("Closing database connection...")
		err := db.Close()
		if err != nil {
			log.Fatalf("Could not close database connection: %v", err)
		}
	}

	log.Println("Server gracefully stopped")
}

func RunServer(port string) *http.Server {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, "Service is running")
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}).Methods("GET")

	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	r.HandleFunc("/users/{uuid}", middleware.JWTMiddleware(handlers.UpdateUser)).Methods("PATCH")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
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

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not start server: %v\n", err)
		}
	}()

	return server
}
