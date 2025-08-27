package validation

import (
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"
	
	"ordm-main/pkg/blockchain"
	"ordm-main/pkg/crypto"
)

// PoSValidator gerencia a validação Proof of Stake
type PoSValidator struct {
	Validators    map[string]*Validator
	StakePools    map[string]*StakePool
	TotalStake    int64
	MinStake      int64
	mu            sync.RWMutex
}

// Validator representa um validador PoS
type Validator struct {
	ID            string    `json:"id"`
	PublicKey     []byte    `json:"public_key"`
	StakeAmount   int64     `json:"stake_amount"`
	Reputation    int       `json:"reputation"`
	IsActive      bool      `json:"is_active"`
	LastActivity  time.Time `json:"last_activity"`
	ValidatedBlocks int64   `json:"validated_blocks"`
	SlashCount    int       `json:"slash_count"`
}

// StakePool representa um pool de stake
type StakePool struct {
	ID            string    `json:"id"`
	ValidatorID   string    `json:"validator_id"`
	TotalStake    int64     `json:"total_stake"`
	Delegators    map[string]int64 `json:"delegators"`
	Rewards       int64     `json:"rewards"`
	CreatedAt     time.Time `json:"created_at"`
}

// BlockValidation representa a validação de um bloco
type BlockValidation struct {
	BlockHash     string    `json:"block_hash"`
	ValidatorID   string    `json:"validator_id"`
	IsValid       bool      `json:"is_valid"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
	StakeUsed     int64     `json:"stake_used"`
}

// NewPoSValidator cria um novo validador PoS
func NewPoSValidator(minStake int64) *PoSValidator {
	return &PoSValidator{
		Validators: make(map[string]*Validator),
		StakePools: make(map[string]*StakePool),
		MinStake:   minStake,
	}
}

// RegisterValidator registra um novo validador
func (pos *PoSValidator) RegisterValidator(minerIdentity *crypto.MinerIdentity, stakeAmount int64) error {
	pos.mu.Lock()
	defer pos.mu.Unlock()
	
	if stakeAmount < pos.MinStake {
		return fmt.Errorf("stake insuficiente: %d < %d", stakeAmount, pos.MinStake)
	}
	
	validator := &Validator{
		ID:           minerIdentity.MinerID,
		PublicKey:    minerIdentity.PublicKey,
		StakeAmount:  stakeAmount,
		Reputation:   100,
		IsActive:     true,
		LastActivity: time.Now(),
	}
	
	pos.Validators[validator.ID] = validator
	pos.TotalStake += stakeAmount
	
	// Criar stake pool
	stakePool := &StakePool{
		ID:          fmt.Sprintf("pool_%s", validator.ID),
		ValidatorID: validator.ID,
		TotalStake:  stakeAmount,
		Delegators:  make(map[string]int64),
		CreatedAt:   time.Now(),
	}
	
	pos.StakePools[stakePool.ID] = stakePool
	
	fmt.Printf("✅ Validador registrado: %s (stake: %d)\n", validator.ID, stakeAmount)
	return nil
}

// ValidateBlock valida um bloco usando PoS
func (pos *PoSValidator) ValidateBlock(block *blockchain.RealBlock, minerSignature []byte) (*BlockValidation, error) {
	pos.mu.Lock()
	defer pos.mu.Unlock()
	
	// Verificar se o minerador é um validador
	validator, exists := pos.Validators[block.Header.MinerID]
	if !exists {
		return &BlockValidation{
			BlockHash:   block.GetBlockHashString(),
			ValidatorID: block.Header.MinerID,
			IsValid:     false,
			Reason:      "minerador não é validador registrado",
			Timestamp:   time.Now(),
		}, nil
	}
	
	if !validator.IsActive {
		return &BlockValidation{
			BlockHash:   block.GetBlockHashString(),
			ValidatorID: block.Header.MinerID,
			IsValid:     false,
			Reason:      "validador inativo",
			Timestamp:   time.Now(),
		}, nil
	}
	
	// Verificar assinatura do bloco
	blockData, err := block.ToJSON()
	if err != nil {
		return &BlockValidation{
			BlockHash:   block.GetBlockHashString(),
			ValidatorID: block.Header.MinerID,
			IsValid:     false,
			Reason:      "erro ao serializar bloco",
			Timestamp:   time.Now(),
		}, nil
	}
	
	if !ed25519.Verify(validator.PublicKey, blockData, minerSignature) {
		return &BlockValidation{
			BlockHash:   block.GetBlockHashString(),
			ValidatorID: block.Header.MinerID,
			IsValid:     false,
			Reason:      "assinatura inválida",
			Timestamp:   time.Now(),
		}, nil
	}
	
	// Verificar PoW do bloco
	if err := block.VerifyBlock(); err != nil {
		return &BlockValidation{
			BlockHash:   block.GetBlockHashString(),
			ValidatorID: block.Header.MinerID,
			IsValid:     false,
			Reason:      "PoW inválido: " + err.Error(),
			Timestamp:   time.Now(),
		}, nil
	}
	
	// Validar transações
	if err := pos.validateTransactions(block); err != nil {
		return &BlockValidation{
			BlockHash:   block.GetBlockHashString(),
			ValidatorID: block.Header.MinerID,
			IsValid:     false,
			Reason:      "transações inválidas: " + err.Error(),
			Timestamp:   time.Now(),
		}, nil
	}
	
	// Bloco válido - atualizar estatísticas
	validator.ValidatedBlocks++
	validator.LastActivity = time.Now()
	validator.Reputation = min(validator.Reputation+1, 100)
	
	validation := &BlockValidation{
		BlockHash:   block.GetBlockHashString(),
		ValidatorID: block.Header.MinerID,
		IsValid:     true,
		Reason:      "bloco validado com sucesso",
		Timestamp:   time.Now(),
		StakeUsed:   validator.StakeAmount,
	}
	
	fmt.Printf("✅ Bloco validado: %s por %s\n", block.GetBlockHashString()[:16], validator.ID)
	return validation, nil
}

// validateTransactions valida as transações de um bloco
func (pos *PoSValidator) validateTransactions(block *blockchain.RealBlock) error {
	// Implementar validação de transações
	// Por enquanto, apenas verificar se há transações
	if len(block.Transactions) == 0 {
		return fmt.Errorf("bloco sem transações")
	}
	
	return nil
}

// SlashValidator penaliza um validador por comportamento malicioso
func (pos *PoSValidator) SlashValidator(validatorID string, reason string) error {
	pos.mu.Lock()
	defer pos.mu.Unlock()
	
	validator, exists := pos.Validators[validatorID]
	if !exists {
		return fmt.Errorf("validador não encontrado: %s", validatorID)
	}
	
	validator.SlashCount++
	validator.Reputation = max(validator.Reputation-10, 0)
	
	if validator.Reputation <= 0 {
		validator.IsActive = false
		fmt.Printf("🚫 Validador desativado: %s (reputação: %d)\n", validatorID, validator.Reputation)
	}
	
	fmt.Printf("⚠️ Validador penalizado: %s (%s)\n", validatorID, reason)
	return nil
}

// GetValidatorStats retorna estatísticas dos validadores
func (pos *PoSValidator) GetValidatorStats() map[string]interface{} {
	pos.mu.RLock()
	defer pos.mu.RUnlock()
	
	activeValidators := 0
	totalValidatedBlocks := int64(0)
	
	for _, validator := range pos.Validators {
		if validator.IsActive {
			activeValidators++
		}
		totalValidatedBlocks += validator.ValidatedBlocks
	}
	
	return map[string]interface{}{
		"total_validators":      len(pos.Validators),
		"active_validators":     activeValidators,
		"total_stake":           pos.TotalStake,
		"min_stake":             pos.MinStake,
		"total_validated_blocks": totalValidatedBlocks,
		"stake_pools":           len(pos.StakePools),
	}
}

// GetTopValidators retorna os validadores com maior stake
func (pos *PoSValidator) GetTopValidators(limit int) []*Validator {
	pos.mu.RLock()
	defer pos.mu.RUnlock()
	
	validators := make([]*Validator, 0, len(pos.Validators))
	for _, validator := range pos.Validators {
		if validator.IsActive {
			validators = append(validators, validator)
		}
	}
	
	// Ordenar por stake (maior primeiro)
	// Implementação simplificada - em produção usar sort.Slice
	
	return validators
}

// min e max são funções auxiliares
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
