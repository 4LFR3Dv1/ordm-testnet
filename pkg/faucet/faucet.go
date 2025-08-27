package faucet

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// FaucetRequest representa uma requisição ao faucet
type FaucetRequest struct {
	Address     string    `json:"address"`
	Amount      int64     `json:"amount"`
	IP          string    `json:"ip"`
	Timestamp   time.Time `json:"timestamp"`
	RequestHash string    `json:"request_hash"`
	Status      string    `json:"status"`
}

// FaucetManager gerencia o faucet da testnet
type FaucetManager struct {
	Requests     map[string]*FaucetRequest `json:"requests"`
	Mutex        sync.RWMutex              `json:"-"`
	Config       *FaucetConfig             `json:"config"`
	FaucetWallet string                    `json:"faucet_wallet"`
	TotalSent    int64                     `json:"total_sent"`
	RequestCount int64                     `json:"request_count"`
}

// FaucetConfig configuração do faucet
type FaucetConfig struct {
	TestnetMode    bool  `json:"testnet_mode"`
	MaxAmount      int64 `json:"max_amount"`      // Máximo por requisição
	DailyLimit     int64 `json:"daily_limit"`     // Limite diário por IP
	RateLimit      int   `json:"rate_limit"`      // Requisições por hora por IP
	FaucetBalance  int64 `json:"faucet_balance"`  // Saldo do faucet
	MinAddressLen  int   `json:"min_address_len"` // Tamanho mínimo do endereço
	MaxAddressLen  int   `json:"max_address_len"` // Tamanho máximo do endereço
	RequestTimeout int   `json:"request_timeout"` // Timeout em segundos
}

// NewFaucetManager cria um novo gerenciador de faucet
func NewFaucetManager() *FaucetManager {
	return &FaucetManager{
		Requests: make(map[string]*FaucetRequest),
		Config: &FaucetConfig{
			TestnetMode:    true,
			MaxAmount:      50,    // 50 tokens por requisição
			DailyLimit:     100,   // 100 tokens por dia por IP
			RateLimit:      1,     // 1 requisição por hora por IP
			FaucetBalance:  10000, // 10k tokens iniciais
			MinAddressLen:  26,    // Endereço mínimo
			MaxAddressLen:  42,    // Endereço máximo
			RequestTimeout: 300,   // 5 minutos
		},
		FaucetWallet: "faucet_testnet_wallet_0000000000000000000000000000000000000000",
		TotalSent:    0,
		RequestCount: 0,
	}
}

// ProcessFaucetRequest processa uma requisição ao faucet
func (fm *FaucetManager) ProcessFaucetRequest(address, ip string, amount int64) (*FaucetRequest, error) {
	fm.Mutex.Lock()
	defer fm.Mutex.Unlock()

	// Validar endereço
	if err := fm.validateAddress(address); err != nil {
		return nil, fmt.Errorf("endereço inválido: %v", err)
	}

	// Validar quantidade
	if err := fm.validateAmount(amount); err != nil {
		return nil, fmt.Errorf("quantidade inválida: %v", err)
	}

	// Verificar rate limiting
	if err := fm.checkRateLimit(ip); err != nil {
		return nil, fmt.Errorf("rate limit excedido: %v", err)
	}

	// Verificar saldo do faucet
	if fm.Config.FaucetBalance < amount {
		return nil, fmt.Errorf("saldo insuficiente no faucet: %d < %d", fm.Config.FaucetBalance, amount)
	}

	// Criar requisição
	request := &FaucetRequest{
		Address:     address,
		Amount:      amount,
		IP:          ip,
		Timestamp:   time.Now(),
		RequestHash: fm.generateRequestHash(address, ip, amount),
		Status:      "pending",
	}

	// Registrar requisição
	fm.Requests[request.RequestHash] = request
	fm.RequestCount++

	// Processar transferência
	if err := fm.processTransfer(request); err != nil {
		request.Status = "failed"
		return request, fmt.Errorf("erro na transferência: %v", err)
	}

	// Atualizar estatísticas
	fm.Config.FaucetBalance -= amount
	fm.TotalSent += amount
	request.Status = "completed"

	log.Printf("💰 Faucet: %d tokens enviados para %s (IP: %s)", amount, address, ip)
	return request, nil
}

// validateAddress valida o endereço da wallet
func (fm *FaucetManager) validateAddress(address string) error {
	if len(address) < fm.Config.MinAddressLen || len(address) > fm.Config.MaxAddressLen {
		return fmt.Errorf("tamanho do endereço inválido: %d (deve ser entre %d e %d)",
			len(address), fm.Config.MinAddressLen, fm.Config.MaxAddressLen)
	}

	// Verificar se é um endereço válido (hex)
	if _, err := hex.DecodeString(address); err != nil {
		return fmt.Errorf("endereço deve ser hexadecimal válido")
	}

	return nil
}

