package wallet

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SecureWallet representa uma wallet física segura
type SecureWallet struct {
	ID           string    `json:"id"`
	PublicKey    string    `json:"public_key"`
	Address      string    `json:"address"`
	Balance      int64     `json:"balance"`
	StakeAmount  int64     `json:"stake_amount"`
	IsValidator  bool      `json:"is_validator"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	Nonce        uint64    `json:"nonce"`
}

// SecureWalletManager gerencia wallets seguras
type SecureWalletManager struct {
	mu      sync.RWMutex
	wallets map[string]*SecureWallet
	genesis bool
}

// NewSecureWalletManager cria um novo gerenciador de wallets
func NewSecureWalletManager() *SecureWalletManager {
	return &SecureWalletManager{
		wallets: make(map[string]*SecureWallet),
		genesis: false,
	}
}

// CreateGenesisWallet cria a primeira wallet (genesis)
func (wm *SecureWalletManager) CreateGenesisWallet() (*SecureWallet, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.genesis {
		return nil, fmt.Errorf("wallet genesis já foi criada")
	}

	// Gerar public key (simulada)
	publicKeyBytes := make([]byte, 32)
	rand.Read(publicKeyBytes)
	publicKey := hex.EncodeToString(publicKeyBytes)

	// Gerar endereço
	address := wm.generateAddress(publicKeyBytes)

	// Verificar se começa com 3 zeros
	if !wm.isGenesisAddress(address) {
		return nil, fmt.Errorf("endereço genesis deve começar com 3 zeros")
	}

	wallet := &SecureWallet{
		ID:           fmt.Sprintf("genesis_%d", time.Now().Unix()),
		PublicKey:    publicKey,
		Address:      address,
		Balance:      1000000, // 1M tokens para genesis
		StakeAmount:  0,
		IsValidator:  false,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Nonce:        0,
	}

	wm.wallets[wallet.ID] = wallet
	wm.genesis = true

	fmt.Printf("🌱 Wallet Genesis criada: %s (Balance: %d)\n", address, wallet.Balance)
	return wallet, nil
}

// CreateWallet cria uma nova wallet
func (wm *SecureWalletManager) CreateWallet() (*SecureWallet, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Gerar public key (simulada)
	publicKeyBytes := make([]byte, 32)
	rand.Read(publicKeyBytes)
	publicKey := hex.EncodeToString(publicKeyBytes)

	// Gerar endereço
	address := wm.generateAddress(publicKeyBytes)

	wallet := &SecureWallet{
		ID:           fmt.Sprintf("wallet_%d", time.Now().UnixNano()),
		PublicKey:    publicKey,
		Address:      address,
		Balance:      0,
		StakeAmount:  0,
		IsValidator:  false,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Nonce:        0,
	}

	wm.wallets[wallet.ID] = wallet

	fmt.Printf("🔑 Nova wallet criada: %s\n", address)
	return wallet, nil
}

// GetWalletByPublicKey retorna wallet pela public key
func (wm *SecureWalletManager) GetWalletByPublicKey(publicKey string) (*SecureWallet, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, wallet := range wm.wallets {
		if wallet.PublicKey == publicKey {
			return wallet, true
		}
	}
	return nil, false
}

// GetWalletByAddress retorna wallet pelo endereço
func (wm *SecureWalletManager) GetWalletByAddress(address string) (*SecureWallet, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, wallet := range wm.wallets {
		if wallet.Address == address {
			return wallet, true
		}
	}
	return nil, false
}

// GetWalletByUserID retorna wallet pelo ID de usuário (8 primeiros dígitos da public key)
func (wm *SecureWalletManager) GetWalletByUserID(userID string) (*SecureWallet, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, wallet := range wm.wallets {
		if len(wallet.PublicKey) >= 8 {
			walletUserID := wallet.PublicKey[:8]
			if walletUserID == userID {
				return wallet, true
			}
		}
	}
	return nil, false
}

// GetAllWallets retorna todas as wallets
func (wm *SecureWalletManager) GetAllWallets() []*SecureWallet {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	wallets := make([]*SecureWallet, 0, len(wm.wallets))
	for _, wallet := range wm.wallets {
		wallets = append(wallets, wallet)
	}
	return wallets
}

// UpdateBalance atualiza o saldo de uma wallet
func (wm *SecureWalletManager) UpdateBalance(address string, amount int64) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, wallet := range wm.wallets {
		if wallet.Address == address {
			wallet.Balance += amount
			wallet.LastActivity = time.Now()
			return nil
		}
	}
	return fmt.Errorf("wallet não encontrada: %s", address)
}

// StakeTokens faz stake de tokens
func (wm *SecureWalletManager) StakeTokens(address string, amount int64) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, wallet := range wm.wallets {
		if wallet.Address == address {
			if wallet.Balance < amount {
				return fmt.Errorf("saldo insuficiente: %d < %d", wallet.Balance, amount)
			}

			wallet.Balance -= amount
			wallet.StakeAmount += amount
			wallet.LastActivity = time.Now()

			// Tornar validador se stake suficiente
			if wallet.StakeAmount >= 1000 {
				wallet.IsValidator = true
			}

			return nil
		}
	}
	return fmt.Errorf("wallet não encontrada: %s", address)
}

// UnstakeTokens remove stake de tokens
func (wm *SecureWalletManager) UnstakeTokens(address string, amount int64) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, wallet := range wm.wallets {
		if wallet.Address == address {
			if wallet.StakeAmount < amount {
				return fmt.Errorf("stake insuficiente: %d < %d", wallet.StakeAmount, amount)
			}

			wallet.StakeAmount -= amount
			wallet.Balance += amount
			wallet.LastActivity = time.Now()

			// Remover validador se stake insuficiente
			if wallet.StakeAmount < 1000 {
				wallet.IsValidator = false
			}

			return nil
		}
	}
	return fmt.Errorf("wallet não encontrada: %s", address)
}

// generateAddress gera endereço a partir da public key
func (wm *SecureWalletManager) generateAddress(publicKeyBytes []byte) string {
	// Hash da public key
	hash := sha256.Sum256(publicKeyBytes)

	// Codificar em hex
	address := hex.EncodeToString(hash[:])

	// Pegar primeiros 20 bytes
	if len(address) > 40 {
		address = address[:40]
	}

	return address
}

// isGenesisAddress verifica se o endereço é válido para genesis
func (wm *SecureWalletManager) isGenesisAddress(address string) bool {
	// Verificar se começa com 3 zeros
	if len(address) >= 3 {
		return address[:3] == "000"
	}
	return false
}

// GetUserIDFromPublicKey extrai ID de usuário da public key
func (wm *SecureWalletManager) GetUserIDFromPublicKey(publicKey string) string {
	if len(publicKey) >= 8 {
		return publicKey[:8]
	}
	return ""
}

// ValidateUserID valida se um userID corresponde a uma public key
func (wm *SecureWalletManager) ValidateUserID(userID, publicKey string) bool {
	if len(publicKey) >= 8 {
		return publicKey[:8] == userID
	}
	return false
}

// GetValidators retorna todas as wallets que são validadores
func (wm *SecureWalletManager) GetValidators() []*SecureWallet {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var validators []*SecureWallet
	for _, wallet := range wm.wallets {
		if wallet.IsValidator {
			validators = append(validators, wallet)
		}
	}
	return validators
}

// GetTotalStake retorna o total de stake na rede
func (wm *SecureWalletManager) GetTotalStake() int64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	total := int64(0)
	for _, wallet := range wm.wallets {
		total += wallet.StakeAmount
	}
	return total
}

// GetTotalSupply retorna o supply total
func (wm *SecureWalletManager) GetTotalSupply() int64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	total := int64(0)
	for _, wallet := range wm.wallets {
		total += wallet.Balance + wallet.StakeAmount
	}
	return total
}
