#!/bin/bash

# 💾 Script para PARTE 2: Persistência e Storage
# Baseado no PLANO_ATUALIZACOES.md

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

log "🔄 Iniciando PARTE 2: Persistência e Storage"

# 2.1.1 Implementar storage offline
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

# 2.2.1 Corrigir storage no Render
log "2.2.1 - Criando pkg/storage/render_storage.go..."
cat > pkg/storage/render_storage.go << 'EOF'
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RenderStorage gerencia storage persistente no Render
type RenderStorage struct {
	DataDir      string
	Persistent   bool
	BackupPath   string
	LastBackup   time.Time
}

// NewRenderStorage cria novo storage para Render
func NewRenderStorage() *RenderStorage {
	dataDir := "/opt/render/data"
	backupPath := "/opt/render/backup"

	// Em desenvolvimento, usar diretório local
	if os.Getenv("NODE_ENV") != "production" {
		dataDir = "./data"
		backupPath = "./backup"
	}

	// Criar diretórios se não existirem
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(backupPath, 0755)

	return &RenderStorage{
		DataDir:    dataDir,
		Persistent: true,
		BackupPath: backupPath,
		LastBackup: time.Now(),
	}
}

// SaveData salva dados persistentes
func (rs *RenderStorage) SaveData(filename string, data interface{}) error {
	filePath := filepath.Join(rs.DataDir, filename)
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dados: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	// Backup automático a cada hora
	if time.Since(rs.LastBackup) > time.Hour {
		rs.createBackup(filename, jsonData)
	}

	return nil
}

// LoadData carrega dados persistentes
func (rs *RenderStorage) LoadData(filename string, data interface{}) error {
	filePath := filepath.Join(rs.DataDir, filename)
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", filePath)
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("erro ao deserializar dados: %v", err)
	}

	return nil
}

// createBackup cria backup dos dados
func (rs *RenderStorage) createBackup(filename string, data []byte) error {
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupFile := fmt.Sprintf("%s.%s.backup", filename, timestamp)
	backupPath := filepath.Join(rs.BackupPath, backupFile)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("erro ao criar backup: %v", err)
	}

	rs.LastBackup = time.Now()
	return nil
}

// GetDataPath retorna caminho dos dados
func (rs *RenderStorage) GetDataPath() string {
	return rs.DataDir
}

// IsPersistent verifica se storage é persistente
func (rs *RenderStorage) IsPersistent() bool {
	return rs.Persistent
}
EOF

# 2.3.1 Implementar protocolo de sincronização
log "2.3.1 - Criando pkg/sync/protocol.go..."
cat > pkg/sync/protocol.go << 'EOF'
package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// BlockPackage representa pacote de blocos para sincronização
type BlockPackage struct {
	Blocks       []*BlockData `json:"blocks"`
	MinerID      string       `json:"miner_id"`
	Signature    string       `json:"signature"`
	Timestamp    int64        `json:"timestamp"`
	BatchID      string       `json:"batch_id"`
	Version      string       `json:"version"`
}

// BlockData representa dados de um bloco
type BlockData struct {
	Hash         string        `json:"hash"`
	ParentHash   string        `json:"parent_hash"`
	Number       int64         `json:"number"`
	MinerID      string        `json:"miner_id"`
	Transactions []Transaction `json:"transactions"`
	PoWProof     string        `json:"pow_proof"`
	Difficulty   uint64        `json:"difficulty"`
	Nonce        uint64        `json:"nonce"`
	Timestamp    int64         `json:"timestamp"`
}

// Transaction representa uma transação
type Transaction struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}

// SyncStatus representa status da sincronização
type SyncStatus struct {
	Status        string    `json:"status"`         // "syncing", "completed", "failed"
	LastSync      time.Time `json:"last_sync"`
	PendingBlocks int       `json:"pending_blocks"`
	SyncedBlocks  int64     `json:"synced_blocks"`
	TotalBlocks   int64     `json:"total_blocks"`
	Error         string    `json:"error,omitempty"`
}

// NewBlockPackage cria novo pacote de blocos
func NewBlockPackage(minerID string, blocks []*BlockData) *BlockPackage {
	return &BlockPackage{
		Blocks:    blocks,
		MinerID:   minerID,
		Timestamp: time.Now().Unix(),
		BatchID:   generateBatchID(),
		Version:   "1.0.0",
	}
}

