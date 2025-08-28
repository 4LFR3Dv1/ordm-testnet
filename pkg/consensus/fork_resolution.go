package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"ordm-main/pkg/blockchain"
)

// ForkResolution implementa resolução de forks e consenso
type ForkResolution struct {
	mu              sync.RWMutex
	knownForks      map[string]*ForkInfo
	resolutionRules []ConsensusRule
	logger          func(string, ...interface{})
}

// ForkInfo representa informações sobre um fork
type ForkInfo struct {
	ID             string    `json:"id"`
	BlockNumber    int64     `json:"block_number"`
	Chains         []Chain   `json:"chains"`
	DetectedAt     time.Time `json:"detected_at"`
	ResolvedAt     time.Time `json:"resolved_at,omitempty"`
	WinningChain   string    `json:"winning_chain,omitempty"`
	Status         string    `json:"status"` // "detected", "resolving", "resolved"
	ResolutionRule string    `json:"resolution_rule,omitempty"`
}

// Chain representa uma cadeia de blocos
type Chain struct {
	ID               string                  `json:"id"`
	Blocks           []*blockchain.RealBlock `json:"blocks"`
	TotalDifficulty  uint64                  `json:"total_difficulty"`
	TotalStake       int64                   `json:"total_stake"`
	TransactionCount int                     `json:"transaction_count"`
	LastTimestamp    int64                   `json:"last_timestamp"`
	Length           int                     `json:"length"`
}

// ConsensusRule define uma regra de consenso
type ConsensusRule struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// NewForkResolution cria um novo sistema de resolução de forks
func NewForkResolution(logger func(string, ...interface{})) *ForkResolution {
	fr := &ForkResolution{
		knownForks: make(map[string]*ForkInfo),
		logger:     logger,
	}

	// Definir regras de consenso
	fr.resolutionRules = []ConsensusRule{
		{
			Name:        "total_difficulty",
			Weight:      0.4,
			Description: "Maior dificuldade total acumulada (PoW)",
		},
		{
			Name:        "total_stake",
			Weight:      0.3,
			Description: "Maior stake total (PoS)",
		},
		{
			Name:        "transaction_count",
			Weight:      0.2,
			Description: "Maior número de transações",
		},
		{
			Name:        "timestamp",
			Weight:      0.1,
			Description: "Timestamp mais recente",
		},
	}

	return fr
}

// DetectFork detecta se há um fork na blockchain
func (fr *ForkResolution) DetectFork(block *blockchain.RealBlock, existingBlocks map[string]*blockchain.RealBlock) (*ForkInfo, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	// Verificar se já existe um bloco com o mesmo número
	var conflictingBlocks []*blockchain.RealBlock
	for _, existingBlock := range existingBlocks {
		if existingBlock.Header.Number == block.Header.Number {
			conflictingBlocks = append(conflictingBlocks, existingBlock)
		}
	}

	// Se não há conflitos, não há fork
	if len(conflictingBlocks) == 0 {
		return nil, nil
	}

	// Criar informações do fork
	forkID := fr.generateForkID(block.Header.Number)
	forkInfo := &ForkInfo{
		ID:          forkID,
		BlockNumber: block.Header.Number,
		DetectedAt:  time.Now(),
		Status:      "detected",
		Chains:      []Chain{},
	}

	// Adicionar o novo bloco aos conflitos
	conflictingBlocks = append(conflictingBlocks, block)

	// Agrupar blocos conflitantes em cadeias
	chains := fr.groupBlocksIntoChains(conflictingBlocks, existingBlocks)
	forkInfo.Chains = chains

	// Registrar o fork
	fr.knownForks[forkID] = forkInfo

	fr.logger("🔍 Fork detectado no bloco #%d com %d cadeias", block.Header.Number, len(chains))
	return forkInfo, nil
}

