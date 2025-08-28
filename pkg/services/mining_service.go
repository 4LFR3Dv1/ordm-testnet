package services

import (
	"fmt"
	"sync"
	"time"

	"ordm-main/pkg/ledger"
	"ordm-main/pkg/state"
	"ordm-main/pkg/wallet"
)

type MiningService struct {
	state    *state.SafeNodeState
	wallet   *wallet.WalletManager
	ledger   *ledger.GlobalLedger
	mu       sync.RWMutex
	isActive bool
	stopChan chan struct{}
}

type MiningStatus struct {
	IsActive     bool      `json:"is_active"`
	TotalBlocks  int64     `json:"total_blocks"`
	HashRate     float64   `json:"hash_rate"`
	LastMined    time.Time `json:"last_mined"`
	TotalRewards int64     `json:"total_rewards"`
	CurrentBlock int64     `json:"current_block"`
}

func NewMiningService(state *state.SafeNodeState, wallet *wallet.WalletManager, ledger *ledger.GlobalLedger) *MiningService {
	return &MiningService{
		state:    state,
		wallet:   wallet,
		ledger:   ledger,
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

	// Atualizar ledger (simulado)
	if ms.ledger != nil {
		// Simular adição de bloco ao ledger
		// Em uma implementação real, seria algo como:
		// ms.ledger.RegisterBlock(block)
		fmt.Printf("📊 Bloco #%d processado pelo ledger\n", nodeInfo.MiningStats.TotalBlocks)
	}
}
