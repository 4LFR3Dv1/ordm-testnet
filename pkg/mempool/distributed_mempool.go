package mempool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Transaction representa uma transação no mempool
type Transaction struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	Nonce     uint64 `json:"nonce"`
	Data      string `json:"data"`
	Signature []byte `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`   // "pending", "confirmed", "failed"
	Received  int64  `json:"received"` // timestamp de quando foi recebida
}

// PeerInfo representa informações de um peer
type PeerInfo struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	TxCount  int       `json:"tx_count"`
	IsActive bool      `json:"is_active"`
}

// DistributedMempool implementa mempool distribuído P2P
type DistributedMempool struct {
	transactions  map[string]*Transaction
	peers         map[string]*PeerInfo
	mu            sync.RWMutex
	maxSize       int
	cleanupTicker *time.Ticker
	stopChan      chan bool
	logger        func(string, ...interface{})
}

// NewDistributedMempool cria um novo mempool distribuído
func NewDistributedMempool(maxSize int, logger func(string, ...interface{})) *DistributedMempool {
	mp := &DistributedMempool{
		transactions: make(map[string]*Transaction),
		peers:        make(map[string]*PeerInfo),
		maxSize:      maxSize,
		stopChan:     make(chan bool),
		logger:       logger,
	}

	// Iniciar limpeza automática
	mp.startCleanupRoutine()

	return mp
}

// AddTransaction adiciona uma transação ao mempool
func (mp *DistributedMempool) AddTransaction(tx *Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Verificar se já existe
	if _, exists := mp.transactions[tx.ID]; exists {
		return fmt.Errorf("transação já existe no mempool")
	}

	// Verificar limite de tamanho
	if len(mp.transactions) >= mp.maxSize {
		// Remover transação mais antiga
		mp.removeOldestTransaction()
	}

	// Validar transação
	if err := mp.validateTransaction(tx); err != nil {
		return fmt.Errorf("transação inválida: %v", err)
	}

	// Adicionar transação
	tx.Received = time.Now().Unix()
	tx.Status = "pending"
	mp.transactions[tx.ID] = tx

	mp.logger("📥 Transação %s adicionada ao mempool", tx.ID)
	return nil
}

// GetTransaction retorna uma transação pelo ID
func (mp *DistributedMempool) GetTransaction(txID string) (*Transaction, bool) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	tx, exists := mp.transactions[txID]
	return tx, exists
}

// GetPendingTransactions retorna transações pendentes ordenadas por fee
func (mp *DistributedMempool) GetPendingTransactions(limit int) []*Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var pending []*Transaction
	for _, tx := range mp.transactions {
		if tx.Status == "pending" {
			pending = append(pending, tx)
		}
	}

	// Ordenar por fee (maior primeiro)
	mp.sortByFee(pending)

	// Limitar resultado
	if limit > 0 && len(pending) > limit {
		pending = pending[:limit]
	}

	return pending
}

// RemoveTransaction remove uma transação do mempool
func (mp *DistributedMempool) RemoveTransaction(txID string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if _, exists := mp.transactions[txID]; exists {
		delete(mp.transactions, txID)
		mp.logger("🗑️ Transação %s removida do mempool", txID)
	}
}

// MarkTransactionConfirmed marca uma transação como confirmada
func (mp *DistributedMempool) MarkTransactionConfirmed(txID string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if tx, exists := mp.transactions[txID]; exists {
		tx.Status = "confirmed"
		mp.logger("✅ Transação %s marcada como confirmada", txID)
	}
}

// GetMempoolStats retorna estatísticas do mempool
func (mp *DistributedMempool) GetMempoolStats() map[string]interface{} {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	pending := 0
	confirmed := 0
	failed := 0
	totalFees := int64(0)

	for _, tx := range mp.transactions {
		switch tx.Status {
		case "pending":
			pending++
		case "confirmed":
			confirmed++
		case "failed":
			failed++
		}
		totalFees += tx.Fee
	}

	return map[string]interface{}{
		"total_transactions": len(mp.transactions),
		"pending":            pending,
		"confirmed":          confirmed,
		"failed":             failed,
		"total_fees":         totalFees,
		"max_size":           mp.maxSize,
		"peers_count":        len(mp.peers),
	}
}

// validateTransaction valida uma transação
func (mp *DistributedMempool) validateTransaction(tx *Transaction) error {
	// Verificar campos obrigatórios
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("endereços de origem e destino são obrigatórios")
	}

	if tx.Amount <= 0 {
		return fmt.Errorf("valor deve ser maior que zero")
	}

	if tx.Fee < 0 {
		return fmt.Errorf("taxa não pode ser negativa")
	}

	// Verificar assinatura (implementação básica)
	if len(tx.Signature) == 0 {
		return fmt.Errorf("assinatura é obrigatória")
	}

	// Verificar timestamp (não muito no futuro)
	currentTime := time.Now().Unix()
	if tx.Timestamp > currentTime+300 { // 5 minutos de tolerância
		return fmt.Errorf("timestamp muito no futuro")
	}

	return nil
}

// removeOldestTransaction remove a transação mais antiga
func (mp *DistributedMempool) removeOldestTransaction() {
	var oldestTx *Transaction
	var oldestTime int64 = time.Now().Unix()

	for _, tx := range mp.transactions {
		if tx.Received < oldestTime {
			oldestTime = tx.Received
			oldestTx = tx
		}
	}

	if oldestTx != nil {
		delete(mp.transactions, oldestTx.ID)
		mp.logger("🗑️ Transação mais antiga %s removida do mempool", oldestTx.ID)
	}
}

// sortByFee ordena transações por fee (maior primeiro)
func (mp *DistributedMempool) sortByFee(transactions []*Transaction) {
	// Implementação simples de ordenação por fee
	for i := 0; i < len(transactions)-1; i++ {
		for j := i + 1; j < len(transactions); j++ {
			if transactions[i].Fee < transactions[j].Fee {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}
}

// startCleanupRoutine inicia rotina de limpeza automática
func (mp *DistributedMempool) startCleanupRoutine() {
	mp.cleanupTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-mp.cleanupTicker.C:
				mp.cleanup()
			case <-mp.stopChan:
				mp.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup remove transações antigas e inválidas
func (mp *DistributedMempool) cleanup() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	currentTime := time.Now().Unix()
	removed := 0

	for txID, tx := range mp.transactions {
		// Remover transações muito antigas (mais de 1 hora)
		if currentTime-tx.Received > 3600 {
			delete(mp.transactions, txID)
			removed++
		}
	}

	if removed > 0 {
		mp.logger("🧹 Limpeza: %d transações antigas removidas", removed)
	}
}

// Stop para a rotina de limpeza
func (mp *DistributedMempool) Stop() {
	close(mp.stopChan)
}

// GenerateTransactionID gera ID único para transação
func GenerateTransactionID(from, to string, amount, fee int64, nonce uint64, timestamp int64) string {
	data := fmt.Sprintf("%s:%s:%d:%d:%d:%d", from, to, amount, fee, nonce, timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
