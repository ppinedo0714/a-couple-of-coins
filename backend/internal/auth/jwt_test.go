package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-secret-that-is-at-least-32-chars-long"

func TestIssueAndValidate(t *testing.T) {
	userID := uuid.New()

	token, err := Issue(testSecret, userID, time.Hour)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if token == "" {
		t.Fatal("Issue() returned empty token")
	}

	got, err := Validate(testSecret, token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got != userID {
		t.Errorf("Validate() = %v, want %v", got, userID)
	}
}

func TestValidateTamperedToken(t *testing.T) {
	userID := uuid.New()
	token, err := Issue(testSecret, userID, time.Hour)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Tamper with the token by appending a character
	tampered := token + "x"
	_, err = Validate(testSecret, tampered)
	if err == nil {
		t.Fatal("Validate() expected error for tampered token, got nil")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	userID := uuid.New()
	token, err := Issue(testSecret, userID, -time.Second)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	_, err = Validate(testSecret, token)
	if err == nil {
		t.Fatal("Validate() expected error for expired token, got nil")
	}
}

func TestValidateWrongSecret(t *testing.T) {
	userID := uuid.New()
	token, err := Issue(testSecret, userID, time.Hour)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	_, err = Validate("wrong-secret-also-at-least-32-characters", token)
	if err == nil {
		t.Fatal("Validate() expected error for wrong secret, got nil")
	}
}
