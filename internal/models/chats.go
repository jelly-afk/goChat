package models

import (
	"database/sql"
	"time"
)

type Chat struct {
	ID        int
	Name      string
	IsPrivate bool
	Created   time.Time
}

type ChatModel struct {
	DB *sql.DB
}

func (m *ChatModel) Insert(name string, isPrivate bool) (int, error) {
	q := `INSERT INTO chats (name, is_private, created) VALUES (?, ?, UTC_TIMESTAMP())`
	result, err := m.DB.Exec(q, name, isPrivate)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
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

func (m *ChatModel) IsPrivate(id int) (bool, error) {
	var isPrivate bool
	q := `SELECT is_private FROM chats WHERE id = ?`
	err := m.DB.QueryRow(q, id).Scan(&isPrivate)
	if err != nil {
		return false, err
	}
	return isPrivate, nil
}
