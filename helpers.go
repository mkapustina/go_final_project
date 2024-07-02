package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Помощник serverError записывает сообщение об ошибке в errorLog и
// затем отправляет пользователю ответ 500 "Внутренняя ошибка сервера".
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

/*
// Помощник clientError отправляет определенный код состояния и
// соответствующее описание пользователю.
func (app *application) clientError(w http.ResponseWriter, status int, statusText string) {
	http.Error(w, statusText, status)
}

// Помощник notFound. Это просто удобная оболочка вокруг clientError,
// которая отправляет пользователю ответ "404 Страница не найдена".
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
*/
