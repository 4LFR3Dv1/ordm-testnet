#!/bin/bash

# 🛡️ Script: Parte 1.3 - Proteção contra Ataques
# Descrição: Implementa CSRF Protection, Input Validation e HTTPS Obrigatório

set -e

echo "🛡️ [$(date)] Iniciando Parte 1.3: Proteção contra Ataques"
echo "========================================================="

# 1.3.1 - CSRF Protection
echo "🛡️ 1.3.1 - Implementando CSRF Protection..."

mkdir -p pkg/middleware
cat > pkg/middleware/csrf.go << 'EOF'
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
EOF

# 1.3.2 - Input Validation
echo "✅ 1.3.2 - Implementando Input Validation..."

mkdir -p pkg/validation
cat > pkg/validation/input.go << 'EOF'
package validation

import (
	"fmt"
	"regexp"
	"unicode"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateWalletAddress(address string) error {
	if len(address) == 0 {
		return fmt.Errorf("endereço da wallet não pode estar vazio")
	}

	if len(address) < 10 {
		return fmt.Errorf("endereço da wallet muito curto")
	}

	if len(address) > 100 {
		return fmt.Errorf("endereço da wallet muito longo")
	}

	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(address) {
		return fmt.Errorf("endereço da wallet contém caracteres inválidos")
	}

	return nil
}

func (v *Validator) ValidatePIN(pin string) error {
	if len(pin) != 8 {
		return fmt.Errorf("PIN deve ter exatamente 8 dígitos")
	}

	for _, char := range pin {
		if !unicode.IsDigit(char) {
			return fmt.Errorf("PIN deve conter apenas números")
		}
	}

	if v.isSequential(pin) {
		return fmt.Errorf("PIN não pode ser sequencial")
	}

	if v.isRepetitive(pin) {
		return fmt.Errorf("PIN não pode ser repetitivo")
	}

	return nil
}

func (v *Validator) ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("nome de usuário deve ter pelo menos 3 caracteres")
	}

	if len(username) > 50 {
		return fmt.Errorf("nome de usuário deve ter no máximo 50 caracteres")
	}

	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(username) {
		return fmt.Errorf("nome de usuário contém caracteres inválidos")
	}

	return nil
}

func (v *Validator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("senha deve ter pelo menos 8 caracteres")
	}

	if len(password) > 128 {
		return fmt.Errorf("senha deve ter no máximo 128 caracteres")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("senha deve conter pelo menos uma letra maiúscula")
	}
	if !hasLower {
		return fmt.Errorf("senha deve conter pelo menos uma letra minúscula")
	}
	if !hasDigit {
		return fmt.Errorf("senha deve conter pelo menos um número")
	}
	if !hasSpecial {
		return fmt.Errorf("senha deve conter pelo menos um caractere especial")
	}

	return nil
}

func (v *Validator) ValidateAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("quantidade deve ser maior que zero")
	}

	if amount > 1000000000 {
		return fmt.Errorf("quantidade muito alta")
	}

	return nil
}

func (v *Validator) SanitizeInput(input string) string {
	dangerous := []string{"<script>", "</script>", "javascript:", "onload=", "onerror="}
	sanitized := input
	
	for _, danger := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, danger, "")
	}
	
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000]
	}
	
	return sanitized
}

func (v *Validator) isSequential(pin string) bool {
	if len(pin) < 3 {
		return false
	}

	ascending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			ascending = false
			break
		}
	}

	descending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]-1 {
			descending = false
			break
		}
	}

	return ascending || descending
}

func (v *Validator) isRepetitive(pin string) bool {
	if len(pin) < 3 {
		return false
	}

	first := rune(pin[0])
	for _, char := range pin {
		if char != first {
			return false
		}
	}

	return true
}
EOF

# 1.3.3 - HTTPS Obrigatório
echo "🔒 1.3.3 - Implementando HTTPS Obrigatório..."

mkdir -p pkg/server
cat > pkg/server/https.go << 'EOF'
package server

import (
	"crypto/tls"
	"context"
	"fmt"
	"net/http"
	"os"

	"ordm-main/pkg/config"
)

type HTTPSServer struct {
	server *http.Server
}

func SetupHTTPS(handler http.Handler, port string) (*HTTPSServer, error) {
	if !config.IsProduction() {
		return &HTTPSServer{
			server: &http.Server{
				Addr:    ":" + port,
				Handler: handler,
			},
		}, nil
	}

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("certificados SSL não configurados")
	}

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	server := &http.Server{
		Addr:      ":" + port,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	return &HTTPSServer{server: server}, nil
}

func (h *HTTPSServer) Start() error {
	if config.IsProduction() {
		return h.server.ListenAndServeTLS("", "")
	}
	return h.server.ListenAndServe()
}

func (h *HTTPSServer) Shutdown() error {
	return h.server.Shutdown(context.Background())
}

func ForceHTTPS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.IsProduction() {
			if r.Header.Get("X-Forwarded-Proto") != "https" && r.TLS == nil {
				httpsURL := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
				return
			}
		}
		next(w, r)
	}
}

func SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		if config.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		next(w, r)
	}
}
EOF

echo "✅ [$(date)] Parte 1.3: Proteção contra Ataques concluída!"
echo "📋 Implementações:"
echo "  ✅ 1.3.1 - CSRF Protection implementada"
echo "  ✅ 1.3.2 - Input Validation criada"
echo "  ✅ 1.3.3 - HTTPS Obrigatório configurado"
echo ""
echo "🎉 Parte 1: Segurança Crítica COMPLETA!"
echo "🚀 Próximo: Execute 'part2a_clean_architecture.sh'"

