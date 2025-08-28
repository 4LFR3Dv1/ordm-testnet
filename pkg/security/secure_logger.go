package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SecureLogger implementa logging seguro com criptografia
type SecureLogger struct {
	EncryptSensitive bool
	MaskAddresses    bool
	LogLevel         string
	LogPath          string
	EncryptionKey    []byte
	mu               chan struct{} // Semáforo para concorrência
}

// LogEntry representa uma entrada de log
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Encrypted bool                   `json:"encrypted"`
	IP        string                 `json:"ip,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Action    string                 `json:"action,omitempty"`
}

// NewSecureLogger cria um novo logger seguro
func NewSecureLogger(logPath string, encryptSensitive bool) (*SecureLogger, error) {
	// Criar diretório de logs se não existir
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de logs: %v", err)
	}

	// Gerar chave de criptografia
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave de criptografia: %v", err)
	}

	return &SecureLogger{
		EncryptSensitive: encryptSensitive,
		MaskAddresses:    true,
		LogLevel:         "INFO",
		LogPath:          logPath,
		EncryptionKey:    encryptionKey,
		mu:               make(chan struct{}, 1), // Semáforo com buffer 1
	}, nil
}

// LogSensitive registra log com dados sensíveis criptografados
func (sl *SecureLogger) LogSensitive(level, message string, data map[string]interface{}) {
	// Adquirir semáforo
	sl.mu <- struct{}{}
	defer func() { <-sl.mu }()

	// Sanitizar dados sensíveis
	sanitizedData := sl.sanitizeData(data)

	// Criar entrada de log
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Data:      sanitizedData,
		Encrypted: sl.EncryptSensitive,
	}

	// Criptografar se necessário
	if sl.EncryptSensitive {
		sl.encryptLogEntry(&entry)
	}

	// Salvar log
	sl.saveLogEntry(entry)
}

// LogSecurityEvent registra eventos de segurança
func (sl *SecureLogger) LogSecurityEvent(eventType, ip, userID, action string, details map[string]interface{}) {
	// Mascarar IP se configurado
	if sl.MaskAddresses {
		ip = sl.maskIP(ip)
	}

	// Adicionar metadados de segurança
	securityData := map[string]interface{}{
		"event_type": eventType,
		"ip":         ip,
		"user_id":    userID,
		"action":     action,
		"details":    details,
	}

	sl.LogSensitive("SECURITY", fmt.Sprintf("Security event: %s", eventType), securityData)
}

// LogAuthentication registra tentativas de autenticação
func (sl *SecureLogger) LogAuthentication(username, ip string, success bool, details map[string]interface{}) {
	// Mascarar dados sensíveis
	maskedUsername := sl.maskUsername(username)
	maskedIP := sl.maskIP(ip)

	authData := map[string]interface{}{
		"username": maskedUsername,
		"ip":       maskedIP,
		"success":  success,
		"details":  details,
	}

	level := "INFO"
	if !success {
		level = "WARNING"
	}

	sl.LogSensitive(level, fmt.Sprintf("Authentication attempt: %s", success), authData)
}

// LogTransaction registra transações
func (sl *SecureLogger) LogTransaction(from, to, amount, txHash string, success bool) {
	// Mascarar endereços
	maskedFrom := sl.maskAddress(from)
	maskedTo := sl.maskAddress(to)

	txData := map[string]interface{}{
		"from":    maskedFrom,
		"to":      maskedTo,
		"amount":  amount,
		"tx_hash": txHash,
		"success": success,
	}

	level := "INFO"
	if !success {
		level = "ERROR"
	}

	sl.LogSensitive(level, fmt.Sprintf("Transaction: %s", success), txData)
}

// sanitizeData remove ou mascarar dados sensíveis
func (sl *SecureLogger) sanitizeData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[key] = sl.sanitizeString(key, v)
		case map[string]interface{}:
			sanitized[key] = sl.sanitizeData(v)
		default:
			sanitized[key] = value
		}
	}

	return sanitized
}

// sanitizeString sanitiza string baseado no tipo
func (sl *SecureLogger) sanitizeString(key, value string) string {
	keyLower := strings.ToLower(key)

	// Mascarar dados sensíveis
	if strings.Contains(keyLower, "password") || strings.Contains(keyLower, "secret") || strings.Contains(keyLower, "key") {
		return "***MASKED***"
	}

	if strings.Contains(keyLower, "ip") || strings.Contains(keyLower, "address") {
		return sl.maskIP(value)
	}

	if strings.Contains(keyLower, "username") || strings.Contains(keyLower, "user") {
		return sl.maskUsername(value)
	}

	if strings.Contains(keyLower, "email") {
		return sl.maskEmail(value)
	}

	return value
}

// maskIP mascara endereço IP
func (sl *SecureLogger) maskIP(ip string) string {
	if ip == "" {
		return "***.***.*.*"
	}

	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.*.*", parts[0], parts[1])
	}

	return "***.***.*.*"
}

// maskUsername mascara nome de usuário
func (sl *SecureLogger) maskUsername(username string) string {
	if username == "" {
		return "***"
	}

	if len(username) <= 2 {
		return "***"
	}

	return username[:2] + "***"
}

// maskEmail mascara email
func (sl *SecureLogger) maskEmail(email string) string {
	if email == "" {
		return "***@***.***"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***.***"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		username = "***"
	} else {
		username = username[:2] + "***"
	}

	domainParts := strings.Split(domain, ".")
	if len(domainParts) >= 2 {
		domain = domainParts[0][:1] + "***." + domainParts[len(domainParts)-1]
	} else {
		domain = "***"
	}

	return username + "@" + domain
}

// maskAddress mascara endereço de wallet
func (sl *SecureLogger) maskAddress(address string) string {
	if address == "" {
		return "***"
	}

	if len(address) <= 8 {
		return "***"
	}

	return address[:4] + "***" + address[len(address)-4:]
}

// encryptLogEntry criptografa entrada de log
func (sl *SecureLogger) encryptLogEntry(entry *LogEntry) {
	// Serializar entrada
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Erro ao serializar entrada de log: %v", err)
		return
	}

	// Criptografar dados
	encrypted, err := sl.encryptData(entryJSON)
	if err != nil {
		log.Printf("Erro ao criptografar entrada de log: %v", err)
		return
	}

	// Substituir dados originais
	entry.Message = base64.StdEncoding.EncodeToString(encrypted)
	entry.Data = nil
}

// encryptData criptografa dados usando AES-256-GCM
func (sl *SecureLogger) encryptData(data []byte) ([]byte, error) {
	// Criar cipher AES
	block, err := aes.NewCipher(sl.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %v", err)
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	// Gerar nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("erro ao gerar nonce: %v", err)
	}

	// Criptografar
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// saveLogEntry salva entrada de log no arquivo
func (sl *SecureLogger) saveLogEntry(entry LogEntry) {
	// Serializar entrada
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Erro ao serializar entrada de log: %v", err)
		return
	}

	// Abrir arquivo de log
	file, err := os.OpenFile(sl.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Erro ao abrir arquivo de log: %v", err)
		return
	}
	defer file.Close()

	// Escrever entrada
	entryLine := string(entryJSON) + "\n"
	if _, err := file.WriteString(entryLine); err != nil {
		log.Printf("Erro ao escrever no arquivo de log: %v", err)
	}
}

// SetLogLevel define nível de log
func (sl *SecureLogger) SetLogLevel(level string) {
	validLevels := map[string]bool{
		"DEBUG":    true,
		"INFO":     true,
		"WARNING":  true,
		"ERROR":    true,
		"SECURITY": true,
	}

	if validLevels[strings.ToUpper(level)] {
		sl.LogLevel = strings.ToUpper(level)
	}
}

// RotateLogs rotaciona logs antigos
func (sl *SecureLogger) RotateLogs(maxSize int64, maxAge time.Duration) error {
	// Verificar tamanho do arquivo
	info, err := os.Stat(sl.LogPath)
	if err != nil {
		return fmt.Errorf("erro ao verificar arquivo de log: %v", err)
	}

	// Rotacionar se necessário
	if info.Size() > maxSize || time.Since(info.ModTime()) > maxAge {
		backupPath := sl.LogPath + "." + time.Now().Format("2006-01-02-15-04-05")
		if err := os.Rename(sl.LogPath, backupPath); err != nil {
			return fmt.Errorf("erro ao rotacionar log: %v", err)
		}
	}

	return nil
}
