package auth

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type SessionClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

type Service struct {
	repo          *Repository
	sessionSecret []byte
}

func NewService(repo *Repository) *Service {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "dev-secret-do-not-use-in-production"
	}
	return &Service{
		repo:          repo,
		sessionSecret: []byte(secret),
	}
}

func (s *Service) Signup(email, password string) (*User, string, error) {
	if !emailRegex.MatchString(email) {
		return nil, "", ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, "", ErrInvalidPassword
	}

	existing, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, "", ErrDuplicateEmail
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(email, string(hash))
	if err != nil {
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) Login(email, password string) (*User, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) ValidateSession(tokenStr string) (int64, error) {
	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.sessionSecret, nil
	})
	if err != nil {
		return 0, ErrNotAuthenticated
	}
	if !token.Valid {
		return 0, ErrNotAuthenticated
	}
	return claims.UserID, nil
}

func (s *Service) issueToken(userID int64) (string, error) {
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.sessionSecret)
}
