package storage

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// BadgerAdapter adapta componentes existentes para usar BadgerDB
type BadgerAdapter struct {
	badgerStore *BadgerStore
	mu          sync.RWMutex
}

// NewBadgerAdapter cria um novo adaptador BadgerDB
func NewBadgerAdapter(dbPath string) (*BadgerAdapter, error) {
	badgerStore, err := NewBadgerStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar BadgerStore: %v", err)
	}

	return &BadgerAdapter{
		badgerStore: badgerStore,
	}, nil
}

// Close fecha a conexão com o BadgerDB
func (ba *BadgerAdapter) Close() error {
	return ba.badgerStore.Close()
}

// LedgerAdapter adapta o GlobalLedger para usar BadgerDB
type LedgerAdapter struct {
	*BadgerAdapter
}

// NewLedgerAdapter cria um adaptador para o GlobalLedger
func NewLedgerAdapter(dbPath string) (*LedgerAdapter, error) {
	adapter, err := NewBadgerAdapter(dbPath)
	if err != nil {
		return nil, err
	}

	return &LedgerAdapter{
		BadgerAdapter: adapter,
	}, nil
}

// SaveLedger salva o ledger no BadgerDB
func (la *LedgerAdapter) SaveLedger(ledgerData interface{}) error {
	la.mu.Lock()
	defer la.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(ledgerData)
	if err != nil {
		return fmt.Errorf("erro ao serializar ledger: %v", err)
	}

	// Salvar no BadgerDB
	key := []byte("ledger:global")
	return la.badgerStore.Set(key, jsonData)
}

// LoadLedger carrega o ledger do BadgerDB
func (la *LedgerAdapter) LoadLedger(ledgerData interface{}) error {
	la.mu.RLock()
	defer la.mu.RUnlock()

	// Carregar do BadgerDB
	key := []byte("ledger:global")
	jsonData, err := la.badgerStore.Get(key)
	if err != nil {
		return fmt.Errorf("erro ao carregar ledger: %v", err)
	}

	// Deserializar dados
	if err := json.Unmarshal(jsonData, ledgerData); err != nil {
		return fmt.Errorf("erro ao deserializar ledger: %v", err)
	}

	return nil
}

