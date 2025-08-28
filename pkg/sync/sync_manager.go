package sync

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ordm-main/pkg/blockchain"
	"ordm-main/pkg/ledger"
)

// BlockchainSyncManager gerencia a sincronização de blocos e transações
type BlockchainSyncManager struct {
	ledger          *ledger.GlobalLedger
	blocks          []*blockchain.Block
	blockValidator  *blockchain.BlockValidator
	mempool         *Mempool
	lastBlockHash   string
	lastBlockNumber int64
	mutex           sync.RWMutex
	logger          func(string, ...interface{})
}

// Mempool gerencia transações pendentes
type Mempool struct {
	transactions map[string]*blockchain.Transaction
	mutex        sync.RWMutex
	maxSize      int
}

// NewBlockchainSyncManager cria um novo gerenciador de sincronização de blockchain
func NewBlockchainSyncManager(ledger *ledger.GlobalLedger, logger func(string, ...interface{})) *SyncManager {
	return &SyncManager{
		ledger:          ledger,
		blocks:          []*blockchain.Block{},
		blockValidator:  blockchain.NewBlockValidator(2), // Dificuldade 2 para testnet
		mempool:         NewMempool(1000),                // Máximo 1000 transações
		lastBlockHash:   "",
		lastBlockNumber: 0,
		logger:          logger,
	}
}

// NewMempool cria um novo mempool
func NewMempool(maxSize int) *Mempool {
	return &Mempool{
		transactions: make(map[string]*blockchain.Transaction),
		maxSize:      maxSize,
	}
}

// AddTransaction adiciona uma transação ao mempool
func (m *Mempool) AddTransaction(tx *blockchain.Transaction) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Verificar se o mempool está cheio
	if len(m.transactions) >= m.maxSize {
		return fmt.Errorf("mempool cheio")
	}

	// Verificar se a transação já existe
	if _, exists := m.transactions[tx.TxHash]; exists {
		return fmt.Errorf("transação já existe no mempool")
	}

	// Validar transação
	if err := validateTransaction(tx); err != nil {
		return fmt.Errorf("transação inválida: %v", err)
	}

	// Adicionar ao mempool
	m.transactions[tx.TxHash] = tx
	return nil
}

// GetTransactions retorna transações do mempool
func (m *Mempool) GetTransactions(limit int) []*blockchain.Transaction {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var transactions []*blockchain.Transaction
	count := 0
	for _, tx := range m.transactions {
		if count >= limit {
			break
		}
		transactions = append(transactions, tx)
		count++
	}

	return transactions
}

// RemoveTransaction remove uma transação do mempool
func (m *Mempool) RemoveTransaction(txHash string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.transactions, txHash)
}

// GetSize retorna o tamanho do mempool
func (m *Mempool) GetSize() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.transactions)
}

// validateTransaction valida uma transação
func validateTransaction(tx *blockchain.Transaction) error {
	// Validar hash
	calculatedHash := calculateTransactionHash(tx)
	if calculatedHash != tx.TxHash {
		return fmt.Errorf("hash inválido")
	}

	// Validar valores
	if tx.Amount <= 0 {
		return fmt.Errorf("valor deve ser positivo")
	}

	if tx.Fee < 0 {
		return fmt.Errorf("fee não pode ser negativo")
	}

	// Validar endereços
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("endereços obrigatórios")
	}

	// Validar timestamp
	currentTime := time.Now().Unix()
	if tx.Timestamp > currentTime+300 {
		return fmt.Errorf("timestamp muito no futuro")
	}

	return nil
}

// calculateTransactionHash calcula o hash de uma transação
func calculateTransactionHash(tx *blockchain.Transaction) string {
	// Usar a função do pacote blockchain
	return blockchain.CalculateTransactionHash(tx)
}

// AddBlock adiciona um bloco à blockchain
func (sm *SyncManager) AddBlock(block *blockchain.Block) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Validar bloco
	var previousBlock *blockchain.Block
	if len(sm.blocks) > 0 {
		previousBlock = sm.blocks[len(sm.blocks)-1]
	}

	if err := sm.blockValidator.ValidateBlock(block, previousBlock); err != nil {
		return fmt.Errorf("bloco inválido: %v", err)
	}

	// Verificar se o bloco já existe
	for _, existingBlock := range sm.blocks {
		if existingBlock.Hash == block.Hash {
			return fmt.Errorf("bloco já existe")
		}
	}

	// Adicionar bloco
	sm.blocks = append(sm.blocks, block)
	sm.lastBlockHash = block.Hash
	sm.lastBlockNumber = int64(len(sm.blocks))

	// Processar transações do bloco
	if err := sm.processBlockTransactions(block); err != nil {
		return fmt.Errorf("erro ao processar transações: %v", err)
	}

	// Remover transações confirmadas do mempool
	sm.removeConfirmedTransactions(block)

	sm.logger("✅ Bloco #%d adicionado: %s", sm.lastBlockNumber, block.Hash[:16])
	return nil
}

