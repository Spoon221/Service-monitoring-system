package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка ping БД: %w", err)
	}

	return &DB{db}, nil
}

func Migrate(db *DB) error {
	// Создание таблицы сервисов
	createServicesTable := `
	CREATE TABLE IF NOT EXISTS services (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		url VARCHAR(500) NOT NULL,
		check_interval INTEGER DEFAULT 30,
		timeout INTEGER DEFAULT 10,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Создание таблицы проверок
	createChecksTable := `
	CREATE TABLE IF NOT EXISTS health_checks (
		id SERIAL PRIMARY KEY,
		service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
		status VARCHAR(50) NOT NULL,
		response_time INTEGER,
		error_message TEXT,
		checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Создание таблицы алертов
	createAlertsTable := `
	CREATE TABLE IF NOT EXISTS alerts (
		id SERIAL PRIMARY KEY,
		service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
		message TEXT NOT NULL,
		severity VARCHAR(50) DEFAULT 'warning',
		is_resolved BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP
	);`

	// Создание индексов
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_health_checks_service_id ON health_checks(service_id);
	CREATE INDEX IF NOT EXISTS idx_health_checks_checked_at ON health_checks(checked_at);
	CREATE INDEX IF NOT EXISTS idx_alerts_service_id ON alerts(service_id);
	CREATE INDEX IF NOT EXISTS idx_alerts_is_resolved ON alerts(is_resolved);
	`

	queries := []string{createServicesTable, createChecksTable, createAlertsTable, createIndexes}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("ошибка миграции: %w", err)
		}
	}

	return nil
}
