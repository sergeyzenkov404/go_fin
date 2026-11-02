package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sergeyzenkov404/go_fin/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var taskDB db.Task

	err := json.NewDecoder(r.Body).Decode(&taskDB)

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if taskDB.Title == "" {
		writeError(w, http.StatusBadRequest, "empty title")
		return
	}

	defer r.Body.Close()

	err = checkDate(&taskDB)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := db.AddTask(&taskDB)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, map[string]any{"id": id})
}

func checkDate(taskDB *db.Task) error {
	now := time.Now()

	if taskDB.Date == "" {
		taskDB.Date = now.Format("20060102")
	}

	t, err := time.Parse("20060102", taskDB.Date)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}
	next := ""

	if taskDB.Repeat != "" {
		next, err = NextDate(now, taskDB.Date, taskDB.Repeat)
		if err != nil {
			return fmt.Errorf("invalid repeat format")
		}
	}

	if afterNow(now, t) {
		if len(taskDB.Repeat) == 0 {
			taskDB.Date = now.Format("20060102")
		} else {
			taskDB.Date = next
		}
	}
	return nil
}

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {

	search := r.URL.Query().Get("search")
	tasks, err := db.Tasks(50, search)

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, TasksResp{
		Tasks: tasks,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJson(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		return err
	}

	return nil
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id пуст")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "get task by id error")
		return
	}
	writeJson(w, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var taskDB db.Task

	err := json.NewDecoder(r.Body).Decode(&taskDB)
	if err != nil {
		writeError(w, http.StatusBadRequest, "decod error")
		return
	}

	defer r.Body.Close()

	if taskDB.Id == 0 {
		writeError(w, http.StatusBadRequest, "id пуст")
		return
	}

	if taskDB.Title == "" {
		writeError(w, http.StatusBadRequest, "invalid title format")
		return
	}

	err = checkDate(&taskDB)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = db.UpdateTask(&taskDB)

	if err != nil {
		writeError(w, http.StatusBadRequest, "update task error")
		return
	}
	writeJson(w, map[string]any{})
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	now := time.Now()
	var next string
	if id == "" {
		writeError(w, http.StatusBadRequest, "id пуст")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "get task by id error")
		return
	}

	if task.Repeat == "" {
		err = db.DeleteTask(id)
	} else {
		next, err = NextDate(now, task.Date, task.Repeat)

		if err != nil {
			writeError(w, http.StatusBadRequest, "next date error")
			return
		}

		err = db.UpdateDate(next, id)
		if err != nil {
			writeError(w, http.StatusBadRequest, "update error")
			return
		}

	}
	writeJson(w, map[string]any{})
}

func daleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		writeError(w, http.StatusBadRequest, "id пуст")
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "delete error")
		return
	}

	writeJson(w, map[string]any{})
}
