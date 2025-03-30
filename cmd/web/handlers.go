package main

import (
	"encoding/json"
	"net/http"

	"go.chat/internal/models"
	"go.chat/internal/validator"
)

type userRegisterForm struct {
	Email    string
	Username string
	Password string
	validator.Validator
}

type userLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from goChat"))
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

	form.CheckField(validator.NotBlank(form.Username), "username", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Email), "username", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Password), "password", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Email), "email", "invalid email")
	form.CheckField(validator.MaxChars(form.Username, 20), "username", "this field cannot have more than 20 characters long")
	emailExist, err := app.users.ExistsEmail(form.Email)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if emailExist {
		form.AddFieldError("email", "email already exists")
	}
	usernameExist, err := app.users.ExistsUsername(form.Username)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if usernameExist {
		form.AddFieldError("username", "username already exists")
	}
	if !form.Valid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(form.FieldErrors)
		if err != nil {
			app.serverError(w, err)
			return
		}
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
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := userLoginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "this field cannot be empty")
	form.CheckField(validator.NotBlank(form.Password), "password", "this field cannot be empty")
	form.CheckField(validator.Matches(validator.EmailRX, form.Email), "email", "invalid email")

	if !form.Valid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(form.FieldErrors)
		if err != nil {
			app.serverError(w, err)
			return
		}
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid email or password",
			})
			return
		}
		app.serverError(w, err)
		return
	}

	token, err := app.jwt.GenerateToken(id, form.Email)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Set the token in a secure HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   true, // Only send cookie over HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}