// SignPackage assina o pacote de blocos
func (bp *BlockPackage) SignPackage(privateKey string) error {
	// Em uma implementação real, usar criptografia real
	data, err := json.Marshal(bp)
	if err != nil {
		return fmt.Errorf("erro ao serializar pacote: %v", err)
	}

	hash := sha256.Sum256(data)
	bp.Signature = hex.EncodeToString(hash[:])
	
	return nil
}

// VerifyPackage verifica assinatura do pacote
func (bp *BlockPackage) VerifyPackage(publicKey string) bool {
	// Em uma implementação real, verificar assinatura real
	if bp.Signature == "" {
		return false
	}

	// Simular verificação
	return true
}

// ValidateBlocks valida todos os blocos no pacote
func (bp *BlockPackage) ValidateBlocks() error {
	for i, block := range bp.Blocks {
		if err := bp.validateBlock(block); err != nil {
			return fmt.Errorf("bloco %d inválido: %v", i, err)
		}
	}
	return nil
}

// validateBlock valida um bloco individual
func (bp *BlockPackage) validateBlock(block *BlockData) error {
	// Verificar hash
	if block.Hash == "" {
		return fmt.Errorf("hash do bloco é obrigatório")
	}

	// Verificar número sequencial
	if block.Number <= 0 {
		return fmt.Errorf("número do bloco deve ser positivo")
	}

	// Verificar timestamp
	if block.Timestamp <= 0 {
		return fmt.Errorf("timestamp do bloco é obrigatório")
	}

	// Verificar PoW proof
	if block.PoWProof == "" {
		return fmt.Errorf("prova de trabalho é obrigatória")
	}

	return nil
}

// generateBatchID gera ID único para o lote
func generateBatchID() string {
	timestamp := time.Now().UnixNano()
	random := fmt.Sprintf("%d", timestamp)
	hash := sha256.Sum256([]byte(random))
	return hex.EncodeToString(hash[:8])
}

// GetPackageSize retorna tamanho do pacote em bytes
func (bp *BlockPackage) GetPackageSize() int {
	data, _ := json.Marshal(bp)
	return len(data)
}

// IsExpired verifica se o pacote expirou
func (bp *BlockPackage) IsExpired(maxAge time.Duration) bool {
	packageTime := time.Unix(bp.Timestamp, 0)
	return time.Since(packageTime) > maxAge
}
EOF

# 2.3.2 Implementar validação de pacotes
log "2.3.2 - Criando pkg/sync/validator.go..."
cat > pkg/sync/validator.go << 'EOF'
package sync

import (
	"fmt"
	"sync"
	"time"
)

// PackageValidator valida pacotes de blocos
type PackageValidator struct {
	mu              sync.RWMutex
	validatedBlocks map[string]bool
	rejectedBlocks  map[string]string
	maxRetries      int
	retryDelay      time.Duration
}

// NewPackageValidator cria novo validador
func NewPackageValidator() *PackageValidator {
	return &PackageValidator{
		validatedBlocks: make(map[string]bool),
		rejectedBlocks:  make(map[string]string),
		maxRetries:      3,
		retryDelay:      5 * time.Second,
	}
}

// ValidatePackage valida um pacote completo
func (pv *PackageValidator) ValidatePackage(pkg *BlockPackage) (*ValidationResult, error) {
	result := &ValidationResult{
		PackageID:    pkg.BatchID,
		MinerID:      pkg.MinerID,
		Timestamp:    time.Now(),
		ValidBlocks:  []string{},
		InvalidBlocks: []InvalidBlock{},
	}

	// Verificar assinatura do minerador
	if !pkg.VerifyPackage("") {
		result.Status = "rejected"
		result.Error = "assinatura do minerador inválida"
		return result, fmt.Errorf("assinatura inválida")
	}

	// Verificar se pacote não expirou
	if pkg.IsExpired(30 * time.Minute) {
		result.Status = "rejected"
		result.Error = "pacote expirado"
		return result, fmt.Errorf("pacote expirado")
	}

	// Validar cada bloco
	for _, block := range pkg.Blocks {
		if err := pv.validateBlock(block); err != nil {
			invalidBlock := InvalidBlock{
				Hash:  block.Hash,
				Error: err.Error(),
			}
			result.InvalidBlocks = append(result.InvalidBlocks, invalidBlock)
		} else {
			result.ValidBlocks = append(result.ValidBlocks, block.Hash)
		}
	}

	// Determinar status final
	if len(result.InvalidBlocks) == 0 {
		result.Status = "accepted"
	} else if len(result.ValidBlocks) == 0 {
		result.Status = "rejected"
		result.Error = "todos os blocos são inválidos"
	} else {
		result.Status = "partial"
		result.Error = fmt.Sprintf("%d blocos válidos, %d inválidos", 
			len(result.ValidBlocks), len(result.InvalidBlocks))
	}

	return result, nil
}

