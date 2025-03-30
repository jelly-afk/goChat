package models

import (
	"database/sql"
	"time"
)

type Chat struct {
	ID        int
	Name      string
	IsPrivate bool
	CreatedAt time.Time
}

type ChatModel struct {
	DB *sql.DB
}

func (m *ChatModel) Insert(name string, isPrivate bool) (int, error) {
	var id int
	q := `INSERT INTO chats (name, is_private) VALUES (?, ?) RETURNING id`
	err := m.DB.QueryRow(q, name, isPrivate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *ChatModel) ExistsId(id int) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT true FROM chats WHERE id = ?)`
	err := m.DB.QueryRow(q, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m *ChatModel) ExistsName(name string) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT true FROM chats WHERE name = ?)`
	err := m.DB.QueryRow(q, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}