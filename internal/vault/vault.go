package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"sync"

	_ "modernc.org/sqlite"
)

type AccountCredentials struct {
	AccountID string
	Platform  string
	APIKey    string
	Token     string
}

type Vault struct {
	db    *sql.DB
	key   []byte
	mu    sync.RWMutex
	cache map[string]AccountCredentials
}

func cacheKey(accountID, platform string) string {
	return fmt.Sprintf("%s:%s", accountID, platform)
}

func NewVault(dsn string, key []byte) (*Vault, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key size: must be 16, 24, or 32 bytes")
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Vault{
		db:    db,
		key:   key,
		cache: make(map[string]AccountCredentials),
	}, nil
}

func initDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS credentials (
		account_id TEXT NOT NULL,
		platform TEXT NOT NULL,
		api_key TEXT,
		token TEXT,
		PRIMARY KEY (account_id, platform)
	);`
	_, err := db.Exec(query)
	return err
}

func (v *Vault) encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (v *Vault) decrypt(ciphertextBase64 string) (string, error) {
	if ciphertextBase64 == "" {
		return "", nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aesgcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertextBytes := ciphertext[:aesgcm.NonceSize()], ciphertext[aesgcm.NonceSize():]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (v *Vault) Store(ctx context.Context, creds AccountCredentials) error {
	encAPIKey, err := v.encrypt(creds.APIKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt APIKey: %w", err)
	}
	encToken, err := v.encrypt(creds.Token)
	if err != nil {
		return fmt.Errorf("failed to encrypt Token: %w", err)
	}

	query := `
	INSERT INTO credentials (account_id, platform, api_key, token)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(account_id, platform) DO UPDATE SET
		api_key=excluded.api_key,
		token=excluded.token;
	`
	_, err = v.db.ExecContext(ctx, query, creds.AccountID, creds.Platform, encAPIKey, encToken)
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	v.mu.Lock()
	v.cache[cacheKey(creds.AccountID, creds.Platform)] = creds
	v.mu.Unlock()

	return nil
}

func (v *Vault) Retrieve(ctx context.Context, accountID, platform string) (AccountCredentials, error) {
	ck := cacheKey(accountID, platform)
	v.mu.RLock()
	creds, ok := v.cache[ck]
	v.mu.RUnlock()
	if ok {
		return creds, nil
	}

	query := `SELECT account_id, platform, api_key, token FROM credentials WHERE account_id = ? AND platform = ?`
	row := v.db.QueryRowContext(ctx, query, accountID, platform)

	var encAPIKey, encToken string
	err := row.Scan(&creds.AccountID, &creds.Platform, &encAPIKey, &encToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return creds, fmt.Errorf("credentials not found")
		}
		return creds, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	creds.APIKey, err = v.decrypt(encAPIKey)
	if err != nil {
		return creds, fmt.Errorf("failed to decrypt APIKey: %w", err)
	}
	creds.Token, err = v.decrypt(encToken)
	if err != nil {
		return creds, fmt.Errorf("failed to decrypt Token: %w", err)
	}

	v.mu.Lock()
	v.cache[ck] = creds
	v.mu.Unlock()

	return creds, nil
}

func (v *Vault) Delete(ctx context.Context, accountID, platform string) error {
	query := `DELETE FROM credentials WHERE account_id = ? AND platform = ?`
	_, err := v.db.ExecContext(ctx, query, accountID, platform)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	v.mu.Lock()
	delete(v.cache, cacheKey(accountID, platform))
	v.mu.Unlock()

	return nil
}

func (v *Vault) Close() error {
	v.mu.Lock()
	v.cache = make(map[string]AccountCredentials)
	v.mu.Unlock()
	return v.db.Close()
}
