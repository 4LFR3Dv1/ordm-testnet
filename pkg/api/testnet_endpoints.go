package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ordm-main/pkg/faucet"
	"ordm-main/pkg/network"
)

// TestnetEndpoints gerencia endpoints específicos da testnet
type TestnetEndpoints struct {
	FaucetManager   *faucet.FaucetManager
	SeedNodeManager *network.SeedNodeManager
	RateLimiter     *RateLimiter
}

// NewTestnetEndpoints cria novos endpoints da testnet
func NewTestnetEndpoints() *TestnetEndpoints {
	return &TestnetEndpoints{
		FaucetManager:   faucet.NewFaucetManager(),
		SeedNodeManager: network.NewSeedNodeManager(),
		RateLimiter:     NewRateLimiter(100, time.Minute), // 100 req/min
	}
}

// RegisterTestnetEndpoints registra endpoints da testnet
func (te *TestnetEndpoints) RegisterTestnetEndpoints(mux *http.ServeMux) {
	// Inicializar seed nodes da testnet
	te.SeedNodeManager.InitializeTestnetSeedNodes()

	// Iniciar serviços em background
	go te.SeedNodeManager.StartSeedNodeHeartbeat()
	go te.FaucetManager.StartFaucetCleanup()

	// Endpoints da testnet
	mux.HandleFunc("/api/testnet/faucet", te.handleFaucet)
	mux.HandleFunc("/api/testnet/faucet/stats", te.handleFaucetStats)
	mux.HandleFunc("/api/testnet/faucet/history", te.handleFaucetHistory)
	mux.HandleFunc("/api/testnet/seed-nodes", te.handleSeedNodes)
	mux.HandleFunc("/api/testnet/network", te.handleNetworkInfo)
	mux.HandleFunc("/api/testnet/peers", te.handleBootstrapPeers)
	mux.HandleFunc("/api/testnet/status", te.handleTestnetStatus)
}

// handleFaucet processa requisições ao faucet
func (te *TestnetEndpoints) handleFaucet(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	if !te.RateLimiter.IsAllowed(r.RemoteAddr) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var request struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validar campos obrigatórios
	if request.Address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	if request.Amount <= 0 {
		request.Amount = 50 // Default amount
	}

	// Extrair IP real (considerando proxy)
	ip := te.getRealIP(r)

	// Processar requisição
	faucetReq, err := te.FaucetManager.ProcessFaucetRequest(request.Address, ip, request.Amount)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"request_hash": faucetReq.RequestHash,
			"address":      faucetReq.Address,
			"amount":       faucetReq.Amount,
			"status":       faucetReq.Status,
			"timestamp":    faucetReq.Timestamp,
		},
		"message": fmt.Sprintf("💰 %d tokens enviados para %s", request.Amount, request.Address),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleFaucetStats retorna estatísticas do faucet
func (te *TestnetEndpoints) handleFaucetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := te.FaucetManager.GetFaucetStats()

	response := map[string]interface{}{
		"success": true,
		"data":    stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleFaucetHistory retorna histórico do faucet
func (te *TestnetEndpoints) handleFaucetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	history := te.FaucetManager.GetRequestHistory(limit)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"requests": history,
			"limit":    limit,
			"total":    len(history),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSeedNodes retorna informações dos seed nodes
func (te *TestnetEndpoints) handleSeedNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activeSeeds := te.SeedNodeManager.GetActiveSeedNodes()
	seedInfo := te.SeedNodeManager.GetSeedNodesInfo()

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"seed_nodes":   activeSeeds,
			"network_info": seedInfo,
			"total_seeds":  seedInfo["total_seeds"],
			"active_seeds": seedInfo["active_seeds"],
			"total_peers":  seedInfo["total_peers"],
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNetworkInfo retorna informações da rede testnet
func (te *TestnetEndpoints) handleNetworkInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	networkInfo := map[string]interface{}{
		"network": "testnet",
		"version": "1.0.0",
		"status":  "active",
		"uptime":  time.Now().Unix(),
		"config": map[string]interface{}{
			"testnet_mode": true,
			"max_peers":    50,
			"heartbeat":    30,
			"timeout":      60,
		},
		"seed_nodes": te.SeedNodeManager.GetSeedNodesInfo(),
		"faucet":     te.FaucetManager.GetFaucetStats(),
	}

	response := map[string]interface{}{
		"success": true,
		"data":    networkInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleBootstrapPeers retorna lista de peers bootstrap
func (te *TestnetEndpoints) handleBootstrapPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peers := te.SeedNodeManager.GetBootstrapPeers()

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"bootstrap_peers": peers,
			"count":           len(peers),
			"network":         "testnet",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTestnetStatus retorna status geral da testnet
func (te *TestnetEndpoints) handleTestnetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Calcular status geral
	seedInfo := te.SeedNodeManager.GetSeedNodesInfo()
	faucetStats := te.FaucetManager.GetFaucetStats()

	activeSeeds := seedInfo["active_seeds"].(int)
	totalSeeds := seedInfo["total_seeds"].(int)

	status := "healthy"
	if activeSeeds < totalSeeds/2 {
		status = "degraded"
	}
	if activeSeeds == 0 {
		status = "down"
	}

	networkStatus := map[string]interface{}{
		"network": "testnet",
		"status":  status,
		"version": "1.0.0",
		"uptime":  time.Now().Unix(),
		"seed_nodes": map[string]interface{}{
			"total":   totalSeeds,
			"active":  activeSeeds,
			"healthy": activeSeeds > 0,
		},
		"faucet": map[string]interface{}{
			"balance":     faucetStats["faucet_balance"],
			"total_sent":  faucetStats["total_sent"],
			"requests":    faucetStats["total_requests"],
			"operational": faucetStats["faucet_balance"].(int64) > 0,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	response := map[string]interface{}{
		"success": true,
		"data":    networkStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getRealIP extrai o IP real considerando proxies
func (te *TestnetEndpoints) getRealIP(r *http.Request) string {
	// Verificar headers de proxy
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// Pegar o primeiro IP da lista
		if commaIndex := strings.Index(ip, ","); commaIndex != -1 {
			return strings.TrimSpace(ip[:commaIndex])
		}
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Client-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fallback para RemoteAddr
	if r.RemoteAddr != "" {
		if colonIndex := strings.LastIndex(r.RemoteAddr, ":"); colonIndex != -1 {
			return r.RemoteAddr[:colonIndex]
		}
		return r.RemoteAddr
	}

	return "unknown"
}
