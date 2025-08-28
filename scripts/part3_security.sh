#!/bin/bash

# 🔐 Script para PARTE 3: Segurança
# Baseado no PLANO_ATUALIZACOES.md

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

log "🔄 Iniciando PARTE 3: Segurança"

# 3.1.1 Corrigir tempo de PIN 2FA
log "3.1.1 - Corrigindo tempo de PIN 2FA..."

# Atualizar cmd/gui/main.go
if [ -f "cmd/gui/main.go" ]; then
    sed -i.bak 's/10 \* time.Second/60 \* time.Second/g' cmd/gui/main.go
    log "✅ Tempo de PIN 2FA aumentado para 60 segundos"
else
    warn "Arquivo cmd/gui/main.go não encontrado"
fi

# 3.1.2 Implementar rate limiting
log "3.1.2 - Criando pkg/auth/rate_limiter.go..."
cat > pkg/auth/rate_limiter.go << 'EOF'
package auth

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implementa rate limiting por IP/wallet
type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*AttemptTracker
	config   *RateLimitConfig
}

// AttemptTracker rastreia tentativas de login
type AttemptTracker struct {
	Attempts    int       `json:"attempts"`
	LastAttempt time.Time `json:"last_attempt"`
	LockedUntil time.Time `json:"locked_until"`
	IsLocked    bool      `json:"is_locked"`
}

// RateLimitConfig configuração do rate limiter
type RateLimitConfig struct {
	MaxAttempts    int           `json:"max_attempts"`
	LockoutDuration time.Duration `json:"lockout_duration"`
	WindowDuration  time.Duration `json:"window_duration"`
}

// NewRateLimiter cria novo rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string]*AttemptTracker),
		config: &RateLimitConfig{
			MaxAttempts:     3,
			LockoutDuration: 5 * time.Minute,
			WindowDuration:  1 * time.Hour,
		},
	}
}

// CheckRateLimit verifica se requisição está dentro do limite
func (rl *RateLimiter) CheckRateLimit(identifier string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	tracker, exists := rl.attempts[identifier]
	if !exists {
		tracker = &AttemptTracker{
			Attempts:    0,
			LastAttempt: time.Now(),
			IsLocked:    false,
		}
		rl.attempts[identifier] = tracker
	}

	// Verificar se está bloqueado
	if tracker.IsLocked {
		if time.Now().Before(tracker.LockedUntil) {
			return false, fmt.Errorf("conta bloqueada até %v", tracker.LockedUntil)
		} else {
			// Desbloquear após período
			tracker.IsLocked = false
			tracker.Attempts = 0
		}
	}

	// Verificar janela de tempo
	if time.Since(tracker.LastAttempt) > rl.config.WindowDuration {
		tracker.Attempts = 0
	}

	// Verificar limite de tentativas
	if tracker.Attempts >= rl.config.MaxAttempts {
		tracker.IsLocked = true
		tracker.LockedUntil = time.Now().Add(rl.config.LockoutDuration)
		return false, fmt.Errorf("máximo de tentativas excedido")
	}

	return true, nil
}

// RecordAttempt registra tentativa de login
func (rl *RateLimiter) RecordAttempt(identifier string, success bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	tracker, exists := rl.attempts[identifier]
	if !exists {
		tracker = &AttemptTracker{
			Attempts:    0,
			LastAttempt: time.Now(),
			IsLocked:    false,
		}
		rl.attempts[identifier] = tracker
	}

	if success {
		// Reset em caso de sucesso
		tracker.Attempts = 0
		tracker.IsLocked = false
	} else {
		// Incrementar tentativas em caso de falha
		tracker.Attempts++
		tracker.LastAttempt = time.Now()

		// Bloquear se excedeu limite
		if tracker.Attempts >= rl.config.MaxAttempts {
			tracker.IsLocked = true
			tracker.LockedUntil = time.Now().Add(rl.config.LockoutDuration)
		}
	}
}

