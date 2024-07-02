package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getNowDate() time.Time {
	nowDate, _ := time.Parse("20060102", time.Now().Format("20060102"))
	return nowDate
}

func checkRepeat(repeat string) (string, []int, []int, error) {
	var num int
	var err error
	var pref string
	var intValue []string
	var suf1, suf2 []int

	if len(repeat) == 0 {
		return pref, suf1, suf2, nil
	}

	if repeat == "y" {
		pref = repeat
		return pref, suf1, suf2, nil
	}

	substrings := strings.Split(repeat, " ")
	if !(len(substrings) == 2 || len(substrings) == 3) {
		return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
	}
	pref = substrings[0]

	switch pref {
	case "d":
		num, err = strconv.Atoi(substrings[1])
		if err != nil {
			return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
		}
		if !(num >= 1 && num <= 400) {
			return pref, suf1, suf2, errors.New("неверное количество дней интервала повторения")
		}
		suf1 = append(suf1, num)
	case "w":
		intValue = strings.Split(substrings[1], ",")
		for i := 0; i < len(intValue); i++ {
			num, err = strconv.Atoi(intValue[i])
			if err != nil {
				return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
			}
			if !(num >= 1 && num <= 7) {
				return pref, suf1, suf2, errors.New("неверный номер дня недели интервала повторения")
			}
			suf1 = append(suf1, num)
		}
	case "m":
		intValue = strings.Split(substrings[1], ",")
		for i := 0; i < len(intValue); i++ {
			num, err = strconv.Atoi(intValue[i])
			if err != nil {
				return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
			}
			if !((num >= 1 && num <= 31) || num == -1 || num == -2) {
				return pref, suf1, suf2, errors.New("неверный номер дня месяца интервала повторения")
			}
			suf1 = append(suf1, num)
		}

		if len(substrings) == 3 {
			intValue = strings.Split(substrings[2], ",")
			for i := 0; i < len(intValue); i++ {
				num, err = strconv.Atoi(intValue[i])
				if err != nil {
					return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
				}
				if !(num >= 1 && num <= 12) {
					return pref, suf1, suf2, errors.New("неверный номер месяца интервала повторения")
				}
				suf2 = append(suf2, num)
			}
		}
	default:
		return pref, suf1, suf2, errors.New("недопустимый тип интервала повторения")
	}
	return pref, suf1, suf2, nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	var retDate, tmpDate time.Time
	var num int
	var isValidMonth bool

	curDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}

	if len(repeat) == 0 {
		return "", errors.New("не задан интервал повторения")
	}

	pref, suf1, suf2, err := checkRepeat(repeat)
	if err != nil {
		return "", err
	}

	switch pref {
	case "y":
		retDate = curDate
		for {
			retDate = retDate.AddDate(1, 0, 0)
			if now.Before(retDate) {
				break
			}
		}
		return retDate.Format("20060102"), nil
	case "d":
		retDate = curDate
		num = suf1[0]
		for {
			retDate = retDate.AddDate(0, 0, num)
			if now.Before(retDate) {
				break
			}
		}
		return retDate.Format("20060102"), nil
	case "w":
		retDate = curDate
		if retDate.Before(now) {
			retDate = now
		}
		for {
			retDate = retDate.AddDate(0, 0, 1)
			num = int(retDate.Weekday())
			if num == 0 {
				num = 7
			}
			if slices.Contains(suf1, num) {
				break
			}
		}
		return retDate.Format("20060102"), nil
	case "m":
		retDate = curDate
		if retDate.Before(now) {
			retDate = now
		}
		for {
			retDate = retDate.AddDate(0, 0, 1)
			if len(suf2) == 0 {
				isValidMonth = true
			} else {
				isValidMonth = slices.Contains(suf2, int(retDate.Month()))
			}
			if isValidMonth {
				if slices.Contains(suf1, retDate.Day()) {
					break
				}
				// ищем 1 день след. месяца
				tmpDate = time.Date(retDate.Year(), retDate.Month(), 1, 0, 0, 0, 0, retDate.Location())
				tmpDate = tmpDate.AddDate(0, 1, 0)
				if slices.Contains(suf1, -1) {
					if retDate == tmpDate.AddDate(0, 0, -1) {
						break
					}
				}
				if slices.Contains(suf1, -2) {
					if retDate == tmpDate.AddDate(0, 0, -2) {
						break
					}
				}
			}
		}
		return retDate.Format("20060102"), nil
	default:
		return "", errors.New("недопустимый тип интервала повторения")
	}
}

