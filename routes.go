package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() *chi.Mux {
	webDir := "./web"

	r := chi.NewRouter()

	staticFileServe := http.FileServer(http.Dir(webDir))

	r.Handle("/css/*", staticFileServe)
	r.Handle("/js/*", staticFileServe)
	r.Handle("/favicon.ico", staticFileServe)
	r.Handle("/index.html", staticFileServe)
	r.Handle("/login.html", staticFileServe)

	r.HandleFunc("/", app.login)

	r.Get("/api/nextdate", app.getNextDate)
	r.Get("/api/tasks", auth(app.getTasks))
	r.Get("/api/task", auth(app.getTask))
	r.Post("/api/task", auth(app.addTask))
	r.Put("/api/task", auth(app.updateTask))
	r.Delete("/api/task", auth(app.deleteTask))
	r.Post("/api/task/done", auth(app.setDoneTask))
	r.Post("/api/signin", app.signIn)

	return r
}
