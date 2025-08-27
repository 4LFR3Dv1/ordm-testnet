package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ordm-main/pkg/ledger"
	"ordm-main/pkg/logger"
	"ordm-main/pkg/storage"
)

// RESTAPI implementa a API REST pública
type RESTAPI struct {
	ledger    *ledger.GlobalLedger
	storage   *storage.BadgerStore
	logger    *logger.Logger
	port      int
	rateLimit *RateLimiter
}

// RateLimiter implementa limitação de taxa
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// IsAllowed verifica se uma requisição é permitida
func (rl *RateLimiter) IsAllowed(clientID string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Limpar requisições antigas
	var validRequests []time.Time
	for _, reqTime := range rl.requests[clientID] {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests[clientID] = validRequests

	// Verificar limite
	if len(validRequests) >= rl.limit {
		return false
	}

	// Adicionar nova requisição
	rl.requests[clientID] = append(rl.requests[clientID], now)
	return true
}

// NewRESTAPI cria uma nova instância da API REST
func NewRESTAPI(ledger *ledger.GlobalLedger, storage *storage.BadgerStore, logger *logger.Logger, port int) *RESTAPI {
	return &RESTAPI{
		ledger:    ledger,
		storage:   storage,
		logger:    logger,
		port:      port,
		rateLimit: NewRateLimiter(100, time.Minute), // 100 req/min
	}
}

// Start inicia o servidor da API
func (api *RESTAPI) Start() error {
	// Configurar rotas
	http.HandleFunc("/api/v1/health", api.handleHealth)
	http.HandleFunc("/api/v1/blocks", api.handleBlocks)
	http.HandleFunc("/api/v1/blocks/", api.handleBlockByID)
	http.HandleFunc("/api/v1/transactions", api.handleTransactions)
	http.HandleFunc("/api/v1/transactions/", api.handleTransactionByID)
	http.HandleFunc("/api/v1/balances", api.handleBalances)
	http.HandleFunc("/api/v1/balances/", api.handleBalanceByAddress)
	http.HandleFunc("/api/v1/stakes", api.handleStakes)
	http.HandleFunc("/api/v1/stakes/", api.handleStakeByAddress)
	http.HandleFunc("/api/v1/nodes", api.handleNodes)
	http.HandleFunc("/api/v1/nodes/", api.handleNodeByID)
	http.HandleFunc("/api/v1/stats", api.handleStats)
	http.HandleFunc("/api/v1/send", api.handleSendTransaction)
	http.HandleFunc("/api/v1/stake", api.handleStake)
	http.HandleFunc("/api/v1/explorer", api.handleExplorer)

	// Middleware de CORS
	http.HandleFunc("/api/", api.corsMiddleware)

	addr := fmt.Sprintf(":%d", api.port)
	api.logger.Info("API REST iniciada", map[string]interface{}{
		"port": api.port,
		"endpoints": []string{
			"/api/v1/health",
			"/api/v1/blocks",
			"/api/v1/transactions",
			"/api/v1/balances",
			"/api/v1/stakes",
			"/api/v1/nodes",
			"/api/v1/stats",
			"/api/v1/send",
			"/api/v1/stake",
			"/api/v1/explorer",
		},
	})

	return http.ListenAndServe(addr, nil)
}

// corsMiddleware adiciona headers CORS
func (api *RESTAPI) corsMiddleware(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// rateLimitMiddleware aplica limitação de taxa
func (api *RESTAPI) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.RemoteAddr
		if !api.rateLimit.IsAllowed(clientID) {
			api.sendError(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// sendResponse envia resposta JSON
func (api *RESTAPI) sendResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// sendError envia erro JSON
func (api *RESTAPI) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success":   false,
		"error":     message,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleHealth verifica saúde da API
func (api *RESTAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	health := map[string]interface{}{
		"status":    "healthy",
		"version":   "1.0.0",
		"uptime":    time.Since(start).String(),
		"timestamp": time.Now().Unix(),
	}

	api.sendResponse(w, health, http.StatusOK)
}

// handleBlocks retorna lista de blocos
func (api *RESTAPI) handleBlocks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parâmetros de paginação
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	if page == 0 {
		page = 1
	}

	// Buscar blocos do storage
	blocks, err := api.storage.GetAllBlocks()
	if err != nil {
		api.logger.Error("Erro ao buscar blocos", err, nil)
		api.sendError(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	// Paginar resultados
	offset := (page - 1) * limit
	blockList := make([]map[string]interface{}, 0)

	count := 0
	for _, blockData := range blocks {
		if count >= offset && len(blockList) < limit {
			var block map[string]interface{}
			if err := json.Unmarshal(blockData, &block); err == nil {
				blockList = append(blockList, block)
			}
		}
		count++
	}

	response := map[string]interface{}{
		"blocks":     blockList,
		"page":       page,
		"limit":      limit,
		"total":      len(blocks),
		"has_more":   len(blockList) == limit,
		"query_time": time.Since(start).Milliseconds(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleBlockByID retorna bloco específico
func (api *RESTAPI) handleBlockByID(w http.ResponseWriter, r *http.Request) {
	blockID := r.URL.Path[len("/api/v1/blocks/"):]

	blockData, err := api.storage.GetBlock(blockID)
	if err != nil {
		api.sendError(w, "Bloco não encontrado", http.StatusNotFound)
		return
	}

	var block map[string]interface{}
	if err := json.Unmarshal(blockData, &block); err != nil {
		api.sendError(w, "Erro ao processar bloco", http.StatusInternalServerError)
		return
	}

	api.sendResponse(w, block, http.StatusOK)
}

// handleTransactions retorna lista de transações
func (api *RESTAPI) handleTransactions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parâmetros de filtro
	status := r.URL.Query().Get("status")
	address := r.URL.Query().Get("address")

	// Buscar transações do storage
	transactions, err := api.storage.GetAllTransactions()
	if err != nil {
		api.logger.Error("Erro ao buscar transações", err, nil)
		api.sendError(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	// Filtrar transações
	txList := make([]map[string]interface{}, 0)
	for _, txData := range transactions {
		var tx map[string]interface{}
		if err := json.Unmarshal(txData, &tx); err == nil {
			// Aplicar filtros
			if status != "" && tx["status"] != status {
				continue
			}
			if address != "" && tx["from"] != address && tx["to"] != address {
				continue
			}
			txList = append(txList, tx)
		}
	}

	response := map[string]interface{}{
		"transactions": txList,
		"total":        len(txList),
		"query_time":   time.Since(start).Milliseconds(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleTransactionByID retorna transação específica
func (api *RESTAPI) handleTransactionByID(w http.ResponseWriter, r *http.Request) {
	txID := r.URL.Path[len("/api/v1/transactions/"):]

	txData, err := api.storage.GetTransaction(txID)
	if err != nil {
		api.sendError(w, "Transação não encontrada", http.StatusNotFound)
		return
	}

	var tx map[string]interface{}
	if err := json.Unmarshal(txData, &tx); err != nil {
		api.sendError(w, "Erro ao processar transação", http.StatusInternalServerError)
		return
	}

	api.sendResponse(w, tx, http.StatusOK)
}

// handleBalances retorna lista de saldos
func (api *RESTAPI) handleBalances(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	balances, err := api.storage.GetAllBalances()
	if err != nil {
		api.logger.Error("Erro ao buscar saldos", err, nil)
		api.sendError(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	balanceList := make([]map[string]interface{}, 0)
	for _, balanceData := range balances {
		var balance map[string]interface{}
		if err := json.Unmarshal(balanceData, &balance); err == nil {
			balanceList = append(balanceList, balance)
		}
	}

	response := map[string]interface{}{
		"balances":   balanceList,
		"total":      len(balanceList),
		"query_time": time.Since(start).Milliseconds(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleBalanceByAddress retorna saldo específico
func (api *RESTAPI) handleBalanceByAddress(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Path[len("/api/v1/balances/"):]

	balance := api.ledger.GetBalance(address)

	response := map[string]interface{}{
		"address":   address,
		"balance":   balance,
		"timestamp": time.Now().Unix(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleStakes retorna lista de stakes
func (api *RESTAPI) handleStakes(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	stakes, err := api.storage.GetAllStakes()
	if err != nil {
		api.logger.Error("Erro ao buscar stakes", err, nil)
		api.sendError(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	stakeList := make([]map[string]interface{}, 0)
	for _, stakeData := range stakes {
		var stake map[string]interface{}
		if err := json.Unmarshal(stakeData, &stake); err == nil {
			stakeList = append(stakeList, stake)
		}
	}

	response := map[string]interface{}{
		"stakes":     stakeList,
		"total":      len(stakeList),
		"query_time": time.Since(start).Milliseconds(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleStakeByAddress retorna stake específico
func (api *RESTAPI) handleStakeByAddress(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Path[len("/api/v1/stakes/"):]

	stakeData, err := api.storage.GetStake(address)
	if err != nil {
		api.sendError(w, "Stake não encontrado", http.StatusNotFound)
		return
	}

	var stake map[string]interface{}
	if err := json.Unmarshal(stakeData, &stake); err != nil {
		api.sendError(w, "Erro ao processar stake", http.StatusInternalServerError)
		return
	}

	api.sendResponse(w, stake, http.StatusOK)
}

// handleNodes retorna lista de nodes
func (api *RESTAPI) handleNodes(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	nodes, err := api.storage.GetAllNodes()
	if err != nil {
		api.logger.Error("Erro ao buscar nodes", err, nil)
		api.sendError(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	nodeList := make([]map[string]interface{}, 0)
	for _, nodeData := range nodes {
		var node map[string]interface{}
		if err := json.Unmarshal(nodeData, &node); err == nil {
			nodeList = append(nodeList, node)
		}
	}

	response := map[string]interface{}{
		"nodes":      nodeList,
		"total":      len(nodeList),
		"query_time": time.Since(start).Milliseconds(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleNodeByID retorna node específico
func (api *RESTAPI) handleNodeByID(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Path[len("/api/v1/nodes/"):]

	nodeData, err := api.storage.GetNode(nodeID)
	if err != nil {
		api.sendError(w, "Node não encontrado", http.StatusNotFound)
		return
	}

	var node map[string]interface{}
	if err := json.Unmarshal(nodeData, &node); err != nil {
		api.sendError(w, "Erro ao processar node", http.StatusInternalServerError)
		return
	}

	api.sendResponse(w, node, http.StatusOK)
}

// handleStats retorna estatísticas gerais
func (api *RESTAPI) handleStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Buscar dados do storage
	blocks, _ := api.storage.GetAllBlocks()
	transactions, _ := api.storage.GetAllTransactions()
	balances, _ := api.storage.GetAllBalances()
	stakes, _ := api.storage.GetAllStakes()
	nodes, _ := api.storage.GetAllNodes()

	// Calcular estatísticas
	stats := api.ledger.GetStats()

	response := map[string]interface{}{
		"blocks":       len(blocks),
		"transactions": len(transactions),
		"addresses":    len(balances),
		"stakes":       len(stakes),
		"nodes":        len(nodes),
		"total_supply": stats["total_supply"],
		"query_time":   time.Since(start).Milliseconds(),
		"timestamp":    time.Now().Unix(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleSendTransaction processa nova transação
func (api *RESTAPI) handleSendTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.sendError(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
		Fee    int64  `json:"fee"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	// Validar dados
	if request.From == "" || request.To == "" || request.Amount <= 0 {
		api.sendError(w, "Dados obrigatórios inválidos", http.StatusBadRequest)
		return
	}

	// Processar transação (implementar lógica real)
	response := map[string]interface{}{
		"tx_hash":   fmt.Sprintf("tx_%d", time.Now().UnixNano()),
		"status":    "pending",
		"timestamp": time.Now().Unix(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleStake processa novo stake
func (api *RESTAPI) handleStake(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.sendError(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	// Validar dados
	if request.Address == "" || request.Amount <= 0 {
		api.sendError(w, "Dados obrigatórios inválidos", http.StatusBadRequest)
		return
	}

	// Processar stake (implementar lógica real)
	response := map[string]interface{}{
		"stake_id":  fmt.Sprintf("stake_%d", time.Now().UnixNano()),
		"status":    "pending",
		"timestamp": time.Now().Unix(),
	}

	api.sendResponse(w, response, http.StatusOK)
}

// handleExplorer retorna dados para explorador de blocos
func (api *RESTAPI) handleExplorer(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Buscar dados para o explorador
	blocks, _ := api.storage.GetAllBlocks()
	transactions, _ := api.storage.GetAllTransactions()
	balances, _ := api.storage.GetAllBalances()

	// Últimos blocos
	recentBlocks := make([]map[string]interface{}, 0)
	count := 0
	for _, blockData := range blocks {
		if count >= 10 { // Últimos 10 blocos
			break
		}
		var block map[string]interface{}
		if err := json.Unmarshal(blockData, &block); err == nil {
			recentBlocks = append(recentBlocks, block)
			count++
		}
	}

	// Últimas transações
	recentTxs := make([]map[string]interface{}, 0)
	count = 0
	for _, txData := range transactions {
		if count >= 20 { // Últimas 20 transações
			break
		}
		var tx map[string]interface{}
		if err := json.Unmarshal(txData, &tx); err == nil {
			recentTxs = append(recentTxs, tx)
			count++
		}
	}

	// Top holders
	topHolders := make([]map[string]interface{}, 0)
	count = 0
	for _, balanceData := range balances {
		if count >= 10 { // Top 10 holders
			break
		}
		var balance map[string]interface{}
		if err := json.Unmarshal(balanceData, &balance); err == nil {
			topHolders = append(topHolders, balance)
			count++
		}
	}

	response := map[string]interface{}{
		"recent_blocks":   recentBlocks,
		"recent_txs":      recentTxs,
		"top_holders":     topHolders,
		"total_blocks":    len(blocks),
		"total_txs":       len(transactions),
		"total_addresses": len(balances),
		"query_time":      time.Since(start).Milliseconds(),
		"timestamp":       time.Now().Unix(),
	}

	api.sendResponse(w, response, http.StatusOK)
}
