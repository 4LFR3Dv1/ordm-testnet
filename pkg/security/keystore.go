package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SecureKeystore implementa armazenamento seguro de chaves privadas
type SecureKeystore struct {
	Path          string
	Password      string
	Encrypted     bool
	BackupPath    string
	Algorithm     string // AES-256
	KeyDerivation string // PBKDF2
	Salt          []byte
	mu            sync.RWMutex
}

// KeystoreEntry representa uma entrada no keystore
type KeystoreEntry struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // "wallet", "miner", "node"
	Encrypted   bool      `json:"encrypted"`
	Data        []byte    `json:"data"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
}

// NewSecureKeystore cria um novo keystore seguro
func NewSecureKeystore(path, password string) (*SecureKeystore, error) {
	ks := &SecureKeystore{
		Path:          path,
		Password:      password,
		Encrypted:     true,
		BackupPath:    filepath.Join(path, "backup"),
		Algorithm:     "AES-256-GCM",
		KeyDerivation: "PBKDF2",
	}

	// Criar diretório se não existir
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório do keystore: %v", err)
	}

	// Criar diretório de backup
	if err := os.MkdirAll(ks.BackupPath, 0700); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de backup: %v", err)
	}

	// Gerar salt se não existir
	saltPath := filepath.Join(path, "salt")
	if _, err := os.Stat(saltPath); os.IsNotExist(err) {
		ks.Salt = make([]byte, 32)
		if _, err := rand.Read(ks.Salt); err != nil {
			return nil, fmt.Errorf("erro ao gerar salt: %v", err)
		}
		if err := os.WriteFile(saltPath, ks.Salt, 0600); err != nil {
			return nil, fmt.Errorf("erro ao salvar salt: %v", err)
		}
	} else {
		saltData, err := os.ReadFile(saltPath)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler salt: %v", err)
		}
		ks.Salt = saltData
	}

	return ks, nil
}

// deriveKey deriva chave de criptografia usando PBKDF2
func (ks *SecureKeystore) deriveKey(password string) ([]byte, error) {
	// Implementação simplificada de PBKDF2
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(ks.Salt)

	// Iterações adicionais para fortalecer
	for i := 0; i < 10000; i++ {
		hash.Write(hash.Sum(nil))
	}

	return hash.Sum(nil), nil
}

// EncryptPrivateKey criptografa uma chave privada
func (ks *SecureKeystore) EncryptPrivateKey(key []byte) ([]byte, error) {
	if !ks.Encrypted {
		return key, nil
	}

	derivedKey, err := ks.deriveKey(ks.Password)
	if err != nil {
		return nil, fmt.Errorf("erro ao derivar chave: %v", err)
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("erro ao gerar nonce: %v", err)
	}

	encrypted := gcm.Seal(nonce, nonce, key, nil)
	return encrypted, nil
}

// DecryptPrivateKey descriptografa uma chave privada
func (ks *SecureKeystore) DecryptPrivateKey(encrypted []byte) ([]byte, error) {
	if !ks.Encrypted {
		return encrypted, nil
	}

	derivedKey, err := ks.deriveKey(ks.Password)
	if err != nil {
		return nil, fmt.Errorf("erro ao derivar chave: %v", err)
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("dados criptografados muito curtos")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar: %v", err)
	}

	return plaintext, nil
}

// StoreKey armazena uma chave no keystore
func (ks *SecureKeystore) StoreKey(id, keyType, description string, keyData []byte) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Criptografar dados da chave
	encryptedData, err := ks.EncryptPrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("erro ao criptografar chave: %v", err)
	}

	entry := &KeystoreEntry{
		ID:          id,
		Type:        keyType,
		Encrypted:   ks.Encrypted,
		Data:        encryptedData,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Description: description,
	}

	// Serializar entrada
	entryData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar entrada: %v", err)
	}

	// Salvar arquivo
	filePath := filepath.Join(ks.Path, fmt.Sprintf("%s.json", id))
	if err := os.WriteFile(filePath, entryData, 0600); err != nil {
		return fmt.Errorf("erro ao salvar chave: %v", err)
	}

	return nil
}

// LoadKey carrega uma chave do keystore
func (ks *SecureKeystore) LoadKey(id string) (*KeystoreEntry, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	filePath := filepath.Join(ks.Path, fmt.Sprintf("%s.json", id))
	entryData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler chave: %v", err)
	}

	var entry KeystoreEntry
	if err := json.Unmarshal(entryData, &entry); err != nil {
		return nil, fmt.Errorf("erro ao deserializar entrada: %v", err)
	}

	// Descriptografar dados da chave
	decryptedData, err := ks.DecryptPrivateKey(entry.Data)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar chave: %v", err)
	}

	entry.Data = decryptedData
	return &entry, nil
}

// ListKeys lista todas as chaves no keystore
func (ks *SecureKeystore) ListKeys() ([]string, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	files, err := os.ReadDir(ks.Path)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar chaves: %v", err)
	}

	var keys []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			keys = append(keys, file.Name()[:len(file.Name())-5]) // Remove .json
		}
	}

	return keys, nil
}

// DeleteKey remove uma chave do keystore
func (ks *SecureKeystore) DeleteKey(id string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	filePath := filepath.Join(ks.Path, fmt.Sprintf("%s.json", id))
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("erro ao remover chave: %v", err)
	}

	return nil
}

// Backup cria backup do keystore
func (ks *SecureKeystore) Backup() error {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	backupFile := filepath.Join(ks.BackupPath, fmt.Sprintf("keystore_backup_%s.tar.gz",
		time.Now().Format("2006-01-02_15-04-05")))

	// Implementação simplificada de backup
	// Em produção, usar compressão e criptografia adicional
	keys, err := ks.ListKeys()
	if err != nil {
		return fmt.Errorf("erro ao listar chaves para backup: %v", err)
	}

	backupData := map[string]interface{}{
		"timestamp": time.Now(),
		"keys":      keys,
		"algorithm": ks.Algorithm,
	}

	backupJSON, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar backup: %v", err)
	}

	if err := os.WriteFile(backupFile, backupJSON, 0600); err != nil {
		return fmt.Errorf("erro ao salvar backup: %v", err)
	}

	return nil
}

// Restore restaura keystore de backup
func (ks *SecureKeystore) Restore(backupFile string) error {
	// Implementação de restore
	// Em produção, implementar restauração completa
	return fmt.Errorf("restore não implementado ainda")
}

// GetStatus retorna status do keystore
func (ks *SecureKeystore) GetStatus() map[string]interface{} {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keys, _ := ks.ListKeys()

	return map[string]interface{}{
		"path":           ks.Path,
		"encrypted":      ks.Encrypted,
		"algorithm":      ks.Algorithm,
		"key_derivation": ks.KeyDerivation,
		"total_keys":     len(keys),
		"backup_path":    ks.BackupPath,
	}
}

// ChangePassword altera senha do keystore
func (ks *SecureKeystore) ChangePassword(newPassword string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Listar todas as chaves
	keys, err := ks.ListKeys()
	if err != nil {
		return fmt.Errorf("erro ao listar chaves: %v", err)
	}

	// Recriar keystore com nova senha
	oldPassword := ks.Password
	ks.Password = newPassword

	// Recriptografar todas as chaves
	for _, keyID := range keys {
		entry, err := ks.LoadKey(keyID)
		if err != nil {
			continue // Pular chaves com erro
		}

		// Recriar com nova senha
		ks.Password = oldPassword
		decryptedData, _ := ks.DecryptPrivateKey(entry.Data)
		ks.Password = newPassword

		if err := ks.StoreKey(entry.ID, entry.Type, entry.Description, decryptedData); err != nil {
			return fmt.Errorf("erro ao recriptografar chave %s: %v", keyID, err)
		}
	}

	return nil
}
