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
