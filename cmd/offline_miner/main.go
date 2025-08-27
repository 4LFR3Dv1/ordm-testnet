package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ordm-main/pkg/blockchain"
	"ordm-main/pkg/crypto"
)

// OfflineMiner representa o minerador offline
type OfflineMiner struct {
	Identity     *crypto.MinerIdentity
	LocalChain   *LocalBlockchain
	MiningPool   *MiningPool
	SyncManager  *SyncManager
	Config       *MinerConfig
	mu           sync.RWMutex
}

// LocalBlockchain gerencia a blockchain local
type LocalBlockchain struct {
	Blocks       map[string]*blockchain.RealBlock `json:"blocks"`
	PendingTxs   []blockchain.RealTransaction     `json:"pending_transactions"`
	LastBlock    *blockchain.RealBlock            `json:"last_block"`
	Difficulty   uint64                           `json:"difficulty"`
	DataPath     string                           `json:"-"`
	mu           sync.RWMutex                     `json:"-"`
}

// MiningPool gerencia o pool de mineração
type MiningPool struct {
	Blocks       []*blockchain.RealBlock `json:"blocks"`
	BatchSize    int                     `json:"batch_size"`
	MaxRetries   int                     `json:"max_retries"`
	mu           sync.RWMutex            `json:"-"`
}

// SyncManager gerencia a sincronização com a rede online
type SyncManager struct {
	SeedNodes    []string                `json:"seed_nodes"`
	SyncInterval time.Duration           `json:"sync_interval"`
	LastSync     time.Time               `json:"last_sync"`
	Status       string                  `json:"status"` // "connected", "disconnected", "syncing"
	mu           sync.RWMutex            `json:"-"`
}

// MinerConfig configurações do minerador
type MinerConfig struct {
	MinerName    string        `json:"miner_name"`
	DataPath     string        `json:"data_path"`
	Difficulty   uint64        `json:"difficulty"`
	BatchSize    int           `json:"batch_size"`
	SyncInterval time.Duration `json:"sync_interval"`
	MaxRetries   int           `json:"max_retries"`
}

// MiningStats estatísticas de mineração
type MiningStats struct {
	TotalBlocks    int64     `json:"total_blocks"`
	ValidBlocks    int64     `json:"valid_blocks"`
	InvalidBlocks  int64     `json:"invalid_blocks"`
	HashRate       float64   `json:"hash_rate"`
	LastMined      time.Time `json:"last_mined"`
	Uptime         time.Duration `json:"uptime"`
	StartTime      time.Time `json:"start_time"`
}

var (
	offlineMiner *OfflineMiner
	miningTicker *time.Ticker
	miningStop   chan bool
	startTime    time.Time
)