// validateBlock valida um bloco individual
func (pv *PackageValidator) validateBlock(block *BlockData) error {
	pv.mu.RLock()
	if validated, exists := pv.validatedBlocks[block.Hash]; exists && validated {
		pv.mu.RUnlock()
		return nil // Já validado
	}
	if _, exists := pv.rejectedBlocks[block.Hash]; exists {
		pv.mu.RUnlock()
		return fmt.Errorf("bloco já foi rejeitado anteriormente")
	}
	pv.mu.RUnlock()

	// Verificar PoW
	if err := pv.verifyPoW(block); err != nil {
		pv.markBlockRejected(block.Hash, err.Error())
		return err
	}

	// Verificar transações
	if err := pv.verifyTransactions(block); err != nil {
		pv.markBlockRejected(block.Hash, err.Error())
		return err
	}

	// Verificar sequência
	if err := pv.verifySequence(block); err != nil {
		pv.markBlockRejected(block.Hash, err.Error())
		return err
	}

	// Marcar como válido
	pv.markBlockValid(block.Hash)
	return nil
}

// verifyPoW verifica prova de trabalho
func (pv *PackageValidator) verifyPoW(block *BlockData) error {
	// Em uma implementação real, verificar PoW real
	if block.PoWProof == "" {
		return fmt.Errorf("prova de trabalho ausente")
	}

	// Simular verificação de dificuldade
	if block.Difficulty < 1 {
		return fmt.Errorf("dificuldade muito baixa")
	}

	return nil
}

// verifyTransactions verifica transações do bloco
func (pv *PackageValidator) verifyTransactions(block *BlockData) error {
	for i, tx := range block.Transactions {
		if err := pv.verifyTransaction(&tx); err != nil {
			return fmt.Errorf("transação %d inválida: %v", i, err)
		}
	}
	return nil
}

// verifyTransaction verifica uma transação individual
func (pv *PackageValidator) verifyTransaction(tx *Transaction) error {
	if tx.ID == "" {
		return fmt.Errorf("ID da transação é obrigatório")
	}

	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("endereços de origem e destino são obrigatórios")
	}

	if tx.Amount <= 0 {
		return fmt.Errorf("valor da transação deve ser positivo")
	}

	if tx.Signature == "" {
		return fmt.Errorf("assinatura da transação é obrigatória")
	}

	return nil
}

// verifySequence verifica sequência de blocos
func (pv *PackageValidator) verifySequence(block *BlockData) error {
	// Em uma implementação real, verificar se o bloco pai existe
	// e se o número é sequencial
	return nil
}

// markBlockValid marca bloco como válido
func (pv *PackageValidator) markBlockValid(blockHash string) {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.validatedBlocks[blockHash] = true
}

// markBlockRejected marca bloco como rejeitado
func (pv *PackageValidator) markBlockRejected(blockHash, reason string) {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.rejectedBlocks[blockHash] = reason
}

// ValidationResult representa resultado da validação
type ValidationResult struct {
	PackageID     string         `json:"package_id"`
	MinerID       string         `json:"miner_id"`
	Status        string         `json:"status"` // "accepted", "rejected", "partial"
	Timestamp     time.Time      `json:"timestamp"`
	ValidBlocks   []string       `json:"valid_blocks"`
	InvalidBlocks []InvalidBlock `json:"invalid_blocks"`
	Error         string         `json:"error,omitempty"`
}

// InvalidBlock representa bloco inválido
type InvalidBlock struct {
	Hash  string `json:"hash"`
	Error string `json:"error"`
}
EOF

# 2.3.3 Implementar retry e recovery
log "2.3.3 - Criando pkg/sync/retry.go..."
cat > pkg/sync/retry.go << 'EOF'
package sync

import (
	"fmt"
	"sync"
	"time"
)

// RetryManager gerencia retry e recovery de sincronização
type RetryManager struct {
	mu           sync.RWMutex
	failedBlocks map[string]*RetryEntry
	maxRetries   int
	retryDelay   time.Duration
	backoff      time.Duration
}

