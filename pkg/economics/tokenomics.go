package economics

import (
	"fmt"
	"time"
)

// Tokenomics implementa economia sustentável da blockchain
type Tokenomics struct {
	InitialReward    int64   `json:"initial_reward"`     // Recompensa inicial por bloco
	HalvingInterval  int64   `json:"halving_interval"`   // Blocos entre halvings
	MaxSupply        int64   `json:"max_supply"`         // Supply máximo
	BurnRate         float64 `json:"burn_rate"`          // Taxa de queima (0.0-1.0)
	StakeRewardRate  float64 `json:"stake_reward_rate"`  // Taxa de recompensa por stake
	ValidatorReward  float64 `json:"validator_reward"`   // Recompensa adicional para validadores
	InflationRate    float64 `json:"inflation_rate"`     // Taxa de inflação anual
	DeflationRate    float64 `json:"deflation_rate"`     // Taxa de deflação anual
	TotalBurned      int64   `json:"total_burned"`       // Total de tokens queimados
	TotalMinted      int64   `json:"total_minted"`       // Total de tokens minerados
	CurrentEpoch     int64   `json:"current_epoch"`      // Época atual
	LastHalvingBlock int64   `json:"last_halving_block"` // Último bloco de halving
	NextHalvingBlock int64   `json:"next_halving_block"` // Próximo bloco de halving
}

// NewTokenomics cria nova instância de tokenomics
func NewTokenomics() *Tokenomics {
	return &Tokenomics{
		InitialReward:    50,       // 50 tokens por bloco inicial
		HalvingInterval:  210000,   // Halving a cada 210k blocos (como Bitcoin)
		MaxSupply:        21000000, // 21M tokens máximo
		BurnRate:         0.1,      // 10% das taxas são queimadas
		StakeRewardRate:  0.05,     // 5% APY para stake
		ValidatorReward:  0.02,     // 2% adicional para validadores
		InflationRate:    0.0,      // Sem inflação
		DeflationRate:    0.02,     // 2% deflação anual
		TotalBurned:      0,
		TotalMinted:      0,
		CurrentEpoch:     0,
		LastHalvingBlock: 0,
		NextHalvingBlock: 210000,
	}
}

// CalculateMiningReward calcula recompensa de mineração com halving
func (t *Tokenomics) CalculateMiningReward(blockNumber int64) int64 {
	// Calcular época atual
	epoch := blockNumber / t.HalvingInterval

	// Aplicar halving
	reward := t.InitialReward
	for i := int64(0); i < epoch; i++ {
		reward = reward / 2
		if reward <= 0 {
			reward = 1 // Mínimo de 1 token
			break
		}
	}

	// Verificar se atingiu supply máximo
	if t.TotalMinted+reward > t.MaxSupply {
		reward = t.MaxSupply - t.TotalMinted
		if reward < 0 {
			reward = 0
		}
	}

	return reward
}

// CalculateStakeReward calcula recompensa por stake
func (t *Tokenomics) CalculateStakeReward(stakeAmount int64, stakeDuration time.Duration) int64 {
	// Calcular APY baseado na duração do stake
	years := stakeDuration.Hours() / 8760 // 8760 horas = 1 ano

	// Recompensa = stake * taxa * anos
	reward := float64(stakeAmount) * t.StakeRewardRate * years

	// Verificar supply máximo
	if t.TotalMinted+int64(reward) > t.MaxSupply {
		reward = float64(t.MaxSupply - t.TotalMinted)
		if reward < 0 {
			reward = 0
		}
	}

	return int64(reward)
}

// CalculateValidatorReward calcula recompensa adicional para validadores
func (t *Tokenomics) CalculateValidatorReward(baseReward int64, validatorStake int64) int64 {
	// Recompensa adicional baseada no stake do validador
	additionalReward := float64(baseReward) * t.ValidatorReward

	// Verificar supply máximo
	if t.TotalMinted+int64(additionalReward) > t.MaxSupply {
		additionalReward = float64(t.MaxSupply - t.TotalMinted)
		if additionalReward < 0 {
			additionalReward = 0
		}
	}

	return int64(additionalReward)
}

