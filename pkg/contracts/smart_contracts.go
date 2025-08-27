package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ContractType define tipos de contratos
type ContractType string

const (
	TimelockContract    ContractType = "timelock"
	MultisigContract    ContractType = "multisig"
	EscrowContract      ContractType = "escrow"
	VestingContract     ContractType = "vesting"
	ConditionalContract ContractType = "conditional"
)

// ContractStatus define status do contrato
type ContractStatus string

const (
	ContractPending   ContractStatus = "pending"
	ContractActive    ContractStatus = "active"
	ContractExecuted  ContractStatus = "executed"
	ContractExpired   ContractStatus = "expired"
	ContractCancelled ContractStatus = "cancelled"
)

// SmartContract representa um contrato inteligente
type SmartContract struct {
	ID              string                 `json:"id"`
	Type            ContractType           `json:"type"`
	Status          ContractStatus         `json:"status"`
	Creator         string                 `json:"creator"`
	Participants    []string               `json:"participants"`
	Amount          int64                  `json:"amount"`
	Conditions      map[string]interface{} `json:"conditions"`
	ExecutedAt      int64                  `json:"executed_at,omitempty"`
	CreatedAt       int64                  `json:"created_at"`
	ExpiresAt       int64                  `json:"expires_at,omitempty"`
	Signatures      []Signature            `json:"signatures"`
	Code            string                 `json:"code"`
	Data            map[string]interface{} `json:"data"`
	GasUsed         int64                  `json:"gas_used"`
	GasLimit        int64                  `json:"gas_limit"`
	GasPrice        int64                  `json:"gas_price"`
	BlockNumber     int64                  `json:"block_number"`
	TransactionHash string                 `json:"transaction_hash"`
}

// Signature representa uma assinatura
type Signature struct {
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}

// TimelockContract representa contrato com timelock
type TimelockContract struct {
	*SmartContract
	UnlockTime int64  `json:"unlock_time"`
	Recipient  string `json:"recipient"`
}

// MultisigContract representa contrato multi-signature
type MultisigContract struct {
	*SmartContract
	RequiredSignatures int      `json:"required_signatures"`
	Signers            []string `json:"signers"`
}

// EscrowContract representa contrato de escrow
type EscrowContract struct {
	*SmartContract
	Buyer      string `json:"buyer"`
	Seller     string `json:"seller"`
	Arbitrator string `json:"arbitrator"`
	ItemHash   string `json:"item_hash"`
}

// VestingContract representa contrato de vesting
type VestingContract struct {
	*SmartContract
	Beneficiary   string `json:"beneficiary"`
	TotalAmount   int64  `json:"total_amount"`
	VestedAmount  int64  `json:"vested_amount"`
	CliffTime     int64  `json:"cliff_time"`
	VestingPeriod int64  `json:"vesting_period"`
	ReleasePeriod int64  `json:"release_period"`
}

// ConditionalContract representa contrato condicional
type ConditionalContract struct {
	*SmartContract
	Condition string `json:"condition"`
	IfTrue    string `json:"if_true"`
	IfFalse   string `json:"if_false"`
}

// ContractManager gerencia contratos inteligentes
type ContractManager struct {
	Contracts map[string]*SmartContract `json:"contracts"`
	GasPrice  int64                     `json:"gas_price"`
	GasLimit  int64                     `json:"gas_limit"`
}

// NewContractManager cria novo gerenciador de contratos
func NewContractManager() *ContractManager {
	return &ContractManager{
		Contracts: make(map[string]*SmartContract),
		GasPrice:  1,       // 1 token por gas
		GasLimit:  1000000, // 1M gas
	}
}

// CreateTimelockContract cria contrato timelock
func (cm *ContractManager) CreateTimelockContract(creator, recipient string, amount int64, unlockTime int64) (*TimelockContract, error) {
	contract := &SmartContract{
		ID:           generateContractID(),
		Type:         TimelockContract,
		Status:       ContractPending,
		Creator:      creator,
		Participants: []string{creator, recipient},
		Amount:       amount,
		Conditions: map[string]interface{}{
			"unlock_time": unlockTime,
			"recipient":   recipient,
		},
		CreatedAt:  time.Now().Unix(),
		ExpiresAt:  unlockTime + 86400, // 24h após unlock
		Signatures: []Signature{},
		Code:       generateTimelockCode(),
		Data:       make(map[string]interface{}),
		GasUsed:    0,
		GasLimit:   cm.GasLimit,
		GasPrice:   cm.GasPrice,
	}

	timelock := &TimelockContract{
		SmartContract: contract,
		UnlockTime:    unlockTime,
		Recipient:     recipient,
	}

	cm.Contracts[contract.ID] = contract
	return timelock, nil
}

