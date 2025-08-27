package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// RealBlock representa um bloco real da blockchain
type RealBlock struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
	MinerProof   MinerProof    `json:"miner_proof"`
	Signature    []byte        `json:"signature"`
	mu           sync.RWMutex  `json:"-"`
}

// BlockHeader contém os metadados do bloco
type BlockHeader struct {
	Number       int64  `json:"number"`
	ParentHash   []byte `json:"parent_hash"`
	Timestamp    int64  `json:"timestamp"`
	Difficulty   uint64 `json:"difficulty"`
	Nonce        uint64 `json:"nonce"`
	MerkleRoot   []byte `json:"merkle_root"`
	MinerID      string `json:"miner_id"`
	Version      uint32 `json:"version"`
}

// MinerProof contém a prova de trabalho do minerador
type MinerProof struct {
	Hash         []byte `json:"hash"`
	Nonce        uint64 `json:"nonce"`
	Difficulty   uint64 `json:"difficulty"`
	Timestamp    int64  `json:"timestamp"`
	MinerID      string `json:"miner_id"`
	Target       []byte `json:"target"`
	WorkDone     string `json:"work_done"`
}

// RealTransaction representa uma transação real
type RealTransaction struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee"`
	Nonce       uint64 `json:"nonce"`
	Data        string `json:"data"`
	Signature   []byte `json:"signature"`
	Timestamp   int64  `json:"timestamp"`
	BlockHash   string `json:"block_hash,omitempty"`
	Status      string `json:"status"` // "pending", "confirmed", "failed"
}

// NewRealBlock cria um novo bloco real
func NewRealBlock(parentHash []byte, number int64, minerID string, difficulty uint64) *RealBlock {
	block := &RealBlock{
		Header: BlockHeader{
			Number:     number,
			ParentHash: parentHash,
			Timestamp:  time.Now().Unix(),
			Difficulty: difficulty,
			Nonce:      0,
			MinerID:    minerID,
			Version:    1,
		},
		Transactions: []Transaction{},
		MinerProof: MinerProof{
			Difficulty: difficulty,
			Timestamp:  time.Now().Unix(),
			MinerID:    minerID,
		},
	}
	
	// Calcular target baseado na dificuldade
	block.MinerProof.Target = calculateTarget(difficulty)
	
	return block
}

// AddTransaction adiciona uma transação ao bloco
func (rb *RealBlock) AddTransaction(tx Transaction) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	// Validar transação
	if err := rb.validateTransaction(tx); err != nil {
		return fmt.Errorf("transação inválida: %v", err)
	}
	
	// Adicionar transação
	rb.Transactions = append(rb.Transactions, tx)
	
	// Recalcular merkle root
	rb.Header.MerkleRoot = rb.calculateMerkleRoot()
	
	return nil
}

// MineBlock executa a mineração do bloco (PoW)
func (rb *RealBlock) MineBlock(targetDifficulty uint64) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	// Calcular hash inicial
	rb.Header.MerkleRoot = rb.calculateMerkleRoot()
	
	// Executar PoW
	target := calculateTarget(targetDifficulty)
	
	for nonce := uint64(0); nonce < 0xffffffffffffffff; nonce++ {
		rb.Header.Nonce = nonce
		rb.MinerProof.Nonce = nonce
		
		// Calcular hash do bloco
		hash := rb.calculateHash()
		
		// Verificar se o hash atende ao target
		if isHashValid(hash, target) {
			rb.MinerProof.Hash = hash
			rb.MinerProof.WorkDone = hex.EncodeToString(hash)
			return nil
		}
	}
	
	return fmt.Errorf("mineração timeout - não foi possível encontrar nonce válido")
}

// SignBlock assina o bloco com a chave privada do minerador
func (rb *RealBlock) SignBlock(signature []byte) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	rb.Signature = signature
}

// VerifyBlock verifica a integridade do bloco
func (rb *RealBlock) VerifyBlock() error {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	// Verificar hash do bloco
	calculatedHash := rb.calculateHash()
	if !isHashValid(calculatedHash, rb.MinerProof.Target) {
		return fmt.Errorf("hash do bloco não atende ao target de dificuldade")
	}
	
	// Verificar merkle root
	calculatedMerkleRoot := rb.calculateMerkleRoot()
	if !bytes.Equal(calculatedMerkleRoot, rb.Header.MerkleRoot) {
		return fmt.Errorf("merkle root inválido")
	}
	
	// Verificar transações
	for i, tx := range rb.Transactions {
		if err := rb.validateTransaction(tx); err != nil {
			return fmt.Errorf("transação %d inválida: %v", i, err)
		}
	}
	
	return nil
}

// GetBlockHash retorna o hash do bloco
func (rb *RealBlock) GetBlockHash() []byte {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	return rb.calculateHash()
}

