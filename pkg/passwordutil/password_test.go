package passwordutil

import "testing"

func TestHashAndCompare(t *testing.T) {
	password := "super-secret"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash == "" {
		t.Fatal("Hash() returned empty hash")
	}

	if hash == password {
		t.Fatal("hash should not equal original password")
	}

	if err := Compare(hash, password); err != nil {
		t.Fatalf("Compare() valid password error = %v", err)
	}

	if err := Compare(hash, "wrong-password"); err == nil {
		t.Fatal("Compare() expected error for wrong password, got nil")
	}
}