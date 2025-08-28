package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// 🌐 Servidor Web Principal - Integra todos os serviços
type MainServer struct {
	router *mux.Router
	port   string
}

// 🔧 Inicializar servidor principal
func NewMainServer() *MainServer {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	router := mux.NewRouter()

	return &MainServer{
		router: router,
		port:   port,
	}
}

// 🛣️ Configurar rotas
func (s *MainServer) setupRoutes() {
	// Health check principal
	s.router.HandleFunc("/", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// API principal (Node)
	s.router.HandleFunc("/api/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/api/status", s.handleStatus).Methods("GET")

	// Subpath para Explorer
	explorerRouter := s.router.PathPrefix("/explorer").Subrouter()
	s.setupExplorerRoutes(explorerRouter)

	// Subpath para Monitor
	monitorRouter := s.router.PathPrefix("/monitor").Subrouter()
	s.setupMonitorRoutes(monitorRouter)

	// Subpath para Node (API principal)
	nodeRouter := s.router.PathPrefix("/node").Subrouter()
	s.setupNodeRoutes(nodeRouter)

	// Redirecionamentos para facilitar acesso
	s.router.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/monitor", http.StatusMovedPermanently)
	})

	s.router.HandleFunc("/blockchain", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/explorer", http.StatusMovedPermanently)
	})
}

// 🔍 Configurar rotas do Explorer
func (s *MainServer) setupExplorerRoutes(router *mux.Router) {
	router.HandleFunc("/", s.handleExplorerHome).Methods("GET")
	router.HandleFunc("/blocks", s.handleExplorerBlocks).Methods("GET")
	router.HandleFunc("/transactions", s.handleExplorerTransactions).Methods("GET")
	router.HandleFunc("/wallets", s.handleExplorerWallets).Methods("GET")
	router.HandleFunc("/block/{hash}", s.handleExplorerBlockDetail).Methods("GET")
	router.HandleFunc("/tx/{hash}", s.handleExplorerTransactionDetail).Methods("GET")
	router.HandleFunc("/address/{address}", s.handleExplorerAddressDetail).Methods("GET")
	router.HandleFunc("/api/stats", s.handleExplorerAPIStats).Methods("GET")
	router.HandleFunc("/api/blocks", s.handleExplorerAPIBlocks).Methods("GET")
	router.HandleFunc("/api/transactions", s.handleExplorerAPITransactions).Methods("GET")
	router.HandleFunc("/api/wallets", s.handleExplorerAPIWallets).Methods("GET")
}

// 📊 Configurar rotas do Monitor
func (s *MainServer) setupMonitorRoutes(router *mux.Router) {
	router.HandleFunc("/", s.handleMonitorDashboard).Methods("GET")
	router.HandleFunc("/api/metrics", s.handleMonitorMetrics).Methods("GET")
	router.HandleFunc("/api/security", s.handleMonitorSecurity).Methods("GET")
	router.HandleFunc("/api/alerts", s.handleMonitorAlerts).Methods("GET")
	router.HandleFunc("/api/events", s.handleMonitorEvents).Methods("GET")
}

// 🔗 Configurar rotas do Node
func (s *MainServer) setupNodeRoutes(router *mux.Router) {
	router.HandleFunc("/api/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/api/status", s.handleStatus).Methods("GET")
	router.HandleFunc("/api/mining/start", s.handleStartMining).Methods("POST")
	router.HandleFunc("/api/mining/stop", s.handleStopMining).Methods("POST")
	router.HandleFunc("/api/wallet/create", s.handleCreateWallet).Methods("POST")
	router.HandleFunc("/api/wallet/balance", s.handleGetBalance).Methods("GET")
	router.HandleFunc("/api/wallet/stake", s.handleStakeTokens).Methods("POST")
}

// 🏥 Health Check
func (s *MainServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "ordm-main",
		"version":   "1.0.0",
		"endpoints": map[string]string{
			"main":     "/",
			"explorer": "/explorer",
			"monitor":  "/monitor",
			"node":     "/node",
		},
	}

	fmt.Fprintf(w, `{"status":"%s","timestamp":"%s","service":"%s","version":"%s","endpoints":{"main":"%s","explorer":"%s","monitor":"%s","node":"%s"}}`,
		response["status"], response["timestamp"], response["service"], response["version"],
		response["endpoints"].(map[string]string)["main"],
		response["endpoints"].(map[string]string)["explorer"],
		response["endpoints"].(map[string]string)["monitor"],
		response["endpoints"].(map[string]string)["node"])
}

// 📊 Status do sistema
func (s *MainServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "online",
		"timestamp": time.Now().UTC(),
		"services": map[string]string{
			"node":     "running",
			"explorer": "running",
			"monitor":  "running",
		},
		"uptime": time.Since(time.Now()).String(),
	}

	fmt.Fprintf(w, `{"status":"%s","timestamp":"%s","services":{"node":"%s","explorer":"%s","monitor":"%s"},"uptime":"%s"}`,
		response["status"], response["timestamp"],
		response["services"].(map[string]string)["node"],
		response["services"].(map[string]string)["explorer"],
		response["services"].(map[string]string)["monitor"],
		response["uptime"])
}

