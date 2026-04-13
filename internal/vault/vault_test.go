package vault

import (
	"context"
	"testing"
)

func TestVault(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256
	v, err := NewVault(":memory:", key)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}
	defer v.Close()

	ctx := context.Background()

	creds := AccountCredentials{
		AccountID: "user-1",
		Platform:  "twitter",
		APIKey:    "key-123",
		Token:     "token-456",
	}

	err = v.Store(ctx, creds)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	retrieved, err := v.Retrieve(ctx, "user-1", "twitter")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if retrieved.APIKey != creds.APIKey || retrieved.Token != creds.Token {
		t.Fatalf("expected %+v, got %+v", creds, retrieved)
	}

	err = v.Delete(ctx, "user-1", "twitter")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = v.Retrieve(ctx, "user-1", "twitter")
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	}
}

func TestVault_EncryptionFailsWithBadKey(t *testing.T) {
	_, err := NewVault(":memory:", []byte("short"))
	if err == nil {
		t.Fatal("expected error with invalid key size, got nil")
	}
}
