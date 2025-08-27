package wallet

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/tyler-smith/go-bip39"
)

// BIP39Wallet implementa wallet BIP-39
type BIP39Wallet struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Mnemonic   string              `json:"mnemonic"`
	Passphrase string              `json:"passphrase,omitempty"`
	Seed       []byte              `json:"seed"`
	MasterKey  *btcec.PrivateKey   `json:"-"`
	Accounts   map[string]*Account `json:"accounts"`
	CreatedAt  int64               `json:"created_at"`
	LastUsed   int64               `json:"last_used"`
	Version    string              `json:"version"`
	Encrypted  bool                `json:"encrypted"`
	FilePath   string              `json:"file_path"`
}

// Account representa uma conta derivada
type Account struct {
	Index      uint32 `json:"index"`
	Path       string `json:"path"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"`
	Balance    int64  `json:"balance"`
	Nonce      uint64 `json:"nonce"`
	CreatedAt  int64  `json:"created_at"`
	LastUsed   int64  `json:"last_used"`
	Label      string `json:"label"`
	IsActive   bool   `json:"is_active"`
}

// WalletManager gerencia múltiplas wallets
type WalletManager struct {
	Wallets     map[string]*BIP39Wallet `json:"wallets"`
	DefaultID   string                  `json:"default_id"`
	WalletsPath string                  `json:"wallets_path"`
	Version     string                  `json:"version"`
}

// NewBIP39Wallet cria uma nova wallet BIP-39
func NewBIP39Wallet(name, passphrase string) (*BIP39Wallet, error) {
	// Gerar mnemonic
	entropy, err := bip39.NewEntropy(256) // 24 palavras
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar entropy: %v", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar mnemonic: %v", err)
	}

	// Gerar seed
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Gerar chave mestra
	masterKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar chave mestra: %v", err)
	}

	// Criar wallet
	wallet := &BIP39Wallet{
		ID:         generateWalletID(),
		Name:       name,
		Mnemonic:   mnemonic,
		Passphrase: passphrase,
		Seed:       seed,
		MasterKey:  masterKey,
		Accounts:   make(map[string]*Account),
		CreatedAt:  time.Now().Unix(),
		LastUsed:   time.Now().Unix(),
		Version:    "1.0.0",
		Encrypted:  false,
	}

	// Criar conta padrão
	if err := wallet.CreateAccount("Default", 0); err != nil {
		return nil, fmt.Errorf("erro ao criar conta padrão: %v", err)
	}

	return wallet, nil
}

// NewWalletManager cria um novo gerenciador de wallets
func NewWalletManager(walletsPath string) *WalletManager {
	return &WalletManager{
		Wallets:     make(map[string]*BIP39Wallet),
		WalletsPath: walletsPath,
		Version:     "1.0.0",
	}
}

// CreateWallet cria uma nova wallet
func (wm *WalletManager) CreateWallet(name, passphrase string) (*BIP39Wallet, error) {
	wallet, err := NewBIP39Wallet(name, passphrase)
	if err != nil {
		return nil, err
	}

	wm.Wallets[wallet.ID] = wallet

	// Definir como padrão se for a primeira
	if wm.DefaultID == "" {
		wm.DefaultID = wallet.ID
	}

	// Salvar wallet
	if err := wm.SaveWallet(wallet); err != nil {
		return nil, fmt.Errorf("erro ao salvar wallet: %v", err)
	}

	return wallet, nil
}

// ImportWallet importa wallet de mnemonic
func (wm *WalletManager) ImportWallet(name, mnemonic, passphrase string) (*BIP39Wallet, error) {
	// Validar mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("mnemonic inválido")
	}

	// Gerar seed
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Gerar chave mestra
	masterKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar chave mestra: %v", err)
	}

	// Criar wallet
	wallet := &BIP39Wallet{
		ID:         generateWalletID(),
		Name:       name,
		Mnemonic:   mnemonic,
		Passphrase: passphrase,
		Seed:       seed,
		MasterKey:  masterKey,
		Accounts:   make(map[string]*Account),
		CreatedAt:  time.Now().Unix(),
		LastUsed:   time.Now().Unix(),
		Version:    "1.0.0",
		Encrypted:  false,
	}

	wm.Wallets[wallet.ID] = wallet

	// Salvar wallet
	if err := wm.SaveWallet(wallet); err != nil {
		return nil, fmt.Errorf("erro ao salvar wallet: %v", err)
	}

	return wallet, nil
}

