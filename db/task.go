package db

import (
	"fmt"
	"time"
)

func Tasks(limit int, search string) ([]*Task, error) {
	now := time.Now().Format("20060102")
	tasks := []*Task{}
	var err error
	var rows interface {
		Next() bool
		Scan(dest ...any) error
		Close() error
		Err() error
	}

	if search != "" {
		d, err2 := time.Parse("02.01.2006", search)
		if err2 == nil {
			dd := d.Format("20060102")
			rows, err = DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT ?", dd, limit)
		} else {
			rows, err = DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?", "%"+search+"%", "%"+search+"%", limit)
		}
	} else {
		rows, err = DB.Query(`SELECT id, date, title, comment, repeat FROM scheduler 
	WHERE date >= ?
	ORDER BY date ASC
	LIMIT ?`, now, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := new(Task)

		err := rows.Scan(&t.Id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	selectID := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = $1`
	t := new(Task)

	err := DB.QueryRow(selectID, id).Scan(&t.Id, &t.Date, &t.Title, &t.Comment, &t.Repeat)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func UpdateTask(task *Task) error {

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ? , repeat = ?
	WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}
func UpdateDate(next string, id string) error {

	query := `UPDATE scheduler SET date = ?	WHERE id = ?`

	res, err := DB.Exec(query, next, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}
