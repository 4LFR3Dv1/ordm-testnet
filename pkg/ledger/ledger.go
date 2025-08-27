package ledger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TokenMovement struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee"`
	Type        string `json:"type"` // "transfer", "mining_reward", "stake_reward"
	BlockHash   string `json:"block_hash,omitempty"`
	Timestamp   int64  `json:"timestamp"`
	Transaction string `json:"transaction_hash,omitempty"`
	Description string `json:"description"`
}

type TokenGeneration struct {
	ID          string `json:"id"`
	Address     string `json:"address"`
	Amount      int64  `json:"amount"`
	Type        string `json:"type"` // "mining_reward", "stake_reward"
	BlockHash   string `json:"block_hash"`
	Timestamp   int64  `json:"timestamp"`
	Description string `json:"description"`
}

// Transaction representa uma transação
type Transaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	Nonce     uint64 `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
}

type GlobalLedger struct {
	Balances      map[string]int64  `json:"balances"`
	Movements     []TokenMovement   `json:"movements"`
	Generations   []TokenGeneration `json:"generations"`
	TotalSupply   int64             `json:"total_supply"`
	Mutex         sync.RWMutex      `json:"-"`
	DataPath      string            `json:"-"`
	WalletManager interface{}       `json:"-"`
}

func NewGlobalLedger(dataPath string, walletManager interface{}) *GlobalLedger {
	return &GlobalLedger{
		Balances:      make(map[string]int64),
		Movements:     []TokenMovement{},
		Generations:   []TokenGeneration{},
		TotalSupply:   0,
		DataPath:      dataPath,
		WalletManager: walletManager,
	}
}

// LoadLedger carrega o ledger do disco
func (gl *GlobalLedger) LoadLedger() error {
	ledgerPath := filepath.Join(gl.DataPath, "global_ledger.json")
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Criar ledger vazio se não existir
			return gl.createEmptyLedger()
		}
		return fmt.Errorf("erro ao ler ledger: %v", err)
	}

	var ledgerData struct {
		Balances    map[string]int64  `json:"balances"`
		Movements   []TokenMovement   `json:"movements"`
		Generations []TokenGeneration `json:"generations"`
		TotalSupply int64             `json:"total_supply"`
	}

	err = json.Unmarshal(data, &ledgerData)
	if err != nil {
		return fmt.Errorf("erro ao decodificar ledger: %v", err)
	}

	gl.Mutex.Lock()
	gl.Balances = ledgerData.Balances
	gl.Movements = ledgerData.Movements
	gl.Generations = ledgerData.Generations
	gl.TotalSupply = ledgerData.TotalSupply
	gl.Mutex.Unlock()

	return nil
}

// createEmptyLedger cria um ledger vazio sem usar locks
func (gl *GlobalLedger) createEmptyLedger() error {
	// Criar diretório se não existir
	err := os.MkdirAll(gl.DataPath, 0755)
	if err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	ledgerData := struct {
		Balances    map[string]int64  `json:"balances"`
		Movements   []TokenMovement   `json:"movements"`
		Generations []TokenGeneration `json:"generations"`
		TotalSupply int64             `json:"total_supply"`
	}{
		Balances:    make(map[string]int64),
		Movements:   []TokenMovement{},
		Generations: []TokenGeneration{},
		TotalSupply: 0,
	}

	data, err := json.MarshalIndent(ledgerData, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao codificar ledger: %v", err)
	}

	ledgerPath := filepath.Join(gl.DataPath, "global_ledger.json")
	err = os.WriteFile(ledgerPath, data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar ledger: %v", err)
	}

	return nil
}

// SaveLedger salva o ledger no disco
func (gl *GlobalLedger) SaveLedger() error {
	gl.Mutex.RLock()

	// Criar uma cópia dos dados para evitar deadlock
	ledgerData := struct {
		Balances    map[string]int64  `json:"balances"`
		Movements   []TokenMovement   `json:"movements"`
		Generations []TokenGeneration `json:"generations"`
		TotalSupply int64             `json:"total_supply"`
	}{
		Balances:    make(map[string]int64),
		Movements:   make([]TokenMovement, len(gl.Movements)),
		Generations: make([]TokenGeneration, len(gl.Generations)),
		TotalSupply: gl.TotalSupply,
	}

	// Copiar dados
	for k, v := range gl.Balances {
		ledgerData.Balances[k] = v
	}
	copy(ledgerData.Movements, gl.Movements)
	copy(ledgerData.Generations, gl.Generations)

	gl.Mutex.RUnlock()

	// Criar diretório se não existir
	err := os.MkdirAll(gl.DataPath, 0755)
	if err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	data, err := json.MarshalIndent(ledgerData, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao codificar ledger: %v", err)
	}

	ledgerPath := filepath.Join(gl.DataPath, "global_ledger.json")
	err = os.WriteFile(ledgerPath, data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar ledger: %v", err)
	}

	return nil
}

// GetBalance retorna o saldo de um endereço
func (gl *GlobalLedger) GetBalance(address string) int64 {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	balance, exists := gl.Balances[address]
	if !exists {
		return 0
	}
	return balance
}

// UpdateBalance atualiza o saldo de um endereço
func (gl *GlobalLedger) UpdateBalance(address string, newBalance int64) error {
	gl.Mutex.Lock()
	gl.Balances[address] = newBalance
	gl.Mutex.Unlock()

	// Salvar ledger sem locks para evitar deadlock
	return gl.SaveLedger()
}

// ProcessTransaction processa uma transação e atualiza os saldos
func (gl *GlobalLedger) ProcessTransaction(tx *Transaction, blockHash string) error {
	gl.Mutex.Lock()

	// Verificar saldo do remetente
	fromBalance := gl.Balances[tx.From]
	if fromBalance < tx.Amount+tx.Fee {
		gl.Mutex.Unlock()
		return fmt.Errorf("saldo insuficiente: %d < %d", fromBalance, tx.Amount+tx.Fee)
	}

	// Atualizar saldos
	gl.Balances[tx.From] -= (tx.Amount + tx.Fee)
	gl.Balances[tx.To] += tx.Amount

	// Registrar movimento
	movement := TokenMovement{
		ID:          fmt.Sprintf("mov_%d", time.Now().UnixNano()),
		From:        tx.From,
		To:          tx.To,
		Amount:      tx.Amount,
		Fee:         tx.Fee,
		Type:        "transfer",
		BlockHash:   blockHash,
		Timestamp:   tx.Timestamp,
		Transaction: tx.Hash,
		Description: fmt.Sprintf("Transferência de %d tokens de %s para %s", tx.Amount, tx.From, tx.To),
	}

	gl.Movements = append(gl.Movements, movement)

	gl.Mutex.Unlock()

	// Salvar ledger sem locks para evitar deadlock
	return gl.SaveLedger()
}

// AddMiningReward adiciona recompensa de mineração
func (gl *GlobalLedger) AddMiningReward(address string, amount int64, blockHash string) error {
	gl.Mutex.Lock()

	// Atualizar saldo
	gl.Balances[address] += amount
	gl.TotalSupply += amount

	// Registrar geração de tokens
	generation := TokenGeneration{
		ID:          fmt.Sprintf("gen_%d", time.Now().UnixNano()),
		Address:     address,
		Amount:      amount,
		Type:        "mining_reward",
		BlockHash:   blockHash,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("Recompensa de mineração de %d tokens para %s", amount, address),
	}

	gl.Generations = append(gl.Generations, generation)

	// Registrar movimento
	movement := TokenMovement{
		ID:          fmt.Sprintf("mov_%d", time.Now().UnixNano()),
		From:        "system",
		To:          address,
		Amount:      amount,
		Fee:         0,
		Type:        "mining_reward",
		BlockHash:   blockHash,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("Recompensa de mineração de %d tokens para %s", amount, address),
	}

	gl.Movements = append(gl.Movements, movement)

	gl.Mutex.Unlock()

	// Salvar ledger sem locks para evitar deadlock
	return gl.SaveLedger()
}

// ProcessStake processa um stake (deduz tokens do saldo)
func (gl *GlobalLedger) ProcessStake(address string, amount int64, stakeHash string) error {
	gl.Mutex.Lock()

	// Verificar saldo suficiente
	currentBalance := gl.Balances[address]
	if currentBalance < amount {
		gl.Mutex.Unlock()
		return fmt.Errorf("saldo insuficiente para stake: %d < %d", currentBalance, amount)
	}

	// Deduzir tokens do saldo (stake é bloqueado)
	gl.Balances[address] -= amount

	// Registrar movimento de stake
	movement := TokenMovement{
		ID:          fmt.Sprintf("stake_%d", time.Now().UnixNano()),
		From:        address,
		To:          "stake_pool",
		Amount:      amount,
		Fee:         0,
		Type:        "stake",
		BlockHash:   stakeHash,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("Stake de %d tokens de %s", amount, address),
	}

	gl.Movements = append(gl.Movements, movement)

	gl.Mutex.Unlock()

	// Salvar ledger sem locks para evitar deadlock
	return gl.SaveLedger()
}

// AddStakeReward adiciona recompensa de stake
func (gl *GlobalLedger) AddStakeReward(address string, amount int64, blockHash string) error {
	gl.Mutex.Lock()

	// Atualizar saldo
	gl.Balances[address] += amount
	gl.TotalSupply += amount

	// Registrar geração de tokens
	generation := TokenGeneration{
		ID:          fmt.Sprintf("gen_%d", time.Now().UnixNano()),
		Address:     address,
		Amount:      amount,
		Type:        "stake_reward",
		BlockHash:   blockHash,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("Recompensa de stake de %d tokens para %s", amount, address),
	}

	gl.Generations = append(gl.Generations, generation)

	// Registrar movimento
	movement := TokenMovement{
		ID:          fmt.Sprintf("mov_%d", time.Now().UnixNano()),
		From:        "system",
		To:          address,
		Amount:      amount,
		Fee:         0,
		Type:        "stake_reward",
		BlockHash:   blockHash,
		Timestamp:   time.Now().Unix(),
		Description: fmt.Sprintf("Recompensa de stake de %d tokens para %s", amount, address),
	}

	gl.Movements = append(gl.Movements, movement)

	gl.Mutex.Unlock()

	// Salvar ledger sem locks para evitar deadlock
	return gl.SaveLedger()
}

// GetMovements retorna movimentações de um endereço
func (gl *GlobalLedger) GetMovements(address string) []TokenMovement {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	var movements []TokenMovement
	for _, movement := range gl.Movements {
		if movement.From == address || movement.To == address {
			movements = append(movements, movement)
		}
	}
	return movements
}

// GetGenerations retorna gerações de tokens de um endereço
func (gl *GlobalLedger) GetGenerations(address string) []TokenGeneration {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	var generations []TokenGeneration
	for _, generation := range gl.Generations {
		if generation.Address == address {
			generations = append(generations, generation)
		}
	}
	return generations
}

// GetAllMovements retorna todas as movimentações
func (gl *GlobalLedger) GetAllMovements() []TokenMovement {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	return gl.Movements
}

// GetAllGenerations retorna todas as gerações
func (gl *GlobalLedger) GetAllGenerations() []TokenGeneration {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	return gl.Generations
}

// GetTotalSupply retorna o supply total
func (gl *GlobalLedger) GetTotalSupply() int64 {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	return gl.TotalSupply
}

// GetStats retorna estatísticas do ledger
func (gl *GlobalLedger) GetStats() map[string]interface{} {
	gl.Mutex.RLock()
	defer gl.Mutex.RUnlock()

	stats := map[string]interface{}{
		"total_supply":      gl.TotalSupply,
		"total_addresses":   len(gl.Balances),
		"total_movements":   len(gl.Movements),
		"total_generations": len(gl.Generations),
	}

	// Calcular top holders
	type holder struct {
		Address string
		Balance int64
	}

	var holders []holder
	for address, balance := range gl.Balances {
		holders = append(holders, holder{Address: address, Balance: balance})
	}

	// Ordenar por saldo (simplificado)
	if len(holders) > 0 {
		// Encontrar maior saldo
		maxBalance := int64(0)
		for _, h := range holders {
			if h.Balance > maxBalance {
				maxBalance = h.Balance
			}
		}
		stats["max_balance"] = maxBalance
	}

	return stats
}

// AddTransfer registra uma transferência entre wallets
func (gl *GlobalLedger) AddTransfer(from string, to string, amount int64, description string) error {
	gl.Mutex.Lock()

	// Verificar se o remetente tem saldo suficiente
	if gl.Balances[from] < amount {
		gl.Mutex.Unlock()
		return fmt.Errorf("saldo insuficiente para transferência: %d < %d", gl.Balances[from], amount)
	}

	// Deduzir do remetente
	gl.Balances[from] -= amount

	// Adicionar ao destinatário
	gl.Balances[to] += amount

	// Registrar movimento de transferência
	movement := TokenMovement{
		ID:          fmt.Sprintf("mov_%d", time.Now().UnixNano()),
		From:        from,
		To:          to,
		Amount:      amount,
		Fee:         0,
		Type:        "transfer",
		BlockHash:   fmt.Sprintf("transfer_%d", time.Now().Unix()),
		Timestamp:   time.Now().Unix(),
		Description: description,
	}

	gl.Movements = append(gl.Movements, movement)

	gl.Mutex.Unlock()

	// Salvar ledger sem locks para evitar deadlock
	return gl.SaveLedger()
}