// CreateMultisigContract cria contrato multi-signature
func (cm *ContractManager) CreateMultisigContract(creator string, signers []string, requiredSignatures int, amount int64) (*MultisigContract, error) {
	if requiredSignatures > len(signers) {
		return nil, fmt.Errorf("signaturas requeridas maior que número de signatários")
	}

	contract := &SmartContract{
		ID:           generateContractID(),
		Type:         MultisigContract,
		Status:       ContractPending,
		Creator:      creator,
		Participants: append([]string{creator}, signers...),
		Amount:       amount,
		Conditions: map[string]interface{}{
			"required_signatures": requiredSignatures,
			"signers":             signers,
		},
		CreatedAt:  time.Now().Unix(),
		Signatures: []Signature{},
		Code:       generateMultisigCode(),
		Data:       make(map[string]interface{}),
		GasUsed:    0,
		GasLimit:   cm.GasLimit,
		GasPrice:   cm.GasPrice,
	}

	multisig := &MultisigContract{
		SmartContract:      contract,
		RequiredSignatures: requiredSignatures,
		Signers:            signers,
	}

	cm.Contracts[contract.ID] = contract
	return multisig, nil
}

// CreateEscrowContract cria contrato de escrow
func (cm *ContractManager) CreateEscrowContract(buyer, seller, arbitrator string, amount int64, itemHash string) (*EscrowContract, error) {
	contract := &SmartContract{
		ID:           generateContractID(),
		Type:         EscrowContract,
		Status:       ContractPending,
		Creator:      buyer,
		Participants: []string{buyer, seller, arbitrator},
		Amount:       amount,
		Conditions: map[string]interface{}{
			"buyer":      buyer,
			"seller":     seller,
			"arbitrator": arbitrator,
			"item_hash":  itemHash,
		},
		CreatedAt:  time.Now().Unix(),
		ExpiresAt:  time.Now().AddDate(0, 1, 0).Unix(), // 1 mês
		Signatures: []Signature{},
		Code:       generateEscrowCode(),
		Data:       make(map[string]interface{}),
		GasUsed:    0,
		GasLimit:   cm.GasLimit,
		GasPrice:   cm.GasPrice,
	}

	escrow := &EscrowContract{
		SmartContract: contract,
		Buyer:         buyer,
		Seller:        seller,
		Arbitrator:    arbitrator,
		ItemHash:      itemHash,
	}

	cm.Contracts[contract.ID] = contract
	return escrow, nil
}

// CreateVestingContract cria contrato de vesting
func (cm *ContractManager) CreateVestingContract(creator, beneficiary string, totalAmount int64, cliffTime, vestingPeriod, releasePeriod int64) (*VestingContract, error) {
	contract := &SmartContract{
		ID:           generateContractID(),
		Type:         VestingContract,
		Status:       ContractActive,
		Creator:      creator,
		Participants: []string{creator, beneficiary},
		Amount:       totalAmount,
		Conditions: map[string]interface{}{
			"beneficiary":    beneficiary,
			"total_amount":   totalAmount,
			"cliff_time":     cliffTime,
			"vesting_period": vestingPeriod,
			"release_period": releasePeriod,
		},
		CreatedAt:  time.Now().Unix(),
		Signatures: []Signature{},
		Code:       generateVestingCode(),
		Data:       make(map[string]interface{}),
		GasUsed:    0,
		GasLimit:   cm.GasLimit,
		GasPrice:   cm.GasPrice,
	}

	vesting := &VestingContract{
		SmartContract: contract,
		Beneficiary:   beneficiary,
		TotalAmount:   totalAmount,
		VestedAmount:  0,
		CliffTime:     cliffTime,
		VestingPeriod: vestingPeriod,
		ReleasePeriod: releasePeriod,
	}

	cm.Contracts[contract.ID] = contract
	return vesting, nil
}

// CreateConditionalContract cria contrato condicional
func (cm *ContractManager) CreateConditionalContract(creator string, condition, ifTrue, ifFalse string, amount int64) (*ConditionalContract, error) {
	contract := &SmartContract{
		ID:           generateContractID(),
		Type:         ConditionalContract,
		Status:       ContractPending,
		Creator:      creator,
		Participants: []string{creator},
		Amount:       amount,
		Conditions: map[string]interface{}{
			"condition": condition,
			"if_true":   ifTrue,
			"if_false":  ifFalse,
		},
		CreatedAt:  time.Now().Unix(),
		ExpiresAt:  time.Now().AddDate(0, 0, 7).Unix(), // 7 dias
		Signatures: []Signature{},
		Code:       generateConditionalCode(),
		Data:       make(map[string]interface{}),
		GasUsed:    0,
		GasLimit:   cm.GasLimit,
		GasPrice:   cm.GasPrice,
	}

	conditional := &ConditionalContract{
		SmartContract: contract,
		Condition:     condition,
		IfTrue:        ifTrue,
		IfFalse:       ifFalse,
	}

	cm.Contracts[contract.ID] = contract
	return conditional, nil
}