// GetBlockHashString retorna o hash do bloco como string
func (rb *RealBlock) GetBlockHashString() string {
	return hex.EncodeToString(rb.GetBlockHash())
}

// GetParentHashString retorna o hash do bloco pai como string
func (rb *RealBlock) GetParentHashString() string {
	return hex.EncodeToString(rb.Header.ParentHash)
}

// GetMerkleRootString retorna o merkle root como string
func (rb *RealBlock) GetMerkleRootString() string {
	return hex.EncodeToString(rb.Header.MerkleRoot)
}

// calculateHash calcula o hash SHA-256 do bloco
func (rb *RealBlock) calculateHash() []byte {
	// Serializar header para hash
	headerData, _ := json.Marshal(rb.Header)
	hash := sha256.Sum256(headerData)
	return hash[:]
}

// calculateMerkleRoot calcula o merkle root das transações
func (rb *RealBlock) calculateMerkleRoot() []byte {
	if len(rb.Transactions) == 0 {
		// Se não há transações, usar hash vazio
		return make([]byte, 32)
	}
	
	// Para simplificar, usar hash da primeira transação
	// Em uma implementação real, seria uma árvore Merkle completa
	txData, _ := json.Marshal(rb.Transactions[0])
	hash := sha256.Sum256(txData)
	return hash[:]
}

// validateTransaction valida uma transação
func (rb *RealBlock) validateTransaction(tx Transaction) error {
	// Verificar campos obrigatórios
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("campos From e To são obrigatórios")
	}
	
	if tx.Amount <= 0 {
		return fmt.Errorf("valor deve ser maior que zero")
	}
	
	if tx.Fee < 0 {
		return fmt.Errorf("taxa não pode ser negativa")
	}
	
	// Verificar assinatura (se fornecida)
	if len(tx.Signature) > 0 {
		// Aqui seria feita a verificação da assinatura
		// Por simplicidade, apenas verificar se existe
	}
	
	return nil
}

// calculateTarget calcula o target baseado na dificuldade
func calculateTarget(difficulty uint64) []byte {
	target := new(big.Int).Lsh(big.NewInt(1), 256-difficulty)
	return target.Bytes()
}

// isHashValid verifica se o hash atende ao target
func isHashValid(hash, target []byte) bool {
	hashInt := new(big.Int).SetBytes(hash)
	targetInt := new(big.Int).SetBytes(target)
	return hashInt.Cmp(targetInt) <= 0
}

// ToJSON converte o bloco para JSON
func (rb *RealBlock) ToJSON() ([]byte, error) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	// Estrutura para serialização
	type blockJSON struct {
		Header       BlockHeader   `json:"header"`
		Transactions []Transaction `json:"transactions"`
		MinerProof   MinerProof    `json:"miner_proof"`
		Signature    string        `json:"signature"`
		BlockHash    string        `json:"block_hash"`
	}
	
	block := blockJSON{
		Header:       rb.Header,
		Transactions: rb.Transactions,
		MinerProof:   rb.MinerProof,
		Signature:    hex.EncodeToString(rb.Signature),
		BlockHash:    rb.GetBlockHashString(),
	}
	
	return json.MarshalIndent(block, "", "  ")
}

// FromJSON carrega o bloco de JSON
func (rb *RealBlock) FromJSON(data []byte) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	// Estrutura para deserialização
	type blockJSON struct {
		Header       BlockHeader   `json:"header"`
		Transactions []Transaction `json:"transactions"`
		MinerProof   MinerProof    `json:"miner_proof"`
		Signature    string        `json:"signature"`
	}
	
	var block blockJSON
	if err := json.Unmarshal(data, &block); err != nil {
		return err
	}
	
	// Converter signature
	signature, err := hex.DecodeString(block.Signature)
	if err != nil {
		return fmt.Errorf("signature inválida: %v", err)
	}
	
	rb.Header = block.Header
	rb.Transactions = block.Transactions
	rb.MinerProof = block.MinerProof
	rb.Signature = signature
	
	return nil
}

// GetBlockInfo retorna informações resumidas do bloco
func (rb *RealBlock) GetBlockInfo() map[string]interface{} {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	return map[string]interface{}{
		"number":         rb.Header.Number,
		"hash":           rb.GetBlockHashString(),
		"parent_hash":    rb.GetParentHashString(),
		"timestamp":      rb.Header.Timestamp,
		"difficulty":     rb.Header.Difficulty,
		"nonce":          rb.Header.Nonce,
		"merkle_root":    rb.GetMerkleRootString(),
		"miner_id":       rb.Header.MinerID,
		"transactions":   len(rb.Transactions),
		"signature":      hex.EncodeToString(rb.Signature),
		"work_done":      rb.MinerProof.WorkDone,
	}
}
