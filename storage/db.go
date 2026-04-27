package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	err1 := godotenv.Load()
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}
	username := os.Getenv("DB_USER")
	dbname := os.Getenv("DBNAME")
	dbpassword := os.Getenv("PASSWORD")
	sslmode := os.Getenv("SSLMODE")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, username, dbpassword, dbname, sslmode)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Error connecting DB:", err)
	}

	log.Println("Connected to PostgreSQL ✅")

}
