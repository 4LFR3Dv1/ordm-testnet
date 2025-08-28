package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ordm-main/pkg/blockchain"
	"ordm-main/pkg/crypto"
	"ordm-main/pkg/mempool"
)

// ORDMNode representa o node principal da blockchain
type ORDMNode struct {
	Config     *NodeConfig
	Blockchain *blockchain.Block
	Mempool    *mempool.DistributedMempool
	RPCServer  *http.Server
	MachineID  *crypto.MachineID
	IsRunning  bool
	DataPath   string
	mu         sync.RWMutex
}

// NodeConfig configurações do node
type NodeConfig struct {
	NetworkID    string        `json:"network_id"`
	Port         string        `json:"port"`
	P2PPort      string        `json:"p2p_port"`
	RPCPort      string        `json:"rpc_port"`
	DataPath     string        `json:"data_path"`
	SeedNodes    []string      `json:"seed_nodes"`
	MaxPeers     int           `json:"max_peers"`
	SyncInterval time.Duration `json:"sync_interval"`
	Difficulty   uint64        `json:"difficulty"`
	BlockTime    time.Duration `json:"block_time"`
	GenesisFile  string        `json:"genesis_file"`
	ConfigFile   string        `json:"config_file"`

	// Configurações de mineração
	MiningEnabled bool   `json:"mining_enabled"`
	MinerKey      string `json:"miner_key"`
	MinerThreads  int    `json:"miner_threads"`
	MinerName     string `json:"miner_name"`
}

// MiningStats estatísticas de mineração
type MiningStats struct {
	TotalBlocks   int64         `json:"total_blocks"`
	ValidBlocks   int64         `json:"valid_blocks"`
	InvalidBlocks int64         `json:"invalid_blocks"`
	HashRate      float64       `json:"hash_rate"`
	LastMined     time.Time     `json:"last_mined"`
	Uptime        time.Duration `json:"uptime"`
	StartTime     time.Time     `json:"start_time"`
}

// NewORDMNode cria um novo node ORDM
func NewORDMNode(config *NodeConfig) (*ORDMNode, error) {
	// Criar diretório de dados se não existir
	if err := os.MkdirAll(config.DataPath, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de dados: %v", err)
	}

	// Inicializar machineID manager
	machineIDManager := crypto.NewMachineIDManager(config.DataPath)

	// Obter ou criar machineID
	machineID, err := machineIDManager.GetOrCreateMachineID()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter machineID: %v", err)
	}

	// Gerar minerID a partir do machineID se necessário
	if config.MiningEnabled && config.MinerKey == "" {
		minerID, err := machineIDManager.GetMinerIDFromMachineID()
		if err != nil {
			return nil, fmt.Errorf("erro ao gerar minerID: %v", err)
		}
		config.MinerKey = minerID
		log.Printf("🔑 MinerID gerado automaticamente: %s", minerID)
	}

	// Inicializar mempool
	logger := func(format string, args ...interface{}) {
		log.Printf(format, args...)
	}
	mempool := mempool.NewDistributedMempool(1000, logger)

	// Criar bloco genesis
	genesisBlock := blockchain.NewBlock("", []blockchain.Transaction{}, "genesis", config.Difficulty)

	return &ORDMNode{
		Config:     config,
		Blockchain: genesisBlock,
		Mempool:    mempool,
		MachineID:  machineID,
		DataPath:   config.DataPath,
	}, nil
}

