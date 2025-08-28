package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Block representa um bloco da blockchain
type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
	Hash         string        `json:"hash"`
	Nonce        uint64        `json:"nonce"`
	Difficulty   uint64        `json:"difficulty"`
	Miner        string        `json:"miner"`
	Timestamp    int64         `json:"timestamp"`
	Signature    string        `json:"signature,omitempty"`
}

// BlockHeader representa o cabeçalho de um bloco
type BlockHeader struct {
	PreviousHash string `json:"previous_hash"`
	MerkleRoot   string `json:"merkle_root"`
	Timestamp    int64  `json:"timestamp"`
	Difficulty   uint64 `json:"difficulty"`
	Nonce        uint64 `json:"nonce"`
}

// Transaction representa uma transação
type Transaction struct {
	TxHash    string `json:"tx_hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	Nonce     uint64 `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
	Status    string `json:"status"` // "pending", "confirmed", "failed"
}

// BlockValidator implementa validação de blocos
type BlockValidator struct {
	Difficulty uint64
}

// NewBlockValidator cria um novo validador de blocos
func NewBlockValidator(difficulty uint64) *BlockValidator {
	return &BlockValidator{
		Difficulty: difficulty,
	}
}

// NewBlock cria um novo bloco
func NewBlock(previousHash string, transactions []Transaction, miner string, difficulty uint64) *Block {
	block := &Block{
		Header: BlockHeader{
			PreviousHash: previousHash,
			MerkleRoot:   calculateMerkleRoot(transactions),
			Timestamp:    time.Now().Unix(),
			Difficulty:   difficulty,
			Nonce:        0,
		},
		Transactions: transactions,
		Miner:        miner,
		Timestamp:    time.Now().Unix(),
		Difficulty:   difficulty,
	}

	// Calcular hash inicial
	block.Hash = block.calculateHash()
	return block
}

// calculateHash calcula o hash do bloco
func (b *Block) calculateHash() string {
	// Criar string para hash (sem o hash atual e nonce)
	blockData := fmt.Sprintf("%s%s%d%d%s",
		b.Header.PreviousHash,
		b.Header.MerkleRoot,
		b.Header.Timestamp,
		b.Header.Difficulty,
		b.Miner,
	)

	// Adicionar nonce
	blockData += strconv.FormatUint(b.Nonce, 10)

	// Calcular hash SHA256
	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

// MineBlock executa a mineração do bloco (Proof of Work)
func (b *Block) MineBlock() error {
	target := b.calculateTarget()

	for {
		b.Hash = b.calculateHash()

		// Verificar se o hash atende ao target de dificuldade
		if b.meetsDifficulty(target) {
			return nil
		}

		b.Nonce++
		b.Header.Nonce = b.Nonce

		// Timeout de segurança
		if b.Nonce > 1000000 {
			return fmt.Errorf("timeout na mineração do bloco")
		}
	}
}

// calculateTarget calcula o target baseado na dificuldade
func (b *Block) calculateTarget() string {
	// Para dificuldade N, precisamos de N zeros no início do hash
	zeros := ""
	for i := uint64(0); i < b.Difficulty; i++ {
		zeros += "0"
	}
	return zeros
}

// meetsDifficulty verifica se o hash atende à dificuldade
func (b *Block) meetsDifficulty(target string) bool {
	return len(b.Hash) >= len(target) && b.Hash[:len(target)] == target
}

// ValidateBlock valida um bloco
func (bv *BlockValidator) ValidateBlock(block *Block, previousBlock *Block) error {
	// 1. Validar hash do bloco
	calculatedHash := block.calculateHash()
	if calculatedHash != block.Hash {
		return fmt.Errorf("hash do bloco inválido")
	}

	// 2. Validar Proof of Work
	if !block.meetsDifficulty(block.calculateTarget()) {
		return fmt.Errorf("proof of work inválido")
	}

	// 3. Validar link com bloco anterior
	if previousBlock != nil && block.Header.PreviousHash != previousBlock.Hash {
		return fmt.Errorf("hash do bloco anterior inválido")
	}

	// 4. Validar timestamp
	currentTime := time.Now().Unix()
	if block.Timestamp > currentTime+300 { // 5 minutos de tolerância
		return fmt.Errorf("timestamp do bloco muito no futuro")
	}

	// 5. Validar Merkle Root
	calculatedMerkleRoot := calculateMerkleRoot(block.Transactions)
	if calculatedMerkleRoot != block.Header.MerkleRoot {
		return fmt.Errorf("merkle root inválido")
	}

	// 6. Validar transações
	for i, tx := range block.Transactions {
		if err := bv.ValidateTransaction(&tx); err != nil {
			return fmt.Errorf("transação %d inválida: %v", i, err)
		}
	}

	return nil
}

// ValidateTransaction valida uma transação
func (bv *BlockValidator) ValidateTransaction(tx *Transaction) error {
	// 1. Validar hash da transação
	calculatedHash := calculateTransactionHash(tx)
	if calculatedHash != tx.TxHash {
		return fmt.Errorf("hash da transação inválido")
	}

	// 2. Validar assinatura (simplificado)
	if tx.Signature == "" {
		return fmt.Errorf("assinatura da transação ausente")
	}

	// 3. Validar valores
	if tx.Amount <= 0 {
		return fmt.Errorf("valor da transação deve ser positivo")
	}

	if tx.Fee < 0 {
		return fmt.Errorf("fee da transação não pode ser negativo")
	}

	// 4. Validar endereços
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("endereços de origem e destino são obrigatórios")
	}

	// 5. Validar timestamp
	currentTime := time.Now().Unix()
	if tx.Timestamp > currentTime+300 { // 5 minutos de tolerância
		return fmt.Errorf("timestamp da transação muito no futuro")
	}

	return nil
}

// calculateMerkleRoot calcula a raiz de Merkle das transações
func calculateMerkleRoot(transactions []Transaction) string {
	if len(transactions) == 0 {
		return ""
	}

	if len(transactions) == 1 {
		return transactions[0].TxHash
	}

	// Criar hashes das transações
	hashes := make([]string, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.TxHash
	}

	// Calcular árvore de Merkle
	for len(hashes) > 1 {
		var newHashes []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				hash := sha256.Sum256([]byte(combined))
				newHashes = append(newHashes, hex.EncodeToString(hash[:]))
			} else {
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}

	return hashes[0]
}

// calculateTransactionHash calcula o hash de uma transação
func calculateTransactionHash(tx *Transaction) string {
	data := fmt.Sprintf("%s%s%d%d%d%d",
		tx.From,
		tx.To,
		tx.Amount,
		tx.Fee,
		tx.Nonce,
		tx.Timestamp,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ToJSON converte o bloco para JSON
func (b *Block) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}

// FromJSON converte JSON para bloco
func FromJSON(data []byte) (*Block, error) {
	var block Block
	err := json.Unmarshal(data, &block)
	return &block, err
}
