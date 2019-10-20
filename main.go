package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	db *sql.DB
)

func main() {
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	err = godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Error loading config.env file")
	}

	dbUser := os.Getenv("db_user")
	dbPass := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err = sql.Open("postgres", psqlInfo)

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/getPost", getPost).Methods("GET")
	r.HandleFunc("/api/v1/getPosts", getPosts).Methods("GET")
	r.HandleFunc("/api/v1/addPost", addPost).Methods("POST")
	r.HandleFunc("/api/v1/updatePost", updatePost).Methods("PUT")
	r.HandleFunc("/api/v1/deletePost", deletePost).Methods("DELETE")

	server := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":8080",
		Handler:      r,
	}

	fmt.Printf("Server successfully started at port %v\n", server.Addr)
	log.Println(server.ListenAndServe())
}
