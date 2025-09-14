package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DBService struct {
	db *sql.DB
}

func NewDB(dataSourceName string) (*DBService, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	service := &DBService{db: db}
	if err = service.createTables(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *DBService) createTables() error {
	createQuery := `
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT NOT NULL,
        user_name TEXT NOT NULL,
        role TEXT NOT NULL,
        content TEXT NOT NULL,
        timestamp DATETIME NOT NULL
    );`
	if _, err := s.db.Exec(createQuery); err != nil {
		return err
	}

	// Simple migration to add the user_name column to older databases.
	alterQuery := `ALTER TABLE messages ADD COLUMN user_name TEXT NOT NULL DEFAULT ''`
	if _, err := s.db.Exec(alterQuery); err != nil {
		// It's safe to ignore the "duplicate column name" error.
		if strings.Contains(err.Error(), "duplicate column name") {
			return nil
		}
		return err
	}

	return nil
}

func (s *DBService) AddMessage(userID, userName, role, content string) error {
	query := `INSERT INTO messages (user_id, user_name, role, content, timestamp) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, userID, userName, role, content, time.Now())
	return err
}

func (s *DBService) GetMessages(userID string) ([]Message, error) {
	query := `SELECT user_name, role, content, timestamp FROM messages WHERE user_id = ? ORDER BY timestamp ASC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var timestamp time.Time
		if err := rows.Scan(&msg.UserName, &msg.Role, &msg.Content, &timestamp); err != nil {
			return nil, err
		}
		msg.Time = timestamp.Format(time.RFC3339)
		messages = append(messages, msg)
	}
	return messages, nil
}

func (s *DBService) TrimHistory(userID string, keepCount int) error {
	var timestamp time.Time
	query := `
        SELECT timestamp FROM messages 
        WHERE user_id = ? 
        ORDER BY timestamp DESC 
        LIMIT 1 OFFSET ?`

	err := s.db.QueryRow(query, userID, keepCount-1).Scan(&timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("could not get trim timestamp: %w", err)
	}

	deleteQuery := `DELETE FROM messages WHERE user_id = ? AND timestamp < ?`
	_, err = s.db.Exec(deleteQuery, userID, timestamp)
	if err != nil {
		return fmt.Errorf("could not delete old messages: %w", err)
	}

	return nil
}

func (s *DBService) Close() {
	s.db.Close()
}
