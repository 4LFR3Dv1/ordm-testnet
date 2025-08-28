#!/bin/bash

# 💾 Script para PARTE 2A: Storage Offline
# Subparte 2.1 do PLANO_ATUALIZACOES.md

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

log "🔄 Iniciando PARTE 2A: Storage Offline"

# 2.1.1 Implementar storage local para mineradores
log "2.1.1 - Criando pkg/storage/offline_storage.go..."
cat > pkg/storage/offline_storage.go << 'EOF'
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
	DataPath     string
	Blockchain   *LocalBlockchain
	MinerState   *MinerState
	SyncQueue    *SyncQueue
	mu           sync.RWMutex
}

// LocalBlockchain representa blockchain local
type LocalBlockchain struct {
	Blocks       map[string]interface{} `json:"blocks"`
	LastBlock    string                 `json:"last_block"`
	Height       int64                  `json:"height"`
	CreatedAt    time.Time              `json:"created_at"`
}

// MinerState representa estado do minerador
type MinerState struct {
	MinerID      string    `json:"miner_id"`
	IsMining     bool      `json:"is_mining"`
	HashRate     float64   `json:"hash_rate"`
	LastMined    time.Time `json:"last_mined"`
	TotalBlocks  int64     `json:"total_blocks"`
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
			CreatedAt: time.Now(),
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
func (os *OfflineStorage) Save() error {
	os.mu.Lock()
	defer os.mu.Unlock()

	data := map[string]interface{}{
		"blockchain": os.Blockchain,
		"miner_state": os.MinerState,
		"sync_queue": os.SyncQueue,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dados: %v", err)
	}

	// Criptografar dados
	encryptedData, err := os.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("erro ao criptografar dados: %v", err)
	}

	// Salvar arquivo
	filePath := filepath.Join(os.DataPath, "offline_data.enc")
	if err := os.WriteFile(filePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return nil
}

// Load carrega dados criptografados
func (os *OfflineStorage) Load() error {
	filePath := filepath.Join(os.DataPath, "offline_data.enc")
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Arquivo não existe, usar dados padrão
		return nil
	}

	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	// Descriptografar dados
	jsonData, err := os.decrypt(encryptedData)
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
			json.Unmarshal(blockchainBytes, &os.Blockchain)
		}
	}

	if minerStateData, ok := data["miner_state"]; ok {
		if minerStateBytes, err := json.Marshal(minerStateData); err == nil {
			json.Unmarshal(minerStateBytes, &os.MinerState)
		}
	}

	if syncQueueData, ok := data["sync_queue"]; ok {
		if syncQueueBytes, err := json.Marshal(syncQueueData); err == nil {
			json.Unmarshal(syncQueueBytes, &os.SyncQueue)
		}
	}

	return nil
}

// AddBlock adiciona bloco à blockchain local
func (os *OfflineStorage) AddBlock(blockHash string, blockData interface{}) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	os.Blockchain.Blocks[blockHash] = blockData
	os.Blockchain.LastBlock = blockHash
	os.Blockchain.Height++

	// Adicionar à fila de sincronização
	os.SyncQueue.PendingBlocks = append(os.SyncQueue.PendingBlocks, blockHash)

	return os.Save()
}

// GetBlock retorna bloco da blockchain local
func (os *OfflineStorage) GetBlock(blockHash string) (interface{}, bool) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	block, exists := os.Blockchain.Blocks[blockHash]
	return block, exists
}

// UpdateMinerState atualiza estado do minerador
func (os *OfflineStorage) UpdateMinerState(state *MinerState) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	os.MinerState = state
	return os.Save()
}

// GetPendingBlocks retorna blocos pendentes de sincronização
func (os *OfflineStorage) GetPendingBlocks() []string {
	os.mu.RLock()
	defer os.mu.RUnlock()

	return append([]string{}, os.SyncQueue.PendingBlocks...)
}

// MarkBlockSynced marca bloco como sincronizado
func (os *OfflineStorage) MarkBlockSynced(blockHash string) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	// Remover da fila de pendentes
	for i, hash := range os.SyncQueue.PendingBlocks {
		if hash == blockHash {
			os.SyncQueue.PendingBlocks = append(
				os.SyncQueue.PendingBlocks[:i],
				os.SyncQueue.PendingBlocks[i+1:]...,
			)
			break
		}
	}

	os.SyncQueue.LastSync = time.Now()
	return os.Save()
}

// encrypt criptografa dados com AES-256
func (os *OfflineStorage) encrypt(data []byte) ([]byte, error) {
	key := os.getEncryptionKey()
	
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
func (os *OfflineStorage) decrypt(data []byte) ([]byte, error) {
	key := os.getEncryptionKey()
	
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
func (os *OfflineStorage) getEncryptionKey() []byte {
	// Em produção, usar variável de ambiente ou keystore
	key := []byte("ordm-offline-storage-key-32bytes!!")
	return key[:32] // AES-256 precisa de 32 bytes
}
EOF

log "✅ PARTE 2A: Storage Offline concluída!"
log "📋 Arquivo criado: pkg/storage/offline_storage.go"