// processBlockTransactions processa as transações de um bloco
func (sm *SyncManager) processBlockTransactions(block *blockchain.Block) error {
	sm.ledger.Mutex.Lock()
	defer sm.ledger.Mutex.Unlock()

	for _, tx := range block.Transactions {
		// Verificar saldo do remetente
		fromBalance := sm.ledger.Balances[tx.From]
		if fromBalance < tx.Amount+tx.Fee {
			return fmt.Errorf("saldo insuficiente para transação %s", tx.TxHash[:16])
		}

		// Executar transferência
		sm.ledger.Balances[tx.From] -= (tx.Amount + tx.Fee)
		sm.ledger.Balances[tx.To] += tx.Amount

		// Adicionar movimento ao ledger
		movement := ledger.TokenMovement{
			ID:          tx.TxHash,
			From:        tx.From,
			To:          tx.To,
			Amount:      tx.Amount,
			Fee:         tx.Fee,
			Type:        "transfer",
			BlockHash:   block.Hash,
			Timestamp:   tx.Timestamp,
			Transaction: tx.TxHash,
			Description: fmt.Sprintf("Transferência de %d tokens", tx.Amount),
		}

		sm.ledger.Movements = append(sm.ledger.Movements, movement)
	}

	// Salvar ledger
	return sm.ledger.SaveLedger()
}

// removeConfirmedTransactions remove transações confirmadas do mempool
func (sm *SyncManager) removeConfirmedTransactions(block *blockchain.Block) {
	for _, tx := range block.Transactions {
		sm.mempool.RemoveTransaction(tx.TxHash)
	}
}

// GetLastBlock retorna o último bloco
func (sm *SyncManager) GetLastBlock() *blockchain.Block {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if len(sm.blocks) == 0 {
		return nil
	}
	return sm.blocks[len(sm.blocks)-1]
}

// GetBlockByHash retorna um bloco pelo hash
func (sm *SyncManager) GetBlockByHash(hash string) *blockchain.Block {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for _, block := range sm.blocks {
		if block.Hash == hash {
			return block
		}
	}
	return nil
}

// GetBlockByNumber retorna um bloco pelo número
func (sm *SyncManager) GetBlockByNumber(number int64) *blockchain.Block {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if number < 0 || int(number) >= len(sm.blocks) {
		return nil
	}
	return sm.blocks[number]
}

// GetBlocks retorna todos os blocos
func (sm *SyncManager) GetBlocks() []*blockchain.Block {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	blocks := make([]*blockchain.Block, len(sm.blocks))
	copy(blocks, sm.blocks)
	return blocks
}

// GetBlockCount retorna o número de blocos
func (sm *SyncManager) GetBlockCount() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return int64(len(sm.blocks))
}

// GetLastBlockHash retorna o hash do último bloco
func (sm *SyncManager) GetLastBlockHash() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.lastBlockHash
}

// GetLastBlockNumber retorna o número do último bloco
func (sm *SyncManager) GetLastBlockNumber() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.lastBlockNumber
}

// CreateNewBlock cria um novo bloco para mineração
func (sm *SyncManager) CreateNewBlock(miner string) (*blockchain.Block, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Obter transações do mempool
	transactions := sm.mempool.GetTransactions(100) // Máximo 100 transações por bloco

	// Converter para slice de Transaction
	var txSlice []blockchain.Transaction
	for _, tx := range transactions {
		txSlice = append(txSlice, *tx)
	}

	// Criar novo bloco
	block := blockchain.NewBlock(sm.lastBlockHash, txSlice, miner, sm.blockValidator.Difficulty)

	return block, nil
}

// SyncWithPeer sincroniza com um peer
func (sm *SyncManager) SyncWithPeer(peerLastBlockHash string, peerLastBlockNumber int64) ([]*blockchain.Block, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Se o peer tem um bloco mais recente, precisamos sincronizar
	if peerLastBlockNumber > sm.lastBlockNumber {
		// Calcular blocos faltantes
		missingBlocks := peerLastBlockNumber - sm.lastBlockNumber
		sm.logger("📡 Sincronizando %d blocos faltantes", missingBlocks)

		// Retornar blocos que precisamos
		return sm.blocks, nil
	}

	return nil, nil
}

// ValidateAndAddBlock valida e adiciona um bloco recebido via P2P
func (sm *SyncManager) ValidateAndAddBlock(blockData map[string]interface{}) error {
	// Converter dados para bloco
	blockJSON, err := json.Marshal(blockData)
	if err != nil {
		return fmt.Errorf("erro ao serializar bloco: %v", err)
	}

	block, err := blockchain.FromJSON(blockJSON)
	if err != nil {
		return fmt.Errorf("erro ao deserializar bloco: %v", err)
	}

	// Adicionar bloco
	return sm.AddBlock(block)
}

// GetMempoolStats retorna estatísticas do mempool
func (sm *SyncManager) GetMempoolStats() map[string]interface{} {
	return map[string]interface{}{
		"size":         sm.mempool.GetSize(),
		"max_size":     sm.mempool.maxSize,
		"transactions": sm.mempool.GetTransactions(10), // Primeiras 10 para preview
	}
}

// GetSyncStats retorna estatísticas de sincronização
func (sm *SyncManager) GetSyncStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return map[string]interface{}{
		"total_blocks":      sm.lastBlockNumber,
		"last_block_hash":   sm.lastBlockHash,
		"last_block_number": sm.lastBlockNumber,
		"mempool_size":      sm.mempool.GetSize(),
		"synced":            true,
	}
}
