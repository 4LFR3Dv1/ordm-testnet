#!/bin/bash

# 🏗️ Script: Parte 2.1 - Separação Frontend/Backend
# Descrição: Implementa API REST separada, middleware chain e service layer

set -e

echo "🏗️ [$(date)] Iniciando Parte 2.1: Separação Frontend/Backend"
echo "============================================================"

# 2.1.1 - API REST Separada
echo "🔗 2.1.1 - Implementando API REST Separada..."

mkdir -p pkg/api
cat > pkg/api/rest.go << 'EOF'
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"ordm-main/pkg/auth"
	"ordm-main/pkg/config"
	"ordm-main/pkg/middleware"
	"ordm-main/pkg/services"
)

type APIServer struct {
	router *mux.Router
	auth   *auth.Manager
	mining *services.MiningService
	wallet *services.WalletService
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewAPIServer(auth *auth.Manager, mining *services.MiningService, wallet *services.WalletService) *APIServer {
	router := mux.NewRouter()
	
	api := &APIServer{
		router: router,
		auth:   auth,
		mining: mining,
		wallet: wallet,
	}
	
	api.setupRoutes()
	return api
}

func (api *APIServer) setupRoutes() {
	// Middleware global
	api.router.Use(middleware.LoggingMiddleware)
	api.router.Use(middleware.CORSMiddleware)
	api.router.Use(middleware.SecurityHeaders)
	
	// Rotas públicas
	api.router.HandleFunc("/api/health", api.healthCheck).Methods("GET")
	api.router.HandleFunc("/api/login", api.login).Methods("POST")
	api.router.HandleFunc("/api/register", api.register).Methods("POST")
	
	// Rotas protegidas
	protected := api.router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(api.auth))
	
	protected.HandleFunc("/status", api.getStatus).Methods("GET")
	protected.HandleFunc("/mining/start", api.startMining).Methods("POST")
	protected.HandleFunc("/mining/stop", api.stopMining).Methods("POST")
	protected.HandleFunc("/wallet/create", api.createWallet).Methods("POST")
	protected.HandleFunc("/wallet/balance", api.getBalance).Methods("GET")
	protected.HandleFunc("/wallet/stake", api.stakeTokens).Methods("POST")
}

