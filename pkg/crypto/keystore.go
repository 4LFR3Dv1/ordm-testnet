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
