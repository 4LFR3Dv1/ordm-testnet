#!/bin/bash

# 🔐 Script: Parte 1.2 - Criptografia de Dados
# Descrição: Implementa criptografia de dados sensíveis, hash seguro e PIN 2FA forte

set -e

echo "🔐 [$(date)] Iniciando Parte 1.2: Criptografia de Dados"
echo "======================================================="

# 1.2.1 - Criptografar dados sensíveis
echo "🔐 1.2.1 - Implementando criptografia de dados sensíveis..."

mkdir -p pkg/crypto
cat > pkg/crypto/wallet_encryption.go << 'EOF'
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type WalletEncryption struct {
	key []byte
}

func NewWalletEncryption(password string, salt []byte) *WalletEncryption {
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return &WalletEncryption{key: key}
}

func (we *WalletEncryption) EncryptWalletData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(we.key)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("erro ao gerar nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (we *WalletEncryption) DecryptWalletData(encryptedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(we.key)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("dados criptografados muito curtos")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar: %v", err)
	}

	return plaintext, nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("erro ao gerar salt: %v", err)
	}
	return salt, nil
}

func HashPassword(password string, salt []byte) string {
	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash)
}

func VerifyPassword(password, hash string, salt []byte) bool {
	computedHash := HashPassword(password, salt)
	return computedHash == hash
}
EOF

# 1.2.2 - Hash seguro de senhas
echo "🔒 1.2.2 - Implementando hash seguro de senhas..."

cat > pkg/auth/password.go << 'EOF'
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
EOF

# 1.2.3 - PIN 2FA forte
echo "🔐 1.2.3 - Implementando PIN 2FA forte..."

cat > pkg/auth/pin_generator.go << 'EOF'
package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type PINGenerator struct {
	length     int
	useLetters bool
	useNumbers bool
	useSymbols bool
}

func NewPINGenerator() *PINGenerator {
	return &PINGenerator{
		length:     8, // Aumentado para 8 dígitos
		useNumbers: true,
	}
}

func (pg *PINGenerator) GeneratePIN() (string, error) {
	const numbers = "0123456789"

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

func (pg *PINGenerator) ValidatePIN(pin string) error {
	if len(pin) != pg.length {
		return fmt.Errorf("PIN deve ter %d dígitos", pg.length)
	}

	for _, char := range pin {
		if char < '0' || char > '9' {
			return fmt.Errorf("PIN deve conter apenas números")
		}
	}

	if pg.isSequential(pin) {
		return fmt.Errorf("PIN não pode ser sequencial")
	}

	if pg.isRepetitive(pin) {
		return fmt.Errorf("PIN não pode ser repetitivo")
	}

	return nil
}

func (pg *PINGenerator) isSequential(pin string) bool {
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

func (pg *PINGenerator) isRepetitive(pin string) bool {
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

echo "✅ [$(date)] Parte 1.2: Criptografia de Dados concluída!"
echo "📋 Implementações:"
echo "  ✅ 1.2.1 - Criptografia de dados sensíveis"
echo "  ✅ 1.2.2 - Hash seguro de senhas"
echo "  ✅ 1.2.3 - PIN 2FA forte (8 dígitos)"
echo ""
echo "🚀 Próximo: Execute 'part1c_attack_protection.sh'"