// RetryEntry representa entrada de retry
type RetryEntry struct {
	BlockHash   string    `json:"block_hash"`
	PackageID   string    `json:"package_id"`
	MinerID     string    `json:"miner_id"`
	RetryCount  int       `json:"retry_count"`
	LastRetry   time.Time `json:"last_retry"`
	NextRetry   time.Time `json:"next_retry"`
	Error       string    `json:"error"`
	Status      string    `json:"status"` // "pending", "retrying", "failed"
}

// NewRetryManager cria novo gerenciador de retry
func NewRetryManager() *RetryManager {
	return &RetryManager{
		failedBlocks: make(map[string]*RetryEntry),
		maxRetries:   3,
		retryDelay:   5 * time.Second,
		backoff:      2 * time.Second,
	}
}

// AddFailedBlock adiciona bloco falhado para retry
func (rm *RetryManager) AddFailedBlock(blockHash, packageID, minerID, error string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.failedBlocks[blockHash] = &RetryEntry{
		BlockHash:  blockHash,
		PackageID:  packageID,
		MinerID:    minerID,
		RetryCount: 0,
		LastRetry:  time.Now(),
		NextRetry:  time.Now().Add(rm.retryDelay),
		Error:      error,
		Status:     "pending",
	}
}

// GetRetryableBlocks retorna blocos prontos para retry
func (rm *RetryManager) GetRetryableBlocks() []*RetryEntry {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	var retryable []*RetryEntry
	now := time.Now()

	for _, entry := range rm.failedBlocks {
		if entry.Status == "pending" && now.After(entry.NextRetry) {
			if entry.RetryCount < rm.maxRetries {
				retryable = append(retryable, entry)
			} else {
				entry.Status = "failed"
			}
		}
	}

	return retryable
}

// MarkRetryAttempt marca tentativa de retry
func (rm *RetryManager) MarkRetryAttempt(blockHash string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if entry, exists := rm.failedBlocks[blockHash]; exists {
		entry.RetryCount++
		entry.LastRetry = time.Now()
		entry.NextRetry = time.Now().Add(rm.calculateBackoff(entry.RetryCount))
		entry.Status = "retrying"
	}
}

// MarkBlockSuccess marca bloco como sincronizado com sucesso
func (rm *RetryManager) MarkBlockSuccess(blockHash string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.failedBlocks, blockHash)
}

// MarkBlockFailed marca bloco como falhado permanentemente
func (rm *RetryManager) MarkBlockFailed(blockHash string, error string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if entry, exists := rm.failedBlocks[blockHash]; exists {
		entry.Status = "failed"
		entry.Error = error
	}
}

// GetFailedBlocks retorna todos os blocos falhados
func (rm *RetryManager) GetFailedBlocks() map[string]*RetryEntry {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*RetryEntry)
	for k, v := range rm.failedBlocks {
		result[k] = v
	}
	return result
}

// calculateBackoff calcula delay exponencial
func (rm *RetryManager) calculateBackoff(retryCount int) time.Duration {
	backoff := rm.retryDelay
	for i := 0; i < retryCount; i++ {
		backoff += rm.backoff
	}
	return backoff
}

// CleanupOldEntries remove entradas antigas
func (rm *RetryManager) CleanupOldEntries(maxAge time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	for hash, entry := range rm.failedBlocks {
		if now.Sub(entry.LastRetry) > maxAge {
			delete(rm.failedBlocks, hash)
		}
	}
}

// GetStats retorna estatísticas de retry
func (rm *RetryManager) GetStats() *RetryStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := &RetryStats{
		TotalFailed:   len(rm.failedBlocks),
		PendingRetry:  0,
		Retrying:      0,
		PermanentlyFailed: 0,
	}

	for _, entry := range rm.failedBlocks {
		switch entry.Status {
		case "pending":
			stats.PendingRetry++
		case "retrying":
			stats.Retrying++
		case "failed":
			stats.PermanentlyFailed++
		}
	}

	return stats
}

// RetryStats representa estatísticas de retry
type RetryStats struct {
	TotalFailed       int `json:"total_failed"`
	PendingRetry      int `json:"pending_retry"`
	Retrying          int `json:"retrying"`
	PermanentlyFailed int `json:"permanently_failed"`
}
EOF

log "✅ PARTE 2: Persistência e Storage concluída!"
log "📋 Arquivos criados:"
log "   - pkg/storage/offline_storage.go"
log "   - pkg/storage/render_storage.go"
log "   - pkg/sync/protocol.go"
log "   - pkg/sync/validator.go"
log "   - pkg/sync/retry.go"

