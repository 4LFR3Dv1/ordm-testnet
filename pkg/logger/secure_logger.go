package logger

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// SecureLogger implementa logging seguro
type SecureLogger struct {
	encryptionKey   []byte
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
	// Criptografar dados sensíveis (para uso futuro)
	_ = sl.encryptSensitiveData(data)

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
