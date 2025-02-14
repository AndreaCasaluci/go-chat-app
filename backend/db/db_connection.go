package database

import (
	"database/sql"
	"fmt"
	"github.com/AndreaCasaluci/go-chat-app/utils"
	_ "github.com/lib/pq"
	"log"
)

var DB *sql.DB

func connect() (*sql.DB, error) {
	config, err := utils.GetConfig()
	if err != nil {
		return nil, err
	}

	if config.DbHost == "" || config.DbUsername == "" || config.DbPassword == "" || config.DbName == "" {
		log.Fatal("Error: Missing required database environment variables")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.DbUsername, config.DbPassword, config.DbHost, config.DbPort, config.DbName)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v\n", err)
		return nil, err
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error pinging the database: %v\n", err)
		return nil, err
	}

	log.Println("Successfully connected to the database.")
	return DB, nil
}

func GetDb() (*sql.DB, error) {
	if DB == nil {
		return connect()
	} else {
		return DB, nil
	}
}
