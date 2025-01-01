package pkg

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5" // PostgreSQLドライバ
)

// DBManager is a database manager.
type DBManager struct {
	mu       sync.RWMutex
	clients  map[string]*pgx.Conn
	settings map[string]DBConfig
}

// DBConfig is a database configuration.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDBManager creates a new DBManager.
func NewDBManager(settings map[string]DBConfig) *DBManager {
	return &DBManager{
		clients:  make(map[string]*pgx.Conn),
		settings: settings,
	}
}

// GetClient returns a database client.
func (m *DBManager) GetClient(dbName string) (*pgx.Conn, error) {
	m.mu.RLock()
	client, exists := m.clients[dbName]
	m.mu.RUnlock()

	if exists {
		return client, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists = m.clients[dbName]
	if exists {
		return client, nil
	}

	config, ok := m.settings[dbName]
	if !ok {
		return nil, fmt.Errorf("database configuration for %s not found", dbName)
	}

	connCfg := pgx.ConnConfig{
		Host:     config.Host,
		Port:     uint16(config.Port),
		User:     config.User,
		Password: config.Password,
		Database: config.DBName,
	}
	db, err := pgx.Connect(connCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", dbName, err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database %s: %w", dbName, err)
	}

	m.clients[dbName] = db
	return db, nil
}

// CloseAll closes all database clients.
func (m *DBManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			log.Printf("failed to close database %s: %v", name, err)
		}
		delete(m.clients, name)
	}
}
