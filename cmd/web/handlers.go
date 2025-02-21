package main

import (
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	w.Write([]byte("Hello from Snippetbox"))
}

func (app *application) userCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	username := "one"
	email := "one@one.com"
	password := "one"

	id, err := app.users.Insert(username, email, password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Print(id)

}
