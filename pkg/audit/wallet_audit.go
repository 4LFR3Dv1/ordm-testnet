package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// WalletAudit representa auditoria de wallet
type WalletAudit struct {
	WalletID         string                 `json:"wallet_id"`
	Address          string                 `json:"address"`
	Balance          int64                  `json:"balance"`
	TotalReceived    int64                  `json:"total_received"`
	TotalSent        int64                  `json:"total_sent"`
	TransactionCount int64                  `json:"transaction_count"`
	FirstSeen        time.Time              `json:"first_seen"`
	LastSeen         time.Time              `json:"last_seen"`
	Movements        []MovementAudit        `json:"movements"`
	RiskScore        float64                `json:"risk_score"`
	Tags             []string               `json:"tags"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// MovementAudit representa auditoria de movimento
type MovementAudit struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // "in", "out", "stake", "reward"
	Amount      int64     `json:"amount"`
	Fee         int64     `json:"fee"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	BlockHash   string    `json:"block_hash"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	RiskLevel   string    `json:"risk_level"` // "low", "medium", "high"
}

// AuditManager gerencia auditoria de wallets
type AuditManager struct {
	Audits     map[string]*WalletAudit `json:"audits"`
	AuditPath  string                  `json:"audit_path"`
	RiskRules  []RiskRule              `json:"risk_rules"`
	LastUpdate time.Time               `json:"last_update"`
}

// RiskRule define regra de risco
type RiskRule struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	RiskScore   float64 `json:"risk_score"`
	Condition   string  `json:"condition"`
}

// NewAuditManager cria novo gerenciador de auditoria
func NewAuditManager(auditPath string) *AuditManager {
	manager := &AuditManager{
		Audits:    make(map[string]*WalletAudit),
		AuditPath: auditPath,
		RiskRules: []RiskRule{
			{
				Name:        "High Transaction Volume",
				Description: "Muitas transações em pouco tempo",
				RiskScore:   0.3,
				Condition:   "transaction_count > 100",
			},
			{
				Name:        "Large Amount Transfer",
				Description: "Transferência de valor alto",
				RiskScore:   0.4,
				Condition:   "amount > 10000",
			},
			{
				Name:        "New Wallet",
				Description: "Wallet criada recentemente",
				RiskScore:   0.2,
				Condition:   "age < 24h",
			},
			{
				Name:        "Suspicious Pattern",
				Description: "Padrão suspeito de movimentação",
				RiskScore:   0.5,
				Condition:   "pattern_suspicious",
			},
		},
		LastUpdate: time.Now(),
	}

	// Criar diretório se não existir
	os.MkdirAll(auditPath, 0755)

	return manager
}

// AddMovement adiciona movimento para auditoria
func (am *AuditManager) AddMovement(walletID, address string, movement MovementAudit) {
	// Criar ou atualizar audit da wallet
	audit, exists := am.Audits[walletID]
	if !exists {
		audit = &WalletAudit{
			WalletID:         walletID,
			Address:          address,
			Balance:          0,
			TotalReceived:    0,
			TotalSent:        0,
			TransactionCount: 0,
			FirstSeen:        movement.Timestamp,
			LastSeen:         movement.Timestamp,
			Movements:        []MovementAudit{},
			RiskScore:        0.0,
			Tags:             []string{},
			Metadata:         make(map[string]interface{}),
		}
		am.Audits[walletID] = audit
	}

	// Atualizar estatísticas
	audit.LastSeen = movement.Timestamp
	audit.TransactionCount++
	audit.Movements = append(audit.Movements, movement)

	// Calcular valores
	if movement.Type == "in" || movement.Type == "reward" {
		audit.TotalReceived += movement.Amount
	} else if movement.Type == "out" || movement.Type == "stake" {
		audit.TotalSent += movement.Amount
	}

	// Calcular saldo
	audit.Balance = audit.TotalReceived - audit.TotalSent

	// Calcular risco
	audit.RiskScore = am.calculateRiskScore(audit)

	// Adicionar tags
	am.addTags(audit)

	am.LastUpdate = time.Now()
}

// GetWalletAudit retorna auditoria de wallet
func (am *AuditManager) GetWalletAudit(walletID string) (*WalletAudit, bool) {
	audit, exists := am.Audits[walletID]
	return audit, exists
}

// GetAllAudits retorna todas as auditorias
func (am *AuditManager) GetAllAudits() []*WalletAudit {
	audits := make([]*WalletAudit, 0, len(am.Audits))
	for _, audit := range am.Audits {
		audits = append(audits, audit)
	}

	// Ordenar por risco (maior primeiro)
	sort.Slice(audits, func(i, j int) bool {
		return audits[i].RiskScore > audits[j].RiskScore
	})

	return audits
}

// GetHighRiskWallets retorna wallets de alto risco
func (am *AuditManager) GetHighRiskWallets() []*WalletAudit {
	var highRisk []*WalletAudit
	for _, audit := range am.Audits {
		if audit.RiskScore > 0.7 {
			highRisk = append(highRisk, audit)
		}
	}
	return highRisk
}

// GetWalletMovements retorna movimentos de wallet
func (am *AuditManager) GetWalletMovements(walletID string, limit int) []MovementAudit {
	audit, exists := am.Audits[walletID]
	if !exists {
		return []MovementAudit{}
	}

	movements := audit.Movements

	// Ordenar por timestamp (mais recente primeiro)
	sort.Slice(movements, func(i, j int) bool {
		return movements[i].Timestamp.After(movements[j].Timestamp)
	})

	// Limitar resultados
	if limit > 0 && len(movements) > limit {
		movements = movements[:limit]
	}

	return movements
}

// GetWalletStats retorna estatísticas de wallet
func (am *AuditManager) GetWalletStats(walletID string) map[string]interface{} {
	audit, exists := am.Audits[walletID]
	if !exists {
		return map[string]interface{}{
			"error": "Wallet não encontrada",
		}
	}

	// Calcular estatísticas
	var totalFees int64
	var avgAmount float64
	var maxAmount int64
	var minAmount int64

	if len(audit.Movements) > 0 {
		minAmount = audit.Movements[0].Amount
		for _, movement := range audit.Movements {
			totalFees += movement.Fee
			if movement.Amount > maxAmount {
				maxAmount = movement.Amount
			}
			if movement.Amount < minAmount {
				minAmount = movement.Amount
			}
		}
		avgAmount = float64(audit.TotalReceived+audit.TotalSent) / float64(len(audit.Movements))
	}

	// Calcular idade da wallet
	age := time.Since(audit.FirstSeen)

	return map[string]interface{}{
		"wallet_id":         audit.WalletID,
		"address":           audit.Address,
		"balance":           audit.Balance,
		"total_received":    audit.TotalReceived,
		"total_sent":        audit.TotalSent,
		"transaction_count": audit.TransactionCount,
		"first_seen":        audit.FirstSeen,
		"last_seen":         audit.LastSeen,
		"age":               age.String(),
		"risk_score":        audit.RiskScore,
		"risk_level":        am.getRiskLevel(audit.RiskScore),
		"tags":              audit.Tags,
		"total_fees":        totalFees,
		"avg_amount":        avgAmount,
		"max_amount":        maxAmount,
		"min_amount":        minAmount,
		"movement_types":    am.getMovementTypes(audit),
	}
}

// GetSystemStats retorna estatísticas do sistema
func (am *AuditManager) GetSystemStats() map[string]interface{} {
	totalWallets := len(am.Audits)
	var totalBalance int64
	var totalTransactions int64
	var highRiskCount int
	var avgRiskScore float64

	for _, audit := range am.Audits {
		totalBalance += audit.Balance
		totalTransactions += audit.TransactionCount
		avgRiskScore += audit.RiskScore

		if audit.RiskScore > 0.7 {
			highRiskCount++
		}
	}

	if totalWallets > 0 {
		avgRiskScore /= float64(totalWallets)
	}

	return map[string]interface{}{
		"total_wallets":      totalWallets,
		"total_balance":      totalBalance,
		"total_transactions": totalTransactions,
		"high_risk_wallets":  highRiskCount,
		"avg_risk_score":     avgRiskScore,
		"last_update":        am.LastUpdate,
	}
}

// calculateRiskScore calcula score de risco
func (am *AuditManager) calculateRiskScore(audit *WalletAudit) float64 {
	riskScore := 0.0

	// Regra 1: Volume de transações
	if audit.TransactionCount > 100 {
		riskScore += 0.3
	}

	// Regra 2: Valor alto
	if audit.TotalReceived > 10000 || audit.TotalSent > 10000 {
		riskScore += 0.4
	}

	// Regra 3: Wallet nova
	age := time.Since(audit.FirstSeen)
	if age < 24*time.Hour {
		riskScore += 0.2
	}

	// Regra 4: Padrão suspeito
	if am.isSuspiciousPattern(audit) {
		riskScore += 0.5
	}

	// Limitar a 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore
}

// isSuspiciousPattern verifica padrão suspeito
func (am *AuditManager) isSuspiciousPattern(audit *WalletAudit) bool {
	if len(audit.Movements) < 3 {
		return false
	}

	// Verificar movimentos muito frequentes
	recentMovements := 0
	for _, movement := range audit.Movements {
		if time.Since(movement.Timestamp) < time.Hour {
			recentMovements++
		}
	}

	return recentMovements > 10
}

// addTags adiciona tags baseadas no comportamento
func (am *AuditManager) addTags(audit *WalletAudit) {
	tags := []string{}

	// Tag por idade
	age := time.Since(audit.FirstSeen)
	if age < 24*time.Hour {
		tags = append(tags, "new")
	} else if age < 7*24*time.Hour {
		tags = append(tags, "recent")
	} else {
		tags = append(tags, "established")
	}

	// Tag por volume
	if audit.TransactionCount > 100 {
		tags = append(tags, "high-volume")
	} else if audit.TransactionCount > 10 {
		tags = append(tags, "active")
	} else {
		tags = append(tags, "low-activity")
	}

	// Tag por risco
	if audit.RiskScore > 0.7 {
		tags = append(tags, "high-risk")
	} else if audit.RiskScore > 0.3 {
		tags = append(tags, "medium-risk")
	} else {
		tags = append(tags, "low-risk")
	}

	// Tag por saldo
	if audit.Balance > 1000 {
		tags = append(tags, "high-balance")
	} else if audit.Balance > 100 {
		tags = append(tags, "medium-balance")
	} else {
		tags = append(tags, "low-balance")
	}

	audit.Tags = tags
}

// getRiskLevel retorna nível de risco
func (am *AuditManager) getRiskLevel(score float64) string {
	if score > 0.7 {
		return "high"
	} else if score > 0.3 {
		return "medium"
	}
	return "low"
}

// getMovementTypes retorna tipos de movimentos
func (am *AuditManager) getMovementTypes(audit *WalletAudit) map[string]int {
	types := make(map[string]int)
	for _, movement := range audit.Movements {
		types[movement.Type]++
	}
	return types
}

// SaveAudits salva auditorias no disco
func (am *AuditManager) SaveAudits() error {
	data, err := json.MarshalIndent(am, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar auditorias: %v", err)
	}

	auditFile := filepath.Join(am.AuditPath, "wallet_audits.json")
	err = os.WriteFile(auditFile, data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar auditorias: %v", err)
	}

	return nil
}

// LoadAudits carrega auditorias do disco
func (am *AuditManager) LoadAudits() error {
	auditFile := filepath.Join(am.AuditPath, "wallet_audits.json")
	data, err := os.ReadFile(auditFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Arquivo não existe ainda
		}
		return fmt.Errorf("erro ao ler auditorias: %v", err)
	}

	err = json.Unmarshal(data, am)
	if err != nil {
		return fmt.Errorf("erro ao deserializar auditorias: %v", err)
	}

	return nil
}
