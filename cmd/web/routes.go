package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Public routes
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodPost, "/user/register", app.userRegister)
	router.HandlerFunc(http.MethodPost, "/user/login", app.userLogin)

	// Protected routes
	protected := alice.New(app.requireAuth)
	router.Handler(http.MethodPost, "/chat/create", protected.ThenFunc(app.createChat))
	router.Handler(http.MethodPost, "/chat/message", protected.ThenFunc(app.sendMessage))
	router.Handler(http.MethodGet, "/chat/messages/:chat_id", protected.ThenFunc(app.getMessages))
	router.Handler(http.MethodPost, "/chat/join", protected.ThenFunc(app.joinChat))
	router.Handler(http.MethodGet, "/ws", protected.ThenFunc(app.handleWebSocket))
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders, app.requestTimeout)
	return standard.Then(router)
}
