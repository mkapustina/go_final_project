package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) Routes() *chi.Mux {
	r := chi.NewRouter()

	staticFileServe := http.FileServer(http.Dir(app.Config.WebDir))

	r.Handle("/css/*", staticFileServe)
	r.Handle("/js/*", staticFileServe)
	r.Handle("/favicon.ico", staticFileServe)
	r.Handle("/index.html", staticFileServe)
	r.Handle("/login.html", staticFileServe)

	r.HandleFunc("/", app.login)

	r.Get("/api/nextdate", app.getNextDate)
	r.Get("/api/tasks", app.auth(app.getTasks))
	r.Get("/api/task", app.auth(app.getTask))
	r.Post("/api/task", app.auth(app.addTask))
	r.Put("/api/task", app.auth(app.updateTask))
	r.Delete("/api/task", app.auth(app.deleteTask))
	r.Post("/api/task/done", app.auth(app.setDoneTask))
	r.Post("/api/signin", app.signIn)

	return r
}