func (app *application) getNextDate(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse("20060102", r.FormValue("now"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	resp, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))

	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func checkDate(taskDate string, taskRepeat string) (string, error) {
	resDate := taskDate
	nowDate := getNowDate()
	if len(resDate) == 0 {
		return nowDate.Format("20060102"), nil
	}
	taskDT, err := time.Parse("20060102", resDate)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}
	if taskDT.Before(nowDate) {
		if len(taskRepeat) == 0 {
			return nowDate.Format("20060102"), nil
		}
		resDate, err = NextDate(nowDate, resDate, taskRepeat)
		if err != nil {
			return "", err
		}
	}
	return resDate, nil
}

func checkTask(r *http.Request) (Task, error) {
	var task Task
	var buf bytes.Buffer
	var err error

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		return task, err
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		return task, err
	}

	if len(task.Title) == 0 {
		return task, errors.New("не указан заголовок задачи")
	}

	if _, _, _, err = checkRepeat(task.Repeat); err != nil {
		return task, err
	}

	if task.Date, err = checkDate(task.Date, task.Repeat); err != nil {
		return task, err
	}

	return task, nil
}

func (app *application) addTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	task, err := checkTask(r)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	id, err := app.tasks.Add(task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	task.ID = fmt.Sprint(id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%d"}`, id)))
}

func (app *application) updateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	task, err := checkTask(r)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	cnt, err := app.tasks.Update(task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if cnt == 0 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Задача не найдена"), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func (app *application) getTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Не задан номер задачи"), http.StatusBadRequest)
		return
	}

	numId, err := strconv.Atoi(id)
	if err != nil || numId < 1 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Некорректный номер задачи"), http.StatusBadRequest)
		return
	}

	task, err := app.tasks.Get(numId)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (app *application) getTasks(w http.ResponseWriter, r *http.Request) {
	var tasks Tasks
	var err error

	w.Header().Set("Content-Type", "application/json")

	search := r.URL.Query().Get("search")
	if len(search) != 0 {
		searchDT, err := time.Parse("02.01.2006", search)
		if err == nil {
			search = searchDT.Format("20060102")
		}
		search = "%" + strings.ToUpper(search) + "%"
	}

	tasks, err = app.tasks.GetAll(search)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (app *application) deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Не задан номер задачи"), http.StatusBadRequest)
		return
	}

	numId, err := strconv.Atoi(id)
	if err != nil || numId < 1 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Некорректный номер задачи"), http.StatusBadRequest)
		return
	}

	err = app.tasks.Delete(numId)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func (app *application) setDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Не задан номер задачи"), http.StatusBadRequest)
		return
	}

	numId, err := strconv.Atoi(id)
	if err != nil || numId < 1 {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "Некорректный номер задачи"), http.StatusBadRequest)
		return
	}

	task, err := app.tasks.Get(numId)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	if len(task.Repeat) == 0 {
		err = app.tasks.Delete(numId)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		task.Date, err = NextDate(getNowDate(), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		cnt, err := app.tasks.Update(task)
		if err != nil || cnt == 0 {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func getToken(password string) (string, error) {
	secret := []byte("my_secret_key")

	pwdHash := sha256.Sum256([]byte(password))
	claims := jwt.MapClaims{
		"pwd": hex.EncodeToString(pwdHash[:]),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString(secret)
	return signedToken, err
}

func checkCookie(r *http.Request) error {
	pwdStored := os.Getenv("TODO_PASSWORD")

	if len(pwdStored) == 0 {
		return nil
	}

	if len(r.Header["Cookie"]) == 0 {
		return errors.New("токен отсутствует")
	}
	cookie, err := r.Cookie("token")
	if err != nil {
		return err
	}
	cookieToken := cookie.Value

	secretKey := []byte("my_secret_key")
	jwtToken, err := jwt.Parse(cookieToken, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}

	if !jwtToken.Valid {
		return errors.New("токен не валиден")
	}

	token, _ := getToken(pwdStored)
	if cookieToken != token {
		return errors.New("токен не валиден")
	}
	return nil
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	url := "./web/login.html"
	err := checkCookie(r)
	if err == nil {
		url = "./web/index.html"
	}
	t, _ := template.ParseFiles(url)
	t.Execute(w, "")
}

func (app *application) signIn(w http.ResponseWriter, r *http.Request) {
	var user User
	var buf bytes.Buffer

	w.Header().Set("Content-Type", "application/json")

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &user)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	pwd := os.Getenv("TODO_PASSWORD")

	if len(pwd) > 0 && pwd != user.Password {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "неверный пароль"), http.StatusUnauthorized)
		return
	}

	signedToken, err := getToken(pwd)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "failed to sign jwt: "+err.Error()), http.StatusUnauthorized)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"token":"%s"}`, signedToken)))
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := checkCookie(r)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
		}
		next(w, r)
	})
}
