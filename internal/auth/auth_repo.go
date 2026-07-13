package auth

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(email, passwordHash string) (*User, error) {
	result, err := r.db.Exec(
		"INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))",
		email, passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id, _ := result.LastInsertId()
	return r.GetUserByID(id)
}

func (r *Repository) GetUserByEmail(email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = ?", email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return u, nil
}

func (r *Repository) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return u, nil
}
