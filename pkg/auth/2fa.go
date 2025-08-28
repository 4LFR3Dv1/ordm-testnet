package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// TwoFactorAuth implementa autenticação 2FA
type TwoFactorAuth struct {
	CurrentPIN      string    `json:"current_pin"`
	GeneratedAt     time.Time `json:"generated_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Attempts        int       `json:"attempts"`
	MaxAttempts     int       `json:"max_attempts"`
	LockedUntil     time.Time `json:"locked_until"`
	SessionToken    string    `json:"session_token"`
	IsAuthenticated bool      `json:"is_authenticated"`
}

// NewTwoFactorAuth cria nova instância de 2FA
func NewTwoFactorAuth() *TwoFactorAuth {
	return &TwoFactorAuth{
		MaxAttempts: 3,
		Attempts:    0,
	}
}

// GeneratePIN gera novo PIN 2FA
func (tfa *TwoFactorAuth) GeneratePIN() string {
	// Gerar 8 dígitos aleatórios (melhorado)
	bytes := make([]byte, 4)
	rand.Read(bytes)

	// Converter para número de 8 dígitos
	num := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	pin := fmt.Sprintf("%08d", num%100000000)

	tfa.CurrentPIN = pin
	tfa.GeneratedAt = time.Now()
	tfa.ExpiresAt = time.Now().Add(60 * time.Second) // 60 segundos (corrigido)
	tfa.Attempts = 0
	tfa.IsAuthenticated = false

	// Gerar token de sessão
	tfa.SessionToken = hex.EncodeToString(bytes)

	return pin
}

// ValidatePIN valida PIN inserido
func (tfa *TwoFactorAuth) ValidatePIN(inputPIN string) (bool, string) {
	// Verificar se está bloqueado
	if !tfa.LockedUntil.IsZero() && time.Now().Before(tfa.LockedUntil) {
		remaining := tfa.LockedUntil.Sub(time.Now())
		return false, fmt.Sprintf("Conta bloqueada por mais %v", remaining.Round(time.Minute))
	}

	// Verificar se PIN expirou
	if time.Now().After(tfa.ExpiresAt) {
		return false, "PIN expirado. Gere um novo PIN."
	}

	// Verificar se excedeu tentativas
	if tfa.Attempts >= tfa.MaxAttempts {
		tfa.LockedUntil = time.Now().Add(15 * time.Minute) // Bloquear por 15 minutos
		return false, "Muitas tentativas. Conta bloqueada por 15 minutos."
	}

	// Validar PIN
	if inputPIN == tfa.CurrentPIN {
		tfa.IsAuthenticated = true
		tfa.Attempts = 0
		return true, "Autenticação bem-sucedida!"
	}

	tfa.Attempts++
	remainingAttempts := tfa.MaxAttempts - tfa.Attempts

	if remainingAttempts == 0 {
		tfa.LockedUntil = time.Now().Add(15 * time.Minute)
		return false, "PIN incorreto. Conta bloqueada por 15 minutos."
	}

	return false, fmt.Sprintf("PIN incorreto. %d tentativas restantes.", remainingAttempts)
}

// IsAuthenticated verifica se está autenticado
func (tfa *TwoFactorAuth) IsUserAuthenticated() bool {
	return tfa.IsAuthenticated
}

// Logout faz logout do usuário
func (tfa *TwoFactorAuth) Logout() {
	tfa.IsAuthenticated = false
	tfa.SessionToken = ""
}

// GetStatus retorna status da autenticação
func (tfa *TwoFactorAuth) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"is_authenticated": tfa.IsAuthenticated,
		"attempts":         tfa.Attempts,
		"max_attempts":     tfa.MaxAttempts,
		"generated_at":     tfa.GeneratedAt,
		"expires_at":       tfa.ExpiresAt,
	}

	if !tfa.LockedUntil.IsZero() {
		status["locked_until"] = tfa.LockedUntil
		status["is_locked"] = time.Now().Before(tfa.LockedUntil)
	}

	return status
}

// GetPINInfo retorna informações do PIN (sem mostrar o PIN)
func (tfa *TwoFactorAuth) GetPINInfo() map[string]interface{} {
	timeLeft := tfa.ExpiresAt.Sub(time.Now())

	return map[string]interface{}{
		"pin_length":    len(tfa.CurrentPIN),
		"generated_at":  tfa.GeneratedAt.Format("15:04:05"),
		"expires_at":    tfa.ExpiresAt.Format("15:04:05"),
		"time_left":     timeLeft.Round(time.Second).String(),
		"is_expired":    time.Now().After(tfa.ExpiresAt),
		"attempts_left": tfa.MaxAttempts - tfa.Attempts,
		"is_locked":     !tfa.LockedUntil.IsZero() && time.Now().Before(tfa.LockedUntil),
	}
}

// ResetPIN reseta PIN e tentativas
func (tfa *TwoFactorAuth) ResetPIN() {
	tfa.CurrentPIN = ""
	tfa.Attempts = 0
	tfa.LockedUntil = time.Time{}
	tfa.IsAuthenticated = false
}
