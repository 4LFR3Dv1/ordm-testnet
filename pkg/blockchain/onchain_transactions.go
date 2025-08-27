package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// OnChainTransaction representa uma transação on-chain
type OnChainTransaction struct {
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Amount      int64     `json:"amount"`
	Fee         int64     `json:"fee"`
	Type        string    `json:"type"` // "transfer", "stake", "unstake", "mining_reward"
	BlockHash   string    `json:"block_hash"`
	Timestamp   int64     `json:"timestamp"`
	Signature   string    `json:"signature"`
	Status      string    `json:"status"` // "pending", "confirmed", "rejected"
	ValidatorID string    `json:"validator_id,omitempty"`
}

// OnChainLedger gerencia transações on-chain
type OnChainLedger struct {
	mu           sync.RWMutex
	transactions map[string]*OnChainTransaction
	blocks       map[string]*Block
	walletManager interface{} // SecureWalletManager
	genesisBlock *Block
	blockHeight  int64
}

// NewOnChainLedger cria um novo ledger on-chain
func NewOnChainLedger(walletManager interface{}) *OnChainLedger {
	return &OnChainLedger{
		transactions: make(map[string]*OnChainTransaction),
		blocks:       make(map[string]*Block),
		walletManager: walletManager,
		blockHeight:  0,
	}
}

// CreateGenesisBlock cria o bloco genesis
func (ocl *OnChainLedger) CreateGenesisBlock() *Block {
	ocl.mu.Lock()
	defer ocl.mu.Unlock()

	genesis := &Block{
		Hash:         "0000000000000000000000000000000000000000000000000000000000000000",
		ParentHash:   "",
		Number:       0,
		Timestamp:    time.Now().Unix(),
		MinerID:      "genesis",
		Reward:       0,
		Difficulty:   0,
		Nonce:        0,
		Transactions: []Transaction{},
		TotalFees:    0,
		BurnedTokens: 0,
	}

	ocl.genesisBlock = genesis
	ocl.blocks[genesis.Hash] = genesis

	fmt.Printf("🌱 Bloco Genesis criado: %s\n", genesis.Hash)
	return genesis
}

// CreateTransaction cria uma nova transação
func (ocl *OnChainLedger) CreateTransaction(from, to string, amount, fee int64, txType string) (*OnChainTransaction, error) {
	// Validar se o remetente tem stake suficiente ou é validador
	if !ocl.canTransact(from, amount+fee) {
		return nil, fmt.Errorf("remetente não tem stake suficiente ou não é validador")
	}

	tx := &OnChainTransaction{
		ID:        ocl.generateTxID(from, to, amount, time.Now().Unix()),
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Type:      txType,
		Timestamp: time.Now().Unix(),
		Status:    "pending",
	}

	ocl.mu.Lock()
	ocl.transactions[tx.ID] = tx
	ocl.mu.Unlock()

	fmt.Printf("📝 Transação criada: %s (%s -> %s: %d)\n", tx.ID[:8], from, to, amount)
	return tx, nil
}

// canTransact verifica se uma wallet pode transacionar
func (ocl *OnChainLedger) canTransact(address string, amount int64) bool {
	// Buscar wallet
	wallet, exists := ocl.getWalletByAddress(address)
	if !exists {
		return false
	}

	// Verificar se é validador usando interface
	if w, ok := wallet.(interface{ IsValidator() bool }); ok && w.IsValidator() {
		return true
	}

	// Verificar se tem stake suficiente usando interface
	if w, ok := wallet.(interface{ GetStakeAmount() int64 }); ok && w.GetStakeAmount() >= amount {
		return true
	}

	return false
}

// getWalletByAddress busca wallet pelo endereço
func (ocl *OnChainLedger) getWalletByAddress(address string) (interface{}, bool) {
	// Interface para acessar o wallet manager
	if wm, ok := ocl.walletManager.(interface {
		GetWalletByAddress(string) (interface{}, bool)
	}); ok {
		return wm.GetWalletByAddress(address)
	}
	return nil, false
}

