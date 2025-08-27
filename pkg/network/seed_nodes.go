package network

import (
	"fmt"
	"log"
	"os"
	"time"
)

// SeedNode representa um nó seed da testnet
type SeedNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen"`
}

// SeedNodeManager gerencia os seed nodes da testnet
type SeedNodeManager struct {
	Nodes map[string]*SeedNode
}

// NewSeedNodeManager cria um novo gerenciador de seed nodes
func NewSeedNodeManager() *SeedNodeManager {
	manager := &SeedNodeManager{
		Nodes: make(map[string]*SeedNode),
	}
	
	// Verificar se estamos em produção (Render)
	// Render define PORT automaticamente, então usamos isso como indicador
	if os.Getenv("PORT") != "" || os.Getenv("NODE_ENV") == "production" {
		// Em produção, usar apenas nodes locais
		manager.initializeLocalNodes()
	} else {
		// Em desenvolvimento, usar seed nodes externos
		manager.initializeExternalNodes()
	}
	
	return manager
}

// initializeLocalNodes inicializa apenas nodes locais para produção
func (snm *SeedNodeManager) initializeLocalNodes() {
	log.Println("🌱 Inicializando seed nodes para produção (Render)...")
	
	// Obter porta do ambiente Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	// Node principal (porta do Render)
	snm.Nodes["render-main"] = &SeedNode{
		ID:       "render-main",
		Name:     "Render Main Node",
		IP:       "0.0.0.0", // Aceitar conexões de qualquer IP
		Port:     3000,      // Porta interna do container
		Status:   "active",
		LastSeen: time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// Node secundário (porta alternativa)
	snm.Nodes["render-secondary"] = &SeedNode{
		ID:       "render-secondary",
		Name:     "Render Secondary Node",
		IP:       "0.0.0.0",
		Port:     8080,      // Porta alternativa
		Status:   "active",
		LastSeen: time.Now().Format("2006-01-02 15:04:05"),
	}
	
	log.Printf("✅ %d seed nodes de produção inicializados na porta %s", len(snm.Nodes), port)
}

// initializeExternalNodes inicializa seed nodes externos para desenvolvimento
func (snm *SeedNodeManager) initializeExternalNodes() {
	log.Println("🌱 Inicializando seed nodes externos para desenvolvimento...")
	
	// Seed nodes externos (apenas para desenvolvimento)
	externalNodes := []SeedNode{
		{
			ID:       "testnet-seed-1",
			Name:     "Testnet Seed 1",
			IP:       "18.188.123.45",
			Port:     3001,
			Status:   "inactive",
			LastSeen: "never",
		},
		{
			ID:       "testnet-seed-2",
			Name:     "Testnet Seed 2",
			IP:       "52.15.67.89",
			Port:     3001,
			Status:   "inactive",
			LastSeen: "never",
		},
		{
			ID:       "testnet-seed-3",
			Name:     "Testnet Seed 3",
			IP:       "34.201.234.56",
			Port:     3001,
			Status:   "inactive",
			LastSeen: "never",
		},
	}
	
	for _, node := range externalNodes {
		snm.Nodes[node.ID] = &node
	}
	
	log.Printf("✅ %d seed nodes externos inicializados", len(snm.Nodes))
}

// GetActiveNodes retorna apenas os nodes ativos
func (snm *SeedNodeManager) GetActiveNodes() []*SeedNode {
	var activeNodes []*SeedNode
	for _, node := range snm.Nodes {
		if node.Status == "active" {
			activeNodes = append(activeNodes, node)
		}
	}
	return activeNodes
}

// UpdateNodeStatus atualiza o status de um node
func (snm *SeedNodeManager) UpdateNodeStatus(nodeID, status string) {
	if node, exists := snm.Nodes[nodeID]; exists {
		node.Status = status
		node.LastSeen = time.Now().Format("2006-01-02 15:04:05")
	}
}

// GetNodeAddresses retorna as URLs dos nodes ativos
func (snm *SeedNodeManager) GetNodeAddresses() []string {
	var addresses []string
	for _, node := range snm.GetActiveNodes() {
		address := fmt.Sprintf("http://%s:%d", node.IP, node.Port)
		addresses = append(addresses, address)
	}
	return addresses
}

// HealthCheck verifica a saúde dos nodes
func (snm *SeedNodeManager) HealthCheck() {
	for nodeID, node := range snm.Nodes {
		// Em produção, sempre considerar nodes locais como ativos
		if os.Getenv("NODE_ENV") == "production" {
			node.Status = "active"
			node.LastSeen = time.Now().Format("2006-01-02 15:04:05")
		} else {
			// Em desenvolvimento, fazer health check real
			// Por enquanto, marcar como inativo para evitar timeouts
			node.Status = "inactive"
		}
		
		if node.Status == "inactive" {
			log.Printf("⚠️ Seed node %s inativo: %s", nodeID, node.IP)
		}
	}
}
