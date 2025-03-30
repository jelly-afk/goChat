package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Username       string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(username, email, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return 0, err
	}
	q := `INSERT INTO users (username, email, hashed_password, created)
          VALUES (?, ?, ?, UTC_TIMESTAMP())`
	result, err := m.DB.Exec(q, username, email, hashedPassword)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	q := `SELECT id, hashed_password FROM users WHERE email = ?`
	err := m.DB.QueryRow(q, email).Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, ErrInvalidCredentials
	}
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, ErrInvalidCredentials
	}
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *UserModel) ExistsId(id int) (bool, error) {
	var exists bool

	q := `SELECT EXISTS(SELECT true FROM users WHERE id = ?);`
	err := m.DB.QueryRow(q, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m *UserModel) ExistsEmail(email string) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT true FROM users WHERE email = ?);`
	err := m.DB.QueryRow(q, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m *UserModel) ExistsUsername(username string) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT true FROM users WHERE username = ?);`
	err := m.DB.QueryRow(q, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
