package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/mkapustina/go_final_project/internal/models"
)

func (app *Application) getNextDate(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse("20060102", r.FormValue("now"))
	if err != nil {
		app.ServerError(w, err)
		return
	}

	resp, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))

	if err != nil {
		app.ServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(resp))
}

func (app *Application) addTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	task, err := CheckTask(r)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	id, err := app.Tasks.Add(task)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	task.ID = fmt.Sprint(id)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"id":"%d"}`, id)))
}

func (app *Application) updateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	task, err := CheckTask(r)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	cnt, err := app.Tasks.Update(task)
	if err != nil {
		app.ServerError(w, err)
		return
	}
	if cnt == 0 {
		app.ClientError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (app *Application) getTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		app.ClientError(w, http.StatusBadRequest, "Не задан номер задачи")
		return
	}

	numId, err := strconv.Atoi(id)
	if err != nil || numId < 1 {
		app.ClientError(w, http.StatusBadRequest, "Некорректный номер задачи")
		return
	}

	task, err := app.Tasks.Get(numId)
	if err != nil {
		app.ClientError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (app *Application) getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tasks, err := app.Tasks.GetAll(r.URL.Query().Get("search"))
	if err != nil {
		app.ServerError(w, err)
		return
	}

	resp, err := json.Marshal(tasks)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (app *Application) deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		app.ClientError(w, http.StatusBadRequest, "Не задан номер задачи")
		return
	}

	numId, err := strconv.Atoi(id)
	if err != nil || numId < 1 {
		app.ClientError(w, http.StatusBadRequest, "Некорректный номер задачи")
		return
	}

	err = app.Tasks.Delete(numId)
	if err != nil {
		app.ClientError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (app *Application) setDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		app.ClientError(w, http.StatusBadRequest, "Не задан номер задачи")
		return
	}

	numId, err := strconv.Atoi(id)
	if err != nil || numId < 1 {
		app.ClientError(w, http.StatusBadRequest, "Некорректный номер задачи")
		return
	}

	task, err := app.Tasks.Get(numId)
	if err != nil {
		app.ClientError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	if len(task.Repeat) == 0 {
		err = app.Tasks.Delete(numId)
		if err != nil {
			app.ServerError(w, err)
			return
		}
	} else {
		task.Date, err = NextDate(getNowDate(), task.Date, task.Repeat)
		if err != nil {
			app.ServerError(w, err)
			return
		}
		cnt, err := app.Tasks.Update(task)
		if err != nil || cnt == 0 {
			app.ServerError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func (app *Application) login(w http.ResponseWriter, r *http.Request) {
	url := filepath.Join(app.Config.WebDir, "login.html")
	err := app.checkCookie(r)
	if err == nil {
		url = filepath.Join(app.Config.WebDir, "index.html")
	}
	t, _ := template.ParseFiles(url)
	t.Execute(w, "")
}

func (app *Application) signIn(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var buf bytes.Buffer

	w.Header().Set("Content-Type", "application/json")

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &user)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	if len(app.Config.Password) > 0 && app.Config.Password != user.Password {
		app.ClientError(w, http.StatusUnauthorized, "неверный пароль")
		return
	}

	signedToken, err := getToken(app.Config.Password)
	if err != nil {
		app.ServerError(w, err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"token":"%s"}`, signedToken)))
}
