package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ConsensusType define o tipo de consenso
type ConsensusType string

const (
	POW    ConsensusType = "POW"
	POS    ConsensusType = "POS"
	HYBRID ConsensusType = "HYBRID"
)

// Block representa um bloco no consenso
type Block struct {
	Hash          string          `json:"hash"`
	ParentHash    string          `json:"parent_hash"`
	Number        int64           `json:"number"`
	Timestamp     int64           `json:"timestamp"`
	MinerID       string          `json:"miner_id"`
	ValidatorID   string          `json:"validator_id,omitempty"`
	Transactions  []Transaction   `json:"transactions"`
	ConsensusType ConsensusType   `json:"consensus_type"`
	Difficulty    int             `json:"difficulty"`
	Nonce         uint64          `json:"nonce"`
	StakeAmount   int64           `json:"stake_amount,omitempty"`
	Signature     string          `json:"signature,omitempty"`
	Validators    []string        `json:"validators,omitempty"`
	Votes         map[string]bool `json:"votes,omitempty"`
}

// Transaction representa uma transação
type Transaction struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Fee    int64  `json:"fee"`
	Data   string `json:"data"`
}

// Validator representa um validador PoS
type Validator struct {
	ID          string    `json:"id"`
	Address     string    `json:"address"`
	StakeAmount int64     `json:"stake_amount"`
	Reputation  float64   `json:"reputation"`
	LastActive  time.Time `json:"last_active"`
	IsActive    bool      `json:"is_active"`
	VotePower   int64     `json:"vote_power"`
}

// HybridConsensus implementa consenso híbrido PoW/PoS
type HybridConsensus struct {
	mu              sync.RWMutex
	validators      map[string]*Validator
	blocks          map[string]*Block
	latestBlock     *Block
	powDifficulty   int
	posThreshold    int64
	blockReward     int64
	stakeReward     int64
	slashingPenalty int64
	consensusType   ConsensusType
}

// NewHybridConsensus cria um novo consenso híbrido
func NewHybridConsensus(consensusType ConsensusType) *HybridConsensus {
	return &HybridConsensus{
		validators:      make(map[string]*Validator),
		blocks:          make(map[string]*Block),
		powDifficulty:   2,
		posThreshold:    1000,
		blockReward:     50,
		stakeReward:     10,
		slashingPenalty: 100,
		consensusType:   consensusType,
	}
}

// MineBlock minera um bloco usando PoW
func (hc *HybridConsensus) MineBlock(parentHash string, transactions []Transaction, minerID string) (*Block, error) {
	block := &Block{
		ParentHash:    parentHash,
		Number:        hc.getNextBlockNumber(),
		Timestamp:     time.Now().Unix(),
		MinerID:       minerID,
		Transactions:  transactions,
		ConsensusType: POW,
		Difficulty:    hc.powDifficulty,
		Votes:         make(map[string]bool),
	}

	// Mineração PoW
	nonce := uint64(0)
	for {
		block.Nonce = nonce
		block.Hash = hc.calculateBlockHash(block)

		if hc.verifyPoW(block.Hash, hc.powDifficulty) {
			break
		}
		nonce++
	}

	// Adicionar bloco
	hc.mu.Lock()
	hc.blocks[block.Hash] = block
	hc.latestBlock = block
	hc.mu.Unlock()

	fmt.Printf("⛏️ Bloco minerado: %s (PoW)\n", block.Hash[:8])
	return block, nil
}

// ValidateBlock valida um bloco usando PoS
func (hc *HybridConsensus) ValidateBlock(block *Block, validatorID string) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// Verificar se o validador existe e tem stake suficiente
	validator, exists := hc.validators[validatorID]
	if !exists {
		return fmt.Errorf("validador não encontrado: %s", validatorID)
	}

	if validator.StakeAmount < hc.posThreshold {
		return fmt.Errorf("stake insuficiente: %d < %d", validator.StakeAmount, hc.posThreshold)
	}

	// Verificar hash do bloco
	if block.Hash != hc.calculateBlockHash(block) {
		return fmt.Errorf("hash do bloco inválido")
	}

	// Verificar transações
	for _, tx := range block.Transactions {
		if err := hc.validateTransaction(tx); err != nil {
			return fmt.Errorf("transação inválida: %v", err)
		}
	}

	// Adicionar voto
	block.Votes[validatorID] = true

	// Verificar se há consenso suficiente
	if hc.hasConsensus(block) {
		block.ConsensusType = HYBRID
		hc.finalizeBlock(block)
	}

	return nil
}

// AddValidator adiciona um validador
func (hc *HybridConsensus) AddValidator(id, address string, stakeAmount int64) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if stakeAmount < hc.posThreshold {
		return fmt.Errorf("stake insuficiente para validador: %d < %d", stakeAmount, hc.posThreshold)
	}

	validator := &Validator{
		ID:          id,
		Address:     address,
		StakeAmount: stakeAmount,
		Reputation:  1.0,
		LastActive:  time.Now(),
		IsActive:    true,
		VotePower:   stakeAmount / 100, // 1 voto por 100 tokens de stake
	}

	hc.validators[id] = validator
	fmt.Printf("🎯 Validador adicionado: %s (stake: %d)\n", id, stakeAmount)
	return nil
}

