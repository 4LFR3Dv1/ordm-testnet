package state

import (
	"sync"
	"time"
)

// MiningStats representa estatísticas de mineração
type MiningStats struct {
	TotalBlocks  int64     `json:"total_blocks"`
	HashRate     float64   `json:"hash_rate"`
	LastMined    time.Time `json:"last_mined"`
	TotalRewards int64     `json:"total_rewards"`
}

// NodeState representa o estado do node
type NodeState struct {
	Status      string      `json:"status"`
	Balance     int64       `json:"balance"`
	MiningStats MiningStats `json:"mining_stats"`
	LastUpdate  time.Time   `json:"last_update"`
}

// SafeNodeState gerencia o estado do node de forma thread-safe
type SafeNodeState struct {
	mu    sync.RWMutex
	state NodeState
}

// NewSafeNodeState cria uma nova instância de SafeNodeState
func NewSafeNodeState() *SafeNodeState {
	return &SafeNodeState{
		state: NodeState{
			Status:      "offline",
			Balance:     0,
			MiningStats: MiningStats{},
			LastUpdate:  time.Now(),
		},
	}
}

// GetState retorna uma cópia do estado atual
func (sns *SafeNodeState) GetState() NodeState {
	sns.mu.RLock()
	defer sns.mu.RUnlock()

	// Retornar cópia para evitar race conditions
	return NodeState{
		Status:      sns.state.Status,
		Balance:     sns.state.Balance,
		MiningStats: sns.state.MiningStats,
		LastUpdate:  sns.state.LastUpdate,
	}
}

// SetState atualiza o estado
func (sns *SafeNodeState) SetState(state NodeState) {
	sns.mu.Lock()
	defer sns.mu.Unlock()

	sns.state = state
	sns.state.LastUpdate = time.Now()
}

// UpdateMiningStats atualiza as estatísticas de mineração
func (sns *SafeNodeState) UpdateMiningStats(stats MiningStats) {
	sns.mu.Lock()
	defer sns.mu.Unlock()

	sns.state.MiningStats = stats
	sns.state.LastUpdate = time.Now()
}

// UpdateBalance atualiza o saldo
func (sns *SafeNodeState) UpdateBalance(balance int64) {
	sns.mu.Lock()
	defer sns.mu.Unlock()

	sns.state.Balance = balance
	sns.state.LastUpdate = time.Now()
}

// UpdateStatus atualiza o status do node
func (sns *SafeNodeState) UpdateStatus(status string) {
	sns.mu.Lock()
	defer sns.mu.Unlock()

	sns.state.Status = status
	sns.state.LastUpdate = time.Now()
}