// CalculateTransactionFee calcula taxa de transação
func (t *Tokenomics) CalculateTransactionFee(amount int64, priority string) int64 {
	// Taxa base
	baseFee := int64(1)

	// Taxa baseada no valor da transação (0.1%)
	valueFee := amount / 1000

	// Taxa baseada na prioridade
	priorityMultiplier := 1.0
	switch priority {
	case "low":
		priorityMultiplier = 0.5
	case "normal":
		priorityMultiplier = 1.0
	case "high":
		priorityMultiplier = 2.0
	case "urgent":
		priorityMultiplier = 5.0
	}

	totalFee := int64(float64(baseFee+valueFee) * priorityMultiplier)

	// Mínimo de 1 token
	if totalFee < 1 {
		totalFee = 1
	}

	return totalFee
}

// BurnTokens queima tokens (reduz supply)
func (t *Tokenomics) BurnTokens(amount int64) {
	t.TotalBurned += amount
}

// BurnTransactionFee queima parte da taxa de transação
func (t *Tokenomics) BurnTransactionFee(fee int64) int64 {
	burnAmount := int64(float64(fee) * t.BurnRate)
	t.BurnTokens(burnAmount)
	return burnAmount
}

// MintTokens cria novos tokens
func (t *Tokenomics) MintTokens(amount int64) bool {
	if t.TotalMinted+amount > t.MaxSupply {
		return false
	}
	t.TotalMinted += amount
	return true
}

// GetCurrentSupply retorna supply atual
func (t *Tokenomics) GetCurrentSupply() int64 {
	return t.TotalMinted - t.TotalBurned
}

// GetCirculatingSupply retorna supply em circulação
func (t *Tokenomics) GetCirculatingSupply() int64 {
	// Supply em circulação = Total minerado - Total queimado - Stake total
	return t.GetCurrentSupply() // Simplificado por enquanto
}

// GetInflationRate calcula taxa de inflação atual
func (t *Tokenomics) GetInflationRate() float64 {
	if t.GetCurrentSupply() == 0 {
		return 0
	}

	// Taxa de inflação = (Novos tokens / Supply atual) * 100
	// Como temos halving, a inflação diminui ao longo do tempo
	currentReward := t.CalculateMiningReward(t.CurrentEpoch * t.HalvingInterval)
	blocksPerYear := int64(52560) // ~10 minutos por bloco
	annualInflation := float64(currentReward*blocksPerYear) / float64(t.GetCurrentSupply()) * 100

	return annualInflation
}

// GetDeflationRate calcula taxa de deflação por queima
func (t *Tokenomics) GetDeflationRate() float64 {
	if t.GetCurrentSupply() == 0 {
		return 0
	}

	// Taxa de deflação = (Tokens queimados / Supply atual) * 100
	return float64(t.TotalBurned) / float64(t.GetCurrentSupply()) * 100
}

// GetHalvingInfo retorna informações sobre halving
func (t *Tokenomics) GetHalvingInfo(blockNumber int64) map[string]interface{} {
	currentEpoch := blockNumber / t.HalvingInterval
	nextHalvingBlock := (currentEpoch + 1) * t.HalvingInterval
	blocksUntilHalving := nextHalvingBlock - blockNumber

	currentReward := t.CalculateMiningReward(blockNumber)
	nextReward := t.CalculateMiningReward(nextHalvingBlock)

	return map[string]interface{}{
		"current_epoch":        currentEpoch,
		"current_reward":       currentReward,
		"next_reward":          nextReward,
		"next_halving_block":   nextHalvingBlock,
		"blocks_until_halving": blocksUntilHalving,
		"halving_interval":     t.HalvingInterval,
		"reduction_percentage": 50.0,
	}
}

// GetEconomicMetrics retorna métricas econômicas
func (t *Tokenomics) GetEconomicMetrics(blockNumber int64) map[string]interface{} {
	return map[string]interface{}{
		"total_supply":       t.MaxSupply,
		"current_supply":     t.GetCurrentSupply(),
		"circulating_supply": t.GetCirculatingSupply(),
		"total_minted":       t.TotalMinted,
		"total_burned":       t.TotalBurned,
		"burn_rate":          t.BurnRate,
		"stake_reward_rate":  t.StakeRewardRate,
		"validator_reward":   t.ValidatorReward,
		"inflation_rate":     t.GetInflationRate(),
		"deflation_rate":     t.GetDeflationRate(),
		"current_reward":     t.CalculateMiningReward(blockNumber),
		"halving_info":       t.GetHalvingInfo(blockNumber),
		"economic_health":    t.CalculateEconomicHealth(),
	}
}

