package transport

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mkapustina/go_final_project/internal/database"

	_ "modernc.org/sqlite"
)

type Application struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	Tasks    *database.TaskStore
	Config   *Config
}

type Config struct {
	Port     string
	DbFile   string
	Password string
	WebDir   string
}

func AppInit() (*Application, *http.Server) {
	cfg := new(Config)

	cfg.Port = os.Getenv("TODO_PORT")
	if len(cfg.Port) == 0 {
		cfg.Port = "7540"
	}

	cfg.DbFile = os.Getenv("TODO_DBFILE")
	if len(cfg.DbFile) == 0 {
		cfg.DbFile = filepath.Join("..", "..", "internal", "database", "scheduler.db")
	}

	cfg.Password = os.Getenv("TODO_PASSWORD")

	cfg.WebDir = filepath.Join("..", "..", "web")

	addr := flag.String("addr", ":"+cfg.Port, "Сетевой адрес веб-сервера")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := database.OpenDB(cfg.DbFile)
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &Application{
		ErrorLog: errorLog,
		InfoLog:  infoLog,
		Tasks:    &database.TaskStore{Db: db},
		Config:   cfg,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.Routes(),
	}

	return app, srv
}
