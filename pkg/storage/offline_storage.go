package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineStorage gerencia storage local para mineradores
type OfflineStorage struct {
	DataPath   string
	Blockchain *LocalBlockchain
	MinerState *MinerState
	SyncQueue  *SyncQueue
	mu         sync.RWMutex
}

// LocalBlockchain representa blockchain local
type LocalBlockchain struct {
	Blocks    map[string]interface{} `json:"blocks"`
	LastBlock string                 `json:"last_block"`
	Height    int64                  `json:"height"`
	CreatedAt time.Time              `json:"created_at"`
}

// MinerState representa estado do minerador
type MinerState struct {
	MinerID     string    `json:"miner_id"`
	IsMining    bool      `json:"is_mining"`
	HashRate    float64   `json:"hash_rate"`
	LastMined   time.Time `json:"last_mined"`
	TotalBlocks int64     `json:"total_blocks"`
}

// SyncQueue representa fila de sincronização
type SyncQueue struct {
	PendingBlocks []string  `json:"pending_blocks"`
	LastSync      time.Time `json:"last_sync"`
	RetryCount    int       `json:"retry_count"`
}

// NewOfflineStorage cria novo storage offline
func NewOfflineStorage(dataPath string) (*OfflineStorage, error) {
	// Criar diretório se não existir
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório: %v", err)
	}

	storage := &OfflineStorage{
		DataPath: dataPath,
		Blockchain: &LocalBlockchain{
			Blocks:    make(map[string]interface{}),
			CreatedAt: time.Now(),
		},
		MinerState: &MinerState{
			LastMined: time.Now(),
		},
		SyncQueue: &SyncQueue{
			PendingBlocks: []string{},
			LastSync:      time.Now(),
		},
	}

	// Carregar dados existentes
	if err := storage.Load(); err != nil {
		return nil, fmt.Errorf("erro ao carregar dados: %v", err)
	}

	return storage, nil
}

// Save salva dados criptografados
func (storage *OfflineStorage) Save() error {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	data := map[string]interface{}{
		"blockchain":  storage.Blockchain,
		"miner_state": storage.MinerState,
		"sync_queue":  storage.SyncQueue,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dados: %v", err)
	}

	// Criptografar dados
	encryptedData, err := storage.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("erro ao criptografar dados: %v", err)
	}

	// Salvar arquivo
	filePath := filepath.Join(storage.DataPath, "offline_data.enc")
	if err := os.WriteFile(filePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return nil
}

// Load carrega dados criptografados
func (storage *OfflineStorage) Load() error {
	filePath := filepath.Join(storage.DataPath, "offline_data.enc")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Arquivo não existe, usar dados padrão
		return nil
	}

	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	// Descriptografar dados
	jsonData, err := storage.decrypt(encryptedData)
	if err != nil {
		return fmt.Errorf("erro ao descriptografar dados: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("erro ao deserializar dados: %v", err)
	}

	// Restaurar dados
	if blockchainData, ok := data["blockchain"]; ok {
		if blockchainBytes, err := json.Marshal(blockchainData); err == nil {
			json.Unmarshal(blockchainBytes, &storage.Blockchain)
		}
	}

	if minerStateData, ok := data["miner_state"]; ok {
		if minerStateBytes, err := json.Marshal(minerStateData); err == nil {
			json.Unmarshal(minerStateBytes, &storage.MinerState)
		}
	}

	if syncQueueData, ok := data["sync_queue"]; ok {
		if syncQueueBytes, err := json.Marshal(syncQueueData); err == nil {
			json.Unmarshal(syncQueueBytes, &storage.SyncQueue)
		}
	}

	return nil
}

// AddBlock adiciona bloco à blockchain local
func (storage *OfflineStorage) AddBlock(blockHash string, blockData interface{}) error {
	storage.mu.Lock()
	storage.Blockchain.Blocks[blockHash] = blockData
	storage.Blockchain.LastBlock = blockHash
	storage.Blockchain.Height++

	// Adicionar à fila de sincronização
	storage.SyncQueue.PendingBlocks = append(storage.SyncQueue.PendingBlocks, blockHash)
	storage.mu.Unlock()

	return storage.Save()
}

// GetBlock retorna bloco da blockchain local
func (storage *OfflineStorage) GetBlock(blockHash string) (interface{}, bool) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	block, exists := storage.Blockchain.Blocks[blockHash]
	return block, exists
}

// UpdateMinerState atualiza estado do minerador
func (storage *OfflineStorage) UpdateMinerState(state *MinerState) error {
	storage.mu.Lock()
	storage.MinerState = state
	storage.mu.Unlock()

	return storage.Save()
}

// GetPendingBlocks retorna blocos pendentes de sincronização
func (storage *OfflineStorage) GetPendingBlocks() []string {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	return append([]string{}, storage.SyncQueue.PendingBlocks...)
}

// MarkBlockSynced marca bloco como sincronizado
func (storage *OfflineStorage) MarkBlockSynced(blockHash string) error {
	storage.mu.Lock()

	// Remover da fila de pendentes
	for i, hash := range storage.SyncQueue.PendingBlocks {
		if hash == blockHash {
			storage.SyncQueue.PendingBlocks = append(
				storage.SyncQueue.PendingBlocks[:i],
				storage.SyncQueue.PendingBlocks[i+1:]...,
			)
			break
		}
	}

	storage.SyncQueue.LastSync = time.Now()
	storage.mu.Unlock()

	return storage.Save()
}

// encrypt criptografa dados com AES-256
func (storage *OfflineStorage) encrypt(data []byte) ([]byte, error) {
	key := storage.getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decrypt descriptografa dados com AES-256
func (storage *OfflineStorage) decrypt(data []byte) ([]byte, error) {
	key := storage.getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("dados criptografados muito curtos")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// getEncryptionKey retorna chave de criptografia
func (storage *OfflineStorage) getEncryptionKey() []byte {
	// Em produção, usar variável de ambiente ou keystore
	key := []byte("ordm-offline-storage-key-32bytes!!")
	return key[:32] // AES-256 precisa de 32 bytes
}