// RemoveValidator remove um validador
func (hc *HybridConsensus) RemoveValidator(id string) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if validator, exists := hc.validators[id]; exists {
		validator.IsActive = false
		fmt.Printf("🚫 Validador removido: %s\n", id)
		return nil
	}

	return fmt.Errorf("validador não encontrado: %s", id)
}

// SlashValidator penaliza um validador
func (hc *HybridConsensus) SlashValidator(id string, reason string) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if validator, exists := hc.validators[id]; exists {
		// Aplicar penalidade
		penalty := hc.slashingPenalty
		if validator.StakeAmount < penalty {
			penalty = validator.StakeAmount
		}
		validator.StakeAmount -= penalty
		validator.Reputation *= 0.5 // Reduzir reputação

		// Desativar se stake ficou muito baixo
		if validator.StakeAmount < hc.posThreshold {
			validator.IsActive = false
		}

		fmt.Printf("⚡ Validador penalizado: %s (razão: %s, penalidade: %d)\n", id, reason, penalty)
		return nil
	}

	return fmt.Errorf("validador não encontrado: %s", id)
}

// GetValidators retorna lista de validadores
func (hc *HybridConsensus) GetValidators() []*Validator {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	validators := make([]*Validator, 0, len(hc.validators))
	for _, validator := range hc.validators {
		validators = append(validators, validator)
	}
	return validators
}

// GetActiveValidators retorna validadores ativos
func (hc *HybridConsensus) GetActiveValidators() []*Validator {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var active []*Validator
	for _, validator := range hc.validators {
		if validator.IsActive {
			active = append(active, validator)
		}
	}
	return active
}

// GetLatestBlock retorna o último bloco
func (hc *HybridConsensus) GetLatestBlock() *Block {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.latestBlock
}

// GetBlock retorna um bloco específico
func (hc *HybridConsensus) GetBlock(hash string) *Block {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.blocks[hash]
}

// calculateBlockHash calcula o hash de um bloco
func (hc *HybridConsensus) calculateBlockHash(block *Block) string {
	data := fmt.Sprintf("%s|%d|%d|%s|%d|%d",
		block.ParentHash,
		block.Number,
		block.Timestamp,
		block.MinerID,
		block.Difficulty,
		block.Nonce,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// verifyPoW verifica se o hash atende à dificuldade PoW
func (hc *HybridConsensus) verifyPoW(hash string, difficulty int) bool {
	zeros := 0
	for _, char := range hash {
		if char == '0' {
			zeros++
		} else {
			break
		}
	}
	return zeros >= difficulty
}

// validateTransaction valida uma transação
func (hc *HybridConsensus) validateTransaction(tx Transaction) error {
	if tx.Amount <= 0 {
		return fmt.Errorf("quantidade inválida: %d", tx.Amount)
	}

	if tx.Fee < 0 {
		return fmt.Errorf("fee inválida: %d", tx.Fee)
	}

	// Verificar se o remetente tem saldo suficiente
	// (implementação simplificada)
	return nil
}

// hasConsensus verifica se há consenso suficiente
func (hc *HybridConsensus) hasConsensus(block *Block) bool {
	activeValidators := hc.GetActiveValidators()
	if len(activeValidators) == 0 {
		return false
	}

	totalVotes := 0
	for _, validator := range activeValidators {
		if block.Votes[validator.ID] {
			totalVotes += int(validator.VotePower)
		}
	}

	totalVotePower := 0
	for _, validator := range activeValidators {
		totalVotePower += int(validator.VotePower)
	}

	// Consenso se 2/3 dos validadores votaram
	return float64(totalVotes) >= float64(totalVotePower)*2.0/3.0
}

// finalizeBlock finaliza um bloco
func (hc *HybridConsensus) finalizeBlock(block *Block) {
	// Distribuir recompensas
	hc.distributeRewards(block)

	// Atualizar reputações
	hc.updateReputations(block)

	fmt.Printf("✅ Bloco finalizado: %s (consenso híbrido)\n", block.Hash[:8])
}

// distributeRewards distribui recompensas
func (hc *HybridConsensus) distributeRewards(block *Block) {
	// Recompensa para minerador
	if validator, exists := hc.validators[block.MinerID]; exists {
		validator.StakeAmount += hc.blockReward
	}

	// Recompensa para validadores
	for validatorID := range block.Votes {
		if validator, exists := hc.validators[validatorID]; exists {
			validator.StakeAmount += hc.stakeReward
		}
	}
}

// updateReputations atualiza reputações dos validadores
func (hc *HybridConsensus) updateReputations(block *Block) {
	for validatorID, voted := range block.Votes {
		if validator, exists := hc.validators[validatorID]; exists {
			if voted {
				validator.Reputation = min(validator.Reputation*1.1, 1.0)
			} else {
				validator.Reputation = max(validator.Reputation*0.9, 0.1)
			}
		}
	}
}

// getNextBlockNumber retorna o próximo número de bloco
func (hc *HybridConsensus) getNextBlockNumber() int64 {
	if hc.latestBlock == nil {
		return 1
	}
	return hc.latestBlock.Number + 1
}

// min retorna o mínimo entre dois valores
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max retorna o máximo entre dois valores
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