// LoadWallets carrega wallets do disco
func (wm *WalletManager) LoadWallets() error {
	// Criar diretório se não existir
	if err := os.MkdirAll(wm.WalletsPath, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	// Listar arquivos de wallet
	files, err := os.ReadDir(wm.WalletsPath)
	if err != nil {
		return fmt.Errorf("erro ao ler diretório: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			walletPath := filepath.Join(wm.WalletsPath, file.Name())

			data, err := os.ReadFile(walletPath)
			if err != nil {
				continue // Ignorar arquivos corrompidos
			}

			var wallet BIP39Wallet
			if err := json.Unmarshal(data, &wallet); err != nil {
				continue
			}

			// Reconstruir chave mestra
			seed := bip39.NewSeed(wallet.Mnemonic, wallet.Passphrase)
			masterKey, err := btcec.NewPrivateKey()
			if err != nil {
				continue
			}
			wallet.MasterKey = masterKey
			wallet.Seed = seed
			wallet.FilePath = walletPath

			wm.Wallets[wallet.ID] = &wallet
		}
	}

	return nil
}

// SaveWallet salva wallet no disco
func (wm *WalletManager) SaveWallet(wallet *BIP39Wallet) error {
	// Criar diretório se não existir
	if err := os.MkdirAll(wm.WalletsPath, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	// Definir caminho do arquivo
	if wallet.FilePath == "" {
		wallet.FilePath = filepath.Join(wm.WalletsPath, fmt.Sprintf("%s.json", wallet.ID))
	}

	// Serializar wallet (sem chaves privadas)
	walletCopy := *wallet
	walletCopy.MasterKey = nil
	walletCopy.Seed = nil

	// Remover chaves privadas das contas
	for _, account := range walletCopy.Accounts {
		account.PrivateKey = ""
	}

	data, err := json.MarshalIndent(walletCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar wallet: %v", err)
	}

	// Salvar arquivo
	if err := os.WriteFile(wallet.FilePath, data, 0600); err != nil {
		return fmt.Errorf("erro ao salvar wallet: %v", err)
	}

	return nil
}

// GetWallet retorna wallet por ID
func (wm *WalletManager) GetWallet(id string) (*BIP39Wallet, bool) {
	wallet, exists := wm.Wallets[id]
	return wallet, exists
}

// GetDefaultWallet retorna wallet padrão
func (wm *WalletManager) GetDefaultWallet() (*BIP39Wallet, bool) {
	if wm.DefaultID == "" {
		return nil, false
	}
	return wm.GetWallet(wm.DefaultID)
}

// SetDefaultWallet define wallet padrão
func (wm *WalletManager) SetDefaultWallet(id string) error {
	if _, exists := wm.Wallets[id]; !exists {
		return fmt.Errorf("wallet não encontrada: %s", id)
	}

	wm.DefaultID = id
	return nil
}

// CreateAccount cria uma nova conta na wallet
func (w *BIP39Wallet) CreateAccount(label string, index uint32) error {
	// Derivar chave privada
	privateKey, err := w.derivePrivateKey(index)
	if err != nil {
		return fmt.Errorf("erro ao derivar chave privada: %v", err)
	}

	// Obter chave pública
	publicKey := privateKey.PubKey()

	// Gerar endereço
	address := w.generateAddress(publicKey)

	// Criar conta
	account := &Account{
		Index:      index,
		Path:       fmt.Sprintf("m/44'/0'/0'/%d", index),
		Address:    address,
		PublicKey:  hex.EncodeToString(publicKey.SerializeCompressed()),
		PrivateKey: hex.EncodeToString(privateKey.Serialize()),
		Balance:    0,
		Nonce:      0,
		CreatedAt:  time.Now().Unix(),
		LastUsed:   time.Now().Unix(),
		Label:      label,
		IsActive:   true,
	}

	w.Accounts[address] = account
	w.LastUsed = time.Now().Unix()

	return nil
}

// GetAccount retorna conta por endereço
func (w *BIP39Wallet) GetAccount(address string) (*Account, bool) {
	account, exists := w.Accounts[address]
	return account, exists
}

// GetAccounts retorna todas as contas
func (w *BIP39Wallet) GetAccounts() []*Account {
	accounts := make([]*Account, 0, len(w.Accounts))
	for _, account := range w.Accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

// GetActiveAccounts retorna contas ativas
func (w *BIP39Wallet) GetActiveAccounts() []*Account {
	var activeAccounts []*Account
	for _, account := range w.Accounts {
		if account.IsActive {
			activeAccounts = append(activeAccounts, account)
		}
	}
	return activeAccounts
}

// UpdateAccountBalance atualiza saldo da conta
func (w *BIP39Wallet) UpdateAccountBalance(address string, balance int64) error {
	account, exists := w.Accounts[address]
	if !exists {
		return fmt.Errorf("conta não encontrada: %s", address)
	}

	account.Balance = balance
	account.LastUsed = time.Now().Unix()
	w.LastUsed = time.Now().Unix()

	return nil
}

// UpdateAccountNonce atualiza nonce da conta
func (w *BIP39Wallet) UpdateAccountNonce(address string, nonce uint64) error {
	account, exists := w.Accounts[address]
	if !exists {
		return fmt.Errorf("conta não encontrada: %s", address)
	}

	account.Nonce = nonce
	account.LastUsed = time.Now().Unix()
	w.LastUsed = time.Now().Unix()

	return nil
}

// SignTransaction assina transação
func (w *BIP39Wallet) SignTransaction(address string, txHash []byte) ([]byte, error) {
	_, exists := w.Accounts[address]
	if !exists {
		return nil, fmt.Errorf("conta não encontrada: %s", address)
	}

	// Assinar transação (simplificado para compatibilidade)
	hash := sha256.Sum256(txHash)
	signature := fmt.Sprintf("sig_%s_%d", hex.EncodeToString(hash[:8]), time.Now().Unix())

	return []byte(signature), nil
}

// VerifySignature verifica assinatura
func (w *BIP39Wallet) VerifySignature(address string, txHash []byte, signature []byte) (bool, error) {
	_, exists := w.Accounts[address]
	if !exists {
		return false, fmt.Errorf("conta não encontrada: %s", address)
	}

	// Verificação simplificada (em produção usar verificação real)
	hash := sha256.Sum256(txHash)
	expectedSignature := fmt.Sprintf("sig_%s_", hex.EncodeToString(hash[:8]))

	return strings.HasPrefix(string(signature), expectedSignature), nil
}

// derivePrivateKey deriva chave privada do índice
func (w *BIP39Wallet) derivePrivateKey(index uint32) (*btcec.PrivateKey, error) {
	// Implementação simplificada - em produção usar BIP-32/44
	// Aqui apenas geramos uma nova chave baseada no seed + índice

	// Hash do seed + índice
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)

	combined := append(w.Seed, indexBytes...)
	hash := sha256.Sum256(combined)

	privateKey, _ := btcec.PrivKeyFromBytes(hash[:])
	return privateKey, nil
}

// generateAddress gera endereço da chave pública
func (w *BIP39Wallet) generateAddress(publicKey *btcec.PublicKey) string {
	// Hash da chave pública
	pubKeyBytes := publicKey.SerializeCompressed()
	hash := sha256.Sum256(pubKeyBytes)

	// Endereço simplificado (em produção usar RIPEMD160 + checksum)
	return hex.EncodeToString(hash[:20])
}

// generateWalletID gera ID único para wallet
func generateWalletID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateMnemonic valida mnemonic
func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// GetMnemonicStrength retorna força do mnemonic
func GetMnemonicStrength(mnemonic string) string {
	words := strings.Fields(mnemonic)
	wordCount := len(words)

	switch wordCount {
	case 12:
		return "128 bits (12 palavras)"
	case 15:
		return "160 bits (15 palavras)"
	case 18:
		return "192 bits (18 palavras)"
	case 21:
		return "224 bits (21 palavras)"
	case 24:
		return "256 bits (24 palavras)"
	default:
		return "Desconhecido"
	}
}

// GetWalletInfo retorna informações da wallet
func (w *BIP39Wallet) GetWalletInfo() map[string]interface{} {
	accounts := w.GetAccounts()
	activeAccounts := w.GetActiveAccounts()

	var totalBalance int64
	for _, account := range accounts {
		totalBalance += account.Balance
	}

	return map[string]interface{}{
		"id":                w.ID,
		"name":              w.Name,
		"version":           w.Version,
		"created_at":        w.CreatedAt,
		"last_used":         w.LastUsed,
		"encrypted":         w.Encrypted,
		"total_accounts":    len(accounts),
		"active_accounts":   len(activeAccounts),
		"total_balance":     totalBalance,
		"mnemonic_words":    len(strings.Fields(w.Mnemonic)),
		"mnemonic_strength": GetMnemonicStrength(w.Mnemonic),
	}
}