// ResolveFork resolve um fork usando as regras de consenso
func (fr *ForkResolution) ResolveFork(forkID string) (*Chain, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	forkInfo, exists := fr.knownForks[forkID]
	if !exists {
		return nil, fmt.Errorf("fork não encontrado: %s", forkID)
	}

	if forkInfo.Status == "resolved" {
		return nil, fmt.Errorf("fork já foi resolvido")
	}

	fr.logger("⚖️ Resolvendo fork %s com %d cadeias", forkID, len(forkInfo.Chains))

	// Calcular scores para cada cadeia
	var bestChain *Chain
	var bestScore float64
	var winningRule string

	for i, chain := range forkInfo.Chains {
		score := fr.calculateChainScore(chain)
		fr.logger("📊 Cadeia %s: score %.2f", chain.ID, score)

		if score > bestScore {
			bestScore = score
			bestChain = &forkInfo.Chains[i]
			winningRule = fr.getWinningRule(chain)
		}
	}

	if bestChain == nil {
		return nil, fmt.Errorf("não foi possível determinar a cadeia vencedora")
	}

	// Marcar fork como resolvido
	forkInfo.Status = "resolved"
	forkInfo.ResolvedAt = time.Now()
	forkInfo.WinningChain = bestChain.ID
	forkInfo.ResolutionRule = winningRule

	fr.logger("✅ Fork %s resolvido: cadeia %s venceu (score: %.2f, regra: %s)",
		forkID, bestChain.ID, bestScore, winningRule)

	return bestChain, nil
}

// calculateChainScore calcula o score de uma cadeia baseado nas regras de consenso
func (fr *ForkResolution) calculateChainScore(chain Chain) float64 {
	var totalScore float64

	for _, rule := range fr.resolutionRules {
		var score float64

		switch rule.Name {
		case "total_difficulty":
			// Normalizar dificuldade (0-1)
			score = float64(chain.TotalDifficulty) / 1000000.0 // Normalizar para milhões
		case "total_stake":
			// Normalizar stake (0-1)
			score = float64(chain.TotalStake) / 10000.0 // Normalizar para 10k
		case "transaction_count":
			// Normalizar número de transações (0-1)
			score = float64(chain.TransactionCount) / 1000.0 // Normalizar para 1k
		case "timestamp":
			// Timestamp mais recente (0-1)
			currentTime := time.Now().Unix()
			age := currentTime - chain.LastTimestamp
			if age < 0 {
				age = 0
			}
			score = 1.0 - (float64(age) / 3600.0) // Normalizar para 1 hora
			if score < 0 {
				score = 0
			}
		}

		totalScore += score * rule.Weight
	}

	return totalScore
}

// getWinningRule determina qual regra foi decisiva para a vitória
func (fr *ForkResolution) getWinningRule(chain Chain) string {
	var bestRule string
	var bestScore float64

	for _, rule := range fr.resolutionRules {
		var score float64

		switch rule.Name {
		case "total_difficulty":
			score = float64(chain.TotalDifficulty) / 1000000.0
		case "total_stake":
			score = float64(chain.TotalStake) / 10000.0
		case "transaction_count":
			score = float64(chain.TransactionCount) / 1000.0
		case "timestamp":
			currentTime := time.Now().Unix()
			age := currentTime - chain.LastTimestamp
			if age < 0 {
				age = 0
			}
			score = 1.0 - (float64(age) / 3600.0)
			if score < 0 {
				score = 0
			}
		}

		if score > bestScore {
			bestScore = score
			bestRule = rule.Name
		}
	}

	return bestRule
}

// groupBlocksIntoChains agrupa blocos conflitantes em cadeias
func (fr *ForkResolution) groupBlocksIntoChains(conflictingBlocks []*blockchain.RealBlock, allBlocks map[string]*blockchain.RealBlock) []Chain {
	var chains []Chain
	processedChains := make(map[string]bool)

	for _, block := range conflictingBlocks {
		chainID := fr.generateChainID(block)

		if processedChains[chainID] {
			continue
		}

		// Construir cadeia a partir deste bloco
		chain := fr.buildChainFromBlock(block, allBlocks)
		chains = append(chains, chain)
		processedChains[chainID] = true
	}

	return chains
}

