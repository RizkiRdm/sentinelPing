package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	user, token, err := h.svc.Signup(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrInvalidPassword):
			writeError(w, http.StatusBadRequest, err.Error(), "validation failed")
		case errors.Is(err, ErrDuplicateEmail):
			writeError(w, http.StatusBadRequest, "duplicate_email", "email already registered")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		}
		return
	}

	setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, authResponse{ID: user.ID, Email: user.Email})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	user, token, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		}
		return
	}

	setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, authResponse{ID: user.ID, Email: user.Email})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "not_authenticated", "not authenticated")
		return
	}

	user, err := h.svc.repo.GetUserByID(userID)
	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "not_authenticated", "not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{ID: user.ID, Email: user.Email})
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(24 * time.Hour / time.Second),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}
