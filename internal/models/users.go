package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID       int
	Username string
	Email    string
	Password string
	Created  time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(username, email, Password string) (int, error) {
	q := `INSERT INTO users (username, email, password, created)
          VALUES (?, ?, ?, UTC_TIMESTAMP())`
	result, err := m.DB.Exec(q, username, email, Password)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}
