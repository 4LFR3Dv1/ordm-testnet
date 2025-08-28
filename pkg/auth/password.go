package auth

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordManager struct{}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{}
}

func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar hash: %v", err)
	}
	return string(hash), nil
}

func (pm *PasswordManager) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (pm *PasswordManager) GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	
	for i := range password {
		randomByte := make([]byte, 1)
		if _, err := rand.Read(randomByte); err != nil {
			return "", fmt.Errorf("erro ao gerar senha: %v", err)
		}
		password[i] = charset[randomByte[0]%byte(len(charset))]
	}
	
	return string(password), nil
}

func (pm *PasswordManager) ValidatePasswordStrength(password string) (bool, []string) {
	var errors []string
	
	if len(password) < 8 {
		errors = append(errors, "Senha deve ter pelo menos 8 caracteres")
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		errors = append(errors, "Senha deve conter pelo menos uma letra maiúscula")
	}
	if !hasLower {
		errors = append(errors, "Senha deve conter pelo menos uma letra minúscula")
	}
	if !hasDigit {
		errors = append(errors, "Senha deve conter pelo menos um número")
	}
	if !hasSpecial {
		errors = append(errors, "Senha deve conter pelo menos um caractere especial")
	}
	
	return len(errors) == 0, errors
}