// ExecuteContract executa um contrato
func (cm *ContractManager) ExecuteContract(contractID string, executor string, params map[string]interface{}) error {
	contract, exists := cm.Contracts[contractID]
	if !exists {
		return fmt.Errorf("contrato não encontrado: %s", contractID)
	}

	// Verificar se pode executar
	if !cm.canExecute(contract, executor) {
		return fmt.Errorf("executor não autorizado: %s", executor)
	}

	// Executar baseado no tipo
	switch contract.Type {
	case TimelockContract:
		return cm.executeTimelock(contract, executor, params)
	case MultisigContract:
		return cm.executeMultisig(contract, executor, params)
	case EscrowContract:
		return cm.executeEscrow(contract, executor, params)
	case VestingContract:
		return cm.executeVesting(contract, executor, params)
	case ConditionalContract:
		return cm.executeConditional(contract, executor, params)
	default:
		return fmt.Errorf("tipo de contrato não suportado: %s", contract.Type)
	}
}

// SignContract assina um contrato
func (cm *ContractManager) SignContract(contractID string, signer string, signature string) error {
	contract, exists := cm.Contracts[contractID]
	if !exists {
		return fmt.Errorf("contrato não encontrado: %s", contractID)
	}

	// Verificar se signer é participante
	if !cm.isParticipant(contract, signer) {
		return fmt.Errorf("signer não é participante: %s", signer)
	}

	// Verificar se já assinou
	for _, sig := range contract.Signatures {
		if sig.Signer == signer {
			return fmt.Errorf("signer já assinou: %s", signer)
		}
	}

	// Adicionar assinatura
	contract.Signatures = append(contract.Signatures, Signature{
		Signer:    signer,
		Signature: signature,
		Timestamp: time.Now().Unix(),
	})

	// Verificar se pode executar
	if cm.canExecute(contract, signer) {
		contract.Status = ContractActive
	}

	return nil
}

// GetContract retorna contrato por ID
func (cm *ContractManager) GetContract(contractID string) (*SmartContract, bool) {
	contract, exists := cm.Contracts[contractID]
	return contract, exists
}

// GetContractsByType retorna contratos por tipo
func (cm *ContractManager) GetContractsByType(contractType ContractType) []*SmartContract {
	var contracts []*SmartContract
	for _, contract := range cm.Contracts {
		if contract.Type == contractType {
			contracts = append(contracts, contract)
		}
	}
	return contracts
}

// GetContractsByParticipant retorna contratos de um participante
func (cm *ContractManager) GetContractsByParticipant(participant string) []*SmartContract {
	var contracts []*SmartContract
	for _, contract := range cm.Contracts {
		if cm.isParticipant(contract, participant) {
			contracts = append(contracts, contract)
		}
	}
	return contracts
}

// canExecute verifica se pode executar contrato
func (cm *ContractManager) canExecute(contract *SmartContract, executor string) bool {
	switch contract.Type {
	case TimelockContract:
		return cm.canExecuteTimelock(contract, executor)
	case MultisigContract:
		return cm.canExecuteMultisig(contract, executor)
	case EscrowContract:
		return cm.canExecuteEscrow(contract, executor)
	case VestingContract:
		return cm.canExecuteVesting(contract, executor)
	case ConditionalContract:
		return cm.canExecuteConditional(contract, executor)
	default:
		return false
	}
}

// canExecuteTimelock verifica se pode executar timelock
func (cm *ContractManager) canExecuteTimelock(contract *SmartContract, executor string) bool {
	// Verificar se é o destinatário
	recipient, ok := contract.Conditions["recipient"].(string)
	if !ok {
		return false
	}

	if executor != recipient {
		return false
	}

	// Verificar se já passou do tempo
	unlockTime, ok := contract.Conditions["unlock_time"].(int64)
	if !ok {
		return false
	}

	return time.Now().Unix() >= unlockTime
}

