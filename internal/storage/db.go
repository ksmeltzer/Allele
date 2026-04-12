package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	tradesTable := `CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY,
		strategy TEXT,
		condition_id TEXT,
		profit REAL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	strategiesTable := `CREATE TABLE IF NOT EXISTS strategies (
		id TEXT PRIMARY KEY,
		min_profit REAL,
		active BOOLEAN,
		fitness REAL
	);`

	if _, err := DB.Exec(tradesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(strategiesTable); err != nil {
		return err
	}
	return nil
}

func LogTrade(strategy, conditionID string, profit float64) error {
	query := `INSERT INTO trades (strategy, condition_id, profit) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, strategy, conditionID, profit)
	return err
}
