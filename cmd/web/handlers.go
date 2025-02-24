package main

import (
	"net/http"

	"go.chat/internal/validator"
)

type userRegisterForm struct {
	email    string
	username string
	password string
	validator.Validator
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Snippetbox"))
}

func (app *application) userRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := userRegisterForm{
		email:    r.PostForm.Get("email"),
		username: r.PostForm.Get("username"),
		password: r.PostForm.Get("password"),
	}

	id, err := app.users.Insert(username, email, password)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.infoLog.Println("id: ", id)

}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {

}
