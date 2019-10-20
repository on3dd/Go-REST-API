package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Post struct {
	Id        int    `json:"id"`
	Author    int    `json:"author"`
	Posted_at string `json:"posted_at"`
	Title     string `json:"title"`
	Text      string `json:"text"`
}

// getPost gets single post from DB by id
func getPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		return
	}

	row := db.QueryRow("SELECT * FROM posts WHERE id = $1", id)

	var post Post
	err := row.Scan(&post.Id, &post.Author, &post.Posted_at, &post.Title, &post.Text)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{'status':'error'}"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		log.Fatal(err)
	}
}

// getPosts gets all posts from DB
func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rows *sql.Rows
	var err error
	num := r.FormValue("num")
	if num == "" {
		rows, err = db.Query("SELECT * FROM posts ORDER BY posted_at DESC")
	} else {
		rows, err = db.Query("SELECT * FROM posts ORDER BY posted_at DESC LIMIT $1", num)
	}
	if err != nil {
		log.Fatal(err)
	}

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.Id, &post.Author, &post.Posted_at, &post.Title, &post.Text)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{'status':'error'}"))
			log.Fatal(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	if len(posts) == 0 {
		err = json.NewEncoder(w).Encode(make([]Post, 0))
	} else {
		err = json.NewEncoder(w).Encode(posts)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}
}

// addPost adds a new post to DB
func addPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		log.Fatal(err)
	}

	var num int
	id := db.QueryRow("SELECT id FROM posts ORDER BY id DESC LIMIT 1")
	err = id.Scan(&num)
	if err == sql.ErrNoRows {
		post.Id = 1
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}
	post.Id = num + 1

	_, err = db.Exec("INSERT INTO posts VALUES($1, $2, $3, $4, $5)",
		post.Id, post.Author, time.Now(), post.Title, post.Text)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{'status':'success'}"))
}

// updatePost updates a single post in DB by id
func updatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		return
	}

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("UPDATE posts SET (title, text) = ($2, $3) WHERE id = ($1)", id, post.Title, post.Text)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{'status':'success'}"))
}

// deletePost deletes a single post from DB by id
func deletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		return
	}

	_, err := db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{'status':'success'}"))
}