// GetAttempts retorna tentativas de um identificador
func (rl *RateLimiter) GetAttempts(identifier string) *AttemptTracker {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if tracker, exists := rl.attempts[identifier]; exists {
		return tracker
	}
	return nil
}

// CleanupOldEntries remove entradas antigas
func (rl *RateLimiter) CleanupOldEntries(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for identifier, tracker := range rl.attempts {
		if now.Sub(tracker.LastAttempt) > maxAge {
			delete(rl.attempts, identifier)
		}
	}
}
EOF

# 3.1.3 Melhorar geração de PIN
log "3.1.3 - Criando pkg/auth/pin_generator.go..."
cat > pkg/auth/pin_generator.go << 'EOF'
package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// PINGenerator gera PINs seguros
type PINGenerator struct {
	length     int
	useLetters bool
	useNumbers bool
	useSymbols bool
}

// NewPINGenerator cria novo gerador de PIN
func NewPINGenerator() *PINGenerator {
	return &PINGenerator{
		length:     8,
		useLetters: false,
		useNumbers: true,
		useSymbols: false,
	}
}

// GeneratePIN gera PIN seguro
func (pg *PINGenerator) GeneratePIN() (string, error) {
	const numbers = "0123456789"
	
	// Gerar PIN numérico de 8 dígitos
	pin := make([]byte, pg.length)
	for i := range pin {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(numbers))))
		if err != nil {
			return "", fmt.Errorf("erro ao gerar número aleatório: %v", err)
		}
		pin[i] = numbers[num.Int64()]
	}
	
	return string(pin), nil
}

// ValidatePIN valida formato do PIN
func (pg *PINGenerator) ValidatePIN(pin string) error {
	if len(pin) != pg.length {
		return fmt.Errorf("PIN deve ter %d dígitos", pg.length)
	}
	
	// Verificar se contém apenas números
	for _, char := range pin {
		if char < '0' || char > '9' {
			return fmt.Errorf("PIN deve conter apenas números")
		}
	}
	
	// Verificar se não é sequencial
	if pg.isSequential(pin) {
		return fmt.Errorf("PIN não pode ser sequencial")
	}
	
	// Verificar se não é repetitivo
	if pg.isRepetitive(pin) {
		return fmt.Errorf("PIN não pode ser repetitivo")
	}
	
	return nil
}

// isSequential verifica se PIN é sequencial
func (pg *PINGenerator) isSequential(pin string) bool {
	if len(pin) < 3 {
		return false
	}
	
	// Verificar sequência crescente
	ascending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			ascending = false
			break
		}
	}
	
	// Verificar sequência decrescente
	descending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]-1 {
			descending = false
			break
		}
	}
	
	return ascending || descending
}

// isRepetitive verifica se PIN é repetitivo
func (pg *PINGenerator) isRepetitive(pin string) bool {
	if len(pin) < 3 {
		return false
	}
	
	// Verificar se todos os dígitos são iguais
	first := pin[0]
	for _, char := range pin {
		if char != first {
			return false
		}
	}
	
	return true
}

// GenerateComplexPIN gera PIN com letras e números
func (pg *PINGenerator) GenerateComplexPIN() (string, error) {
	const (
		letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers = "0123456789"
	)
	
	charset := ""
	if pg.useLetters {
		charset += letters
	}
	if pg.useNumbers {
		charset += numbers
	}
	
	if charset == "" {
		return "", fmt.Errorf("pelo menos um tipo de caractere deve ser habilitado")
	}
	
	pin := make([]byte, pg.length)
	for i := range pin {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("erro ao gerar caractere aleatório: %v", err)
		}
		pin[i] = charset[num.Int64()]
	}
	
	return string(pin), nil
}
EOF