func (api *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Message: "API funcionando",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		},
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) login(w http.ResponseWriter, r *http.Request) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		api.sendError(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	
	session, err := api.auth.AuthenticateUser(loginData.Username, loginData.Password, r.RemoteAddr, r.UserAgent())
	if err != nil {
		api.sendError(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: "Login realizado com sucesso",
		Data: map[string]interface{}{
			"token": session.Token,
			"user_id": session.UserID,
		},
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) getStatus(w http.ResponseWriter, r *http.Request) {
	status := api.mining.GetStatus()
	
	response := APIResponse{
		Success: true,
		Data:    status,
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) startMining(w http.ResponseWriter, r *http.Request) {
	err := api.mining.StartMining()
	if err != nil {
		api.sendError(w, "Erro ao iniciar mineração", http.StatusInternalServerError)
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: "Mineração iniciada",
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) stopMining(w http.ResponseWriter, r *http.Request) {
	err := api.mining.StopMining()
	if err != nil {
		api.sendError(w, "Erro ao parar mineração", http.StatusInternalServerError)
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: "Mineração parada",
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) createWallet(w http.ResponseWriter, r *http.Request) {
	wallet, err := api.wallet.CreateWallet()
	if err != nil {
		api.sendError(w, "Erro ao criar wallet", http.StatusInternalServerError)
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: "Wallet criada com sucesso",
		Data:    wallet,
	}
	
	api.sendJSON(w, response, http.StatusCreated)
}

func (api *APIServer) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	balance, err := api.wallet.GetBalance(userID)
	if err != nil {
		api.sendError(w, "Erro ao obter saldo", http.StatusInternalServerError)
		return
	}
	
	response := APIResponse{
		Success: true,
		Data:    balance,
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) stakeTokens(w http.ResponseWriter, r *http.Request) {
	var stakeData struct {
		Amount int64 `json:"amount"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&stakeData); err != nil {
		api.sendError(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	
	userID := r.Context().Value("user_id").(string)
	err := api.wallet.StakeTokens(userID, stakeData.Amount)
	if err != nil {
		api.sendError(w, "Erro ao fazer stake", http.StatusInternalServerError)
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: "Stake realizado com sucesso",
	}
	
	api.sendJSON(w, response, http.StatusOK)
}

func (api *APIServer) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (api *APIServer) sendError(w http.ResponseWriter, message string, status int) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	api.sendJSON(w, response, status)
}

func (api *APIServer) GetRouter() *mux.Router {
	return api.router
}
EOF

# 2.1.2 - Middleware Chain
echo "🔗 2.1.2 - Implementando Middleware Chain..."

cat > pkg/middleware/chain.go << 'EOF'
package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"ordm-main/pkg/auth"
)

func AuthMiddleware(authManager *auth.Manager) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Token não fornecido", http.StatusUnauthorized)
				return
			}
			
			// Remover "Bearer " se presente
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
			
			session, valid := authManager.ValidateSession(token)
			if !valid {
				http.Error(w, "Token inválido", http.StatusUnauthorized)
				return
			}
			
			// Adicionar dados do usuário ao contexto
			ctx := context.WithValue(r.Context(), "user_id", session.UserID)
			ctx = context.WithValue(ctx, "session", session)
			
			next(w, r.WithContext(ctx))
		}
	}
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrapper para capturar status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next(ww, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d - %v", r.Method, r.RemoteAddr, r.URL.Path, ww.statusCode, duration)
	}
}

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

func RateLimitMiddleware(rateLimiter *auth.RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			identifier := r.RemoteAddr
			
			result := rateLimiter.CheckRateLimit(identifier)
			if !result.Allowed {
				http.Error(w, "Rate limit excedido", http.StatusTooManyRequests)
				return
			}
			
			next(w, r)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
EOF

# 2.1.3 - Service Layer
echo "🔧 2.1.3 - Implementando Service Layer..."

mkdir -p pkg/services
cat > pkg/services/mining_service.go << 'EOF'
package services

import (
	"sync"
	"time"

	"ordm-main/pkg/state"
	"ordm-main/pkg/wallet"
	"ordm-main/pkg/ledger"
)

type MiningService struct {
	state    *state.SafeNodeState
	wallet   *wallet.Manager
	ledger   *ledger.GlobalLedger
	mu       sync.RWMutex
	isActive bool
	stopChan chan struct{}
}

type MiningStatus struct {
	IsActive      bool      `json:"is_active"`
	TotalBlocks   int64     `json:"total_blocks"`
	HashRate      float64   `json:"hash_rate"`
	LastMined     time.Time `json:"last_mined"`
	TotalRewards  int64     `json:"total_rewards"`
	CurrentBlock  int64     `json:"current_block"`
}

func NewMiningService(state *state.SafeNodeState, wallet *wallet.Manager, ledger *ledger.GlobalLedger) *MiningService {
	return &MiningService{
		state:   state,
		wallet:  wallet,
		ledger:  ledger,
		stopChan: make(chan struct{}),
	}
}

func (ms *MiningService) StartMining() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if ms.isActive {
		return nil // Já está minerando
	}
	
	ms.isActive = true
	ms.stopChan = make(chan struct{})
	
	// Iniciar goroutine de mineração
	go ms.miningWorker()
	
	return nil
}

func (ms *MiningService) StopMining() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if !ms.isActive {
		return nil // Já está parado
	}
	
	ms.isActive = false
	close(ms.stopChan)
	
	return nil
}

func (ms *MiningService) GetStatus() *MiningStatus {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	nodeInfo := ms.state.GetState()
	
	return &MiningStatus{
		IsActive:     ms.isActive,
		TotalBlocks:  nodeInfo.MiningStats.TotalBlocks,
		HashRate:     nodeInfo.MiningStats.HashRate,
		LastMined:    nodeInfo.MiningStats.LastMined,
		TotalRewards: nodeInfo.MiningStats.TotalRewards,
		CurrentBlock: nodeInfo.MiningStats.TotalBlocks + 1,
	}
}

func (ms *MiningService) miningWorker() {
	ticker := time.NewTicker(10 * time.Second) // Mineração a cada 10 segundos
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if ms.isActive {
				ms.mineBlock()
			}
		case <-ms.stopChan:
			return
		}
	}
}

func (ms *MiningService) mineBlock() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	// Simular mineração de bloco
	nodeInfo := ms.state.GetState()
	nodeInfo.MiningStats.TotalBlocks++
	nodeInfo.MiningStats.LastMined = time.Now()
	nodeInfo.MiningStats.TotalRewards += 50 // Recompensa por bloco
	nodeInfo.MiningStats.HashRate = 1234.56 // Hash rate simulado
	
	ms.state.SetState(nodeInfo)
	
	// Atualizar ledger
	if ms.ledger != nil {
		ms.ledger.AddBlock(nodeInfo.MiningStats.TotalBlocks)
	}
}
EOF

cat > pkg/services/wallet_service.go << 'EOF'
package services

import (
	"ordm-main/pkg/wallet"
)

type WalletService struct {
	walletManager *wallet.Manager
}

func NewWalletService(walletManager *wallet.Manager) *WalletService {
	return &WalletService{
		walletManager: walletManager,
	}
}

func (ws *WalletService) CreateWallet() (*wallet.Wallet, error) {
	return ws.walletManager.CreateWallet()
}

func (ws *WalletService) GetBalance(userID string) (map[string]interface{}, error) {
	wallet, err := ws.walletManager.GetWallet(userID)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"balance": wallet.Balance,
		"address": wallet.Address,
		"staked":  wallet.StakedAmount,
	}, nil
}

func (ws *WalletService) StakeTokens(userID string, amount int64) error {
	return ws.walletManager.StakeTokens(userID, amount)
}

func (ws *WalletService) GetWallet(userID string) (*wallet.Wallet, error) {
	return ws.walletManager.GetWallet(userID)
}
EOF

echo "✅ [$(date)] Parte 2.1: Separação Frontend/Backend concluída!"
echo "📋 Implementações:"
echo "  ✅ 2.1.1 - API REST Separada criada"
echo "  ✅ 2.1.2 - Middleware Chain implementada"
echo "  ✅ 2.1.3 - Service Layer criada"
echo ""
echo "🚀 Próximo: Execute 'part2b_thread_safe.sh'"

