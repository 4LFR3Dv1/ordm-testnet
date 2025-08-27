package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ordm-main/pkg/backend"
)

var globalDB *backend.GlobalDatabase

func main() {
	fmt.Println("🚀 Iniciando Backend da Blockchain 2-Layer...")

	// Inicializar banco de dados global
	globalDB = backend.NewGlobalDatabase()

	// Configurar rotas da API
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/wallets", handleWallets)
	http.HandleFunc("/api/transactions", handleTransactions)
	http.HandleFunc("/api/blocks", handleBlocks)
	http.HandleFunc("/api/nodes", handleNodes)
	http.HandleFunc("/api/state", handleGlobalState)
	http.HandleFunc("/api/audit", handleAudit)

	// Rota para registro de nodes
	http.HandleFunc("/api/nodes/register", handleNodeRegistration)

	// Rota para autenticação de nodes
	http.HandleFunc("/api/nodes/auth", handleNodeAuth)

	// Rota para sincronização de dados
	http.HandleFunc("/api/sync", handleSync)

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("📡 Backend rodando na porta %s\n", port)
	fmt.Printf("🌐 API disponível em: http://localhost:%s/api\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleHealth verifica saúde do backend
func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "2.0.0",
		"service":   "blockchain-2layer-backend",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleWallets gerencia wallets
func handleWallets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Listar todas as wallets
		wallets := globalDB.GetAllWallets()
		response := map[string]interface{}{
			"wallets": wallets,
			"count":   len(wallets),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case "POST":
		// Registrar nova wallet
		var wallet backend.GlobalWallet
		if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		if err := globalDB.RegisterWallet(&wallet); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Wallet registrada com sucesso",
			"wallet":  wallet,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

// handleTransactions gerencia transações
func handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Listar todas as transações
		transactions := globalDB.GetAllTransactions()
		response := map[string]interface{}{
			"transactions": transactions,
			"count":        len(transactions),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case "POST":
		// Registrar nova transação
		var tx backend.GlobalTransaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		if err := globalDB.RegisterTransaction(&tx); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		response := map[string]interface{}{
			"success":     true,
			"message":     "Transação registrada com sucesso",
			"transaction": tx,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

// handleBlocks gerencia blocos
func handleBlocks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Listar todos os blocos
		blocks := globalDB.GetAllBlocks()
		response := map[string]interface{}{
			"blocks": blocks,
			"count":  len(blocks),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case "POST":
		// Registrar novo bloco
		var block backend.GlobalBlock
		if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		if err := globalDB.RegisterBlock(&block); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Bloco registrado com sucesso",
			"block":   block,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

// handleNodes gerencia nodes
func handleNodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Listar todos os nodes
		nodes := globalDB.GetAllNodes()
		response := map[string]interface{}{
			"nodes": nodes,
			"count": len(nodes),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

// handleGlobalState retorna estado global
func handleGlobalState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	state := globalDB.GetGlobalState()
	response := map[string]interface{}{
		"state": state,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAudit retorna log de auditoria
func handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	auditLog := globalDB.GetAuditLog()
	response := map[string]interface{}{
		"audit_log": auditLog,
		"count":     len(auditLog),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNodeRegistration registra novo node
func handleNodeRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var node backend.RegisteredNode
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if err := globalDB.RegisterNode(&node); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Node registrado com sucesso",
		"node":    node,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNodeAuth autentica node
func handleNodeAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var authData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	nodeID, ok := authData["node_id"].(string)
	if !ok {
		http.Error(w, "node_id obrigatório", http.StatusBadRequest)
		return
	}

	// Verificar se node existe
	node, exists := globalDB.GetNode(nodeID)
	if !exists {
		http.Error(w, "Node não encontrado", http.StatusNotFound)
		return
	}

	// TODO: Implementar verificação de assinatura
	// Por enquanto, aceitar qualquer autenticação

	response := map[string]interface{}{
		"success": true,
		"message": "Node autenticado com sucesso",
		"node":    node,
		"token":   fmt.Sprintf("token_%s_%d", nodeID, time.Now().Unix()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSync sincroniza dados
func handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Retornar snapshot completo do banco
	syncData := map[string]interface{}{
		"wallets":      globalDB.GetAllWallets(),
		"transactions": globalDB.GetAllTransactions(),
		"blocks":       globalDB.GetAllBlocks(),
		"nodes":        globalDB.GetAllNodes(),
		"state":        globalDB.GetGlobalState(),
		"timestamp":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(syncData)
}