// validateAmount valida a quantidade solicitada
func (fm *FaucetManager) validateAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("quantidade deve ser maior que zero")
	}

	if amount > fm.Config.MaxAmount {
		return fmt.Errorf("quantidade excede o máximo permitido: %d > %d", amount, fm.Config.MaxAmount)
	}

	return nil
}

// checkRateLimit verifica rate limiting por IP
func (fm *FaucetManager) checkRateLimit(ip string) error {
	now := time.Now()
	hourAgo := now.Add(-time.Hour)
	dailyLimit := now.Add(-24 * time.Hour)

	hourlyCount := 0
	dailyAmount := int64(0)

	for _, req := range fm.Requests {
		if req.IP == ip {
			// Contar requisições na última hora
			if req.Timestamp.After(hourAgo) {
				hourlyCount++
			}

			// Somar quantidade nas últimas 24h
			if req.Timestamp.After(dailyLimit) && req.Status == "completed" {
				dailyAmount += req.Amount
			}
		}
	}

	// Verificar limite por hora
	if hourlyCount >= fm.Config.RateLimit {
		return fmt.Errorf("limite de %d requisições por hora excedido", fm.Config.RateLimit)
	}

	// Verificar limite diário
	if dailyAmount >= fm.Config.DailyLimit {
		return fmt.Errorf("limite diário de %d tokens excedido", fm.Config.DailyLimit)
	}

	return nil
}

// generateRequestHash gera hash único para a requisição
func (fm *FaucetManager) generateRequestHash(address, ip string, amount int64) string {
	data := fmt.Sprintf("%s:%s:%d:%d", address, ip, amount, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// processTransfer processa a transferência do faucet
func (fm *FaucetManager) processTransfer(request *FaucetRequest) error {
	// Aqui você integraria com o sistema de ledger
	// Por enquanto, simulamos a transferência

	log.Printf("🔄 Processando transferência: %d tokens de %s para %s",
		request.Amount, fm.FaucetWallet, request.Address)

	// Simular delay de processamento
	time.Sleep(100 * time.Millisecond)

	return nil
}

// GetFaucetStats retorna estatísticas do faucet
func (fm *FaucetManager) GetFaucetStats() map[string]interface{} {
	fm.Mutex.RLock()
	defer fm.Mutex.RUnlock()

	// Calcular estatísticas
	completedCount := 0
	failedCount := 0
	pendingCount := 0

	for _, req := range fm.Requests {
		switch req.Status {
		case "completed":
			completedCount++
		case "failed":
			failedCount++
		case "pending":
			pendingCount++
		}
	}

	return map[string]interface{}{
		"faucet_balance":     fm.Config.FaucetBalance,
		"total_sent":         fm.TotalSent,
		"total_requests":     fm.RequestCount,
		"completed_requests": completedCount,
		"failed_requests":    failedCount,
		"pending_requests":   pendingCount,
		"config":             fm.Config,
		"faucet_wallet":      fm.FaucetWallet,
	}
}

// GetRequestHistory retorna histórico de requisições
func (fm *FaucetManager) GetRequestHistory(limit int) []*FaucetRequest {
	fm.Mutex.RLock()
	defer fm.Mutex.RUnlock()

	var requests []*FaucetRequest
	for _, req := range fm.Requests {
		requests = append(requests, req)
	}

	// Ordenar por timestamp (mais recente primeiro)
	// Implementar ordenação se necessário

	if limit > 0 && len(requests) > limit {
		return requests[:limit]
	}

	return requests
}

// RefillFaucet recarrega o saldo do faucet
func (fm *FaucetManager) RefillFaucet(amount int64) error {
	fm.Mutex.Lock()
	defer fm.Mutex.Unlock()

	if amount <= 0 {
		return fmt.Errorf("quantidade deve ser maior que zero")
	}

	fm.Config.FaucetBalance += amount
	log.Printf("💰 Faucet recarregado: +%d tokens (saldo: %d)", amount, fm.Config.FaucetBalance)

	return nil
}

// CleanupOldRequests remove requisições antigas
func (fm *FaucetManager) CleanupOldRequests() {
	fm.Mutex.Lock()
	defer fm.Mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	deleted := 0

	for hash, req := range fm.Requests {
		if req.Timestamp.Before(cutoff) {
			delete(fm.Requests, hash)
			deleted++
		}
	}

	if deleted > 0 {
		log.Printf("🧹 Limpeza do faucet: %d requisições antigas removidas", deleted)
	}
}

// StartFaucetCleanup inicia limpeza automática
func (fm *FaucetManager) StartFaucetCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		fm.CleanupOldRequests()
	}
}
