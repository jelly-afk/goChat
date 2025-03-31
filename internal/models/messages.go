package models

import (
	"database/sql"
	"time"
)

type Message struct {
	ID       int
	ChatID   int
	SenderID int
	Content  string
	Created  time.Time
}

type MessageModel struct {
	DB *sql.DB
}

func (m *MessageModel) Insert(chatID, senderID int, content string) (int, error) {
	q := `INSERT INTO messages (chat_id, sender_id, content, created) VALUES (?, ?, ?, UTC_TIMESTAMP())`
	result, err := m.DB.Exec(q, chatID, senderID, content)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *MessageModel) GetByChatID(chatID int) ([]*Message, error) {
	q := `SELECT id, chat_id, sender_id, content, created FROM messages WHERE chat_id = ? ORDER BY created ASC`
	rows, err := m.DB.Query(q, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.Created)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
