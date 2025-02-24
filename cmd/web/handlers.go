package main

import (
	"encoding/json"
	"net/http"

	"go.chat/internal/validator"
)

type userRegisterForm struct {
	Email    string
	Username string
	Password string
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
		Email:    r.PostForm.Get("email"),
		Username: r.PostForm.Get("username"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "invalid email")
	form.CheckField(validator.NotBlank(form.Username), "username", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Password), "password", "this field cannot be empty")
	form.CheckField(validator.MaxChars(form.Username, 20), "username", "this field cannot have more than 20 characters long")

	if !form.Valid() {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(form.FieldErrors)
		if err != nil {
			app.serverError(w, err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := app.users.Insert(form.Username, form.Email, form.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.infoLog.Println("id: ", id)

}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {

}