// CalculateEconomicHealth calcula saúde econômica
func (t *Tokenomics) CalculateEconomicHealth() string {
	supplyUtilization := float64(t.GetCurrentSupply()) / float64(t.MaxSupply) * 100
	inflationRate := t.GetInflationRate()
	deflationRate := t.GetDeflationRate()

	// Critérios de saúde econômica
	if supplyUtilization < 50 && inflationRate < 2 && deflationRate > 0.5 {
		return "excellent"
	} else if supplyUtilization < 70 && inflationRate < 5 && deflationRate > 0.2 {
		return "good"
	} else if supplyUtilization < 90 && inflationRate < 10 {
		return "fair"
	} else {
		return "poor"
	}
}

// PredictFutureSupply prevê supply futuro
func (t *Tokenomics) PredictFutureSupply(blocksAhead int64) int64 {
	predictedMinted := t.TotalMinted
	predictedBurned := t.TotalBurned

	// Simular mineração futura
	for i := int64(0); i < blocksAhead; i++ {
		blockNumber := t.CurrentEpoch*t.HalvingInterval + i
		reward := t.CalculateMiningReward(blockNumber)
		predictedMinted += reward
	}

	// Simular queima futura (estimativa)
	estimatedBurnRate := float64(t.TotalBurned) / float64(t.TotalMinted) // Taxa histórica
	predictedBurned += int64(float64(predictedMinted-t.TotalMinted) * estimatedBurnRate)

	return predictedMinted - predictedBurned
}

// GetTokenDistribution retorna distribuição de tokens
func (t *Tokenomics) GetTokenDistribution() map[string]interface{} {
	totalSupply := t.GetCurrentSupply()

	if totalSupply == 0 {
		return map[string]interface{}{
			"miners":      0,
			"validators":  0,
			"burned":      0,
			"circulating": 0,
		}
	}

	return map[string]interface{}{
		"miners":      float64(t.TotalMinted) / float64(totalSupply) * 100,
		"validators":  float64(t.TotalMinted*int64(t.StakeRewardRate*100)) / float64(totalSupply),
		"burned":      float64(t.TotalBurned) / float64(totalSupply) * 100,
		"circulating": float64(t.GetCirculatingSupply()) / float64(totalSupply) * 100,
	}
}

// ValidateTransaction valida transação economicamente
func (t *Tokenomics) ValidateTransaction(amount, fee int64, fromBalance int64) error {
	// Verificar saldo suficiente
	if fromBalance < amount+fee {
		return fmt.Errorf("saldo insuficiente: %d < %d", fromBalance, amount+fee)
	}

	// Verificar limites de transação
	if amount <= 0 {
		return fmt.Errorf("valor da transação deve ser positivo")
	}

	if fee <= 0 {
		return fmt.Errorf("taxa deve ser positiva")
	}

	// Verificar se não excede supply máximo
	if t.GetCurrentSupply()+amount > t.MaxSupply {
		return fmt.Errorf("transação excederia supply máximo")
	}

	return nil
}

// GetStakingMetrics retorna métricas de staking
func (t *Tokenomics) GetStakingMetrics(totalStaked int64) map[string]interface{} {
	totalSupply := t.GetCurrentSupply()

	if totalSupply == 0 {
		return map[string]interface{}{
			"staking_ratio":         0.0,
			"apy":                   t.StakeRewardRate * 100,
			"validator_apy":         (t.StakeRewardRate + t.ValidatorReward) * 100,
			"total_staked":          0,
			"staking_participation": 0.0,
		}
	}

	stakingRatio := float64(totalStaked) / float64(totalSupply) * 100

	return map[string]interface{}{
		"staking_ratio":         stakingRatio,
		"apy":                   t.StakeRewardRate * 100,
		"validator_apy":         (t.StakeRewardRate + t.ValidatorReward) * 100,
		"total_staked":          totalStaked,
		"staking_participation": stakingRatio,
		"min_stake":             1000,              // Stake mínimo
		"max_stake":             totalSupply / 100, // Máximo 1% do supply
	}
}

// String retorna representação string do tokenomics
func (t *Tokenomics) String() string {
	return fmt.Sprintf(
		"Tokenomics{Supply: %d/%d, Burned: %d, Epoch: %d, Reward: %d}",
		t.GetCurrentSupply(), t.MaxSupply, t.TotalBurned, t.CurrentEpoch,
		t.CalculateMiningReward(t.CurrentEpoch*t.HalvingInterval),
	)
}
