package auth

import "testing"

func TestHashAndCompareCorrectPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty string")
	}

	if err := Compare(hash, password); err != nil {
		t.Errorf("Compare() error = %v, want nil", err)
	}
}

func TestHashAndCompareWrongPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if err := Compare(hash, "wrongpassword"); err == nil {
		t.Error("Compare() expected error for wrong password, got nil")
	}
}

func TestHashProducesUniqueValues(t *testing.T) {
	password := "samepassword"

	hash1, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	hash2, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash() should produce different hashes for same password (bcrypt salting)")
	}
}