// Start inicia o node
func (n *ORDMNode) Start() error {
	log.Printf("🚀 Iniciando ORDM Node (Network: %s)", n.Config.NetworkID)
	log.Printf("📁 Data Path: %s", n.DataPath)
	log.Printf("🌐 P2P Port: %s", n.Config.P2PPort)
	log.Printf("🔌 RPC Port: %s", n.Config.RPCPort)

	// Carregar configuração da testnet se especificada
	if n.Config.ConfigFile != "" {
		if err := n.loadTestnetConfig(); err != nil {
			log.Printf("⚠️  Erro ao carregar config da testnet: %v", err)
		}
	}

	// Carregar bloco genesis se especificado
	if n.Config.GenesisFile != "" {
		if err := n.loadGenesisBlock(); err != nil {
			log.Printf("⚠️  Erro ao carregar bloco genesis: %v", err)
		}
	}

	// Iniciar servidor RPC
	if err := n.startRPCServer(); err != nil {
		return fmt.Errorf("erro ao iniciar servidor RPC: %v", err)
	}

	// Iniciar sincronização automática
	go n.startAutoSync()

	// Iniciar mineração se habilitada
	if n.Config.MiningEnabled {
		log.Printf("⛏️  Mineração habilitada com %d threads", n.Config.MinerThreads)
		for i := 0; i < n.Config.MinerThreads; i++ {
			go n.miningWorker(i)
		}
	} else if n.Config.BlockTime > 0 {
		// Mineração automática por tempo
		go n.startAutoMining()
	}

	n.mu.Lock()
	n.IsRunning = true
	n.mu.Unlock()

	log.Printf("✅ ORDM Node iniciado com sucesso!")

	return nil
}

// Stop para o node
func (n *ORDMNode) Stop() error {
	log.Printf("🛑 Parando ORDM Node...")

	n.mu.Lock()
	n.IsRunning = false
	n.mu.Unlock()

	// Parar servidor RPC
	if n.RPCServer != nil {
		if err := n.RPCServer.Close(); err != nil {
			log.Printf("⚠️  Erro ao parar servidor RPC: %v", err)
		}
	}

	log.Printf("✅ ORDM Node parado com sucesso!")
	return nil
}

// startRPCServer inicia o servidor RPC
func (n *ORDMNode) startRPCServer() error {
	mux := http.NewServeMux()

	// Endpoint de informações da blockchain
	mux.HandleFunc("/api/v1/blockchain/info", n.handleBlockchainInfo)

	// Endpoint de status da blockchain
	mux.HandleFunc("/api/v1/blockchain/status", n.handleBlockchainStatus)

	// Endpoint de transações pendentes
	mux.HandleFunc("/api/v1/transactions/pending", n.handlePendingTransactions)

	// Endpoint para submeter blocos
	mux.HandleFunc("/api/v1/blocks/submit", n.handleSubmitBlock)

	n.RPCServer = &http.Server{
		Addr:    ":" + n.Config.RPCPort,
		Handler: mux,
	}

	go func() {
		log.Printf("🔌 Servidor RPC iniciado na porta %s", n.Config.RPCPort)
		if err := n.RPCServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Erro no servidor RPC: %v", err)
		}
	}()

	return nil
}