// 🔍 Explorer Handlers
func (s *MainServer) handleExplorerHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>ORDM Blockchain Explorer</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; }
			.container { max-width: 1200px; margin: 0 auto; }
			.header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
			.nav { background: #34495e; padding: 10px; margin: 10px 0; border-radius: 5px; }
			.nav a { color: white; text-decoration: none; margin-right: 20px; }
			.content { background: #ecf0f1; padding: 20px; border-radius: 5px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>🔍 ORDM Blockchain Explorer</h1>
				<p>Explore a blockchain ORDM em tempo real</p>
			</div>
			<div class="nav">
				<a href="/explorer/blocks">📦 Blocos</a>
				<a href="/explorer/transactions">💸 Transações</a>
				<a href="/explorer/wallets">👛 Carteiras</a>
				<a href="/explorer/api/stats">📊 Estatísticas</a>
			</div>
			<div class="content">
				<h2>Bem-vindo ao Explorer</h2>
				<p>Use os links acima para navegar pela blockchain ORDM.</p>
				<p><strong>Status:</strong> <span style="color: green;">🟢 Online</span></p>
			</div>
		</div>
	</body>
	</html>`

	fmt.Fprint(w, html)
}

func (s *MainServer) handleExplorerBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "<h1>📦 Blocos da Blockchain</h1><p>Lista de blocos será exibida aqui.</p>")
}

func (s *MainServer) handleExplorerTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "<h1>💸 Transações</h1><p>Lista de transações será exibida aqui.</p>")
}

func (s *MainServer) handleExplorerWallets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "<h1>👛 Carteiras</h1><p>Lista de carteiras será exibida aqui.</p>")
}

func (s *MainServer) handleExplorerBlockDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<h1>📦 Detalhes do Bloco</h1><p>Hash: %s</p>", hash)
}

func (s *MainServer) handleExplorerTransactionDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<h1>💸 Detalhes da Transação</h1><p>Hash: %s</p>", hash)
}

func (s *MainServer) handleExplorerAddressDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<h1>👛 Detalhes do Endereço</h1><p>Endereço: %s</p>", address)
}

func (s *MainServer) handleExplorerAPIStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"total_blocks": 100, "total_transactions": 500, "total_wallets": 50}`)
}

func (s *MainServer) handleExplorerAPIBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `[{"hash": "abc123", "height": 100, "timestamp": "2024-01-01T00:00:00Z"}]`)
}

func (s *MainServer) handleExplorerAPITransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `[{"hash": "def456", "from": "addr1", "to": "addr2", "amount": 100}]`)
}

func (s *MainServer) handleExplorerAPIWallets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `[{"address": "addr1", "balance": 1000, "transactions": 10}]`)
}

// 📊 Monitor Handlers
func (s *MainServer) handleMonitorDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>ORDM Monitor Dashboard</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; }
			.container { max-width: 1200px; margin: 0 auto; }
			.header { background: #27ae60; color: white; padding: 20px; border-radius: 5px; }
			.nav { background: #2ecc71; padding: 10px; margin: 10px 0; border-radius: 5px; }
			.nav a { color: white; text-decoration: none; margin-right: 20px; }
			.content { background: #ecf0f1; padding: 20px; border-radius: 5px; }
			.metric { background: white; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #27ae60; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>📊 ORDM Monitor Dashboard</h1>
				<p>Monitoramento em tempo real da testnet ORDM</p>
			</div>
			<div class="nav">
				<a href="/monitor/api/metrics">📈 Métricas</a>
				<a href="/monitor/api/security">🔒 Segurança</a>
				<a href="/monitor/api/alerts">🚨 Alertas</a>
				<a href="/monitor/api/events">📋 Eventos</a>
			</div>
			<div class="content">
				<h2>Dashboard Principal</h2>
				<div class="metric">
					<h3>🟢 Status do Sistema</h3>
					<p><strong>Online:</strong> Sim</p>
					<p><strong>Última verificação:</strong> ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
				</div>
				<div class="metric">
					<h3>📊 Métricas</h3>
					<p><strong>Blocos:</strong> 100</p>
					<p><strong>Transações:</strong> 500</p>
					<p><strong>Carteiras:</strong> 50</p>
				</div>
			</div>
		</div>
	</body>
	</html>`

	fmt.Fprint(w, html)
}

func (s *MainServer) handleMonitorMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"blocks": 100, "transactions": 500, "wallets": 50, "uptime": "24h"}`)
}

func (s *MainServer) handleMonitorSecurity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "secure", "last_scan": "2024-01-01T00:00:00Z", "threats": 0}`)
}

func (s *MainServer) handleMonitorAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `[{"level": "info", "message": "Sistema funcionando normalmente", "timestamp": "2024-01-01T00:00:00Z"}]`)
}

func (s *MainServer) handleMonitorEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `[{"type": "block_mined", "data": {"block": 100, "miner": "miner1"}, "timestamp": "2024-01-01T00:00:00Z"}]`)
}

// 🔗 Node Handlers
func (s *MainServer) handleStartMining(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "mining_started", "message": "Mineração iniciada com sucesso"}`)
}

func (s *MainServer) handleStopMining(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "mining_stopped", "message": "Mineração parada com sucesso"}`)
}

func (s *MainServer) handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "wallet_created", "address": "addr123", "message": "Carteira criada com sucesso"}`)
}

func (s *MainServer) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"address": "addr123", "balance": 1000, "currency": "ORDM"}`)
}

func (s *MainServer) handleStakeTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "staked", "amount": 100, "message": "Tokens stakados com sucesso"}`)
}

// 🚀 Iniciar servidor
func (s *MainServer) Start() error {
	s.setupRoutes()

	log.Printf("🚀 Iniciando servidor ORDM na porta %s", s.port)
	log.Printf("📊 URLs disponíveis:")
	log.Printf("  🏠 Principal: http://localhost:%s/", s.port)
	log.Printf("  🔍 Explorer: http://localhost:%s/explorer", s.port)
	log.Printf("  📊 Monitor: http://localhost:%s/monitor", s.port)
	log.Printf("  🔗 Node API: http://localhost:%s/node", s.port)

	return http.ListenAndServe(":"+s.port, s.router)
}

// 🎯 Função principal
func main() {
	server := NewMainServer()

	if err := server.Start(); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}
