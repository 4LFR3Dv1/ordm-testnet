package network

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// FailoverSeedNode representa um seed node da rede para failover
type FailoverSeedNode struct {
	URL          string        `json:"url"`
	Port         string        `json:"port"`
	Region       string        `json:"region"`
	Priority     int           `json:"priority"` // Prioridade (1 = mais alta)
	Weight       int           `json:"weight"`   // Peso para load balancing
	IsActive     bool          `json:"is_active"`
	LastPing     time.Time     `json:"last_ping"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount   int           `json:"error_count"`
	MaxErrors    int           `json:"max_errors"`
	HealthStatus string        `json:"health_status"` // "healthy", "degraded", "unhealthy"
	mu           sync.RWMutex
}

// FailoverManager gerencia failover automático entre seed nodes
type FailoverManager struct {
	seedNodes    []*FailoverSeedNode
	activeNode   *FailoverSeedNode
	backupNodes  []*FailoverSeedNode
	healthCheck  *HealthChecker
	loadBalancer *LoadBalancer
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// HealthChecker verifica saúde dos seed nodes
type HealthChecker struct {
	checkInterval time.Duration
	timeout       time.Duration
	httpClient    *http.Client
	mu            sync.RWMutex
}

// LoadBalancer distribui carga entre seed nodes
type LoadBalancer struct {
	strategy     string // "round_robin", "weighted", "least_connections"
	currentIndex int
	mu           sync.RWMutex
}

// NewFailoverSeedNode cria um novo seed node
func NewFailoverSeedNode(url, port, region string, priority, weight int) *FailoverSeedNode {
	return &FailoverSeedNode{
		URL:          url,
		Port:         port,
		Region:       region,
		Priority:     priority,
		Weight:       weight,
		IsActive:     true,
		LastPing:     time.Now(),
		MaxErrors:    5,
		HealthStatus: "healthy",
	}
}

// NewFailoverManager cria um novo gerenciador de failover
func NewFailoverManager() *FailoverManager {
	ctx, cancel := context.WithCancel(context.Background())

	fm := &FailoverManager{
		seedNodes:   []*FailoverSeedNode{},
		backupNodes: []*FailoverSeedNode{},
		ctx:         ctx,
		cancel:      cancel,
	}

	fm.healthCheck = NewHealthChecker()
	fm.loadBalancer = NewLoadBalancer("weighted")

	return fm
}

// AddSeedNode adiciona um seed node ao gerenciador
func (fm *FailoverManager) AddSeedNode(node *FailoverSeedNode) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.seedNodes = append(fm.seedNodes, node)

	// Ordenar por prioridade
	fm.sortNodesByPriority()

	// Definir primeiro node como ativo se não houver nenhum
	if fm.activeNode == nil {
		fm.activeNode = node
	}
}

// sortNodesByPriority ordena nodes por prioridade
func (fm *FailoverManager) sortNodesByPriority() {
	// Implementação de ordenação por prioridade
	// (maior prioridade = menor número)
	for i := 0; i < len(fm.seedNodes)-1; i++ {
		for j := i + 1; j < len(fm.seedNodes); j++ {
			if fm.seedNodes[i].Priority > fm.seedNodes[j].Priority {
				fm.seedNodes[i], fm.seedNodes[j] = fm.seedNodes[j], fm.seedNodes[i]
			}
		}
	}
}

// StartHealthCheck inicia verificação de saúde
func (fm *FailoverManager) StartHealthCheck() {
	go fm.healthCheckLoop()
}

// healthCheckLoop loop principal de verificação de saúde
func (fm *FailoverManager) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second) // Verificar a cada 30 segundos
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.checkAllNodes()
		}
	}
}

// checkAllNodes verifica saúde de todos os nodes
func (fm *FailoverManager) checkAllNodes() {
	fm.mu.RLock()
	nodes := make([]*FailoverSeedNode, len(fm.seedNodes))
	copy(nodes, fm.seedNodes)
	fm.mu.RUnlock()

	for _, node := range nodes {
		go fm.checkNodeHealth(node)
	}
}

// checkNodeHealth verifica saúde de um node específico
func (fm *FailoverManager) checkNodeHealth(node *FailoverSeedNode) {
	healthy := fm.healthCheck.CheckHealth(node)

	node.mu.Lock()
	defer node.mu.Unlock()

	node.LastPing = time.Now()

	if healthy {
		node.ErrorCount = 0
		node.HealthStatus = "healthy"
	} else {
		node.ErrorCount++
		if node.ErrorCount >= node.MaxErrors {
			node.HealthStatus = "unhealthy"
			node.IsActive = false
			fm.handleNodeFailure(node)
		} else {
			node.HealthStatus = "degraded"
		}
	}
}

// handleNodeFailure trata falha de um node
func (fm *FailoverManager) handleNodeFailure(failedNode *FailoverSeedNode) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Se o node falhado é o ativo, fazer failover
	if fm.activeNode == failedNode {
		fm.performFailover()
	}
}

// performFailover executa failover para o próximo node saudável
func (fm *FailoverManager) performFailover() {
	// Encontrar próximo node saudável
	for _, node := range fm.seedNodes {
		if node != fm.activeNode && node.IsActive && node.HealthStatus == "healthy" {
			oldNode := fm.activeNode
			fm.activeNode = node

			fmt.Printf("🔄 Failover: %s -> %s\n", oldNode.URL, node.URL)
			return
		}
	}

	// Se não encontrar node saudável, tentar nodes degradados
	for _, node := range fm.seedNodes {
		if node != fm.activeNode && node.IsActive && node.HealthStatus == "degraded" {
			oldNode := fm.activeNode
			fm.activeNode = node

			fmt.Printf("⚠️ Failover para node degradado: %s -> %s\n", oldNode.URL, node.URL)
			return
		}
	}

	fmt.Printf("🚨 Nenhum node saudável disponível para failover\n")
}

// GetActiveNode retorna o node ativo
func (fm *FailoverManager) GetActiveNode() *FailoverSeedNode {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.activeNode
}

// GetLoadBalancedNode retorna um node usando load balancing
func (fm *FailoverManager) GetLoadBalancedNode() *FailoverSeedNode {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	healthyNodes := []*FailoverSeedNode{}
	for _, node := range fm.seedNodes {
		if node.IsActive && node.HealthStatus == "healthy" {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return fm.activeNode // Fallback para node ativo
	}

	return fm.loadBalancer.SelectNode(healthyNodes)
}

// NewHealthChecker cria um novo verificador de saúde
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checkInterval: 30 * time.Second,
		timeout:       10 * time.Second,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckHealth verifica saúde de um node
func (hc *HealthChecker) CheckHealth(node *FailoverSeedNode) bool {
	start := time.Now()

	// Tentar conectar ao endpoint de saúde
	url := fmt.Sprintf("%s:%s/api/health", node.URL, node.Port)

	resp, err := hc.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Calcular tempo de resposta
	responseTime := time.Since(start)

	node.mu.Lock()
	node.ResponseTime = responseTime
	node.mu.Unlock()

	// Considerar saudável se status 200-299
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// NewLoadBalancer cria um novo load balancer
func NewLoadBalancer(strategy string) *LoadBalancer {
	return &LoadBalancer{
		strategy:     strategy,
		currentIndex: 0,
	}
}

// SelectNode seleciona um node usando a estratégia de load balancing
func (lb *LoadBalancer) SelectNode(nodes []*FailoverSeedNode) *FailoverSeedNode {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(nodes) == 0 {
		return nil
	}

	switch lb.strategy {
	case "round_robin":
		return lb.roundRobin(nodes)
	case "weighted":
		return lb.weightedRoundRobin(nodes)
	case "least_connections":
		return lb.leastConnections(nodes)
	default:
		return nodes[0]
	}
}

// roundRobin seleção round-robin simples
func (lb *LoadBalancer) roundRobin(nodes []*FailoverSeedNode) *FailoverSeedNode {
	node := nodes[lb.currentIndex%len(nodes)]
	lb.currentIndex++
	return node
}

// weightedRoundRobin seleção round-robin com pesos
func (lb *LoadBalancer) weightedRoundRobin(nodes []*FailoverSeedNode) *FailoverSeedNode {
	// Implementação simplificada - retorna node com maior peso
	maxWeight := 0
	selectedNode := nodes[0]

	for _, node := range nodes {
		if node.Weight > maxWeight {
			maxWeight = node.Weight
			selectedNode = node
		}
	}

	return selectedNode
}

// leastConnections seleção por menor número de conexões
func (lb *LoadBalancer) leastConnections(nodes []*FailoverSeedNode) *FailoverSeedNode {
	// Implementação simplificada - retorna node com menor tempo de resposta
	minResponseTime := nodes[0].ResponseTime
	selectedNode := nodes[0]

	for _, node := range nodes {
		if node.ResponseTime < minResponseTime {
			minResponseTime = node.ResponseTime
			selectedNode = node
		}
	}

	return selectedNode
}

// GetStatus retorna status do failover manager
func (fm *FailoverManager) GetStatus() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	status := map[string]interface{}{
		"active_node":     fm.activeNode.URL,
		"total_nodes":     len(fm.seedNodes),
		"healthy_nodes":   0,
		"degraded_nodes":  0,
		"unhealthy_nodes": 0,
	}

	for _, node := range fm.seedNodes {
		switch node.HealthStatus {
		case "healthy":
			status["healthy_nodes"] = status["healthy_nodes"].(int) + 1
		case "degraded":
			status["degraded_nodes"] = status["degraded_nodes"].(int) + 1
		case "unhealthy":
			status["unhealthy_nodes"] = status["unhealthy_nodes"].(int) + 1
		}
	}

	return status
}

// Stop para o failover manager
func (fm *FailoverManager) Stop() {
	fm.cancel()
}

// Reconnect tenta reconectar um node
func (fm *FailoverManager) Reconnect(nodeURL string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, node := range fm.seedNodes {
		if node.URL == nodeURL {
			node.mu.Lock()
			node.ErrorCount = 0
			node.HealthStatus = "healthy"
			node.IsActive = true
			node.mu.Unlock()

			fmt.Printf("✅ Node reconectado: %s\n", nodeURL)
			return
		}
	}
}
