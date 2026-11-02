package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const cteateTable = `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT(128)
	);`

var DB *sql.DB

func Init(dbFile string) error {
	var err error

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	//defer DB.Close()

	_, err = DB.Exec(cteateTable)

	if err != nil {
		return err
	}
	return nil
}

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
	Id      int64  `json:"id,string"`
	Error   string `json:"error"`
}

func AddTask(task *Task) (int64, error) {

	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}

	return id, err
}
