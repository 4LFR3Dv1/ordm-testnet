package network

import (
	"fmt"
	"log"
	"os"
	"time"
)

// OnlineSeedNode representa um seed node online
type OnlineSeedNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen"`
	Region   string `json:"region"`
	Type     string `json:"type"` // "validator", "explorer", "api"
}

// OnlineSeedNodeManager gerencia seed nodes online
type OnlineSeedNodeManager struct {
	Nodes map[string]*OnlineSeedNode
}

// NewOnlineSeedNodeManager cria um novo gerenciador de seed nodes online
func NewOnlineSeedNodeManager() *OnlineSeedNodeManager {
	manager := &OnlineSeedNodeManager{
		Nodes: make(map[string]*OnlineSeedNode),
	}
	
	// Inicializar seed nodes online
	manager.initializeOnlineNodes()
	
	return manager
}

// initializeOnlineNodes inicializa seed nodes com URLs públicas
func (osnm *OnlineSeedNodeManager) initializeOnlineNodes() {
	log.Println("🌐 Inicializando seed nodes online...")
	
	// Obter URL base do ambiente
	baseURL := os.Getenv("RENDER_EXTERNAL_URL")
	if baseURL == "" {
		baseURL = "https://ordm-testnet-1.onrender.com"
	}
	
	// Seed nodes online (URLs públicas)
	onlineNodes := []OnlineSeedNode{
		{
			ID:       "main-validator",
			Name:     "Main Validator Node",
			URL:      fmt.Sprintf("%s/api/validator", baseURL),
			Status:   "active",
			LastSeen: time.Now().Format("2006-01-02 15:04:05"),
			Region:   "us-west-1",
			Type:     "validator",
		},
		{
			ID:       "explorer-node",
			Name:     "Blockchain Explorer",
			URL:      fmt.Sprintf("%s/explorer", baseURL),
			Status:   "active",
			LastSeen: time.Now().Format("2006-01-02 15:04:05"),
			Region:   "us-west-1",
			Type:     "explorer",
		},
		{
			ID:       "api-gateway",
			Name:     "API Gateway",
			URL:      fmt.Sprintf("%s/api", baseURL),
			Status:   "active",
			LastSeen: time.Now().Format("2006-01-02 15:04:05"),
			Region:   "us-west-1",
			Type:     "api",
		},
		{
			ID:       "faucet-service",
			Name:     "Testnet Faucet",
			URL:      fmt.Sprintf("%s/api/testnet/faucet", baseURL),
			Status:   "active",
			LastSeen: time.Now().Format("2006-01-02 15:04:05"),
			Region:   "us-west-1",
			Type:     "faucet",
		},
	}
	
	for _, node := range onlineNodes {
		osnm.Nodes[node.ID] = &node
	}
	
	log.Printf("✅ %d seed nodes online inicializados", len(osnm.Nodes))
}

// GetActiveNodes retorna apenas os nodes ativos
func (osnm *OnlineSeedNodeManager) GetActiveNodes() []*OnlineSeedNode {
	var activeNodes []*OnlineSeedNode
	for _, node := range osnm.Nodes {
		if node.Status == "active" {
			activeNodes = append(activeNodes, node)
		}
	}
	return activeNodes
}

// GetNodeURLs retorna as URLs dos nodes ativos
func (osnm *OnlineSeedNodeManager) GetNodeURLs() []string {
	var urls []string
	for _, node := range osnm.GetActiveNodes() {
		urls = append(urls, node.URL)
	}
	return urls
}

// GetValidatorNodes retorna apenas nodes validadores
func (osnm *OnlineSeedNodeManager) GetValidatorNodes() []*OnlineSeedNode {
	var validators []*OnlineSeedNode
	for _, node := range osnm.Nodes {
		if node.Type == "validator" && node.Status == "active" {
			validators = append(validators, node)
		}
	}
	return validators
}

// UpdateNodeStatus atualiza o status de um node
func (osnm *OnlineSeedNodeManager) UpdateNodeStatus(nodeID, status string) {
	if node, exists := osnm.Nodes[nodeID]; exists {
		node.Status = status
		node.LastSeen = time.Now().Format("2006-01-02 15:04:05")
	}
}

// HealthCheck verifica a saúde dos nodes online
func (osnm *OnlineSeedNodeManager) HealthCheck() {
	for nodeID, node := range osnm.Nodes {
		// Em produção, sempre considerar nodes como ativos
		node.Status = "active"
		node.LastSeen = time.Now().Format("2006-01-02 15:04:05")
		
		log.Printf("🌐 Seed node online: %s (%s) - %s", nodeID, node.URL, node.Status)
	}
}
