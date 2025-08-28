package security

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AuditLogger implementa sistema de logs de auditoria
type AuditLogger struct {
	LogPath       string
	MaxFileSize   int64
	MaxFileAge    time.Duration
	EncryptLogs   bool
	EncryptionKey []byte
	mu            sync.RWMutex
	file          *os.File
	encoder       *json.Encoder
}

// AuditEvent representa um evento de auditoria
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"` // success, failure, error
	Details   map[string]interface{} `json:"details,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Severity  string                 `json:"severity"` // low, medium, high, critical
	Hash      string                 `json:"hash"`
}

// AuditConfig configuração do audit logger
type AuditConfig struct {
	LogPath      string        `json:"log_path"`
	MaxFileSize  int64         `json:"max_file_size"`
	MaxFileAge   time.Duration `json:"max_file_age"`
	EncryptLogs  bool          `json:"encrypt_logs"`
	RotationSize int64         `json:"rotation_size"`
}

// NewAuditLogger cria novo audit logger
func NewAuditLogger(config *AuditConfig) (*AuditLogger, error) {
	if config == nil {
		config = &AuditConfig{
			LogPath:      "logs/audit/audit.log",
			MaxFileSize:  100 * 1024 * 1024,   // 100MB
			MaxFileAge:   30 * 24 * time.Hour, // 30 dias
			EncryptLogs:  true,
			RotationSize: 50 * 1024 * 1024, // 50MB
		}
	}

	// Criar diretório se não existir
	if err := os.MkdirAll(filepath.Dir(config.LogPath), 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de audit: %v", err)
	}

	// Gerar chave de criptografia se necessário
	var encryptionKey []byte
	if config.EncryptLogs {
		encryptionKey = make([]byte, 32)
		// Em produção, carregar de variável de ambiente
		// encryptionKey = []byte(os.Getenv("AUDIT_ENCRYPTION_KEY"))
	}

	audit := &AuditLogger{
		LogPath:       config.LogPath,
		MaxFileSize:   config.MaxFileSize,
		MaxFileAge:    config.MaxFileAge,
		EncryptLogs:   config.EncryptLogs,
		EncryptionKey: encryptionKey,
	}

	// Abrir arquivo de log
	if err := audit.openLogFile(); err != nil {
		return nil, err
	}

	return audit, nil
}

// LogAction registra ação de auditoria
func (al *AuditLogger) LogAction(eventType, userID, ip, userAgent, action, resource, result string, details map[string]interface{}) error {
	// Gerar ID único do evento
	eventID := generateAuditEventID()

	// Criar evento de auditoria
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventID:   eventID,
		EventType: eventType,
		UserID:    maskUserID(userID),
		IP:        maskIP(ip),
		UserAgent: maskUserAgent(userAgent),
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   sanitizeDetails(details),
		Severity:  determineSeverity(eventType, result),
	}

	// Gerar hash do evento
	event.Hash = generateEventHash(event)

	// Registrar evento
	return al.writeEvent(event)
}

// LogSecurityEvent registra evento de segurança
func (al *AuditLogger) LogSecurityEvent(eventType, userID, ip, userAgent, action string, details map[string]interface{}) error {
	return al.LogAction("security", userID, ip, userAgent, action, "security", "success", details)
}

// LogAuthentication registra tentativa de autenticação
func (al *AuditLogger) LogAuthentication(userID, ip, userAgent string, success bool, details map[string]interface{}) error {
	result := "success"
	if !success {
		result = "failure"
	}
	return al.LogAction("authentication", userID, ip, userAgent, "login", "auth", result, details)
}

// LogTransaction registra transação
func (al *AuditLogger) LogTransaction(userID, ip, userAgent, action, resource string, success bool, details map[string]interface{}) error {
	result := "success"
	if !success {
		result = "failure"
	}
	return al.LogAction("transaction", userID, ip, userAgent, action, resource, result, details)
}

// LogAdminAction registra ação administrativa
func (al *AuditLogger) LogAdminAction(userID, ip, userAgent, action, resource string, details map[string]interface{}) error {
	return al.LogAction("admin", userID, ip, userAgent, action, resource, "success", details)
}

// writeEvent escreve evento no arquivo de log
func (al *AuditLogger) writeEvent(event *AuditEvent) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Verificar se precisa rotacionar arquivo
	if err := al.checkRotation(); err != nil {
		return fmt.Errorf("erro na rotação: %v", err)
	}

	// Serializar evento
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %v", err)
	}

	// Criptografar se necessário
	if al.EncryptLogs {
		eventJSON, err = al.encryptData(eventJSON)
		if err != nil {
			return fmt.Errorf("erro ao criptografar evento: %v", err)
		}
	}

	// Escrever no arquivo
	eventLine := string(eventJSON) + "\n"
	if _, err := al.file.WriteString(eventLine); err != nil {
		return fmt.Errorf("erro ao escrever evento: %v", err)
	}

	// Forçar flush para disco
	if err := al.file.Sync(); err != nil {
		return fmt.Errorf("erro ao sincronizar arquivo: %v", err)
	}

	return nil
}

// openLogFile abre arquivo de log
func (al *AuditLogger) openLogFile() error {
	file, err := os.OpenFile(al.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de log: %v", err)
	}

	al.file = file
	al.encoder = json.NewEncoder(file)
	return nil
}

// checkRotation verifica se precisa rotacionar arquivo
func (al *AuditLogger) checkRotation() error {
	info, err := al.file.Stat()
	if err != nil {
		return err
	}

	// Verificar tamanho do arquivo
	if info.Size() > al.MaxFileSize {
		return al.rotateFile()
	}

	// Verificar idade do arquivo
	if time.Since(info.ModTime()) > al.MaxFileAge {
		return al.rotateFile()
	}

	return nil
}

// rotateFile rotaciona arquivo de log
func (al *AuditLogger) rotateFile() error {
	// Fechar arquivo atual
	if err := al.file.Close(); err != nil {
		return fmt.Errorf("erro ao fechar arquivo: %v", err)
	}

	// Renomear arquivo atual
	backupPath := al.LogPath + "." + time.Now().Format("2006-01-02-15-04-05")
	if err := os.Rename(al.LogPath, backupPath); err != nil {
		return fmt.Errorf("erro ao renomear arquivo: %v", err)
	}

	// Abrir novo arquivo
	return al.openLogFile()
}

// encryptData criptografa dados
func (al *AuditLogger) encryptData(data []byte) ([]byte, error) {
	// Implementação simplificada - em produção usar AES-256-GCM
	// Por enquanto, apenas codificar em base64
	return []byte(base64.StdEncoding.EncodeToString(data)), nil
}

// generateAuditEventID gera ID único para evento de auditoria
func generateAuditEventID() string {
	timestamp := time.Now().UnixNano()
	random := make([]byte, 8)
	// Em produção, usar crypto/rand
	// rand.Read(random)

	return fmt.Sprintf("%d-%x", timestamp, random)
}

// generateEventHash gera hash do evento
func generateEventHash(event *AuditEvent) string {
	// Criar string para hash (sem o próprio hash)
	eventData := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s",
		event.Timestamp.Format(time.RFC3339),
		event.EventID,
		event.EventType,
		event.UserID,
		event.Action,
		event.Resource,
		event.Result,
		event.Severity,
	)

	hash := sha256.Sum256([]byte(eventData))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// maskUserID mascara ID do usuário
func maskUserID(userID string) string {
	if userID == "" {
		return "anonymous"
	}
	if len(userID) <= 2 {
		return "***"
	}
	return userID[:2] + "***"
}

// maskIP mascara endereço IP
func maskIP(ip string) string {
	if ip == "" {
		return "***.***.*.*"
	}
	parts := []string{}
	for _, part := range strings.Split(ip, ".") {
		if len(parts) < 2 {
			parts = append(parts, part)
		} else {
			parts = append(parts, "*")
		}
	}
	return strings.Join(parts, ".")
}

// maskUserAgent mascara User-Agent
func maskUserAgent(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}
	if len(userAgent) <= 10 {
		return userAgent
	}
	return userAgent[:10] + "***"
}

// sanitizeDetails sanitiza detalhes do evento
func sanitizeDetails(details map[string]interface{}) map[string]interface{} {
	if details == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	for key, value := range details {
		switch v := value.(type) {
		case string:
			// Mascarar dados sensíveis
			if strings.Contains(strings.ToLower(key), "password") ||
				strings.Contains(strings.ToLower(key), "secret") ||
				strings.Contains(strings.ToLower(key), "token") {
				sanitized[key] = "***MASKED***"
			} else {
				sanitized[key] = v
			}
		default:
			sanitized[key] = v
		}
	}

	return sanitized
}

// determineSeverity determina severidade do evento
func determineSeverity(eventType, result string) string {
	switch eventType {
	case "security":
		if result == "failure" {
			return "high"
		}
		return "medium"
	case "authentication":
		if result == "failure" {
			return "medium"
		}
		return "low"
	case "admin":
		return "high"
	case "transaction":
		if result == "failure" {
			return "medium"
		}
		return "low"
	default:
		return "low"
	}
}

// GetAuditStats retorna estatísticas de auditoria
func (al *AuditLogger) GetAuditStats() map[string]interface{} {
	al.mu.RLock()
	defer al.mu.RUnlock()

	info, err := al.file.Stat()
	if err != nil {
		return map[string]interface{}{
			"error": "erro ao obter estatísticas",
		}
	}

	return map[string]interface{}{
		"log_file_size": info.Size(),
		"log_file_age":  time.Since(info.ModTime()).String(),
		"encrypted":     al.EncryptLogs,
		"max_file_size": al.MaxFileSize,
		"max_file_age":  al.MaxFileAge.String(),
	}
}

// Close fecha o audit logger
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.file != nil {
		return al.file.Close()
	}
	return nil
}
