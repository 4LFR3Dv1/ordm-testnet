package backend

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// GlobalDatabase representa o banco de dados global da blockchain
type GlobalDatabase struct {
	mu sync.RWMutex

	// Wallets globais
	Wallets map[string]*GlobalWallet `json:"wallets"`

	// Transações globais
	Transactions map[string]*GlobalTransaction `json:"transactions"`

	// Blocos globais
	Blocks map[string]*GlobalBlock `json:"blocks"`

	// Nodes registrados
	Nodes map[string]*RegisteredNode `json:"nodes"`

	// Estado global da blockchain
	GlobalState *GlobalState `json:"global_state"`

	// Auditoria
	AuditLog []*AuditEntry `json:"audit_log"`
}

// GlobalWallet representa uma wallet no banco global
type GlobalWallet struct {
	Address   string            `json:"address"`
	PublicKey string            `json:"public_key"`
	Balance   int64             `json:"balance"`
	Nonce     uint64            `json:"nonce"`
	CreatedAt time.Time         `json:"created_at"`
	LastUsed  time.Time         `json:"last_used"`
	OwnerNode string            `json:"owner_node"`
	IsActive  bool              `json:"is_active"`
	Metadata  map[string]string `json:"metadata"`
}

// GlobalTransaction representa uma transação global
type GlobalTransaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    int64     `json:"amount"`
	Fee       int64     `json:"fee"`
	Nonce     uint64    `json:"nonce"`
	BlockHash string    `json:"block_hash"`
	Status    string    `json:"status"` // pending, confirmed, failed
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"`
	GasUsed   int64     `json:"gas_used"`
	GasPrice  int64     `json:"gas_price"`
}

// GlobalBlock representa um bloco global
type GlobalBlock struct {
	Hash         string    `json:"hash"`
	ParentHash   string    `json:"parent_hash"`
	Number       uint64    `json:"number"`
	Timestamp    time.Time `json:"timestamp"`
	Miner        string    `json:"miner"`
	Difficulty   uint64    `json:"difficulty"`
	GasLimit     int64     `json:"gas_limit"`
	GasUsed      int64     `json:"gas_used"`
	Transactions []string  `json:"transactions"`
	Reward       int64     `json:"reward"`
	Layer        int       `json:"layer"` // 1 = PoW, 2 = PoS
	IsValid      bool      `json:"is_valid"`
}

