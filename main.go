package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
	"GO-REST-API/api"
)

var db *sql.DB

func main() {
	f, err := openLogfile()
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config.env file: %v", err)
	}

	db, err := initDatabase(config)
	if err = db.Ping(); err != nil {
		log.Fatalf("Error initializing DB: %v", err)
	}

	server := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":8080",
	}

	apiHandler := api.New(db)
	http.Handle("/api/", apiHandler)

	log.Printf("Server successfully started at port %v\n", server.Addr)
	log.Println(server.ListenAndServe())
}

// Config represents structure of the config.env
type Config struct {
	dbUser string
	dbPass string
	dbName string
	dbHost string
	dbPort string
}

func openLogfile() (f *os.File, err error) {
	f, err = os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return f, err
}

func loadConfig() (config *Config, err error) {
	err = godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Error loading config.env file")
	}

	config = &Config {
		dbUser : os.Getenv("db_user"),
		dbPass : os.Getenv("db_pass"),
		dbName : os.Getenv("db_name"),
		dbHost : os.Getenv("db_host"),
		dbPort : os.Getenv("db_port"),
	}
	return config, err
}

func initDatabase(c *Config) (db *sql.DB, err error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.dbHost, c.dbPort, c.dbUser, c.dbPass, c.dbName)

	db, err = sql.Open("postgres", psqlInfo)
	return db, err
}