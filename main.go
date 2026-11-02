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

	dbFile := "scheduler.db"
	err := db.Init(dbFile)
	if err != nil {
		log.Fatalf("fail to init db: %v", err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	port = ":" + port

	r := chi.NewRouter()

	api.Init(r)

	r.Handle("/*", http.FileServer(http.Dir("./web")))
	//r.Handle("/js/scripts.min.js", http.FileServer(http.Dir("./web/js/scripts.min.js")))
	//r.Handle("/css/style.css", http.FileServer(http.Dir("./web/css/style.css")))
	//	r.Handle("/favicon.ico", http.FileServer(http.Dir("./web/favicon.ico")))
	log.Printf("Listening on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}

}
