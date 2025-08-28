package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ordm-main/pkg/consensus"
	"ordm-main/pkg/mempool"
)

// SimplePublicAPI implementa APIs públicas simplificadas para usuários externos
type SimplePublicAPI struct {
	mempool        *mempool.DistributedMempool
	forkResolution *consensus.ForkResolution
	logger         func(string, ...interface{})
}

// NewSimplePublicAPI cria uma nova instância da API pública simplificada
func NewSimplePublicAPI(mempool *mempool.DistributedMempool, forkResolution *consensus.ForkResolution, logger func(string, ...interface{})) *SimplePublicAPI {
	return &SimplePublicAPI{
		mempool:        mempool,
		forkResolution: forkResolution,
		logger:         logger,
	}
}

// SetupRoutes configura as rotas públicas da API
func (api *SimplePublicAPI) SetupRoutes() {
	// Informações da blockchain
	http.HandleFunc("/api/v1/blockchain/info", api.handleBlockchainInfo)
	http.HandleFunc("/api/v1/blockchain/status", api.handleBlockchainStatus)

	// Transações
	http.HandleFunc("/api/v1/transactions/send", api.handleSendTransaction)
	http.HandleFunc("/api/v1/transactions/pending", api.handleGetPendingTransactions)

	// Mempool
	http.HandleFunc("/api/v1/mempool/status", api.handleMempoolStatus)
	http.HandleFunc("/api/v1/mempool/transactions", api.handleMempoolTransactions)

	// Consenso e forks
	http.HandleFunc("/api/v1/consensus/status", api.handleConsensusStatus)
	http.HandleFunc("/api/v1/consensus/forks", api.handleGetForks)

	// Mineração
	http.HandleFunc("/api/v1/mining/status", api.handleMiningStatus)
	http.HandleFunc("/api/v1/mining/start", api.handleStartMining)
	http.HandleFunc("/api/v1/mining/stop", api.handleStopMining)

	// Estatísticas
	http.HandleFunc("/api/v1/stats", api.handleGetStats)
	http.HandleFunc("/api/v1/stats/network", api.handleGetNetworkStats)

	// Wallets
	http.HandleFunc("/api/v1/wallets/create", api.handleCreateWallet)
}

// ===== HANDLERS DA BLOCKCHAIN =====

// handleBlockchainInfo retorna informações gerais da blockchain
func (api *SimplePublicAPI) handleBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]interface{}{
		"name":               "ORDM Blockchain 2-Layer",
		"version":            "1.0.0",
		"consensus":          "PoW + PoS Hybrid",
		"block_time":         "10 seconds",
		"difficulty":         "Dynamic",
		"total_blocks":       0,
		"total_transactions": 0,
		"network_status":     "active",
		"timestamp":          time.Now().Unix(),
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    info,
	})
}

// handleBlockchainStatus retorna o status atual da blockchain
func (api *SimplePublicAPI) handleBlockchainStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"synced":          true,
		"last_block":      0,
		"last_block_time": time.Now().Unix(),
		"peers_count":     0,
		"difficulty":      4,
		"hash_rate":       0.0,
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// ===== HANDLERS DE TRANSAÇÕES =====

