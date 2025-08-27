package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// BlockCalculator calcula blocos minerados e recompensas
type BlockCalculator struct {
	mu              sync.RWMutex
	genesisBlock    *Block
	blocks          map[string]*Block
	blockHeight     int64
	totalSupply     int64
	initialReward   int64
	halvingInterval int64
	maxSupply       int64
	burnRate        float64
}

// Block representa um bloco na blockchain
type Block struct {
	Hash         string        `json:"hash"`
	ParentHash   string        `json:"parent_hash"`
	Number       int64         `json:"number"`
	Timestamp    int64         `json:"timestamp"`
	MinerID      string        `json:"miner_id"`
	Reward       int64         `json:"reward"`
	Difficulty   int           `json:"difficulty"`
	Nonce        uint64        `json:"nonce"`
	Transactions []Transaction `json:"transactions"`
	TotalFees    int64         `json:"total_fees"`
	BurnedTokens int64         `json:"burned_tokens"`
}

// Transaction representa uma transação
type Transaction struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Fee    int64  `json:"fee"`
	Data   string `json:"data"`
}

// NewBlockCalculator cria um novo calculador de blocos
func NewBlockCalculator() *BlockCalculator {
	// Criar bloco genesis
	genesis := &Block{
		Hash:       "0000000000000000000000000000000000000000000000000000000000000000",
		ParentHash: "",
		Number:     0,
		Timestamp:  time.Now().Unix(),
		MinerID:    "genesis",
		Reward:     0,
		Difficulty: 0,
		Nonce:      0,
	}

	return &BlockCalculator{
		genesisBlock:    genesis,
		blocks:          make(map[string]*Block),
		blockHeight:     0,
		totalSupply:     0,
		initialReward:   50,       // 50 tokens por bloco inicial
		halvingInterval: 210000,   // Halving a cada 210k blocos (como Bitcoin)
		maxSupply:       21000000, // 21M tokens máximo
		burnRate:        0.1,      // 10% das taxas são queimadas
	}
}

// CreateGenesisBlock cria o bloco genesis
func (bc *BlockCalculator) CreateGenesisBlock() *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.blocks[bc.genesisBlock.Hash] = bc.genesisBlock
	bc.blockHeight = 0

	fmt.Printf("🌱 Bloco Genesis criado: %s\n", bc.genesisBlock.Hash[:8])
	return bc.genesisBlock
}

// MineBlock minera um novo bloco
func (bc *BlockCalculator) MineBlock(parentHash string, transactions []Transaction, minerID string, difficulty int) (*Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Verificar se o bloco pai existe
	if parentHash != bc.genesisBlock.Hash {
		if _, exists := bc.blocks[parentHash]; !exists {
			return nil, fmt.Errorf("bloco pai não encontrado: %s", parentHash)
		}
	}

	// Calcular número do bloco
	blockNumber := bc.blockHeight + 1

	// Calcular recompensa com halving
	reward := bc.calculateMiningReward(blockNumber)

	// Calcular taxas totais
	totalFees := int64(0)
	for _, tx := range transactions {
		totalFees += tx.Fee
	}

	// Calcular tokens queimados (10% das taxas)
	burnedTokens := int64(float64(totalFees) * bc.burnRate)

	// Criar bloco
	block := &Block{
		ParentHash:   parentHash,
		Number:       blockNumber,
		Timestamp:    time.Now().Unix(),
		MinerID:      minerID,
		Reward:       reward,
		Difficulty:   difficulty,
		Transactions: transactions,
		TotalFees:    totalFees,
		BurnedTokens: burnedTokens,
	}

	// Mineração PoW
	nonce := uint64(0)
	for {
		block.Nonce = nonce
		block.Hash = bc.calculateBlockHash(block)

		if bc.verifyPoW(block.Hash, difficulty) {
			break
		}
		nonce++
	}

	// Adicionar bloco à blockchain
	bc.blocks[block.Hash] = block
	bc.blockHeight = blockNumber

	// Atualizar supply total
	bc.totalSupply += reward
	bc.totalSupply -= burnedTokens // Queimar tokens

	fmt.Printf("⛏️ Bloco #%d minerado: %s | Recompensa: %d | Taxas: %d | Queimados: %d | Supply: %d\n",
		blockNumber, block.Hash[:8], reward, totalFees, burnedTokens, bc.totalSupply)

	return block, nil
}

// calculateMiningReward calcula recompensa de mineração com halving
func (bc *BlockCalculator) calculateMiningReward(blockNumber int64) int64 {
	// Calcular época atual
	epoch := blockNumber / bc.halvingInterval

	// Aplicar halving
	reward := bc.initialReward
	for i := int64(0); i < epoch; i++ {
		reward = reward / 2
		if reward <= 0 {
			reward = 1 // Mínimo de 1 token
			break
		}
	}

	// Verificar se atingiu supply máximo
	if bc.totalSupply+reward > bc.maxSupply {
		reward = bc.maxSupply - bc.totalSupply
		if reward < 0 {
			reward = 0
		}
	}

	return reward
}