func main() {
	// Inicializar minerador offline
	offlineMiner = NewOfflineMiner()
	
	// Configurar rotas HTTP
	setupRoutes()
	
	// Iniciar servidor
	port := "8081" // Porta diferente do online
	fmt.Printf("🏭 Minerador Offline iniciado na porta %s\n", port)
	fmt.Printf("📁 Dados salvos em: %s\n", offlineMiner.Config.DataPath)
	fmt.Printf("🔑 Minerador ID: %s\n", offlineMiner.Identity.MinerID)
	
	startTime = time.Now()
	
	// Iniciar servidor HTTP
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// NewOfflineMiner cria um novo minerador offline
func NewOfflineMiner() *OfflineMiner {
	// Configurar caminho de dados
	dataPath := "./offline_data"
	if os.Getenv("OFFLINE_DATA_PATH") != "" {
		dataPath = os.Getenv("OFFLINE_DATA_PATH")
	}
	
	// Criar diretório se não existir
	os.MkdirAll(dataPath, 0755)
	
	// Inicializar gerenciador de chaves
	keyManager := crypto.NewMinerKeyManager(filepath.Join(dataPath, "keys"))
	
	// Gerar ou carregar identidade do minerador
	var identity *crypto.MinerIdentity
	identities := keyManager.ListIdentities()
	
	if len(identities) > 0 {
		// Usar identidade existente
		identity = identities[0]
		fmt.Printf("🔑 Identidade carregada: %s\n", identity.MinerID)
	} else {
		// Gerar nova identidade
		var err error
		identity, err = keyManager.GenerateMinerIdentity("offline_miner")
		if err != nil {
			log.Fatalf("Erro ao gerar identidade: %v", err)
		}
		fmt.Printf("🔑 Nova identidade criada: %s\n", identity.MinerID)
	}
	
	// Configurações padrão
	config := &MinerConfig{
		MinerName:    "OfflineMiner",
		DataPath:     dataPath,
		Difficulty:   4, // Dificuldade inicial baixa
		BatchSize:    10,
		SyncInterval: 5 * time.Minute,
		MaxRetries:   3,
	}
	
	// Inicializar blockchain local
	localChain := NewLocalBlockchain(dataPath)
	
	// Inicializar pool de mineração
	miningPool := &MiningPool{
		Blocks:     []*blockchain.RealBlock{},
		BatchSize:  config.BatchSize,
		MaxRetries: config.MaxRetries,
	}
	
	// Inicializar gerenciador de sincronização
	syncManager := &SyncManager{
		SeedNodes:    []string{"https://ordm-testnet-1.onrender.com/api/sync"},
		SyncInterval: config.SyncInterval,
		Status:       "disconnected",
	}
	
	return &OfflineMiner{
		Identity:    identity,
		LocalChain:  localChain,
		MiningPool:  miningPool,
		SyncManager: syncManager,
		Config:      config,
	}
}

// NewLocalBlockchain cria uma nova blockchain local
func NewLocalBlockchain(dataPath string) *LocalBlockchain {
	chain := &LocalBlockchain{
		Blocks:     make(map[string]*blockchain.RealBlock),
		PendingTxs: []blockchain.RealTransaction{},
		Difficulty: 4,
		DataPath:   dataPath,
	}
	
	// Carregar blockchain existente
	chain.LoadBlockchain()
	
	// Criar bloco genesis se não existir
	if chain.LastBlock == nil {
		chain.CreateGenesisBlock()
	}
	
	return chain
}

// CreateGenesisBlock cria o bloco genesis
func (lc *LocalBlockchain) CreateGenesisBlock() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	
	// Criar bloco genesis
	genesis := blockchain.NewRealBlock(
		make([]byte, 32), // Parent hash vazio
		0,                 // Número 0
		"genesis",         // Minerador genesis
		1,                 // Dificuldade baixa
	)
	
	// Adicionar transação genesis
	genesisTx := blockchain.RealTransaction{
		ID:        "genesis_tx",
		From:      "genesis",
		To:        "genesis",
		Amount:    0,
		Fee:       0,
		Nonce:     0,
		Data:      "Genesis Block",
		Timestamp: time.Now().Unix(),
		Status:    "confirmed",
	}
	
	genesis.AddTransaction(genesisTx)
	
	// Salvar bloco genesis
	lc.Blocks[genesis.GetBlockHashString()] = genesis
	lc.LastBlock = genesis
	
	// Persistir
	lc.SaveBlockchain()
	
	fmt.Printf("🌱 Bloco Genesis criado: %s\n", genesis.GetBlockHashString())
}

// MineNextBlock minera o próximo bloco
func (lc *LocalBlockchain) MineNextBlock(minerID string) (*blockchain.RealBlock, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	
	// Obter hash do último bloco
	var parentHash []byte
	if lc.LastBlock != nil {
		parentHash = lc.LastBlock.GetBlockHash()
	}
	
	// Número do próximo bloco
	nextNumber := int64(0)
	if lc.LastBlock != nil {
		nextNumber = lc.LastBlock.Header.Number + 1
	}
	
	// Criar novo bloco
	block := blockchain.NewRealBlock(parentHash, nextNumber, minerID, lc.Difficulty)
	
	// Adicionar transações pendentes
	for _, tx := range lc.PendingTxs {
		block.AddTransaction(tx)
	}
	
	// Minerar o bloco
	fmt.Printf("⛏️ Minerando bloco #%d...\n", nextNumber)
	startTime := time.Now()
	
	err := block.MineBlock(lc.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("erro na mineração: %v", err)
	}
	
	miningTime := time.Since(startTime)
	fmt.Printf("✅ Bloco #%d minerado em %v\n", nextNumber, miningTime)
	
	// Adicionar bloco à blockchain
	lc.Blocks[block.GetBlockHashString()] = block
	lc.LastBlock = block
	
	// Limpar transações pendentes
	lc.PendingTxs = []blockchain.RealTransaction{}
	
	// Persistir
	lc.SaveBlockchain()
	
	return block, nil
}

