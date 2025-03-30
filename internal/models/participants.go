package models

import (
	"database/sql"
	"time"
)

type Participant struct {
	ID      int
	ChatID  int
	UserID  int
	Created time.Time
}

type ParticipantModel struct {
	DB *sql.DB
}

func (m *ParticipantModel) Insert(chatID, userID int) (int, error) {
	var id int
	q := `INSERT INTO participants (chat_id, user_id) VALUES (?, ?) RETURNING id`
	err := m.DB.QueryRow(q, chatID, userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *ParticipantModel) GetByChatID(chatID int) ([]*Participant, error) {
	q := `SELECT id, chat_id, user_id, created FROM participants WHERE chat_id = ?`
	rows, err := m.DB.Query(q, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := []*Participant{}
	for rows.Next() {
		var p Participant
		err := rows.Scan(&p.ID, &p.ChatID, &p.UserID, &p.Created)
		if err != nil {
			return nil, err
		}
		participants = append(participants, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return participants, nil
}

func (m *ParticipantModel) GetByUserID(userID int) ([]*Participant, error) {
	stmt := `SELECT id, chat_id, user_id, created FROM participants WHERE user_id = ?`
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := []*Participant{}
	for rows.Next() {
		var p Participant
		err := rows.Scan(&p.ID, &p.ChatID, &p.UserID, &p.Created)
		if err != nil {
			return nil, err
		}
		participants = append(participants, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return participants, nil
}