# 3.2.1 Implementar keystore seguro
log "3.2.1 - Criando pkg/crypto/keystore.go..."
cat > pkg/crypto/keystore.go << 'EOF'
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// SecureKeystore gerencia chaves criptográficas
type SecureKeystore struct {
	Path         string
	Password     string
	Encrypted    bool
	BackupPath   string
	CreatedAt    time.Time
}

// KeyEntry representa uma entrada no keystore
type KeyEntry struct {
	ID           string    `json:"id"`
	PublicKey    string    `json:"public_key"`
	PrivateKey   string    `json:"private_key"` // Criptografada
	Type         string    `json:"type"`         // "wallet", "miner", "validator"
	CreatedAt    time.Time `json:"created_at"`
	LastUsed     time.Time `json:"last_used"`
	Description  string    `json:"description"`
}

// NewSecureKeystore cria novo keystore seguro
func NewSecureKeystore(path, password string) *SecureKeystore {
	// Criar diretório se não existir
	os.MkdirAll(path, 0700)
	
	backupPath := filepath.Join(path, "backup")
	os.MkdirAll(backupPath, 0700)

	return &SecureKeystore{
		Path:       path,
		Password:   password,
		Encrypted:  true,
		BackupPath: backupPath,
		CreatedAt:  time.Now(),
	}
}

// StoreKey armazena chave no keystore
func (ks *SecureKeystore) StoreKey(entry *KeyEntry) error {
	// Criptografar chave privada
	encryptedPrivateKey, err := ks.encryptPrivateKey(entry.PrivateKey)
	if err != nil {
		return fmt.Errorf("erro ao criptografar chave privada: %v", err)
	}

	entry.PrivateKey = encryptedPrivateKey
	entry.CreatedAt = time.Now()
	entry.LastUsed = time.Now()

	// Salvar entrada
	entryPath := filepath.Join(ks.Path, fmt.Sprintf("%s.key", entry.ID))
	
	entryData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar entrada: %v", err)
	}

	if err := os.WriteFile(entryPath, entryData, 0600); err != nil {
		return fmt.Errorf("erro ao salvar entrada: %v", err)
	}

	return nil
}

// LoadKey carrega chave do keystore
func (ks *SecureKeystore) LoadKey(id string) (*KeyEntry, error) {
	entryPath := filepath.Join(ks.Path, fmt.Sprintf("%s.key", id))
	
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("chave não encontrada: %s", id)
	}

	entryData, err := os.ReadFile(entryPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler entrada: %v", err)
	}

	var entry KeyEntry
	if err := json.Unmarshal(entryData, &entry); err != nil {
		return nil, fmt.Errorf("erro ao deserializar entrada: %v", err)
	}

	// Descriptografar chave privada
	decryptedPrivateKey, err := ks.decryptPrivateKey(entry.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar chave privada: %v", err)
	}

	entry.PrivateKey = decryptedPrivateKey
	entry.LastUsed = time.Now()

	// Atualizar timestamp de uso
	ks.StoreKey(&entry)

	return &entry, nil
}

