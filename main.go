package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const createTable string = `
	CREATE TABLE IF NOT EXISTS scheduler (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			date TEXT(8) NOT NULL,
			title TEXT(50) NOT NULL,
			comment TEXT(100),
			repeat TEXT(128)
	);
	CREATE INDEX IF NOT EXISTS scheduler_date_IDX ON scheduler (date);
`

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	tasks    *TaskStore
}

func main() {
	var err error

	port := os.Getenv("TODO_PORT")
	if len(port) == 0 {
		port = "7540"
	}

	dsn := os.Getenv("TODO_DBFILE")
	if len(dsn) == 0 {
		dsn = "scheduler.db"
	}

	addr := flag.String("addr", ":"+port, "Сетевой адрес веб-сервера")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		tasks:    &TaskStore{db: db},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Запуск сервера на %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), dsn)
	_, err = os.Stat(dbFile)
	install := err != nil

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	if install {
		if _, err := db.Exec(createTable); err != nil {
			return nil, err
		}
	}
	return db, nil
}
