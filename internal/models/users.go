package models

import (
	"database/sql"
	"time"
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
	q := `INSERT INTO users (username, email, hashed_password, created)
          VALUES (?, ?, ?, UTC_TIMESTAMP())`
	result, err := m.DB.Exec(q, username, email, password)
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
	return 0, nil
}
func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}
