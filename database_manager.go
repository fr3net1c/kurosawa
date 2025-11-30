package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type DatabaseManager struct {
	dataDir string
	dbs     map[string]*DBService
	mu      sync.Mutex
}

func NewDatabaseManager(dataDir string) (*DatabaseManager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create data directory: %w", err)
	}
	return &DatabaseManager{
		dataDir: dataDir,
		dbs:     make(map[string]*DBService),
	}, nil
}

func (m *DatabaseManager) GetUserDB(userID string) (*DBService, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if db, ok := m.dbs[userID]; ok {
		return db, nil
	}

	dbPath := filepath.Join(m.dataDir, fmt.Sprintf("%s.db", userID))
	db, err := NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not create new DB service for user %s: %w", userID, err)
	}

	m.dbs[userID] = db
	return db, nil
}

func (m *DatabaseManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, db := range m.dbs {
		db.Close()
	}
}

func (m *DatabaseManager) ClearUserHistory(userID string) error {
	db, err := m.GetUserDB(userID)
	if err != nil {
		return err
	}
	return db.ClearHistory()
}

func (m *DatabaseManager) DeleteUserDB(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if db, ok := m.dbs[userID]; ok {
		db.Close()
		delete(m.dbs, userID)
	}

	dbPath := filepath.Join(m.dataDir, fmt.Sprintf("%s.db", userID))
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(dbPath)
}
