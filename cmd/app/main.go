package main

import (
	"github.com/mkapustina/go_final_project/internal/transport"
)

func main() {
	app, srv := transport.AppInit()
	defer app.Tasks.Db.Close()

	app.InfoLog.Printf("Запуск сервера на %s", srv.Addr)
	err := srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}
