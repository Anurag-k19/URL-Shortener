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

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", username, dbpassword, dbname, sslmode)
	fmt.Println("Conn with:  ", connStr)
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