// generateTxID gera ID único para transação
func (ocl *OnChainLedger) generateTxID(from, to string, amount int64, timestamp int64) string {
	data := fmt.Sprintf("%s|%s|%d|%d", from, to, amount, timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// MineBlock minera um bloco com transações pendentes
func (ocl *OnChainLedger) MineBlock(minerID string, difficulty int) (*Block, error) {
	ocl.mu.Lock()
	defer ocl.mu.Unlock()

	// Buscar transações pendentes
	pendingTxs := ocl.getPendingTransactions()
	if len(pendingTxs) == 0 {
		return nil, fmt.Errorf("nenhuma transação pendente")
	}

	// Converter para Transaction
	txs := make([]Transaction, 0, len(pendingTxs))
	for _, tx := range pendingTxs {
		txs = append(txs, Transaction{
			ID:     tx.ID,
			From:   tx.From,
			To:     tx.To,
			Amount: tx.Amount,
			Fee:    tx.Fee,
			Data:   tx.Type,
		})
	}

	// Obter hash do bloco pai
	parentHash := ocl.genesisBlock.Hash
	if ocl.blockHeight > 0 {
		if latest := ocl.getLatestBlock(); latest != nil {
			parentHash = latest.Hash
		}
	}

	// Criar bloco
	block := &Block{
		ParentHash:   parentHash,
		Number:       ocl.blockHeight + 1,
		Timestamp:    time.Now().Unix(),
		MinerID:      minerID,
		Transactions: txs,
		Difficulty:   difficulty,
	}

	// Mineração PoW
	nonce := uint64(0)
	for {
		block.Nonce = nonce
		block.Hash = ocl.calculateBlockHash(block)
		
		if ocl.verifyPoW(block.Hash, difficulty) {
			break
		}
		nonce++
	}

	// Calcular recompensa e taxas
	block.Reward = ocl.calculateMiningReward(block.Number)
	totalFees := int64(0)
	for _, tx := range txs {
		totalFees += tx.Fee
	}
	block.TotalFees = totalFees
	block.BurnedTokens = int64(float64(totalFees) * 0.1) // 10% queimados

	// Adicionar bloco
	ocl.blocks[block.Hash] = block
	ocl.blockHeight = block.Number

	// Processar transações
	ocl.processBlockTransactions(block)

	fmt.Printf("⛏️ Bloco #%d minerado: %s | Recompensa: %d | Taxas: %d | Queimados: %d\n",
		block.Number, block.Hash[:8], block.Reward, totalFees, block.BurnedTokens)

	return block, nil
}

// getPendingTransactions retorna transações pendentes
func (ocl *OnChainLedger) getPendingTransactions() []*OnChainTransaction {
	var pending []*OnChainTransaction
	for _, tx := range ocl.transactions {
		if tx.Status == "pending" {
			pending = append(pending, tx)
		}
	}
	return pending
}

// getLatestBlock retorna o último bloco
func (ocl *OnChainLedger) getLatestBlock() *Block {
	var latest *Block
	for _, block := range ocl.blocks {
		if latest == nil || block.Number > latest.Number {
			latest = block
		}
	}
	return latest
}

// calculateBlockHash calcula hash do bloco
func (ocl *OnChainLedger) calculateBlockHash(block *Block) string {
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

// verifyPoW verifica se hash atende à dificuldade
func (ocl *OnChainLedger) verifyPoW(hash string, difficulty int) bool {
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

// calculateMiningReward calcula recompensa de mineração
func (ocl *OnChainLedger) calculateMiningReward(blockNumber int64) int64 {
	initialReward := int64(50)
	halvingInterval := int64(210000)
	
	epoch := blockNumber / halvingInterval
	reward := initialReward
	
	for i := int64(0); i < epoch; i++ {
		reward = reward / 2
		if reward <= 0 {
			reward = 1
			break
		}
	}
	
	return reward
}

// processBlockTransactions processa transações de um bloco
func (ocl *OnChainLedger) processBlockTransactions(block *Block) {
	for _, tx := range block.Transactions {
		// Atualizar status da transação
		if onchainTx, exists := ocl.transactions[tx.ID]; exists {
			onchainTx.Status = "confirmed"
			onchainTx.BlockHash = block.Hash
		}

		// Atualizar saldos das wallets
		ocl.updateWalletBalance(tx.From, -(tx.Amount + tx.Fee))
		ocl.updateWalletBalance(tx.To, tx.Amount)
	}

	// Adicionar recompensa do minerador
	ocl.updateWalletBalance(block.MinerID, block.Reward)
}

// updateWalletBalance atualiza saldo de uma wallet
func (ocl *OnChainLedger) updateWalletBalance(address string, amount int64) {
	if wm, ok := ocl.walletManager.(interface {
		UpdateBalance(string, int64) error
	}); ok {
		wm.UpdateBalance(address, amount)
	}
}

// GetTransactionHistory retorna histórico de transações
func (ocl *OnChainLedger) GetTransactionHistory() []*OnChainTransaction {
	ocl.mu.RLock()
	defer ocl.mu.RUnlock()

	var history []*OnChainTransaction
	for _, tx := range ocl.transactions {
		history = append(history, tx)
	}
	return history
}

// GetBlockHistory retorna histórico de blocos
func (ocl *OnChainLedger) GetBlockHistory() []*Block {
	ocl.mu.RLock()
	defer ocl.mu.RUnlock()

	var history []*Block
	for _, block := range ocl.blocks {
		history = append(history, block)
	}
	return history
}

// GetBlockchainStats retorna estatísticas da blockchain
func (ocl *OnChainLedger) GetBlockchainStats() map[string]interface{} {
	ocl.mu.RLock()
	defer ocl.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["block_height"] = ocl.blockHeight
	stats["total_blocks"] = len(ocl.blocks)
	stats["total_transactions"] = len(ocl.transactions)
	stats["pending_transactions"] = len(ocl.getPendingTransactions())

	// Calcular supply total
	if wm, ok := ocl.walletManager.(interface {
		GetTotalSupply() int64
	}); ok {
		stats["total_supply"] = wm.GetTotalSupply()
	}

	return stats
}
