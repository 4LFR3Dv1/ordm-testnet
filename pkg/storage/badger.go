package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerStore implementa persistência usando BadgerDB
type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore cria uma nova instância do BadgerDB
func NewBadgerStore(dbPath string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Desabilitar logs do BadgerDB

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir BadgerDB: %v", err)
	}

	return &BadgerStore{db: db}, nil
}

// Close fecha a conexão com o banco
func (bs *BadgerStore) Close() error {
	return bs.db.Close()
}

// Set salva um valor no banco
func (bs *BadgerStore) Set(key, value []byte) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get recupera um valor do banco
func (bs *BadgerStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

// Delete remove um valor do banco
func (bs *BadgerStore) Delete(key []byte) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// PrefixScan busca valores com prefixo específico
func (bs *BadgerStore) PrefixScan(prefix []byte) (map[string][]byte, error) {
	results := make(map[string][]byte)

	err := bs.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			key := item.Key()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			results[string(key)] = value
		}
		return nil
	})

	return results, err
}

// Backup cria backup do banco
func (bs *BadgerStore) Backup(backupPath string) error {
	file, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = bs.db.Backup(file, 0)
	return err
}

// Compact compacta o banco para otimizar performance
func (bs *BadgerStore) Compact() error {
	return bs.db.RunValueLogGC(0.7)
}

// Stats retorna estatísticas do banco
func (bs *BadgerStore) Stats() map[string]interface{} {
	lsmSize, vlogSize := bs.db.Size()
	return map[string]interface{}{
		"lsm_size":  lsmSize,
		"vlog_size": vlogSize,
		"is_closed": bs.db.IsClosed(),
	}
}

// Chaves específicas para diferentes tipos de dados
const (
	KeyPrefixBlock       = "block:"
	KeyPrefixTransaction = "tx:"
	KeyPrefixBalance     = "balance:"
	KeyPrefixStake       = "stake:"
	KeyPrefixNode        = "node:"
	KeyPrefixConfig      = "config:"
	KeyPrefixStats       = "stats:"
)

// BlockStorage métodos específicos para blocos
func (bs *BadgerStore) SaveBlock(blockID string, blockData []byte) error {
	key := []byte(KeyPrefixBlock + blockID)
	return bs.Set(key, blockData)
}

func (bs *BadgerStore) GetBlock(blockID string) ([]byte, error) {
	key := []byte(KeyPrefixBlock + blockID)
	return bs.Get(key)
}

func (bs *BadgerStore) GetAllBlocks() (map[string][]byte, error) {
	return bs.PrefixScan([]byte(KeyPrefixBlock))
}

// TransactionStorage métodos específicos para transações
func (bs *BadgerStore) SaveTransaction(txID string, txData []byte) error {
	key := []byte(KeyPrefixTransaction + txID)
	return bs.Set(key, txData)
}

func (bs *BadgerStore) GetTransaction(txID string) ([]byte, error) {
	key := []byte(KeyPrefixTransaction + txID)
	return bs.Get(key)
}

func (bs *BadgerStore) GetAllTransactions() (map[string][]byte, error) {
	return bs.PrefixScan([]byte(KeyPrefixTransaction))
}

// BalanceStorage métodos específicos para saldos
func (bs *BadgerStore) SaveBalance(address string, balanceData []byte) error {
	key := []byte(KeyPrefixBalance + address)
	return bs.Set(key, balanceData)
}

func (bs *BadgerStore) GetBalance(address string) ([]byte, error) {
	key := []byte(KeyPrefixBalance + address)
	return bs.Get(key)
}

func (bs *BadgerStore) GetAllBalances() (map[string][]byte, error) {
	return bs.PrefixScan([]byte(KeyPrefixBalance))
}

// StakeStorage métodos específicos para stake
func (bs *BadgerStore) SaveStake(address string, stakeData []byte) error {
	key := []byte(KeyPrefixStake + address)
	return bs.Set(key, stakeData)
}

func (bs *BadgerStore) GetStake(address string) ([]byte, error) {
	key := []byte(KeyPrefixStake + address)
	return bs.Get(key)
}

func (bs *BadgerStore) GetAllStakes() (map[string][]byte, error) {
	return bs.PrefixScan([]byte(KeyPrefixStake))
}

// NodeStorage métodos específicos para nodes
func (bs *BadgerStore) SaveNode(nodeID string, nodeData []byte) error {
	key := []byte(KeyPrefixNode + nodeID)
	return bs.Set(key, nodeData)
}

func (bs *BadgerStore) GetNode(nodeID string) ([]byte, error) {
	key := []byte(KeyPrefixNode + nodeID)
	return bs.Get(key)
}

func (bs *BadgerStore) GetAllNodes() (map[string][]byte, error) {
	return bs.PrefixScan([]byte(KeyPrefixNode))
}

// ConfigStorage métodos específicos para configurações
func (bs *BadgerStore) SaveConfig(key string, configData []byte) error {
	dbKey := []byte(KeyPrefixConfig + key)
	return bs.Set(dbKey, configData)
}

func (bs *BadgerStore) GetConfig(key string) ([]byte, error) {
	dbKey := []byte(KeyPrefixConfig + key)
	return bs.Get(dbKey)
}

// StatsStorage métodos específicos para estatísticas
func (bs *BadgerStore) SaveStats(statsData []byte) error {
	key := []byte(KeyPrefixStats + time.Now().Format("2006-01-02"))
	return bs.Set(key, statsData)
}

func (bs *BadgerStore) GetStats(date string) ([]byte, error) {
	key := []byte(KeyPrefixStats + date)
	return bs.Get(key)
}

// Maintenance métodos de manutenção
func (bs *BadgerStore) RunMaintenance() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if err := bs.Compact(); err != nil {
				log.Printf("Erro na compactação: %v", err)
			}
		}
	}()
}

// BatchOperations operações em lote para melhor performance
func (bs *BadgerStore) BatchSet(items map[string][]byte) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		for key, value := range items {
			if err := txn.Set([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bs *BadgerStore) BatchDelete(keys []string) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			if err := txn.Delete([]byte(key)); err != nil {
				return err
			}
		}
		return nil
	})
}
