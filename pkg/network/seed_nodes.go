package network

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SeedNode representa um nó seed público da testnet
type SeedNode struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	Port       int       `json:"port"`
	PublicKey  string    `json:"public_key"`
	IsActive   bool      `json:"is_active"`
	LastSeen   time.Time `json:"last_seen"`
	Version    string    `json:"version"`
	PeersCount int       `json:"peers_count"`
	Uptime     int64     `json:"uptime"`
	Region     string    `json:"region"`
	Provider   string    `json:"provider"`
}

// SeedNodeManager gerencia os seed nodes da testnet
type SeedNodeManager struct {
	SeedNodes map[string]*SeedNode `json:"seed_nodes"`
	Mutex     sync.RWMutex         `json:"-"`
	Config    *SeedNodeConfig      `json:"config"`
}

// SeedNodeConfig configuração dos seed nodes
type SeedNodeConfig struct {
	TestnetMode    bool     `json:"testnet_mode"`
	BootstrapPeers []string `json:"bootstrap_peers"`
	MaxPeers       int      `json:"max_peers"`
	Heartbeat      int      `json:"heartbeat"`
	Timeout        int      `json:"timeout"`
}

// NewSeedNodeManager cria um novo gerenciador de seed nodes
func NewSeedNodeManager() *SeedNodeManager {
	return &SeedNodeManager{
		SeedNodes: make(map[string]*SeedNode),
		Config: &SeedNodeConfig{
			TestnetMode: true,
			BootstrapPeers: []string{
				"/ip4/18.188.123.45/tcp/3001/p2p/QmSeedNode1",
				"/ip4/52.15.67.89/tcp/3001/p2p/QmSeedNode2",
				"/ip4/34.201.234.56/tcp/3001/p2p/QmSeedNode3",
			},
			MaxPeers:  50,
			Heartbeat: 30,
			Timeout:   60,
		},
	}
}

// InitializeTestnetSeedNodes inicializa os seed nodes da testnet
func (snm *SeedNodeManager) InitializeTestnetSeedNodes() {
	snm.Mutex.Lock()
	defer snm.Mutex.Unlock()

	// Seed nodes da testnet (VPS públicos)
	testnetSeeds := []*SeedNode{
		{
			ID:         "testnet-seed-1",
			Address:    "18.188.123.45",
			Port:       3001,
			PublicKey:  "QmSeedNode1",
			IsActive:   true,
			LastSeen:   time.Now(),
			Version:    "1.0.0-testnet",
			PeersCount: 0,
			Uptime:     time.Now().Unix(),
			Region:     "us-east-1",
			Provider:   "AWS",
		},
		{
			ID:         "testnet-seed-2",
			Address:    "52.15.67.89",
			Port:       3001,
			PublicKey:  "QmSeedNode2",
			IsActive:   true,
			LastSeen:   time.Now(),
			Version:    "1.0.0-testnet",
			PeersCount: 0,
			Uptime:     time.Now().Unix(),
			Region:     "us-west-2",
			Provider:   "AWS",
		},
		{
			ID:         "testnet-seed-3",
			Address:    "34.201.234.56",
			Port:       3001,
			PublicKey:  "QmSeedNode3",
			IsActive:   true,
			LastSeen:   time.Now(),
			Version:    "1.0.0-testnet",
			PeersCount: 0,
			Uptime:     time.Now().Unix(),
			Region:     "eu-west-1",
			Provider:   "AWS",
		},
	}

	// Adicionar seed nodes
	for _, seed := range testnetSeeds {
		snm.SeedNodes[seed.ID] = seed
	}

	log.Printf("🌱 Testnet seed nodes inicializados: %d nós", len(testnetSeeds))
}

// GetBootstrapPeers retorna a lista de peers bootstrap
func (snm *SeedNodeManager) GetBootstrapPeers() []string {
	snm.Mutex.RLock()
	defer snm.Mutex.RUnlock()

	var peers []string
	for _, seed := range snm.SeedNodes {
		if seed.IsActive {
			peerAddr := fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s",
				seed.Address, seed.Port, seed.PublicKey)
			peers = append(peers, peerAddr)
		}
	}
	return peers
}

// GetActiveSeedNodes retorna seed nodes ativos
func (snm *SeedNodeManager) GetActiveSeedNodes() []*SeedNode {
	snm.Mutex.RLock()
	defer snm.Mutex.RUnlock()

	var active []*SeedNode
	for _, seed := range snm.SeedNodes {
		if seed.IsActive {
			active = append(active, seed)
		}
	}
	return active
}

// UpdateSeedNodeStatus atualiza status de um seed node
func (snm *SeedNodeManager) UpdateSeedNodeStatus(id string, isActive bool, peersCount int) {
	snm.Mutex.Lock()
	defer snm.Mutex.Unlock()

	if seed, exists := snm.SeedNodes[id]; exists {
		seed.IsActive = isActive
		seed.LastSeen = time.Now()
		seed.PeersCount = peersCount
	}
}

// StartSeedNodeHeartbeat inicia heartbeat dos seed nodes
func (snm *SeedNodeManager) StartSeedNodeHeartbeat() {
	ticker := time.NewTicker(time.Duration(snm.Config.Heartbeat) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		snm.checkSeedNodesHealth()
	}
}

// checkSeedNodesHealth verifica saúde dos seed nodes
func (snm *SeedNodeManager) checkSeedNodesHealth() {
	snm.Mutex.Lock()
	defer snm.Mutex.Unlock()

	for id, seed := range snm.SeedNodes {
		// Verificar se o seed node está respondendo
		url := fmt.Sprintf("http://%s:%d/health", seed.Address, seed.Port)

		client := &http.Client{Timeout: time.Duration(snm.Config.Timeout) * time.Second}
		resp, err := client.Get(url)

		if err != nil || resp.StatusCode != 200 {
			seed.IsActive = false
			log.Printf("⚠️ Seed node %s inativo: %v", id, err)
		} else {
			seed.IsActive = true
			seed.LastSeen = time.Now()
			resp.Body.Close()
		}
	}
}

// GetSeedNodesInfo retorna informações dos seed nodes para API
func (snm *SeedNodeManager) GetSeedNodesInfo() map[string]interface{} {
	snm.Mutex.RLock()
	defer snm.Mutex.RUnlock()

	activeCount := 0
	totalPeers := 0

	for _, seed := range snm.SeedNodes {
		if seed.IsActive {
			activeCount++
			totalPeers += seed.PeersCount
		}
	}

	return map[string]interface{}{
		"total_seeds":     len(snm.SeedNodes),
		"active_seeds":    activeCount,
		"total_peers":     totalPeers,
		"bootstrap_peers": snm.GetBootstrapPeers(),
		"config":          snm.Config,
	}
}
