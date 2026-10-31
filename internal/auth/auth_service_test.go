package auth

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	return db
}

func setupService(t *testing.T) *Service {
	t.Helper()
	db := setupTestDB(t)
	repo := NewRepository(db)
	os.Setenv("SESSION_SECRET", "test-secret-1234567890123456")
	return NewService(repo)
}

func TestSignup_Success(t *testing.T) {
	svc := setupService(t)
	user, token, err := svc.Signup("test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", user.Email)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	svc := setupService(t)
	_, _, err := svc.Signup("dup@example.com", "password123")
	if err != nil {
		t.Fatalf("first signup failed: %v", err)
	}
	_, _, err = svc.Signup("dup@example.com", "password123")
	if err != ErrDuplicateEmail {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestSignup_InvalidEmail(t *testing.T) {
	svc := setupService(t)
	_, _, err := svc.Signup("not-an-email", "password123")
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestSignup_ShortPassword(t *testing.T) {
	svc := setupService(t)
	_, _, err := svc.Signup("test@example.com", "short")
	if err != ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := setupService(t)
	_, _, err := svc.Signup("login@example.com", "correct-password")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	user, token, err := svc.Login("login@example.com", "correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := setupService(t)
	_, _, err := svc.Signup("wrongpw@example.com", "correct-password")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	_, _, err = svc.Login("wrongpw@example.com", "wrong-password")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	svc := setupService(t)
	_, _, err := svc.Login("nobody@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateSession_Valid(t *testing.T) {
	svc := setupService(t)
	_, token, err := svc.Signup("session@example.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	userID, err := svc.ValidateSession(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID == 0 {
		t.Fatal("expected non-zero user ID")
	}
}

func TestValidateSession_Invalid(t *testing.T) {
	svc := setupService(t)
	_, err := svc.ValidateSession("invalid-token")
	if err != ErrNotAuthenticated {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}
