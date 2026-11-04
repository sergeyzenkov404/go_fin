package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sergeyzenkov404/go_fin/api"
	"github.com/sergeyzenkov404/go_fin/db"
	_ "modernc.org/sqlite"
)

func main() {

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	err := db.Init(dbFile)
	if err != nil {
		log.Fatalf("fail to init db: %v", err)
	}
	defer db.DB.Close()
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7545"
	}
	port = ":" + port

	r := chi.NewRouter()

	api.Init(r)

	r.Handle("/*", http.FileServer(http.Dir("./web")))

	log.Printf("Listening on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}

}