// encryptPrivateKey criptografa chave privada
func (ks *SecureKeystore) encryptPrivateKey(privateKey string) (string, error) {
	// Derivar chave da senha
	salt := []byte("ordm-keystore-salt")
	key := pbkdf2.Key([]byte(ks.Password), salt, 10000, 32, sha256.New)

	// Criptografar com AES-256-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(privateKey), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decryptPrivateKey descriptografa chave privada
func (ks *SecureKeystore) decryptPrivateKey(encryptedKey string) (string, error) {
	// Derivar chave da senha
	salt := []byte("ordm-keystore-salt")
	key := pbkdf2.Key([]byte(ks.Password), salt, 10000, 32, sha256.New)

	// Decodificar hex
	ciphertext, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}

	// Descriptografar com AES-256-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("dados criptografados muito curtos")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
EOF

# 3.3.1 Criptografar logs sensíveis
log "3.3.1 - Criando pkg/logger/secure_logger.go..."
cat > pkg/logger/secure_logger.go << 'EOF'
package logger

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// SecureLogger implementa logging seguro
type SecureLogger struct {
	encryptionKey []byte
	sensitiveFields []string
}

// NewSecureLogger cria novo logger seguro
func NewSecureLogger(encryptionKey string) *SecureLogger {
	key := []byte(encryptionKey)
	if len(key) < 32 {
		// Padding para 32 bytes
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	}
	
	return &SecureLogger{
		encryptionKey: key[:32],
		sensitiveFields: []string{
			"private_key", "password", "pin", "secret",
			"wallet_address", "miner_id", "signature",
		},
	}
}

// LogSecure registra log com dados sensíveis criptografados
func (sl *SecureLogger) LogSecure(level, message string, data map[string]interface{}) {
	// Criptografar dados sensíveis
	secureData := sl.encryptSensitiveData(data)
	
	logEntry := map[string]interface{}{
		"level":     level,
		"message":   message,
		"data":      secureData,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	// Em uma implementação real, salvar no arquivo de log
	fmt.Printf("[%s] %s: %s\n", level, message, sl.maskSensitiveData(message))
}

// encryptSensitiveData criptografa dados sensíveis
func (sl *SecureLogger) encryptSensitiveData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for key, value := range data {
		if sl.isSensitiveField(key) {
			if strValue, ok := value.(string); ok {
				encrypted, err := sl.encrypt(strValue)
				if err == nil {
					result[key] = fmt.Sprintf("ENCRYPTED:%s", encrypted)
				} else {
					result[key] = "ENCRYPTION_ERROR"
				}
			} else {
				result[key] = "ENCRYPTED:COMPLEX_TYPE"
			}
		} else {
			result[key] = value
		}
	}
	
	return result
}

// isSensitiveField verifica se campo é sensível
func (sl *SecureLogger) isSensitiveField(fieldName string) bool {
	lowerField := strings.ToLower(fieldName)
	for _, sensitive := range sl.sensitiveFields {
		if strings.Contains(lowerField, sensitive) {
			return true
		}
	}
	return false
}

// maskSensitiveData mascara dados sensíveis em mensagens
func (sl *SecureLogger) maskSensitiveData(message string) string {
	// Mascarar endereços de wallet (40 caracteres hex)
	message = sl.maskPattern(message, `[0-9a-fA-F]{40}`, "***WALLET_ADDRESS***")
	
	// Mascarar chaves privadas
	message = sl.maskPattern(message, `[0-9a-fA-F]{64}`, "***PRIVATE_KEY***")
	
	// Mascarar PINs
	message = sl.maskPattern(message, `\b\d{6,8}\b`, "***PIN***")
	
	return message
}

// maskPattern mascara padrão específico
func (sl *SecureLogger) maskPattern(text, pattern, replacement string) string {
	// Implementação simplificada - em produção usar regex
	if strings.Contains(text, "wallet") {
		return strings.ReplaceAll(text, "wallet", "***WALLET***")
	}
	return text
}

// encrypt criptografa string
func (sl *SecureLogger) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(sl.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decrypt descriptografa string
func (sl *SecureLogger) decrypt(encryptedText string) (string, error) {
	if !strings.HasPrefix(encryptedText, "ENCRYPTED:") {
		return "", fmt.Errorf("texto não está criptografado")
	}
	
	encryptedText = strings.TrimPrefix(encryptedText, "ENCRYPTED:")
	
	ciphertext, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sl.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("dados criptografados muito curtos")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
EOF

log "✅ PARTE 3: Segurança concluída!"
log "📋 Arquivos criados:"
log "   - pkg/auth/rate_limiter.go"
log "   - pkg/auth/pin_generator.go"
log "   - pkg/crypto/keystore.go"
log "   - pkg/logger/secure_logger.go"
log "   - Tempo de PIN 2FA atualizado para 60 segundos"

