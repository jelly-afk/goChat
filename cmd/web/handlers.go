package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
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

type createChatForm struct {
	Name       string
	IsPrivate  bool
	ReceiverID int
	validator.Validator
}

type createMessageForm struct {
	Content string
	ChatID  int
	validator.Validator
}

type joinChatForm struct {
	ChatID int
	validator.Validator
}

var upgrader = websocket.Upgrader{}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from goChat"))
}

func (app *application) userRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Printf("Error parsing form in userRegister: %v", err)
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
		app.errorLog.Printf("Error checking email existence: %v", err)
		app.serverError(w, err)
		return
	}
	if emailExist {
		form.AddFieldError("email", "email already exists")
	}
	usernameExist, err := app.users.ExistsUsername(form.Username)
	if err != nil {
		app.errorLog.Printf("Error checking username existence: %v", err)
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
			app.errorLog.Printf("Error encoding form errors: %v", err)
			app.serverError(w, err)
			return
		}
		return
	}

	id, err := app.users.Insert(form.Username, form.Email, form.Password)
	if err != nil {
		app.errorLog.Printf("Error inserting user: %v", err)
		app.serverError(w, err)
		return
	}
	app.infoLog.Printf("User registered successfully with ID: %d", id)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Printf("Error parsing form in userLogin: %v", err)
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
			app.errorLog.Printf("Error encoding form errors: %v", err)
			app.serverError(w, err)
			return
		}
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			app.errorLog.Printf("Invalid login attempt for email: %s", form.Email)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid email or password",
			})
			return
		}
		app.errorLog.Printf("Error during authentication: %v", err)
		app.serverError(w, err)
		return
	}

	token, err := app.jwt.GenerateToken(id, form.Email)
	if err != nil {
		app.errorLog.Printf("Error generating JWT token: %v", err)
		app.serverError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	app.infoLog.Printf("User logged in successfully with ID: %d", id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (app *application) createChat(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Printf("Error parsing form in createChat: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	receiverID, err := strconv.Atoi(r.PostForm.Get("receiver_id"))
	if err != nil {
		app.errorLog.Printf("Error parsing receiver_id: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := createChatForm{
		Name:       r.PostForm.Get("name"),
		IsPrivate:  r.PostForm.Get("is_private") == "true",
		ReceiverID: receiverID,
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "this field cannot be empty")
	form.CheckField(validator.MaxChars(form.Name, 50), "name", "this field cannot have more than 50 characters")

	if form.IsPrivate {
		form.CheckField(form.ReceiverID > 0, "receiver_id", "receiver ID is required for private chats")
		exists, err := app.users.ExistsId(form.ReceiverID)
		if err != nil {
			app.errorLog.Printf("Error checking user existence: %v", err)
			app.serverError(w, err)
			return
		}
		if !exists {
			form.AddFieldError("receiver_id", "receiver does not exist")
		}
	}

	if !form.Valid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(form.FieldErrors)
		if err != nil {
			app.errorLog.Printf("Error encoding form errors: %v", err)
			app.serverError(w, err)
			return
		}
		return
	}

	exists, err := app.chats.ExistsName(form.Name)
	if err != nil {
		app.errorLog.Printf("Error checking chat name existence: %v", err)
		app.serverError(w, err)
		return
	}
	if exists {
		form.AddFieldError("name", "chat name already exists")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.FieldErrors)
		return
	}

	id, err := app.chats.Insert(form.Name, form.IsPrivate)
	if err != nil {
		app.errorLog.Printf("Error creating chat: %v", err)
		app.serverError(w, err)
		return
	}

	userID := r.Context().Value("user_id").(int)
	_, err = app.participants.Insert(id, userID)
	if err != nil {
		app.errorLog.Printf("Error adding creator as participant: %v", err)
		app.serverError(w, err)
		return
	}

	if form.IsPrivate {
		_, err = app.participants.Insert(id, form.ReceiverID)
		if err != nil {
			app.errorLog.Printf("Error adding receiver as participant: %v", err)
			app.serverError(w, err)
			return
		}
	}

	app.infoLog.Printf("Chat created successfully with ID: %d", id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (app *application) sendMessage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Printf("Error parsing form in sendMessage: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	chatID, err := strconv.Atoi(r.PostForm.Get("chat_id"))
	if err != nil {
		app.errorLog.Printf("Error parsing chat_id: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := createMessageForm{
		Content: r.PostForm.Get("content"),
		ChatID:  chatID,
	}

	form.CheckField(validator.NotBlank(form.Content), "content", "this field cannot be empty")
	form.CheckField(validator.MaxChars(form.Content, 500), "content", "this field cannot have more than 500 characters")

	if !form.Valid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(form.FieldErrors)
		if err != nil {
			app.errorLog.Printf("Error encoding form errors: %v", err)
			app.serverError(w, err)
			return
		}
		return
	}

	exists, err := app.chats.ExistsId(form.ChatID)
	if err != nil {
		app.errorLog.Printf("Error checking chat existence: %v", err)
		app.serverError(w, err)
		return
	}
	if !exists {
		app.errorLog.Printf("Chat not found with ID: %d", form.ChatID)
		app.clientError(w, http.StatusNotFound)
		return
	}

	userID := r.Context().Value("user_id").(int)
	participants, err := app.participants.GetByChatID(form.ChatID)
	if err != nil {
		app.errorLog.Printf("Error getting chat participants: %v", err)
		app.serverError(w, err)
		return
	}

	isParticipant := false
	for _, p := range participants {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		app.errorLog.Printf("User %d is not a participant in chat %d", userID, form.ChatID)
		app.clientError(w, http.StatusForbidden)
		return
	}

	id, err := app.messages.Insert(form.ChatID, userID, form.Content)
	if err != nil {
		app.errorLog.Printf("Error inserting message: %v", err)
		app.serverError(w, err)
		return
	}

	// Broadcast message to connected clients
	message := Message{
		Type:    "message",
		Content: form.Content,
		ChatID:  form.ChatID,
		UserID:  userID,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		app.errorLog.Printf("Error marshaling message for broadcast: %v", err)
		app.serverError(w, err)
		return
	}

	app.hub.broadcast <- messageBytes

	app.infoLog.Printf("Message sent successfully with ID: %d in chat: %d", id, form.ChatID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (app *application) getMessages(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	chatID, err := strconv.Atoi(params.ByName("chat_id"))
	if err != nil {
		app.errorLog.Printf("Error parsing chat_id: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	exists, err := app.chats.ExistsId(chatID)
	if err != nil {
		app.errorLog.Printf("Error checking chat existence: %v", err)
		app.serverError(w, err)
		return
	}
	if !exists {
		app.errorLog.Printf("Chat not found with ID: %d", chatID)
		app.clientError(w, http.StatusNotFound)
		return
	}

	userID := r.Context().Value("user_id").(int)
	participants, err := app.participants.GetByChatID(chatID)
	if err != nil {
		app.errorLog.Printf("Error getting chat participants: %v", err)
		app.serverError(w, err)
		return
	}

	isParticipant := false
	for _, p := range participants {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		app.errorLog.Printf("User %d is not a participant in chat %d", userID, chatID)
		app.clientError(w, http.StatusForbidden)
		return
	}

	messages, err := app.messages.GetByChatID(chatID)
	if err != nil {
		app.errorLog.Printf("Error getting messages: %v", err)
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("Retrieved %d messages for chat: %d", len(messages), chatID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"messages": messages,
	})
}

func (app *application) joinChat(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Printf("Error parsing form in joinChat: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	chatID, err := strconv.Atoi(r.PostForm.Get("chat_id"))
	if err != nil {
		app.errorLog.Printf("Error parsing chat_id: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := joinChatForm{
		ChatID: chatID,
	}

	form.CheckField(validator.NotBlank(strconv.Itoa(form.ChatID)), "chat_id", "this field cannot be empty")

	if !form.Valid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(form.FieldErrors)
		if err != nil {
			app.errorLog.Printf("Error encoding form errors: %v", err)
			app.serverError(w, err)
			return
		}
		return
	}
	private, err := app.chats.IsPrivate(chatID)
	if err != nil {
		app.errorLog.Printf("Error checking if chat is private: %v", err)
		app.serverError(w, err)
		return
	}
	if private {
		app.errorLog.Printf("Attempt to join private chat %d", chatID)
		app.clientError(w, http.StatusForbidden)
		return
	}

	userID := r.Context().Value("user_id").(int)
	_, err = app.participants.Insert(form.ChatID, userID)
	if err != nil {
		app.errorLog.Printf("Error adding user to chat: %v", err)
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("User %d joined chat %d successfully", userID, form.ChatID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id": form.ChatID,
	})
}

func (app *application) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection: %v", err)
		return
	}

	userID := r.Context().Value("user_id").(int)


	client := &Client{
		hub:    app.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