// buildChainFromBlock constrói uma cadeia a partir de um bloco
func (fr *ForkResolution) buildChainFromBlock(startBlock *blockchain.RealBlock, allBlocks map[string]*blockchain.RealBlock) Chain {
	var blocks []*blockchain.RealBlock
	var totalDifficulty uint64
	var totalStake int64
	var transactionCount int
	var lastTimestamp int64

	// Adicionar bloco inicial
	blocks = append(blocks, startBlock)
	totalDifficulty += startBlock.Header.Difficulty
	transactionCount += len(startBlock.Transactions)
	lastTimestamp = startBlock.Header.Timestamp

	// Construir cadeia para trás
	currentBlock := startBlock
	for {
		// Procurar bloco pai
		var parentBlock *blockchain.RealBlock
		for _, block := range allBlocks {
			if fr.blocksAreConnected(currentBlock, block) {
				parentBlock = block
				break
			}
		}

		if parentBlock == nil {
			break
		}

		blocks = append([]*blockchain.RealBlock{parentBlock}, blocks...)
		totalDifficulty += parentBlock.Header.Difficulty
		transactionCount += len(parentBlock.Transactions)
		lastTimestamp = parentBlock.Header.Timestamp
		currentBlock = parentBlock
	}

	chainID := fr.generateChainID(startBlock)
	return Chain{
		ID:               chainID,
		Blocks:           blocks,
		TotalDifficulty:  totalDifficulty,
		TotalStake:       totalStake, // TODO: Implementar cálculo de stake
		TransactionCount: transactionCount,
		LastTimestamp:    lastTimestamp,
		Length:           len(blocks),
	}
}

// blocksAreConnected verifica se dois blocos estão conectados (pai-filho)
func (fr *ForkResolution) blocksAreConnected(child, parent *blockchain.RealBlock) bool {
	// Verificar se o hash do pai do filho corresponde ao hash do pai
	childParentHash := hex.EncodeToString(child.Header.ParentHash)
	parentHash := fr.calculateBlockHash(parent)

	return childParentHash == parentHash
}

// calculateBlockHash calcula o hash de um bloco
func (fr *ForkResolution) calculateBlockHash(block *blockchain.RealBlock) string {
	data := fmt.Sprintf("%d:%d:%s:%d",
		block.Header.Number,
		block.Header.Timestamp,
		block.Header.MinerID,
		block.Header.Nonce,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// generateForkID gera um ID único para um fork
func (fr *ForkResolution) generateForkID(blockNumber int64) string {
	data := fmt.Sprintf("fork_%d_%d", blockNumber, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // Primeiros 8 bytes
}

// generateChainID gera um ID único para uma cadeia
func (fr *ForkResolution) generateChainID(block *blockchain.RealBlock) string {
	data := fmt.Sprintf("chain_%d_%s", block.Header.Number, block.Header.MinerID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // Primeiros 8 bytes
}

// GetForkInfo retorna informações sobre um fork
func (fr *ForkResolution) GetForkInfo(forkID string) (*ForkInfo, bool) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	forkInfo, exists := fr.knownForks[forkID]
	return forkInfo, exists
}

// GetAllForks retorna todos os forks conhecidos
func (fr *ForkResolution) GetAllForks() []*ForkInfo {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	var forks []*ForkInfo
	for _, fork := range fr.knownForks {
		forks = append(forks, fork)
	}

	return forks
}

// GetForkStats retorna estatísticas dos forks
func (fr *ForkResolution) GetForkStats() map[string]interface{} {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	totalForks := len(fr.knownForks)
	resolvedForks := 0
	detectedForks := 0

	for _, fork := range fr.knownForks {
		if fork.Status == "resolved" {
			resolvedForks++
		} else if fork.Status == "detected" {
			detectedForks++
		}
	}

	return map[string]interface{}{
		"total_forks":     totalForks,
		"resolved_forks":  resolvedForks,
		"detected_forks":  detectedForks,
		"resolution_rate": float64(resolvedForks) / float64(totalForks),
	}
}