// SaveBalance salva saldo no BadgerDB
func (la *LedgerAdapter) SaveBalance(address string, balance int64) error {
	balanceData := map[string]interface{}{
		"address": address,
		"balance": balance,
		"updated": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(balanceData)
	if err != nil {
		return fmt.Errorf("erro ao serializar saldo: %v", err)
	}

	return la.badgerStore.SaveBalance(address, jsonData)
}

// GetBalance obtém saldo do BadgerDB
func (la *LedgerAdapter) GetBalance(address string) (int64, error) {
	jsonData, err := la.badgerStore.GetBalance(address)
	if err != nil {
		return 0, err
	}

	var balanceData map[string]interface{}
	if err := json.Unmarshal(jsonData, &balanceData); err != nil {
		return 0, fmt.Errorf("erro ao deserializar saldo: %v", err)
	}

	if balance, ok := balanceData["balance"].(float64); ok {
		return int64(balance), nil
	}

	return 0, fmt.Errorf("formato de saldo inválido")
}

// WalletAdapter adapta o WalletManager para usar BadgerDB
type WalletAdapter struct {
	*BadgerAdapter
}

// NewWalletAdapter cria um adaptador para o WalletManager
func NewWalletAdapter(dbPath string) (*WalletAdapter, error) {
	adapter, err := NewBadgerAdapter(dbPath)
	if err != nil {
		return nil, err
	}

	return &WalletAdapter{
		BadgerAdapter: adapter,
	}, nil
}

// SaveWallet salva wallet no BadgerDB
func (wa *WalletAdapter) SaveWallet(walletID string, walletData interface{}) error {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(walletData)
	if err != nil {
		return fmt.Errorf("erro ao serializar wallet: %v", err)
	}

	// Salvar no BadgerDB
	key := fmt.Sprintf("wallet:%s", walletID)
	return wa.badgerStore.Set([]byte(key), jsonData)
}

// LoadWallet carrega wallet do BadgerDB
func (wa *WalletAdapter) LoadWallet(walletID string, walletData interface{}) error {
	wa.mu.RLock()
	defer wa.mu.RUnlock()

	// Carregar do BadgerDB
	key := fmt.Sprintf("wallet:%s", walletID)
	jsonData, err := wa.badgerStore.Get([]byte(key))
	if err != nil {
		return fmt.Errorf("erro ao carregar wallet: %v", err)
	}

	// Deserializar dados
	if err := json.Unmarshal(jsonData, walletData); err != nil {
		return fmt.Errorf("erro ao deserializar wallet: %v", err)
	}

	return nil
}

// GetAllWallets obtém todas as wallets do BadgerDB
func (wa *WalletAdapter) GetAllWallets() (map[string][]byte, error) {
	return wa.badgerStore.PrefixScan([]byte("wallet:"))
}

// UserAdapter adapta o UserManager para usar BadgerDB
type UserAdapter struct {
	*BadgerAdapter
}

// NewUserAdapter cria um adaptador para o UserManager
func NewUserAdapter(dbPath string) (*UserAdapter, error) {
	adapter, err := NewBadgerAdapter(dbPath)
	if err != nil {
		return nil, err
	}

	return &UserAdapter{
		BadgerAdapter: adapter,
	}, nil
}

// SaveUsers salva usuários no BadgerDB
func (ua *UserAdapter) SaveUsers(usersData interface{}) error {
	ua.mu.Lock()
	defer ua.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(usersData)
	if err != nil {
		return fmt.Errorf("erro ao serializar usuários: %v", err)
	}

	// Salvar no BadgerDB
	key := []byte("users:global")
	return ua.badgerStore.Set(key, jsonData)
}

// LoadUsers carrega usuários do BadgerDB
func (ua *UserAdapter) LoadUsers(usersData interface{}) error {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	// Carregar do BadgerDB
	key := []byte("users:global")
	jsonData, err := ua.badgerStore.Get(key)
	if err != nil {
		return fmt.Errorf("erro ao carregar usuários: %v", err)
	}

	// Deserializar dados
	if err := json.Unmarshal(jsonData, usersData); err != nil {
		return fmt.Errorf("erro ao deserializar usuários: %v", err)
	}

	return nil
}

// BlockAdapter adapta o armazenamento de blocos para usar BadgerDB
type BlockAdapter struct {
	*BadgerAdapter
}

// NewBlockAdapter cria um adaptador para blocos
func NewBlockAdapter(dbPath string) (*BlockAdapter, error) {
	adapter, err := NewBadgerAdapter(dbPath)
	if err != nil {
		return nil, err
	}

	return &BlockAdapter{
		BadgerAdapter: adapter,
	}, nil
}

// SaveBlock salva bloco no BadgerDB
func (ba *BlockAdapter) SaveBlock(blockHash string, blockData interface{}) error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(blockData)
	if err != nil {
		return fmt.Errorf("erro ao serializar bloco: %v", err)
	}

	// Salvar no BadgerDB
	return ba.badgerStore.SaveBlock(blockHash, jsonData)
}

// LoadBlock carrega bloco do BadgerDB
func (ba *BlockAdapter) LoadBlock(blockHash string, blockData interface{}) error {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	// Carregar do BadgerDB
	jsonData, err := ba.badgerStore.GetBlock(blockHash)
	if err != nil {
		return fmt.Errorf("erro ao carregar bloco: %v", err)
	}

	// Deserializar dados
	if err := json.Unmarshal(jsonData, blockData); err != nil {
		return fmt.Errorf("erro ao deserializar bloco: %v", err)
	}

	return nil
}