// handleSendTransaction envia uma nova transação
func (api *SimplePublicAPI) handleSendTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		From      string `json:"from"`
		To        string `json:"to"`
		Amount    int64  `json:"amount"`
		Fee       int64  `json:"fee"`
		Data      string `json:"data"`
		Signature string `json:"signature"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validações básicas
	if request.From == "" || request.To == "" {
		http.Error(w, "from and to are required", http.StatusBadRequest)
		return
	}

	if request.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if request.Signature == "" {
		http.Error(w, "signature is required", http.StatusBadRequest)
		return
	}

	var response map[string]interface{}

	if api.mempool != nil {
		// Gerar ID da transação
		txID := mempool.GenerateTransactionID(
			request.From, request.To, request.Amount, request.Fee, 0, time.Now().Unix(),
		)

		// Criar transação
		tx := &mempool.Transaction{
			ID:        txID,
			From:      request.From,
			To:        request.To,
			Amount:    request.Amount,
			Fee:       request.Fee,
			Data:      request.Data,
			Signature: []byte(request.Signature),
			Timestamp: time.Now().Unix(),
			Status:    "pending",
		}

		// Adicionar ao mempool
		if err := api.mempool.AddTransaction(tx); err != nil {
			response = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			response = map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"tx_id":   txID,
					"status":  "pending",
					"message": "Transaction added to mempool",
				},
			}
		}
	} else {
		response = map[string]interface{}{
			"success": false,
			"error":   "Mempool not available",
		}
	}

	api.sendJSONResponse(w, response)
}

// handleGetPendingTransactions retorna transações pendentes
func (api *SimplePublicAPI) handleGetPendingTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var transactions []map[string]interface{}

	if api.mempool != nil {
		pendingTxs := api.mempool.GetPendingTransactions(limit)
		for _, tx := range pendingTxs {
			transactions = append(transactions, map[string]interface{}{
				"id":        tx.ID,
				"from":      tx.From,
				"to":        tx.To,
				"amount":    tx.Amount,
				"fee":       tx.Fee,
				"timestamp": tx.Timestamp,
				"status":    tx.Status,
			})
		}
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"transactions": transactions,
			"total":        len(transactions),
			"limit":        limit,
		},
	})
}

// ===== HANDLERS DE MINERAÇÃO =====

// handleMiningStatus retorna o status da mineração
func (api *SimplePublicAPI) handleMiningStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"mining":     true, // TODO: Conectar com status real do node
		"difficulty": 4,
		"hash_rate":  0.0,
		"miner_id":   "",
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// ===== HANDLERS DE ESTATÍSTICAS =====

// handleGetStats retorna estatísticas gerais
func (api *SimplePublicAPI) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"total_blocks":         0,
		"total_transactions":   0,
		"pending_transactions": 0,
		"total_wallets":        0,
		"network_uptime":       0,
		"difficulty":           4,
		"hash_rate":            0.0,
	}

	// Adicionar dados do mempool se disponível
	if api.mempool != nil {
		mempoolStats := api.mempool.GetMempoolStats()
		if pending, ok := mempoolStats["pending"].(int); ok {
			stats["pending_transactions"] = pending
		}
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ===== HANDLERS AUXILIARES =====

// sendJSONResponse envia uma resposta JSON
func (api *SimplePublicAPI) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode ...int) {
	w.Header().Set("Content-Type", "application/json")

	if len(statusCode) > 0 {
		w.WriteHeader(statusCode[0])
	}

	json.NewEncoder(w).Encode(data)
}

// handleMempoolStatus retorna status do mempool
func (api *SimplePublicAPI) handleMempoolStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stats map[string]interface{}

	if api.mempool != nil {
		stats = api.mempool.GetMempoolStats()
	} else {
		stats = map[string]interface{}{
			"total_transactions": 0,
			"pending":            0,
			"confirmed":          0,
			"failed":             0,
		}
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// handleMempoolTransactions retorna transações do mempool
func (api *SimplePublicAPI) handleMempoolTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var transactions []map[string]interface{}

	if api.mempool != nil {
		pendingTxs := api.mempool.GetPendingTransactions(limit)
		for _, tx := range pendingTxs {
			transactions = append(transactions, map[string]interface{}{
				"id":        tx.ID,
				"from":      tx.From,
				"to":        tx.To,
				"amount":    tx.Amount,
				"fee":       tx.Fee,
				"timestamp": tx.Timestamp,
				"status":    tx.Status,
			})
		}
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"transactions": transactions,
			"total":        len(transactions),
			"limit":        limit,
		},
	})
}

// handleConsensusStatus retorna status do consenso
func (api *SimplePublicAPI) handleConsensusStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stats map[string]interface{}

	if api.forkResolution != nil {
		stats = api.forkResolution.GetForkStats()
	} else {
		stats = map[string]interface{}{
			"total_forks":     0,
			"resolved_forks":  0,
			"detected_forks":  0,
			"resolution_rate": 0.0,
		}
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// handleGetForks retorna forks conhecidos
func (api *SimplePublicAPI) handleGetForks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var forks []interface{}

	if api.forkResolution != nil {
		forkList := api.forkResolution.GetAllForks()
		for _, fork := range forkList {
			forks = append(forks, fork)
		}
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"forks": forks,
			"count": len(forks),
		},
	})
}

// handleStartMining inicia mineração
func (api *SimplePublicAPI) handleStartMining(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Mining started",
	})
}

// handleStopMining para mineração
func (api *SimplePublicAPI) handleStopMining(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Mining stopped",
	})
}

// handleGetNetworkStats retorna estatísticas da rede
func (api *SimplePublicAPI) handleGetNetworkStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"peers_count":    0,
		"connections":    0,
		"network_uptime": 0,
		"sync_status":    "synced",
		"last_sync":      time.Now().Unix(),
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// handleCreateWallet cria uma nova wallet
func (api *SimplePublicAPI) handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Gerar nova wallet (implementação básica)
	walletData := map[string]interface{}{
		"address":     fmt.Sprintf("wallet_%d", time.Now().Unix()),
		"private_key": "generated_private_key",
		"public_key":  "generated_public_key",
		"created_at":  time.Now().Unix(),
	}

	api.sendJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    walletData,
	})
}