// AddPendingTransaction adiciona uma transação pendente
func (lc *LocalBlockchain) AddPendingTransaction(tx blockchain.RealTransaction) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	
	// Validar transação
	if tx.From == "" || tx.To == "" {
		return fmt.Errorf("transação inválida: campos obrigatórios")
	}
	
	if tx.Amount <= 0 {
		return fmt.Errorf("valor deve ser maior que zero")
	}
	
	// Adicionar à lista de pendentes
	lc.PendingTxs = append(lc.PendingTxs, tx)
	
	// Persistir
	lc.SaveBlockchain()
	
	return nil
}

// LoadBlockchain carrega a blockchain do disco
func (lc *LocalBlockchain) LoadBlockchain() error {
	filePath := filepath.Join(lc.DataPath, "local_blockchain.json")
	
	// Se o arquivo não existir, retornar sem erro
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler blockchain: %v", err)
	}
	
	// Estrutura temporária para deserialização
	type tempBlockchain struct {
		Blocks       map[string]map[string]interface{} `json:"blocks"`
		PendingTxs   []blockchain.RealTransaction      `json:"pending_transactions"`
		Difficulty   uint64                            `json:"difficulty"`
		LastBlockHash string                           `json:"last_block_hash"`
	}
	
	var temp tempBlockchain
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("erro ao deserializar: %v", err)
	}
	
	// Converter blocos
	for hash, blockData := range temp.Blocks {
		blockJSON, _ := json.Marshal(blockData)
		block := &blockchain.RealBlock{}
		if err := block.FromJSON(blockJSON); err != nil {
			continue
		}
		lc.Blocks[hash] = block
	}
	
	// Definir último bloco
	if temp.LastBlockHash != "" {
		if lastBlock, exists := lc.Blocks[temp.LastBlockHash]; exists {
			lc.LastBlock = lastBlock
		}
	}
	
	lc.PendingTxs = temp.PendingTxs
	lc.Difficulty = temp.Difficulty
	
	fmt.Printf("📁 Blockchain local carregada: %d blocos\n", len(lc.Blocks))
	return nil
}

// SaveBlockchain salva a blockchain no disco
func (lc *LocalBlockchain) SaveBlockchain() error {
	// Estrutura temporária para serialização
	tempBlocks := make(map[string]map[string]interface{})
	
	for hash, block := range lc.Blocks {
		blockJSON, _ := block.ToJSON()
		var blockData map[string]interface{}
		json.Unmarshal(blockJSON, &blockData)
		tempBlocks[hash] = blockData
	}
	
	lastBlockHash := ""
	if lc.LastBlock != nil {
		lastBlockHash = lc.LastBlock.GetBlockHashString()
	}
	
	data := map[string]interface{}{
		"blocks":           tempBlocks,
		"pending_transactions": lc.PendingTxs,
		"difficulty":       lc.Difficulty,
		"last_block_hash":  lastBlockHash,
	}
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar: %v", err)
	}
	
	filePath := filepath.Join(lc.DataPath, "local_blockchain.json")
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("erro ao salvar: %v", err)
	}
	
	return nil
}

// GetMiningStats retorna estatísticas de mineração
func GetMiningStats() MiningStats {
	stats := MiningStats{
		StartTime: startTime,
		Uptime:    time.Since(startTime),
	}
	
	if offlineMiner != nil && offlineMiner.LocalChain != nil {
		offlineMiner.LocalChain.mu.RLock()
		stats.TotalBlocks = int64(len(offlineMiner.LocalChain.Blocks))
		offlineMiner.LocalChain.mu.RUnlock()
		
		// Calcular hash rate (simplificado)
		stats.HashRate = float64(stats.TotalBlocks) * 0.5
	}
	
	return stats
}