// GetAllBlocks obtém todos os blocos do BadgerDB
func (ba *BlockAdapter) GetAllBlocks() (map[string][]byte, error) {
	return ba.badgerStore.GetAllBlocks()
}

// TransactionAdapter adapta o armazenamento de transações para usar BadgerDB
type TransactionAdapter struct {
	*BadgerAdapter
}

// NewTransactionAdapter cria um adaptador para transações
func NewTransactionAdapter(dbPath string) (*TransactionAdapter, error) {
	adapter, err := NewBadgerAdapter(dbPath)
	if err != nil {
		return nil, err
	}

	return &TransactionAdapter{
		BadgerAdapter: adapter,
	}, nil
}

// SaveTransaction salva transação no BadgerDB
func (ta *TransactionAdapter) SaveTransaction(txHash string, txData interface{}) error {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(txData)
	if err != nil {
		return fmt.Errorf("erro ao serializar transação: %v", err)
	}

	// Salvar no BadgerDB
	return ta.badgerStore.SaveTransaction(txHash, jsonData)
}

// LoadTransaction carrega transação do BadgerDB
func (ta *TransactionAdapter) LoadTransaction(txHash string, txData interface{}) error {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	// Carregar do BadgerDB
	jsonData, err := ta.badgerStore.GetTransaction(txHash)
	if err != nil {
		return fmt.Errorf("erro ao carregar transação: %v", err)
	}

	// Deserializar dados
	if err := json.Unmarshal(jsonData, txData); err != nil {
		return fmt.Errorf("erro ao deserializar transação: %v", err)
	}

	return nil
}

// GetAllTransactions obtém todas as transações do BadgerDB
func (ta *TransactionAdapter) GetAllTransactions() (map[string][]byte, error) {
	return ta.badgerStore.GetAllTransactions()
}

// MigrationManager gerencia migração de JSON para BadgerDB
type MigrationManager struct {
	badgerAdapter *BadgerAdapter
	mu            sync.RWMutex
}

// NewMigrationManager cria um gerenciador de migração
func NewMigrationManager(dbPath string) (*MigrationManager, error) {
	adapter, err := NewBadgerAdapter(dbPath)
	if err != nil {
		return nil, err
	}

	return &MigrationManager{
		badgerAdapter: adapter,
	}, nil
}

// MigrateLedger migra ledger de JSON para BadgerDB
func (mm *MigrationManager) MigrateLedger(ledgerData interface{}) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(ledgerData)
	if err != nil {
		return fmt.Errorf("erro ao serializar ledger para migração: %v", err)
	}

	// Salvar no BadgerDB
	key := []byte("ledger:global")
	return mm.badgerAdapter.badgerStore.Set(key, jsonData)
}

// MigrateWallets migra wallets de JSON para BadgerDB
func (mm *MigrationManager) MigrateWallets(wallets map[string]interface{}) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for walletID, walletData := range wallets {
		// Serializar dados
		jsonData, err := json.Marshal(walletData)
		if err != nil {
			return fmt.Errorf("erro ao serializar wallet %s: %v", walletID, err)
		}

		// Salvar no BadgerDB
		key := fmt.Sprintf("wallet:%s", walletID)
		if err := mm.badgerAdapter.badgerStore.Set([]byte(key), jsonData); err != nil {
			return fmt.Errorf("erro ao salvar wallet %s: %v", walletID, err)
		}
	}

	return nil
}

// MigrateUsers migra usuários de JSON para BadgerDB
func (mm *MigrationManager) MigrateUsers(usersData interface{}) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Serializar dados
	jsonData, err := json.Marshal(usersData)
	if err != nil {
		return fmt.Errorf("erro ao serializar usuários para migração: %v", err)
	}

	// Salvar no BadgerDB
	key := []byte("users:global")
	return mm.badgerAdapter.badgerStore.Set(key, jsonData)
}

// Close fecha a conexão com o BadgerDB
func (mm *MigrationManager) Close() error {
	return mm.badgerAdapter.Close()
}
