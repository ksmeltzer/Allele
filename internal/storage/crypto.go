package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var encryptionKey []byte
const encPrefix = "enc:v1:"

func InitCrypto(dbPath string) error {
	keyPath := filepath.Join(filepath.Dir(dbPath), "master.key")
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate new 32-byte key
			newKey := make([]byte, 32)
			if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
				return fmt.Errorf("failed to generate random key: %w", err)
			}
			keyHex := hex.EncodeToString(newKey)
			if err := os.WriteFile(keyPath, []byte(keyHex), 0600); err != nil {
				return fmt.Errorf("failed to write master.key: %w", err)
			}
			encryptionKey = newKey
			return nil
		}
		return fmt.Errorf("failed to read master.key: %w", err)
	}
	
	decoded, err := hex.DecodeString(strings.TrimSpace(string(keyBytes)))
	if err != nil || len(decoded) != 32 {
		return fmt.Errorf("invalid master.key format, expected 32-byte hex string")
	}
	encryptionKey = decoded
	return nil
}

func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(encryptionKey)
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
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptSecret(ciphertextBase64 string) (string, error) {
	if !strings.HasPrefix(ciphertextBase64, encPrefix) {
		return ciphertextBase64, nil // Not encrypted, return as plaintext for backwards compatibility
	}
	
	rawBase64 := strings.TrimPrefix(ciphertextBase64, encPrefix)
	ciphertext, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}
	block, err := aes.NewCipher(encryptionKey)
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
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}
