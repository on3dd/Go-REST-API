package api

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

type API struct {
	db *sql.DB
}

// db.New returns a new API instance
func New(db *sql.DB) *API {
	return &API{
		db: db,
	}
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/getPost", logHandlerCall(api.getPost)).Methods("GET")
	router.HandleFunc("/api/v1/getPosts", logHandlerCall(api.getPosts)).Methods("GET")
	router.HandleFunc("/api/v1/addPost", logHandlerCall(api.addPost)).Methods("POST")
	router.HandleFunc("/api/v1/updatePost", logHandlerCall(api.updatePost)).Methods("PUT")
	router.HandleFunc("/api/v1/deletePost", logHandlerCall(api.deletePost)).Methods("DELETE")

	router.ServeHTTP(w, r)
}

// logHandlerCall logs any handler call
func logHandlerCall(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		log.Printf("Handler function called: %v", name)
		handler(w, r)
	}
}

// Post represents a Post instance in the DB
type Post struct {
	Id        int    `json:"id"`
	Author    int    `json:"author"`
	Posted_at string `json:"posted_at"`
	Title     string `json:"title"`
	Text      string `json:"text"`
}

// getPost gets single post from DB by id
func (api *API) getPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		return
	}

	row := api.db.QueryRow("SELECT * FROM posts WHERE id = $1", id)

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
func (api *API) getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var rows *sql.Rows
	var err error
	num := r.FormValue("num")
	if num == "" {
		rows, err = api.db.Query("SELECT * FROM posts ORDER BY posted_at DESC")
	} else {
		rows, err = api.db.Query("SELECT * FROM posts ORDER BY posted_at DESC LIMIT $1", num)
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
func (api *API) addPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		log.Fatal(err)
	}

	var num int
	id := api.db.QueryRow("SELECT id FROM posts ORDER BY id DESC LIMIT 1")
	err = id.Scan(&num)
	if err == sql.ErrNoRows {
		post.Id = 1
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}
	post.Id = num + 1

	_, err = api.db.Exec("INSERT INTO posts VALUES($1, $2, $3, $4, $5)",
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
func (api *API) updatePost(w http.ResponseWriter, r *http.Request) {
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

	_, err = api.db.Exec("UPDATE posts SET (title, text) = ($2, $3) WHERE id = ($1)", id, post.Title, post.Text)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{'status':'success'}"))
}

// deletePost deletes a single post from DB by id
func (api *API) deletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		return
	}

	_, err := api.db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{'status':'error'}"))
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{'status':'success'}"))
}
