package api

import (
	"encoding/json"
	"net/http"
	"time"

	"ordm-main/pkg/auth"
	"ordm-main/pkg/services"

	"github.com/gorilla/mux"
)

type APIServer struct {
	router *mux.Router
	auth   *auth.UserManager
	mining *services.MiningService
	wallet *services.WalletService
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewAPIServer(auth *auth.UserManager, mining *services.MiningService, wallet *services.WalletService) *APIServer {
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
	// Rotas públicas
	api.router.HandleFunc("/api/health", api.healthCheck).Methods("GET")
	api.router.HandleFunc("/api/login", api.login).Methods("POST")

	// Rotas protegidas (simplificadas)
	api.router.HandleFunc("/api/status", api.getStatus).Methods("GET")
	api.router.HandleFunc("/api/mining/start", api.startMining).Methods("POST")
	api.router.HandleFunc("/api/mining/stop", api.stopMining).Methods("POST")
	api.router.HandleFunc("/api/wallet/create", api.createWallet).Methods("POST")
	api.router.HandleFunc("/api/wallet/balance", api.getBalance).Methods("GET")
	api.router.HandleFunc("/api/wallet/stake", api.stakeTokens).Methods("POST")
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

	user, err := api.auth.Login(loginData.Username, loginData.Password)
	if err != nil {
		api.sendError(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Login realizado com sucesso",
		Data: map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
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
