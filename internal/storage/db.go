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

	if err := InitCrypto(dbPath); err != nil {
		return fmt.Errorf("failed to initialize encryption layer: %w", err)
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

	// This table replaces the .env file. The UI collects config params defined
	// by the plugin's Manifest and stores them here so the engine can inject them.
	configTable := `CREATE TABLE IF NOT EXISTS plugin_config (
		plugin_name TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY (plugin_name, key)
	);`

	if _, err := DB.Exec(tradesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(strategiesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(configTable); err != nil {
		return err
	}
	return nil
}

func LogTrade(strategy, conditionID string, profit float64) error {
	query := `INSERT INTO trades (strategy, condition_id, profit) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, strategy, conditionID, profit)
	return err
}

// HasPluginConfig checks if a configuration key exists in the database for a given plugin.
func HasPluginConfig(pluginName, key string) bool {
	query := `SELECT 1 FROM plugin_config WHERE plugin_name = ? AND key = ?`
	var exists int
	err := DB.QueryRow(query, pluginName, key).Scan(&exists)
	return err == nil
}

// GetPluginConfig returns the configuration for a given plugin key, avoiding the need for .env variables.
func GetPluginConfig(pluginName, key string) (string, error) {
	query := `SELECT value FROM plugin_config WHERE plugin_name = ? AND key = ?`
	var val string
	err := DB.QueryRow(query, pluginName, key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // config not set yet
		}
		return "", err
	}
	decrypted, err := DecryptSecret(val)
	if err == nil {
		return decrypted, nil
	}
	return val, nil
}

// SetPluginConfig stores a UI-collected parameter into the database.
func SetPluginConfig(pluginName, key, value string, isSecret bool) error {
	if isSecret {
		enc, err := EncryptSecret(value)
		if err == nil {
			value = enc
		}
	}
	query := `INSERT INTO plugin_config (plugin_name, key, value) 
		VALUES (?, ?, ?) 
		ON CONFLICT(plugin_name, key) DO UPDATE SET value=excluded.value;`
	_, err := DB.Exec(query, pluginName, key, value)
	return err
}
