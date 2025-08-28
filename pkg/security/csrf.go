package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CSRFProtection implementa proteção contra ataques CSRF
type CSRFProtection struct {
	SecretKey    string
	TokenLength  int
	TokenTTL     time.Duration
	Tokens       map[string]*CSRFToken
	mu           sync.RWMutex
	CleanupTimer *time.Timer
}

// CSRFToken representa um token CSRF
type CSRFToken struct {
	Token     string
	UserID    string
	IP        string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

// CSRFConfig configuração do CSRF
type CSRFConfig struct {
	SecretKey   string        `json:"secret_key"`
	TokenLength int           `json:"token_length"`
	TokenTTL    time.Duration `json:"token_ttl"`
	CleanupTTL  time.Duration `json:"cleanup_ttl"`
}

// NewCSRFProtection cria nova instância de proteção CSRF
func NewCSRFProtection(config *CSRFConfig) *CSRFProtection {
	if config == nil {
		config = &CSRFConfig{
			SecretKey:   generateCSRFSecret(),
			TokenLength: 32,
			TokenTTL:    30 * time.Minute,
			CleanupTTL:  1 * time.Hour,
		}
	}

	csrf := &CSRFProtection{
		SecretKey:   config.SecretKey,
		TokenLength: config.TokenLength,
		TokenTTL:    config.TokenTTL,
		Tokens:      make(map[string]*CSRFToken),
	}

	// Iniciar limpeza automática
	csrf.startCleanup(config.CleanupTTL)

	return csrf
}

// GenerateToken gera novo token CSRF
func (csrf *CSRFProtection) GenerateToken(userID, ip, userAgent string) (string, error) {
	// Gerar token aleatório
	tokenBytes := make([]byte, csrf.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("erro ao gerar token: %v", err)
	}

	// Codificar em base64
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Criar entrada de token
	csrfToken := &CSRFToken{
		Token:     token,
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(csrf.TokenTTL),
		Used:      false,
	}

	// Armazenar token
	csrf.mu.Lock()
	csrf.Tokens[token] = csrfToken
	csrf.mu.Unlock()

	return token, nil
}

// ValidateToken valida token CSRF
func (csrf *CSRFProtection) ValidateToken(token, userID, ip, userAgent string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token CSRF não fornecido")
	}

	csrf.mu.RLock()
	csrfToken, exists := csrf.Tokens[token]
	csrf.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("token CSRF inválido")
	}

	// Verificar expiração
	if time.Now().After(csrfToken.ExpiresAt) {
		csrf.removeToken(token)
		return false, fmt.Errorf("token CSRF expirado")
	}

	// Verificar se já foi usado
	if csrfToken.Used {
		csrf.removeToken(token)
		return false, fmt.Errorf("token CSRF já foi usado")
	}

	// Verificar usuário
	if csrfToken.UserID != userID {
		return false, fmt.Errorf("token CSRF não pertence ao usuário")
	}

	// Verificar IP (opcional, pode ser muito restritivo)
	if csrfToken.IP != ip {
		// Log de segurança, mas não falha
		fmt.Printf("⚠️ CSRF IP mismatch: expected %s, got %s\n", csrfToken.IP, ip)
	}

	// Verificar User-Agent (opcional)
	if csrfToken.UserAgent != userAgent {
		// Log de segurança, mas não falha
		fmt.Printf("⚠️ CSRF User-Agent mismatch: expected %s, got %s\n", csrfToken.UserAgent, userAgent)
	}

	// Marcar como usado
	csrf.mu.Lock()
	csrfToken.Used = true
	csrf.mu.Unlock()

	return true, nil
}

// removeToken remove token do cache
func (csrf *CSRFProtection) removeToken(token string) {
	csrf.mu.Lock()
	delete(csrf.Tokens, token)
	csrf.mu.Unlock()
}

