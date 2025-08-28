package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"ordm-main/pkg/config"
)

type CSRFProtection struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

func NewCSRFProtection() *CSRFProtection {
	return &CSRFProtection{
		tokens: make(map[string]time.Time),
	}
}

func (csrf *CSRFProtection) GenerateToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	csrf.mu.Lock()
	csrf.tokens[token] = time.Now().Add(30 * time.Minute)
	csrf.mu.Unlock()

	return token, nil
}

func (csrf *CSRFProtection) ValidateToken(token string) bool {
	csrf.mu.Lock()
	defer csrf.mu.Unlock()

	expiresAt, exists := csrf.tokens[token]
	if !exists {
		return false
	}

	if time.Now().After(expiresAt) {
		delete(csrf.tokens, token)
		return false
	}

	delete(csrf.tokens, token)
	return true
}

func (csrf *CSRFProtection) CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			token, err := csrf.GenerateToken()
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					Secure:   config.IsProduction(),
					SameSite: http.SameSiteStrictMode,
					MaxAge:   1800,
				})
			}
			next(w, r)
			return
		}

		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("csrf_token")
			}

			if !csrf.ValidateToken(token) {
				http.Error(w, "CSRF token inválido", http.StatusForbidden)
				return
			}
		}

		next(w, r)
	}
}