// handleBlockchainInfo retorna informações da blockchain
func (n *ORDMNode) handleBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]interface{}{
		"network_id":    n.Config.NetworkID,
		"version":       "1.0.0",
		"block_height":  1,
		"current_block": n.Blockchain.Hash,
		"difficulty":    n.Config.Difficulty,
		"is_running":    n.IsRunning,
		"mining":        n.Config.MiningEnabled,
		"machine_id":    n.MachineID.ID,
		"miner_id":      n.Config.MinerKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleBlockchainStatus retorna status da blockchain
func (n *ORDMNode) handleBlockchainStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"status":       "running",
		"uptime":       "0s",
		"peers":        0,
		"mempool_size": 0,
		"last_block":   n.Blockchain.Hash,
		"network_id":   n.Config.NetworkID,
		"mining":       n.Config.MiningEnabled,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handlePendingTransactions retorna transações pendentes
func (n *ORDMNode) handlePendingTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Retornar transações vazias por enquanto
	response := map[string]interface{}{
		"transactions": []interface{}{},
		"count":        0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSubmitBlock recebe blocos minerados
func (n *ORDMNode) handleSubmitBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Por enquanto, apenas aceitar o bloco
	response := map[string]interface{}{
		"status":  "accepted",
		"message": "Bloco recebido com sucesso",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// loadTestnetConfig carrega configuração da testnet
func (n *ORDMNode) loadTestnetConfig() error {
	// Implementar carregamento de config/testnet.json
	log.Printf("📋 Carregando configuração da testnet: %s", n.Config.ConfigFile)
	return nil
}

// loadGenesisBlock carrega bloco genesis
func (n *ORDMNode) loadGenesisBlock() error {
	// Implementar carregamento de genesis/testnet.json
	log.Printf("🌱 Carregando bloco genesis: %s", n.Config.GenesisFile)
	return nil
}

// startAutoSync inicia sincronização automática
func (n *ORDMNode) startAutoSync() {
	ticker := time.NewTicker(n.Config.SyncInterval)
	defer ticker.Stop()

	for {
		n.mu.RLock()
		if !n.IsRunning {
			n.mu.RUnlock()
			break
		}
		n.mu.RUnlock()

		select {
		case <-ticker.C:
			if err := n.syncWithPeers(); err != nil {
				log.Printf("⚠️  Erro na sincronização: %v", err)
			}
		}
	}
}

// startAutoMining inicia mineração automática
func (n *ORDMNode) startAutoMining() {
	ticker := time.NewTicker(n.Config.BlockTime)
	defer ticker.Stop()

	for {
		n.mu.RLock()
		if !n.IsRunning {
			n.mu.RUnlock()
			break
		}
		n.mu.RUnlock()

		select {
		case <-ticker.C:
			if err := n.mineBlockAuto(); err != nil {
				log.Printf("⚠️  Erro na mineração: %v", err)
			}
		}
	}
}

// miningWorker worker de mineração
func (n *ORDMNode) miningWorker(workerID int) {
	log.Printf("🧵 Worker de mineração %d iniciado", workerID)

	for {
		n.mu.RLock()
		if !n.IsRunning {
			n.mu.RUnlock()
			break
		}
		n.mu.RUnlock()

		// Criar bloco candidato
		block := n.createCandidateBlock()

		// Minerar bloco
		if err := n.mineBlock(block, workerID); err != nil {
			log.Printf("⚠️  Worker %d: Erro na mineração: %v", workerID, err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("🎉 Worker %d: Bloco minerado!", workerID)
	}
}

// createCandidateBlock cria um bloco candidato
func (n *ORDMNode) createCandidateBlock() *blockchain.Block {
	// Criar bloco com transações do mempool
	block := blockchain.NewBlock(n.Blockchain.Hash, []blockchain.Transaction{}, n.Config.MinerName, n.Config.Difficulty)
	return block
}

// mineBlock executa a mineração do bloco
func (n *ORDMNode) mineBlock(block *blockchain.Block, workerID int) error {
	startTime := time.Now()
	nonce := uint64(0)

	for {
		n.mu.RLock()
		if !n.IsRunning {
			n.mu.RUnlock()
			return fmt.Errorf("mineração interrompida")
		}
		n.mu.RUnlock()

		block.Header.Nonce = nonce
		block.Nonce = nonce

		// Calcular hash (simplificado)
		hash := fmt.Sprintf("hash_%d_%d", block.Header.Timestamp, nonce)

		// Verificar se o hash atende à dificuldade
		if n.checkDifficulty(hash) {
			block.Hash = hash
			duration := time.Since(startTime)
			log.Printf("🎯 Worker %d: Bloco minerado em %v (nonce: %d)", workerID, duration, nonce)
			return nil
		}

		nonce++
	}
}

// mineBlockAuto executa mineração automática (sem workerID)
func (n *ORDMNode) mineBlockAuto() error {
	block := n.createCandidateBlock()
	return n.mineBlock(block, 0)
}

// checkDifficulty verifica se o hash atende à dificuldade
func (n *ORDMNode) checkDifficulty(hash string) bool {
	// Verificar se o hash começa com zeros suficientes
	zeros := 0
	for _, char := range hash {
		if char == '0' {
			zeros++
		} else {
			break
		}
	}
	return zeros >= int(n.Config.Difficulty)
}

// syncWithPeers sincroniza com peers
func (n *ORDMNode) syncWithPeers() error {
	// Implementar sincronização com peers P2P
	return nil
}

// GetStatus retorna status do node
func (n *ORDMNode) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"network_id": n.Config.NetworkID,
		"is_running": n.IsRunning,
		"block_hash": n.Blockchain.Hash,
		"data_path":  n.DataPath,
		"mining":     n.Config.MiningEnabled,
		"machine_id": n.MachineID.ID,
		"miner_id":   n.Config.MinerKey,
	}
}

func main() {
	// Definir flags
	networkID := flag.String("network", "testnet", "ID da rede (testnet/mainnet)")
	port := flag.String("port", "8080", "Porta HTTP")
	p2pPort := flag.String("p2p-port", "3000", "Porta P2P")
	rpcPort := flag.String("rpc-port", "8081", "Porta RPC")
	dataPath := flag.String("data", "./data", "Caminho para dados")
	configFile := flag.String("config", "", "Arquivo de configuração")
	genesisFile := flag.String("genesis", "", "Arquivo do bloco genesis")
	maxPeers := flag.Int("max-peers", 50, "Máximo de peers")
	blockTime := flag.Duration("block-time", 0, "Tempo entre blocos (0 = sem mineração automática)")
	difficulty := flag.Uint64("difficulty", 4, "Dificuldade de mineração")

	// Flags de mineração
	miningEnabled := flag.Bool("mining", false, "Habilitar mineração")
	minerKey := flag.String("miner-key", "", "Chave privada do minerador")
	minerThreads := flag.Int("miner-threads", 1, "Número de threads de mineração")
	minerName := flag.String("miner-name", "ordm-node", "Nome do minerador")

	// Modo de operação
	mode := flag.String("mode", "node", "Modo de operação (node/miner/both)")

	flag.Parse()

	// Configurar caminho de dados padrão para testnet
	if *networkID == "testnet" && *dataPath == "./data" {
		*dataPath = "./data/testnet"
	}

	// Configurar arquivos padrão para testnet
	if *networkID == "testnet" {
		if *configFile == "" {
			*configFile = "config/testnet.json"
		}
		if *genesisFile == "" {
			*genesisFile = "genesis/testnet.json"
		}
	}

	// Configurar mineração baseado no modo
	if *mode == "miner" || *mode == "both" {
		*miningEnabled = true
		if *minerKey == "" {
			*minerKey = "miner_key_default"
		}
	}

	// Criar configuração do node
	config := &NodeConfig{
		NetworkID:    *networkID,
		Port:         *port,
		P2PPort:      *p2pPort,
		RPCPort:      *rpcPort,
		DataPath:     *dataPath,
		MaxPeers:     *maxPeers,
		BlockTime:    *blockTime,
		Difficulty:   *difficulty,
		GenesisFile:  *genesisFile,
		ConfigFile:   *configFile,
		SyncInterval: 30 * time.Second,
		SeedNodes:    []string{}, // Será carregado do config file

		// Configurações de mineração
		MiningEnabled: *miningEnabled,
		MinerKey:      *minerKey,
		MinerThreads:  *minerThreads,
		MinerName:     *minerName,
	}

	// Criar node
	node, err := NewORDMNode(config)
	if err != nil {
		log.Fatalf("❌ Erro ao criar node: %v", err)
	}

	// Configurar graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar node
	if err := node.Start(); err != nil {
		log.Fatalf("❌ Erro ao iniciar node: %v", err)
	}

	// Aguardar sinal de parada
	<-sigChan
	log.Printf("📡 Recebido sinal de parada...")

	// Parar node
	if err := node.Stop(); err != nil {
		log.Fatalf("❌ Erro ao parar node: %v", err)
	}
}
