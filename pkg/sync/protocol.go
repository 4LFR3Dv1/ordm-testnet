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
