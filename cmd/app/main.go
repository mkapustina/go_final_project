package main

import (
	"os"

	"github.com/mkapustina/go_final_project/cmd/config"
	"github.com/mkapustina/go_final_project/internal/transport"
)

func main() {
	logger := config.InitLogger()

	cfg, err := config.ParseConfig()
	if err != nil {
		logger.ErrorLog.Fatal("Error parse config", err)
		os.Exit(1)
	}

	app, srv := transport.AppInit(&logger, &cfg)
	defer app.Tasks.Db.Close()

	app.InfoLog.Printf("Запуск сервера на %s", srv.Addr)
	err = srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}