// calculateBlockHash calcula o hash de um bloco
func (bc *BlockCalculator) calculateBlockHash(block *Block) string {
	data := fmt.Sprintf("%s|%d|%d|%s|%d|%d|%d|%d",
		block.ParentHash,
		block.Number,
		block.Timestamp,
		block.MinerID,
		block.Difficulty,
		block.Nonce,
		block.TotalFees,
		block.BurnedTokens,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// verifyPoW verifica se o hash atende à dificuldade PoW
func (bc *BlockCalculator) verifyPoW(hash string, difficulty int) bool {
	zeros := 0
	for _, char := range hash {
		if char == '0' {
			zeros++
		} else {
			break
		}
	}
	return zeros >= difficulty
}

// GetBlockHeight retorna a altura atual da blockchain
func (bc *BlockCalculator) GetBlockHeight() int64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blockHeight
}

// GetTotalSupply retorna o supply total
func (bc *BlockCalculator) GetTotalSupply() int64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.totalSupply
}

// GetBlock retorna um bloco específico
func (bc *BlockCalculator) GetBlock(hash string) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blocks[hash]
}

// GetBlockByNumber retorna um bloco pelo número
func (bc *BlockCalculator) GetBlockByNumber(number int64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.blocks {
		if block.Number == number {
			return block
		}
	}
	return nil
}

// GetLatestBlock retorna o último bloco
func (bc *BlockCalculator) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var latest *Block
	for _, block := range bc.blocks {
		if latest == nil || block.Number > latest.Number {
			latest = block
		}
	}
	return latest
}

// GetMiningStats retorna estatísticas de mineração
func (bc *BlockCalculator) GetMiningStats() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["block_height"] = bc.blockHeight
	stats["total_supply"] = bc.totalSupply
	stats["max_supply"] = bc.maxSupply
	stats["remaining_supply"] = bc.maxSupply - bc.totalSupply
	stats["current_reward"] = bc.calculateMiningReward(bc.blockHeight + 1)
	stats["next_halving_block"] = ((bc.blockHeight / bc.halvingInterval) + 1) * bc.halvingInterval
	stats["blocks_until_halving"] = stats["next_halving_block"].(int64) - bc.blockHeight

	// Calcular estatísticas por minerador
	minerStats := make(map[string]int64)
	totalRewards := int64(0)
	for _, block := range bc.blocks {
		if block.Number > 0 { // Excluir genesis
			minerStats[block.MinerID] += block.Reward
			totalRewards += block.Reward
		}
	}
	stats["miner_stats"] = minerStats
	stats["total_rewards"] = totalRewards

	return stats
}

// GetBlockchainInfo retorna informações da blockchain
func (bc *BlockCalculator) GetBlockchainInfo() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	info := make(map[string]interface{})
	info["genesis_hash"] = bc.genesisBlock.Hash
	info["current_height"] = bc.blockHeight
	info["total_blocks"] = len(bc.blocks)
	info["total_supply"] = bc.totalSupply
	info["max_supply"] = bc.maxSupply
	info["burn_rate"] = bc.burnRate
	info["initial_reward"] = bc.initialReward
	info["halving_interval"] = bc.halvingInterval

	// Calcular taxa de inflação
	if bc.blockHeight > 0 {
		blocksPerYear := int64(365 * 24 * 60 * 60 / 10) // Assumindo 10s por bloco
		annualRewards := bc.calculateMiningReward(bc.blockHeight+1) * blocksPerYear
		inflationRate := float64(annualRewards) / float64(bc.totalSupply) * 100
		info["inflation_rate_percent"] = inflationRate
	}

	return info
}

// ValidateBlockchain valida a integridade da blockchain
func (bc *BlockCalculator) ValidateBlockchain() error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.blocks {
		if block.Number == 0 {
			continue // Genesis block
		}

		// Verificar hash do bloco
		calculatedHash := bc.calculateBlockHash(block)
		if calculatedHash != block.Hash {
			return fmt.Errorf("hash inválido no bloco #%d", block.Number)
		}

		// Verificar se o bloco pai existe
		if _, exists := bc.blocks[block.ParentHash]; !exists {
			return fmt.Errorf("bloco pai não encontrado para bloco #%d", block.Number)
		}

		// Verificar se o número do bloco está correto
		parentBlock := bc.blocks[block.ParentHash]
		if block.Number != parentBlock.Number+1 {
			return fmt.Errorf("número de bloco incorreto: #%d", block.Number)
		}
	}

	fmt.Println("✅ Blockchain validada com sucesso!")
	return nil
}