// canExecuteMultisig verifica se pode executar multisig
func (cm *ContractManager) canExecuteMultisig(contract *SmartContract, executor string) bool {
	// Verificar se é signatário
	signers, ok := contract.Conditions["signers"].([]string)
	if !ok {
		return false
	}

	isSigner := false
	for _, signer := range signers {
		if signer == executor {
			isSigner = true
			break
		}
	}

	if !isSigner {
		return false
	}

	// Verificar se tem assinaturas suficientes
	requiredSignatures, ok := contract.Conditions["required_signatures"].(int)
	if !ok {
		return false
	}

	return len(contract.Signatures) >= requiredSignatures
}

// canExecuteEscrow verifica se pode executar escrow
func (cm *ContractManager) canExecuteEscrow(contract *SmartContract, executor string) bool {
	buyer, _ := contract.Conditions["buyer"].(string)
	seller, _ := contract.Conditions["seller"].(string)
	arbitrator, _ := contract.Conditions["arbitrator"].(string)

	return executor == buyer || executor == seller || executor == arbitrator
}

// canExecuteVesting verifica se pode executar vesting
func (cm *ContractManager) canExecuteVesting(contract *SmartContract, executor string) bool {
	beneficiary, ok := contract.Conditions["beneficiary"].(string)
	if !ok {
		return false
	}

	return executor == beneficiary
}

// canExecuteConditional verifica se pode executar condicional
func (cm *ContractManager) canExecuteConditional(contract *SmartContract, executor string) bool {
	return executor == contract.Creator
}

// isParticipant verifica se é participante
func (cm *ContractManager) isParticipant(contract *SmartContract, participant string) bool {
	for _, p := range contract.Participants {
		if p == participant {
			return true
		}
	}
	return false
}

// executeTimelock executa contrato timelock
func (cm *ContractManager) executeTimelock(contract *SmartContract, executor string, params map[string]interface{}) error {
	contract.Status = ContractExecuted
	contract.ExecutedAt = time.Now().Unix()
	contract.GasUsed = 21000 // Gas básico

	// Transferir tokens para destinatário
	recipient, _ := contract.Conditions["recipient"].(string)
	contract.Data["transfer_to"] = recipient
	contract.Data["amount"] = contract.Amount

	return nil
}

// executeMultisig executa contrato multisig
func (cm *ContractManager) executeMultisig(contract *SmartContract, executor string, params map[string]interface{}) error {
	contract.Status = ContractExecuted
	contract.ExecutedAt = time.Now().Unix()
	contract.GasUsed = 50000 // Gas para multisig

	// Processar transferência
	if recipient, ok := params["recipient"].(string); ok {
		contract.Data["transfer_to"] = recipient
		contract.Data["amount"] = contract.Amount
	}

	return nil
}

// executeEscrow executa contrato escrow
func (cm *ContractManager) executeEscrow(contract *SmartContract, executor string, params map[string]interface{}) error {
	action, ok := params["action"].(string)
	if !ok {
		return fmt.Errorf("ação não especificada")
	}

	contract.GasUsed = 30000 // Gas para escrow

	switch action {
	case "release":
		contract.Status = ContractExecuted
		contract.ExecutedAt = time.Now().Unix()
		seller, _ := contract.Conditions["seller"].(string)
		contract.Data["transfer_to"] = seller
		contract.Data["action"] = "released_to_seller"
	case "refund":
		contract.Status = ContractExecuted
		contract.ExecutedAt = time.Now().Unix()
		buyer, _ := contract.Conditions["buyer"].(string)
		contract.Data["transfer_to"] = buyer
		contract.Data["action"] = "refunded_to_buyer"
	case "dispute":
		contract.Status = ContractActive
		contract.Data["action"] = "disputed"
		contract.Data["disputed_by"] = executor
	default:
		return fmt.Errorf("ação inválida: %s", action)
	}

	return nil
}

// executeVesting executa contrato vesting
func (cm *ContractManager) executeVesting(contract *SmartContract, executor string, params map[string]interface{}) error {
	// Calcular tokens vestidos
	cliffTime, _ := contract.Conditions["cliff_time"].(int64)
	vestingPeriod, _ := contract.Conditions["vesting_period"].(int64)
	totalAmount, _ := contract.Conditions["total_amount"].(int64)

	now := time.Now().Unix()

	if now < cliffTime {
		return fmt.Errorf("cliff ainda não atingido")
	}

	// Calcular tokens vestidos
	elapsed := now - cliffTime
	if elapsed > vestingPeriod {
		elapsed = vestingPeriod
	}

	vestedAmount := (totalAmount * elapsed) / vestingPeriod
	contract.Data["vested_amount"] = vestedAmount
	contract.Data["elapsed_time"] = elapsed

	contract.GasUsed = 25000 // Gas para vesting

	return nil
}