// startCleanup inicia limpeza automática de tokens expirados
func (csrf *CSRFProtection) startCleanup(cleanupTTL time.Duration) {
	csrf.CleanupTimer = time.AfterFunc(cleanupTTL, func() {
		csrf.cleanupExpiredTokens()
		csrf.startCleanup(cleanupTTL) // Agendar próxima limpeza
	})
}

// cleanupExpiredTokens remove tokens expirados
func (csrf *CSRFProtection) cleanupExpiredTokens() {
	now := time.Now()
	expiredTokens := []string{}

	csrf.mu.RLock()
	for token, csrfToken := range csrf.Tokens {
		if now.After(csrfToken.ExpiresAt) {
			expiredTokens = append(expiredTokens, token)
		}
	}
	csrf.mu.RUnlock()

	// Remover tokens expirados
	csrf.mu.Lock()
	for _, token := range expiredTokens {
		delete(csrf.Tokens, token)
	}
	csrf.mu.Unlock()

	if len(expiredTokens) > 0 {
		fmt.Printf("🧹 CSRF: %d tokens expirados removidos\n", len(expiredTokens))
	}
}

// generateCSRFSecret gera chave secreta para CSRF
func generateCSRFSecret() string {
	secret := make([]byte, 32)
	rand.Read(secret)
	return base64.URLEncoding.EncodeToString(secret)
}

// CSRFMiddleware middleware para proteção CSRF
func CSRFMiddleware(csrf *CSRFProtection) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Aplicar apenas em métodos que modificam dados
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next(w, r)
				return
			}

			// Extrair token do header ou form
			token := extractCSRFToken(r)
			if token == "" {
				http.Error(w, "Token CSRF não fornecido", http.StatusForbidden)
				return
			}

			// Extrair informações do usuário
			userID := extractUserID(r)
			ip := extractIP(r)
			userAgent := r.UserAgent()

			// Validar token
			valid, err := csrf.ValidateToken(token, userID, ip, userAgent)
			if !valid {
				http.Error(w, fmt.Sprintf("Token CSRF inválido: %v", err), http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

// extractCSRFToken extrai token CSRF da requisição
func extractCSRFToken(r *http.Request) string {
	// Tentar extrair do header
	token := r.Header.Get("X-CSRF-Token")
	if token != "" {
		return token
	}

	// Tentar extrair do form
	token = r.FormValue("csrf_token")
	if token != "" {
		return token
	}

	// Tentar extrair do JSON body (para APIs)
	if r.Header.Get("Content-Type") == "application/json" {
		// Implementar parsing de JSON se necessário
		return ""
	}

	return ""
}

// extractUserID extrai ID do usuário da requisição
func extractUserID(r *http.Request) string {
	// Tentar extrair de diferentes fontes
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		return userID
	}

	userID = r.FormValue("user_id")
	if userID != "" {
		return userID
	}

	// Tentar extrair de JWT token se disponível
	// Implementar conforme necessário

	return "anonymous"
}

// extractIP extrai IP real do cliente
func extractIP(r *http.Request) string {
	// Verificar headers de proxy
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// Pegar primeiro IP da lista
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Client-IP")
	if ip != "" {
		return ip
	}

	// Usar IP remoto
	return r.RemoteAddr
}

// GetTokenStats retorna estatísticas dos tokens
func (csrf *CSRFProtection) GetTokenStats() map[string]interface{} {
	csrf.mu.RLock()
	defer csrf.mu.RUnlock()

	now := time.Now()
	active := 0
	expired := 0
	used := 0

	for _, token := range csrf.Tokens {
		if now.After(token.ExpiresAt) {
			expired++
		} else {
			active++
		}
		if token.Used {
			used++
		}
	}

	return map[string]interface{}{
		"total_tokens":   len(csrf.Tokens),
		"active_tokens":  active,
		"expired_tokens": expired,
		"used_tokens":    used,
		"token_ttl":      csrf.TokenTTL.String(),
	}
}

// Shutdown para a limpeza automática
func (csrf *CSRFProtection) Shutdown() {
	if csrf.CleanupTimer != nil {
		csrf.CleanupTimer.Stop()
	}
}
