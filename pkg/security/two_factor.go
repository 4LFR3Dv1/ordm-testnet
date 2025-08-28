package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// TwoFactorAuth implementa autenticação de dois fatores com TOTP
type TwoFactorAuth struct {
	SecretKey   string
	Algorithm   string // TOTP
	Digits      int    // 6 dígitos
	Period      int    // 30 segundos
	Window      int    // Janela de tolerância
	BackupCodes []string
	LastUsed    time.Time
	Attempts    int
	LockedUntil time.Time
}

// TwoFactorConfig configuração do 2FA
type TwoFactorConfig struct {
	Digits      int           `json:"digits"`       // Número de dígitos (6 ou 8)
	Period      int           `json:"period"`       // Período em segundos (30)
	Window      int           `json:"window"`       // Janela de tolerância
	MaxAttempts int           `json:"max_attempts"` // Máximo de tentativas
	LockoutTime time.Duration `json:"lockout_time"` // Tempo de bloqueio
	BackupCodes int           `json:"backup_codes"` // Número de códigos de backup
}

// NewTwoFactorAuth cria nova instância de 2FA
func NewTwoFactorAuth(config *TwoFactorConfig) *TwoFactorAuth {
	if config == nil {
		config = &TwoFactorConfig{
			Digits:      6,
			Period:      30,
			Window:      1,
			MaxAttempts: 5,
			LockoutTime: 15 * time.Minute,
			BackupCodes: 10,
		}
	}

	// Gerar chave secreta
	secretKey := generateSecretKey()

	// Gerar códigos de backup
	backupCodes := generateBackupCodes(config.BackupCodes)

	return &TwoFactorAuth{
		SecretKey:   secretKey,
		Algorithm:   "TOTP",
		Digits:      config.Digits,
		Period:      config.Period,
		Window:      config.Window,
		BackupCodes: backupCodes,
		Attempts:    0,
	}
}

// GenerateCode gera código TOTP atual
func (tfa *TwoFactorAuth) GenerateCode() string {
	return tfa.generateTOTP(time.Now().Unix())
}

// ValidateCode valida código TOTP
func (tfa *TwoFactorAuth) ValidateCode(code string) (bool, error) {
	// Verificar se está bloqueado
	if time.Now().Before(tfa.LockedUntil) {
		return false, fmt.Errorf("2FA bloqueado até %v", tfa.LockedUntil)
	}

	// Verificar se é código de backup
	if tfa.validateBackupCode(code) {
		tfa.resetAttempts()
		return true, nil
	}

	// Validar código TOTP
	now := time.Now().Unix()
	valid := false

	// Verificar janela de tolerância
	for i := -tfa.Window; i <= tfa.Window; i++ {
		expectedCode := tfa.generateTOTP(now + int64(i*tfa.Period))
		if code == expectedCode {
			valid = true
			break
		}
	}

	if valid {
		tfa.resetAttempts()
		tfa.LastUsed = time.Now()
		return true, nil
	}

	// Incrementar tentativas
	tfa.Attempts++
	if tfa.Attempts >= 5 {
		tfa.LockedUntil = time.Now().Add(15 * time.Minute)
		return false, fmt.Errorf("muitas tentativas inválidas, 2FA bloqueado por 15 minutos")
	}

	return false, fmt.Errorf("código inválido, tentativas restantes: %d", 5-tfa.Attempts)
}

// generateTOTP gera código TOTP para timestamp específico
func (tfa *TwoFactorAuth) generateTOTP(timestamp int64) string {
	// Calcular contador
	counter := timestamp / int64(tfa.Period)

	// Converter para bytes
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(counter))

	// Decodificar chave secreta
	secret, err := base32.StdEncoding.DecodeString(strings.ToUpper(tfa.SecretKey))
	if err != nil {
		return ""
	}

	// Calcular HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Gerar código usando algoritmo TOTP
	offset := hash[len(hash)-1] & 0xf
	code := ((int(hash[offset]) & 0x7f) << 24) |
		((int(hash[offset+1]) & 0xff) << 16) |
		((int(hash[offset+2]) & 0xff) << 8) |
		(int(hash[offset+3]) & 0xff)

	// Aplicar módulo para obter dígitos corretos
	modulo := int(math.Pow10(tfa.Digits))
	code = code % modulo

	// Formatar com zeros à esquerda
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", tfa.Digits), code)
}

// generateSecretKey gera chave secreta para TOTP
func generateSecretKey() string {
	// Gerar 20 bytes aleatórios
	secret := make([]byte, 20)
	rand.Read(secret)

	// Codificar em base32
	return base32.StdEncoding.EncodeToString(secret)
}

// generateBackupCodes gera códigos de backup
func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Gerar 8 dígitos aleatórios
		code := make([]byte, 4)
		rand.Read(code)

		// Converter para número de 8 dígitos
		num := binary.BigEndian.Uint32(code) % 100000000
		codes[i] = fmt.Sprintf("%08d", num)
	}
	return codes
}

// validateBackupCode valida código de backup
func (tfa *TwoFactorAuth) validateBackupCode(code string) bool {
	for i, backupCode := range tfa.BackupCodes {
		if code == backupCode {
			// Remover código usado
			tfa.BackupCodes = append(tfa.BackupCodes[:i], tfa.BackupCodes[i+1:]...)
			return true
		}
	}
	return false
}

// resetAttempts reseta contador de tentativas
func (tfa *TwoFactorAuth) resetAttempts() {
	tfa.Attempts = 0
	tfa.LockedUntil = time.Time{}
}

// GetQRCodeURL gera URL para QR Code
func (tfa *TwoFactorAuth) GetQRCodeURL(issuer, account string) string {
	// Formatar chave para URL
	secret := strings.ReplaceAll(tfa.SecretKey, "=", "")

	// Criar URL otpauth
	url := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		issuer, account, secret, issuer, tfa.Digits, tfa.Period)

	return url
}

// GetBackupCodes retorna códigos de backup
func (tfa *TwoFactorAuth) GetBackupCodes() []string {
	return tfa.BackupCodes
}

// RegenerateBackupCodes regenera códigos de backup
func (tfa *TwoFactorAuth) RegenerateBackupCodes() {
	tfa.BackupCodes = generateBackupCodes(10)
}

// IsLocked verifica se 2FA está bloqueado
func (tfa *TwoFactorAuth) IsLocked() bool {
	return time.Now().Before(tfa.LockedUntil)
}

// GetRemainingAttempts retorna tentativas restantes
func (tfa *TwoFactorAuth) GetRemainingAttempts() int {
	return 5 - tfa.Attempts
}

// GetLockoutTime retorna tempo restante de bloqueio
func (tfa *TwoFactorAuth) GetLockoutTime() time.Duration {
	if tfa.IsLocked() {
		return tfa.LockedUntil.Sub(time.Now())
	}
	return 0
}