// executeConditional executa contrato condicional
func (cm *ContractManager) executeConditional(contract *SmartContract, executor string, params map[string]interface{}) error {
	condition, _ := contract.Conditions["condition"].(string)
	ifTrue, _ := contract.Conditions["if_true"].(string)
	ifFalse, _ := contract.Conditions["if_false"].(string)

	// Avaliar condição (simplificado)
	result := cm.evaluateCondition(condition, params)

	contract.Status = ContractExecuted
	contract.ExecutedAt = time.Now().Unix()
	contract.GasUsed = 40000 // Gas para condicional

	if result {
		contract.Data["result"] = ifTrue
		contract.Data["condition_met"] = true
	} else {
		contract.Data["result"] = ifFalse
		contract.Data["condition_met"] = false
	}

	return nil
}

// evaluateCondition avalia condição (simplificado)
func (cm *ContractManager) evaluateCondition(condition string, params map[string]interface{}) bool {
	// Implementação simplificada - em produção usar parser de expressões
	if strings.Contains(condition, ">") {
		parts := strings.Split(condition, ">")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			if val, ok := params[left]; ok {
				if num, err := strconv.ParseInt(fmt.Sprintf("%v", val), 10, 64); err == nil {
					if target, err := strconv.ParseInt(right, 10, 64); err == nil {
						return num > target
					}
				}
			}
		}
	}

	return false
}

// generateContractID gera ID único para contrato
func generateContractID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:16])
}

// generateTimelockCode gera código para timelock
func generateTimelockCode() string {
	return `
		contract Timelock {
			uint256 public unlockTime;
			address public recipient;
			
			constructor(uint256 _unlockTime, address _recipient) {
				unlockTime = _unlockTime;
				recipient = _recipient;
			}
			
			function withdraw() public {
				require(block.timestamp >= unlockTime, "Not yet unlocked");
				require(msg.sender == recipient, "Only recipient can withdraw");
				// Transfer logic here
			}
		}
	`
}

// generateMultisigCode gera código para multisig
func generateMultisigCode() string {
	return `
		contract Multisig {
			address[] public signers;
			uint256 public requiredSignatures;
			mapping(address => bool) public hasSigned;
			uint256 public signatureCount;
			
			constructor(address[] memory _signers, uint256 _requiredSignatures) {
				signers = _signers;
				requiredSignatures = _requiredSignatures;
			}
			
			function sign() public {
				require(isSigner(msg.sender), "Not a signer");
				require(!hasSigned[msg.sender], "Already signed");
				hasSigned[msg.sender] = true;
				signatureCount++;
			}
			
			function execute() public {
				require(signatureCount >= requiredSignatures, "Not enough signatures");
				// Execute logic here
			}
		}
	`
}

// generateEscrowCode gera código para escrow
func generateEscrowCode() string {
	return `
		contract Escrow {
			address public buyer;
			address public seller;
			address public arbitrator;
			string public itemHash;
			
			constructor(address _buyer, address _seller, address _arbitrator, string memory _itemHash) {
				buyer = _buyer;
				seller = _seller;
				arbitrator = _arbitrator;
				itemHash = _itemHash;
			}
			
			function release() public {
				require(msg.sender == buyer || msg.sender == arbitrator, "Not authorized");
				// Release to seller
			}
			
			function refund() public {
				require(msg.sender == seller || msg.sender == arbitrator, "Not authorized");
				// Refund to buyer
			}
		}
	`
}

// generateVestingCode gera código para vesting
func generateVestingCode() string {
	return `
		contract Vesting {
			address public beneficiary;
			uint256 public totalAmount;
			uint256 public cliffTime;
			uint256 public vestingPeriod;
			
			constructor(address _beneficiary, uint256 _totalAmount, uint256 _cliffTime, uint256 _vestingPeriod) {
				beneficiary = _beneficiary;
				totalAmount = _totalAmount;
				cliffTime = _cliffTime;
				vestingPeriod = _vestingPeriod;
			}
			
			function claim() public {
				require(msg.sender == beneficiary, "Only beneficiary can claim");
				require(block.timestamp >= cliffTime, "Cliff not reached");
				// Calculate and transfer vested amount
			}
		}
	`
}

// generateConditionalCode gera código para condicional
func generateConditionalCode() string {
	return `
		contract Conditional {
			string public condition;
			string public ifTrue;
			string public ifFalse;
			
			constructor(string memory _condition, string memory _ifTrue, string memory _ifFalse) {
				condition = _condition;
				ifTrue = _ifTrue;
				ifFalse = _ifFalse;
			}
			
			function execute() public {
				// Evaluate condition and execute accordingly
			}
		}
	`
}
