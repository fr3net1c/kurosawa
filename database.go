package main

import (
	"database/sql"
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
			// Continue anyway
		} else {
			return err
		}
	}

	// Create user preferences table
	userPrefsQuery := `
	CREATE TABLE IF NOT EXISTS user_preferences (
		user_id TEXT PRIMARY KEY,
		provider TEXT NOT NULL DEFAULT 'none',
		model TEXT NOT NULL DEFAULT 'none'
	);`
	if _, err := s.db.Exec(userPrefsQuery); err != nil {
		return err
	}

	// Add provider and model columns to user_preferences if they don't exist
	alterProviderQuery := `ALTER TABLE user_preferences ADD COLUMN provider TEXT NOT NULL DEFAULT 'none'`
	if _, err := s.db.Exec(alterProviderQuery); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			// Continue anyway - table might not exist yet
		}
	}

	alterModelQuery := `ALTER TABLE user_preferences ADD COLUMN model TEXT NOT NULL DEFAULT 'none'`
	if _, err := s.db.Exec(alterModelQuery); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			// Continue anyway
		}
	}

	return nil
}

func (s *DBService) AddMessage(userID, userName, role, content string) error {
	query := `INSERT INTO messages (user_id, user_name, role, content, timestamp) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, userID, userName, role, content, time.Now())
	return err
}

func (s *DBService) GetMessages() ([]Message, error) {
	query := `SELECT user_name, role, content, timestamp FROM messages ORDER BY timestamp ASC`
	rows, err := s.db.Query(query)
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

func (s *DBService) Close() {
	s.db.Close()
}

func (s *DBService) ClearHistory() error {
	query := `DELETE FROM messages`
	_, err := s.db.Exec(query)
	return err
}

func (s *DBService) SetUserPreference(userID, provider, model string) error {
	query := `INSERT INTO user_preferences (user_id, provider, model) VALUES (?, ?, ?)
	         ON CONFLICT(user_id) DO UPDATE SET provider = ?, model = ?`
	_, err := s.db.Exec(query, userID, provider, model, provider, model)
	return err
}

func (s *DBService) GetUserPreference(userID string) (provider, model string, err error) {
	query := `SELECT provider, model FROM user_preferences WHERE user_id = ?`
	err = s.db.QueryRow(query, userID).Scan(&provider, &model)
	if err == sql.ErrNoRows {
		return "none", "none", nil
	}
	return provider, model, err
}
