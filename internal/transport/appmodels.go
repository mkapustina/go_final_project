package transport

import (
	"flag"
	"log"
	"net/http"

	"github.com/mkapustina/go_final_project/cmd/config"
	"github.com/mkapustina/go_final_project/internal/database"

	_ "modernc.org/sqlite"
)

type Application struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	Tasks    *database.TaskStore
	Config   *config.Config
}

func InitApp(log *config.Logger, cfg *config.Config) (*Application, *http.Server) {
	addr := flag.String("addr", ":"+cfg.Port, "Сетевой адрес веб-сервера")
	flag.Parse()

	db, err := database.OpenDB(cfg.DbFile)
	if err != nil {
		log.ErrorLog.Fatal(err)
	}

	app := &Application{
		ErrorLog: log.ErrorLog,
		InfoLog:  log.InfoLog,
		Tasks:    &database.TaskStore{Db: db},
		Config:   cfg,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: log.ErrorLog,
		Handler:  app.Routes(),
	}

	return app, srv
}
