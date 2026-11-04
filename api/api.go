package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
)

const format = "20060102"

func Init(r chi.Router) {
	r.Get("/api/nextdate", nextDayHandler)
	r.Post("/api/task", auth(addTaskHandler))
	r.Get("/api/tasks", auth(tasksHandler))
	r.Get("/api/task", auth(getTaskHandler))
	r.Put("/api/task", auth(updateTaskHandler))
	r.Post("/api/task/done", auth(doneTaskHandler))
	r.Delete("/api/task", auth(daleteTaskHandler))
	r.Post("/api/signin", signinHandler)
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")
	now := time.Now()
	var err error

	if nowParam != "" {
		now, err = time.Parse(format, nowParam)

		if err != nil {
			http.Error(w, "invalid format", http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, "invalid format", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(next))
}

func afterNow(date, now time.Time) bool {
	cNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	if cDate.After(cNow) {
		return true
	} else {
		return false
	}
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	date, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", err
	}

	if repeat == "" {
		return "", fmt.Errorf("invalid repeat format")
	}

	rul := strings.Split(repeat, " ")

	if rul[0] == "y" {
		if len(rul) > 1 {
			return "", fmt.Errorf("invalid year format")
		}

		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}

		return date.Format("20060102"), nil
	}

	if rul[0] == "d" {
		if len(rul) != 2 {
			return "", fmt.Errorf("invalid day format")
		}

		day, err := strconv.Atoi(rul[1])
		if err != nil {
			return "", err
		}

		if day <= 0 || day > 400 {
			return "", fmt.Errorf("invalid day format")
		}

		for {
			date = date.AddDate(0, 0, day)
			if afterNow(date, now) {
				break
			}
		}

		return date.Format("20060102"), nil
	}

	if rul[0] == "m" {
		var dd []int
		mm := make(map[int]bool)
		if len(rul) == 1 || len(rul) > 3 {
			return "", fmt.Errorf("invalid month format")
		}

		//дни точно есть
		days := strings.Split(rul[1], ",")
		for _, day := range days {
			d, err := strconv.Atoi(day)

			if err != nil {
				return "", fmt.Errorf("invalid day format in mounth repeat")
			}
			if (d < 1 || d > 31) && d != -1 && d != -2 {
				return "", fmt.Errorf("invalid day format in mounth repeat")
			}
			dd = append(dd, d)
		}

		//мб есть месяцы
		if len(rul) == 3 {
			mounths := strings.Split(rul[2], ",")
			for _, mounth := range mounths {

				m, err := strconv.Atoi(mounth)

				if err != nil {
					return "", fmt.Errorf("invalid mounth format in mounth repeat")
				}
				if m < 1 || m > 12 {
					return "", fmt.Errorf("invalid mounth format in mounth repeat")
				}
				mm[m] = true
			}
		} else {
			for m := 1; m <= 12; m++ {
				mm[m] = true
			}
		}

		startY := date.Year()
		if now.Year() < startY {
			startY = now.Year()
		}

		for y := startY; y < startY+10; y++ {
			for m := 1; m <= 12; m++ {
				if !mm[m] {
					continue
				}

				lastDayOfMounth := time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()

				//если отриц перевод дня
				var ddd []int
				for _, d := range dd {

					if d == -1 {
						ddd = append(ddd, lastDayOfMounth)
					} else if d == -2 {
						ddd = append(ddd, lastDayOfMounth-1)
					} else if d >= 1 && d <= lastDayOfMounth {

						ddd = append(ddd, d)
					}

				}
				sort.Ints(ddd)
				for _, day := range ddd {

					dateM := time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)

					if dateM.Before(date) || dateM.Equal(date) {
						continue
					}
					if afterNow(dateM, now) {
						return dateM.Format(format), nil
					}
				}
			}
		}

		return "", fmt.Errorf("unsupported repeat rule")
	}

	if rul[0] == "w" {
		weekDay := map[time.Weekday]bool{
			time.Saturday:  false,
			time.Sunday:    false,
			time.Monday:    false,
			time.Tuesday:   false,
			time.Wednesday: false,
			time.Thursday:  false,
			time.Friday:    false,
		}
		if len(rul) != 2 {
			return "", fmt.Errorf("inval week form")
		}

		days := strings.Split(rul[1], ",")
		for _, d := range days {
			day, err := strconv.Atoi(d)
			if err != nil || day < 1 || day > 7 {
				return "", fmt.Errorf("invalid day format in weeek repeat")
			}
			if day == 7 {
				weekDay[0] = true
			} else {
				weekDay[time.Weekday(day)] = true
			}
		}

		for {
			if weekDay[date.Weekday()] && afterNow(date, now) {
				break
			}
			date = date.AddDate(0, 0, 1)
		}
		return date.Format(format), nil
	}

	return "", fmt.Errorf("unsupported repeat rule")
}

var pass = os.Getenv("TODO_PASSWORD")

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if len(pass) > 0 {
			var tokenC string

			cookie, err := r.Cookie("token")

			if err != nil || cookie.Value == "" {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			tokenC = cookie.Value

			jwtToken, err := jwt.Parse(tokenC, func(t *jwt.Token) (interface{}, error) {

				return []byte(pass), nil

			})

			if err != nil || !jwtToken.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

		}
		next(w, r)
	})
}

func signinHandler(w http.ResponseWriter, r *http.Request) {

	var passR signinR

	err := json.NewDecoder(r.Body).Decode(&passR)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid req formar")
		return
	}

	defer r.Body.Close()

	if pass != passR.Pass {
		writeError(w, http.StatusBadRequest, "invalid password")
		return
	}

	secret := []byte(pass)

	jwtToken := jwt.New(jwt.SigningMethodHS256)

	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		writeError(w, http.StatusBadRequest, "generate token error")
	}

	writeJson(w, map[string]string{"token": signedToken})
}

type signinR struct {
	Pass string `json:"password"`
}
