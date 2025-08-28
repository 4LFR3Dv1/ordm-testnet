package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// EncryptWithAES256 criptografa dados usando AES-256-GCM
func EncryptWithAES256(data []byte, password string) ([]byte, error) {
	// Derivar chave da senha
	key := deriveKeyFromPassword(password)

	// Criar cipher AES-256
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %v", err)
	}

	// Criar GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	// Gerar nonce único
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("erro ao gerar nonce: %v", err)
	}

	// Criptografar dados
	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

// DecryptWithAES256 descriptografa dados usando AES-256-GCM
func DecryptWithAES256(encrypted []byte, password string) ([]byte, error) {
	// Derivar chave da senha
	key := deriveKeyFromPassword(password)

	// Criar cipher AES-256
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %v", err)
	}

	// Criar GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	// Extrair nonce
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("dados criptografados muito curtos")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// Descriptografar dados
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar: %v", err)
	}

	return plaintext, nil
}

// DeriveKeyWithPBKDF2 deriva chave usando PBKDF2 (implementação simplificada)
func DeriveKeyWithPBKDF2(password string, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		// Gerar salt se não fornecido
		salt = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("erro ao gerar salt: %v", err)
		}
	}

	// Implementação simplificada de PBKDF2
	// Em produção, usar golang.org/x/crypto/pbkdf2
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(salt)

	// Iterações para fortalecer
	for i := 0; i < 10000; i++ {
		hash.Write(hash.Sum(nil))
	}

	return hash.Sum(nil), nil
}

// deriveKeyFromPassword deriva chave de 32 bytes da senha
func deriveKeyFromPassword(password string) []byte {
	hash := sha256.New()
	hash.Write([]byte(password))
	key := hash.Sum(nil)

	// Garantir que a chave tem 32 bytes (256 bits)
	if len(key) < 32 {
		// Estender chave se necessário
		extended := make([]byte, 32)
		copy(extended, key)
		for i := len(key); i < 32; i++ {
			extended[i] = byte(i % 256)
		}
		return extended
	}

	return key[:32]
}

// EncryptString criptografa string e retorna base64
func EncryptString(text string, password string) (string, error) {
	encrypted, err := EncryptWithAES256([]byte(text), password)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString descriptografa string de base64
func DecryptString(encryptedText string, password string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar base64: %v", err)
	}

	decrypted, err := DecryptWithAES256(encrypted, password)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// GenerateRandomKey gera chave aleatória de 32 bytes
func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave aleatória: %v", err)
	}
	return key, nil
}

// GenerateRandomSalt gera salt aleatório
func GenerateRandomSalt(size int) ([]byte, error) {
	if size <= 0 {
		size = 32 // Tamanho padrão
	}

	salt := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("erro ao gerar salt: %v", err)
	}
	return salt, nil
}

// HashPassword cria hash seguro de senha
func HashPassword(password string) (string, error) {
	salt, err := GenerateRandomSalt(16)
	if err != nil {
		return "", err
	}

	// Derivar hash da senha
	key, err := DeriveKeyWithPBKDF2(password, salt)
	if err != nil {
		return "", err
	}

	// Combinar salt + hash
	combined := append(salt, key...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword verifica senha contra hash
func VerifyPassword(password, hash string) (bool, error) {
	// Decodificar hash
	combined, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("erro ao decodificar hash: %v", err)
	}

	if len(combined) < 16 {
		return false, fmt.Errorf("hash inválido")
	}

	// Extrair salt e hash original
	salt := combined[:16]
	originalHash := combined[16:]

	// Derivar hash da senha fornecida
	derivedKey, err := DeriveKeyWithPBKDF2(password, salt)
	if err != nil {
		return false, err
	}

	// Comparar hashes
	return string(derivedKey) == string(originalHash), nil
}

// EncryptFile criptografa arquivo
func EncryptFile(inputPath, outputPath, password string) error {
	// Ler arquivo de entrada
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	// Criptografar dados
	encrypted, err := EncryptWithAES256(data, password)
	if err != nil {
		return fmt.Errorf("erro ao criptografar: %v", err)
	}

	// Salvar arquivo criptografado
	if err := os.WriteFile(outputPath, encrypted, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return nil
}

// DecryptFile descriptografa arquivo
func DecryptFile(inputPath, outputPath, password string) error {
	// Ler arquivo criptografado
	encrypted, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	// Descriptografar dados
	decrypted, err := DecryptWithAES256(encrypted, password)
	if err != nil {
		return fmt.Errorf("erro ao descriptografar: %v", err)
	}

	// Salvar arquivo descriptografado
	if err := os.WriteFile(outputPath, decrypted, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return nil
}
