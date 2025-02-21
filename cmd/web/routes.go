package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/users/create", app.userCreate)

	return app.logRequest(secureHeaders(mux))
}
