package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	WalletIDs    []string  `json:"wallet_ids"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
	IsActive     bool      `json:"is_active"`
}

type WalletAuth struct {
	PublicKey     string `json:"public_key"`
	PrivateKey    string `json:"private_key"` // Criptografada
	WalletPIN     string `json:"wallet_pin"`  // PIN único da wallet
	LastLogin     int64  `json:"last_login"`
	LoginAttempts int    `json:"login_attempts"`
	LockedUntil   int64  `json:"locked_until"`
}

type UserManager struct {
	Users        map[string]*User       `json:"users"`
	WalletAuths  map[string]*WalletAuth `json:"wallet_auths"`
	FilePath     string                 `json:"file_path"`
	ActiveUser   *User                  `json:"-"`
	ActiveWallet *WalletAuth            `json:"-"`
}

func NewUserManager(dataPath string) *UserManager {
	// Em produção (Render), usar diretório temporário
	if os.Getenv("NODE_ENV") == "production" {
		dataPath = "/tmp/ordm-data"
	}
	
	// Criar diretório se não existir
	os.MkdirAll(dataPath, 0755)
	
	um := &UserManager{
		Users:       make(map[string]*User),
		WalletAuths: make(map[string]*WalletAuth),
		FilePath:    filepath.Join(dataPath, "users.json"),
	}
	
	// Carregar dados existentes
	um.LoadUsers()
	
	// Criar usuário padrão se não existir
	if len(um.Users) == 0 {
		um.CreateDefaultUser()
	}
	
	return um
}

func (um *UserManager) CreateDefaultUser() {
	// Garantir que os mapas estão inicializados
	if um.Users == nil {
		um.Users = make(map[string]*User)
	}
	if um.WalletAuths == nil {
		um.WalletAuths = make(map[string]*WalletAuth)
	}

	defaultUser := &User{
		ID:           "default_user",
		Username:     "admin",
		PasswordHash: "admin123", // Em produção, usar hash real
		WalletIDs:    []string{},
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
		IsActive:     true,
	}

	um.Users[defaultUser.ID] = defaultUser
	um.SaveUsers()
}

func (um *UserManager) Login(username, password string) (*User, error) {
	for _, user := range um.Users {
		if user.Username == username && user.PasswordHash == password {
			user.LastLogin = time.Now()
			um.ActiveUser = user
			um.SaveUsers()
			return user, nil
		}
	}
	return nil, fmt.Errorf("credenciais inválidas")
}

func (um *UserManager) GetActiveUser() *User {
	return um.ActiveUser
}

func (um *UserManager) Logout() {
	um.ActiveUser = nil
}

func (um *UserManager) AddWalletToUser(walletID string) error {
	if um.ActiveUser == nil {
		return fmt.Errorf("nenhum usuário ativo")
	}

	// Verificar se wallet já está associada
	for _, id := range um.ActiveUser.WalletIDs {
		if id == walletID {
			return nil // Já existe
		}
	}

	um.ActiveUser.WalletIDs = append(um.ActiveUser.WalletIDs, walletID)
	um.SaveUsers()
	return nil
}

func (um *UserManager) GetUserWallets() []string {
	if um.ActiveUser == nil {
		return []string{}
	}
	return um.ActiveUser.WalletIDs
}

func (um *UserManager) LoadUsers() error {
	if _, err := os.Stat(um.FilePath); os.IsNotExist(err) {
		return nil // Arquivo não existe, usar usuários padrão
	}

	data, err := os.ReadFile(um.FilePath)
	if err != nil {
		return err
	}

	// Estrutura temporária para carregar dados
	var tempData struct {
		Users       map[string]*User       `json:"users"`
		WalletAuths map[string]*WalletAuth `json:"wallet_auths"`
	}

	err = json.Unmarshal(data, &tempData)
	if err != nil {
		// Tentar carregar formato antigo (apenas Users)
		return json.Unmarshal(data, &um.Users)
	}

	// Carregar dados do novo formato
	um.Users = tempData.Users
	um.WalletAuths = tempData.WalletAuths

	return nil
}

func (um *UserManager) SaveUsers() error {
	os.MkdirAll(filepath.Dir(um.FilePath), 0755)

	// Salvar tanto Users quanto WalletAuths
	data, err := json.MarshalIndent(map[string]interface{}{
		"users":        um.Users,
		"wallet_auths": um.WalletAuths,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(um.FilePath, data, 0644)
}

// LoginWallet autentica uma wallet usando public key e PIN único
func (um *UserManager) LoginWallet(publicKey, pin string) (*WalletAuth, error) {
	fmt.Printf("🔍 Tentando login: Public Key: %s, PIN: %s\n", publicKey, pin)
	fmt.Printf("📊 Wallets registradas: %d\n", len(um.WalletAuths))

	walletAuth, exists := um.WalletAuths[publicKey]
	if !exists {
		fmt.Printf("❌ Wallet não encontrada: %s\n", publicKey)
		// Listar wallets disponíveis para debug
		for pk := range um.WalletAuths {
			fmt.Printf("   - %s\n", pk)
		}
		return nil, fmt.Errorf("wallet não encontrada")
	}

	// Verificar se wallet está bloqueada
	if walletAuth.LockedUntil > time.Now().Unix() {
		return nil, fmt.Errorf("wallet bloqueada por %d segundos", walletAuth.LockedUntil-time.Now().Unix())
	}

	// Verificar PIN
	if walletAuth.WalletPIN != pin {
		walletAuth.LoginAttempts++

		// Bloquear após 3 tentativas
		if walletAuth.LoginAttempts >= 3 {
			walletAuth.LockedUntil = time.Now().Unix() + 300 // 5 minutos
		}

		um.SaveUsers()
		return nil, fmt.Errorf("PIN incorreto. Tentativas restantes: %d", 3-walletAuth.LoginAttempts)
	}

	// Login bem-sucedido
	walletAuth.LastLogin = time.Now().Unix()
	walletAuth.LoginAttempts = 0
	um.ActiveWallet = walletAuth
	um.SaveUsers()

	return walletAuth, nil
}

// CreateWalletAuth cria uma nova autenticação de wallet
func (um *UserManager) CreateWalletAuth(publicKey, privateKey string) (*WalletAuth, error) {
	// Gerar PIN único para a wallet (6 dígitos)
	pin := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	walletAuth := &WalletAuth{
		PublicKey:     publicKey,
		PrivateKey:    privateKey, // Em produção, criptografar
		WalletPIN:     pin,
		LastLogin:     time.Now().Unix(),
		LoginAttempts: 0,
		LockedUntil:   0,
	}

	um.WalletAuths[publicKey] = walletAuth
	fmt.Printf("✅ Wallet auth criada: %s com PIN: %s\n", publicKey, pin)
	fmt.Printf("📊 Total de wallets: %d\n", len(um.WalletAuths))

	err := um.SaveUsers()
	if err != nil {
		fmt.Printf("❌ Erro ao salvar: %v\n", err)
		return nil, err
	}

	return walletAuth, nil
}

// GetActiveWallet retorna a wallet ativa
func (um *UserManager) GetActiveWallet() *WalletAuth {
	return um.ActiveWallet
}

// LogoutWallet faz logout da wallet ativa
func (um *UserManager) LogoutWallet() {
	um.ActiveWallet = nil
}