// RegisteredNode representa um node registrado
type RegisteredNode struct {
	NodeID      string    `json:"node_id"`
	PublicKey   string    `json:"public_key"`
	Address     string    `json:"address"`
	Port        int       `json:"port"`
	IsActive    bool      `json:"is_active"`
	LastSeen    time.Time `json:"last_seen"`
	StakeAmount int64     `json:"stake_amount"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

// GlobalState representa o estado global da blockchain
type GlobalState struct {
	TotalSupply       int64     `json:"total_supply"`
	CirculatingSupply int64     `json:"circulating_supply"`
	BurnedSupply      int64     `json:"burned_supply"`
	ActiveNodes       int       `json:"active_nodes"`
	TotalBlocks       uint64    `json:"total_blocks"`
	TotalTransactions int64     `json:"total_transactions"`
	CurrentDifficulty uint64    `json:"current_difficulty"`
	LastBlockHash     string    `json:"last_block_hash"`
	LastBlockNumber   uint64    `json:"last_block_number"`
	LastUpdate        time.Time `json:"last_update"`
}

// AuditEntry representa uma entrada de auditoria
type AuditEntry struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Action       string                 `json:"action"`
	NodeID       string                 `json:"node_id"`
	Details      map[string]interface{} `json:"details"`
	Hash         string                 `json:"hash"`
	PreviousHash string                 `json:"previous_hash"`
}

// NewGlobalDatabase cria nova instância do banco global
func NewGlobalDatabase() *GlobalDatabase {
	return &GlobalDatabase{
		Wallets:      make(map[string]*GlobalWallet),
		Transactions: make(map[string]*GlobalTransaction),
		Blocks:       make(map[string]*GlobalBlock),
		Nodes:        make(map[string]*RegisteredNode),
		GlobalState: &GlobalState{
			TotalSupply:       21000000, // 21M tokens
			CirculatingSupply: 0,
			BurnedSupply:      0,
			ActiveNodes:       0,
			TotalBlocks:       0,
			TotalTransactions: 0,
			CurrentDifficulty: 1,
			LastUpdate:        time.Now(),
		},
		AuditLog: make([]*AuditEntry, 0),
	}
}

// RegisterWallet registra uma wallet no banco global
func (db *GlobalDatabase) RegisterWallet(wallet *GlobalWallet) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Verificar se wallet já existe
	if _, exists := db.Wallets[wallet.Address]; exists {
		return fmt.Errorf("wallet já registrada: %s", wallet.Address)
	}

	// Registrar wallet
	db.Wallets[wallet.Address] = wallet

	// Atualizar estado global
	db.GlobalState.CirculatingSupply += wallet.Balance
	db.GlobalState.LastUpdate = time.Now()

	// Registrar auditoria
	db.addAuditEntry("wallet_registered", wallet.OwnerNode, map[string]interface{}{
		"address":    wallet.Address,
		"balance":    wallet.Balance,
		"owner_node": wallet.OwnerNode,
	})

	return nil
}

// UpdateWalletBalance atualiza saldo de uma wallet
func (db *GlobalDatabase) UpdateWalletBalance(address string, newBalance int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	wallet, exists := db.Wallets[address]
	if !exists {
		return fmt.Errorf("wallet não encontrada: %s", address)
	}

	oldBalance := wallet.Balance
	wallet.Balance = newBalance
	wallet.LastUsed = time.Now()

	// Atualizar estado global
	db.GlobalState.CirculatingSupply += (newBalance - oldBalance)
	db.GlobalState.LastUpdate = time.Now()

	// Registrar auditoria
	db.addAuditEntry("balance_updated", wallet.OwnerNode, map[string]interface{}{
		"address":     address,
		"old_balance": oldBalance,
		"new_balance": newBalance,
		"difference":  newBalance - oldBalance,
	})

	return nil
}

// RegisterTransaction registra uma transação no banco global
func (db *GlobalDatabase) RegisterTransaction(tx *GlobalTransaction) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Verificar se transação já existe
	if _, exists := db.Transactions[tx.Hash]; exists {
		return fmt.Errorf("transação já registrada: %s", tx.Hash)
	}

	// Registrar transação
	db.Transactions[tx.Hash] = tx
	db.GlobalState.TotalTransactions++
	db.GlobalState.LastUpdate = time.Now()

	// Registrar auditoria
	db.addAuditEntry("transaction_registered", "", map[string]interface{}{
		"hash":   tx.Hash,
		"from":   tx.From,
		"to":     tx.To,
		"amount": tx.Amount,
		"status": tx.Status,
	})

	return nil
}

// RegisterBlock registra um bloco no banco global
func (db *GlobalDatabase) RegisterBlock(block *GlobalBlock) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Verificar se bloco já existe
	if _, exists := db.Blocks[block.Hash]; exists {
		return fmt.Errorf("bloco já registrado: %s", block.Hash)
	}

	// Registrar bloco
	db.Blocks[block.Hash] = block
	db.GlobalState.TotalBlocks++
	db.GlobalState.LastBlockHash = block.Hash
	db.GlobalState.LastBlockNumber = block.Number
	db.GlobalState.LastUpdate = time.Now()

	// Atualizar supply se for bloco de mineração
	if block.Layer == 1 && block.Reward > 0 {
		db.GlobalState.CirculatingSupply += block.Reward
	}

	// Registrar auditoria
	db.addAuditEntry("block_registered", block.Miner, map[string]interface{}{
		"hash":               block.Hash,
		"number":             block.Number,
		"miner":              block.Miner,
		"layer":              block.Layer,
		"reward":             block.Reward,
		"transactions_count": len(block.Transactions),
	})

	return nil
}

// RegisterNode registra um node no banco global
func (db *GlobalDatabase) RegisterNode(node *RegisteredNode) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Verificar se node já existe
	if _, exists := db.Nodes[node.NodeID]; exists {
		return fmt.Errorf("node já registrado: %s", node.NodeID)
	}

	// Registrar node
	db.Nodes[node.NodeID] = node
	db.GlobalState.ActiveNodes++
	db.GlobalState.LastUpdate = time.Now()

	// Registrar auditoria
	db.addAuditEntry("node_registered", node.NodeID, map[string]interface{}{
		"node_id":      node.NodeID,
		"address":      node.Address,
		"port":         node.Port,
		"stake_amount": node.StakeAmount,
	})

	return nil
}

// GetWallet retorna uma wallet do banco global
func (db *GlobalDatabase) GetWallet(address string) (*GlobalWallet, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	wallet, exists := db.Wallets[address]
	return wallet, exists
}

// GetTransaction retorna uma transação do banco global
func (db *GlobalDatabase) GetTransaction(hash string) (*GlobalTransaction, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	tx, exists := db.Transactions[hash]
	return tx, exists
}

// GetBlock retorna um bloco do banco global
func (db *GlobalDatabase) GetBlock(hash string) (*GlobalBlock, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	block, exists := db.Blocks[hash]
	return block, exists
}

// GetNode retorna um node do banco global
func (db *GlobalDatabase) GetNode(nodeID string) (*RegisteredNode, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	node, exists := db.Nodes[nodeID]
	return node, exists
}

// GetAllWallets retorna todas as wallets
func (db *GlobalDatabase) GetAllWallets() map[string]*GlobalWallet {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(map[string]*GlobalWallet)
	for addr, wallet := range db.Wallets {
		result[addr] = wallet
	}
	return result
}

// GetAllTransactions retorna todas as transações
func (db *GlobalDatabase) GetAllTransactions() map[string]*GlobalTransaction {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(map[string]*GlobalTransaction)
	for hash, tx := range db.Transactions {
		result[hash] = tx
	}
	return result
}

// GetAllBlocks retorna todos os blocos
func (db *GlobalDatabase) GetAllBlocks() map[string]*GlobalBlock {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(map[string]*GlobalBlock)
	for hash, block := range db.Blocks {
		result[hash] = block
	}
	return result
}

// GetAllNodes retorna todos os nodes
func (db *GlobalDatabase) GetAllNodes() map[string]*RegisteredNode {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(map[string]*RegisteredNode)
	for nodeID, node := range db.Nodes {
		result[nodeID] = node
	}
	return result
}

// GetGlobalState retorna o estado global
func (db *GlobalDatabase) GetGlobalState() *GlobalState {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.GlobalState
}

// GetAuditLog retorna o log de auditoria
func (db *GlobalDatabase) GetAuditLog() []*AuditEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make([]*AuditEntry, len(db.AuditLog))
	copy(result, db.AuditLog)
	return result
}

// addAuditEntry adiciona entrada de auditoria
func (db *GlobalDatabase) addAuditEntry(action, nodeID string, details map[string]interface{}) {
	entry := &AuditEntry{
		ID:        fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Action:    action,
		NodeID:    nodeID,
		Details:   details,
		Hash:      fmt.Sprintf("hash_%d", time.Now().UnixNano()),
	}

	if len(db.AuditLog) > 0 {
		entry.PreviousHash = db.AuditLog[len(db.AuditLog)-1].Hash
	}

	db.AuditLog = append(db.AuditLog, entry)
}

// ExportData exporta dados para JSON
func (db *GlobalDatabase) ExportData() ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return json.MarshalIndent(db, "", "  ")
}

// ImportData importa dados de JSON
func (db *GlobalDatabase) ImportData(data []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return json.Unmarshal(data, db)
}
