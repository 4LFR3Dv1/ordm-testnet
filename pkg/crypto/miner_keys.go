package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MinerIdentity representa a identidade criptográfica de um minerador
type MinerIdentity struct {
	PublicKey     ed25519.PublicKey `json:"public_key"`
	PrivateKey    ed25519.PrivateKey `json:"private_key"`
	MinerID       string            `json:"miner_id"`
	Reputation    int               `json:"reputation"`
	TotalBlocks   int64             `json:"total_blocks"`
	ValidBlocks   int64             `json:"valid_blocks"`
	InvalidBlocks int64             `json:"invalid_blocks"`
	StakeAmount   int64             `json:"stake_amount"`
	IsValidator   bool              `json:"is_validator"`
	CreatedAt     time.Time         `json:"created_at"`
	LastActivity  time.Time         `json:"last_activity"`
	mu            sync.RWMutex      `json:"-"`
}

// MinerKeyManager gerencia as chaves dos mineradores
type MinerKeyManager struct {
	Identities map[string]*MinerIdentity `json:"identities"`
	DataPath   string                    `json:"-"`
	mu         sync.RWMutex              `json:"-"`
}

// NewMinerKeyManager cria um novo gerenciador de chaves
func NewMinerKeyManager(dataPath string) *MinerKeyManager {
	manager := &MinerKeyManager{
		Identities: make(map[string]*MinerIdentity),
		DataPath:   dataPath,
	}
	
	// Criar diretório se não existir
	os.MkdirAll(dataPath, 0755)
	
	// Carregar identidades existentes
	manager.LoadIdentities()
	
	return manager
}

// GenerateMinerIdentity gera uma nova identidade de minerador
func (mkm *MinerKeyManager) GenerateMinerIdentity(minerName string) (*MinerIdentity, error) {
	mkm.mu.Lock()
	defer mkm.mu.Unlock()
	
	// Gerar par de chaves Ed25519
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar chaves: %v", err)
	}
	
	// Gerar ID único do minerador
	minerID := generateMinerID(publicKey)
	
	// Criar identidade
	identity := &MinerIdentity{
		PublicKey:     publicKey,
		PrivateKey:    privateKey,
		MinerID:       minerID,
		Reputation:    100, // Reputação inicial
		TotalBlocks:   0,
		ValidBlocks:   0,
		InvalidBlocks: 0,
		StakeAmount:   0,
		IsValidator:   false,
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
	}
	
	// Salvar identidade
	mkm.Identities[minerID] = identity
	
	// Persistir no disco
	if err := mkm.SaveIdentities(); err != nil {
		return nil, fmt.Errorf("erro ao salvar identidade: %v", err)
	}
	
	fmt.Printf("🔑 Nova identidade de minerador criada: %s\n", minerID)
	return identity, nil
}

// SignBlock assina um bloco com a chave privada do minerador
func (mi *MinerIdentity) SignBlock(blockData []byte) ([]byte, error) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	
	// Calcular hash do bloco
	hash := sha256.Sum256(blockData)
	
	// Assinar o hash
	signature := ed25519.Sign(mi.PrivateKey, hash[:])
	
	// Atualizar última atividade
	mi.LastActivity = time.Now()
	
	return signature, nil
}

// VerifySignature verifica a assinatura de um bloco
func (mi *MinerIdentity) VerifySignature(blockData []byte, signature []byte) bool {
	// Calcular hash do bloco
	hash := sha256.Sum256(blockData)
	
	// Verificar assinatura
	return ed25519.Verify(mi.PublicKey, hash[:], signature)
}

// GetPublicKeyString retorna a chave pública como string
func (mi *MinerIdentity) GetPublicKeyString() string {
	return hex.EncodeToString(mi.PublicKey)
}

// GetPrivateKeyString retorna a chave privada como string (para backup)
func (mi *MinerIdentity) GetPrivateKeyString() string {
	return hex.EncodeToString(mi.PrivateKey)
}

// UpdateReputation atualiza a reputação do minerador
func (mi *MinerIdentity) UpdateReputation(validBlock bool) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	
	mi.TotalBlocks++
	mi.LastActivity = time.Now()
	
	if validBlock {
		mi.ValidBlocks++
		mi.Reputation = min(1000, mi.Reputation+10) // Máximo 1000
	} else {
		mi.InvalidBlocks++
		mi.Reputation = max(0, mi.Reputation-50) // Mínimo 0
	}
}

// UpdateStake atualiza o stake do minerador
func (mi *MinerIdentity) UpdateStake(amount int64) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	
	mi.StakeAmount = amount
	mi.IsValidator = amount >= 1000 // Mínimo 1000 tokens para ser validator
	mi.LastActivity = time.Now()
}

// LoadIdentities carrega identidades do disco
func (mkm *MinerKeyManager) LoadIdentities() error {
	filePath := filepath.Join(mkm.DataPath, "miner_identities.json")
	
	// Se o arquivo não existir, retornar sem erro
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}
	
	// Estrutura temporária para deserialização
	type tempIdentity struct {
		PublicKey     string    `json:"public_key"`
		PrivateKey    string    `json:"private_key"`
		MinerID       string    `json:"miner_id"`
		Reputation    int       `json:"reputation"`
		TotalBlocks   int64     `json:"total_blocks"`
		ValidBlocks   int64     `json:"valid_blocks"`
		InvalidBlocks int64     `json:"invalid_blocks"`
		StakeAmount   int64     `json:"stake_amount"`
		IsValidator   bool      `json:"is_validator"`
		CreatedAt     time.Time `json:"created_at"`
		LastActivity  time.Time `json:"last_activity"`
	}
	
	var tempIdentities map[string]tempIdentity
	if err := json.Unmarshal(data, &tempIdentities); err != nil {
		return fmt.Errorf("erro ao deserializar: %v", err)
	}
	
	// Converter para MinerIdentity
	for id, temp := range tempIdentities {
		publicKey, err := hex.DecodeString(temp.PublicKey)
		if err != nil {
			continue
		}
		
		privateKey, err := hex.DecodeString(temp.PrivateKey)
		if err != nil {
			continue
		}
		
		identity := &MinerIdentity{
			PublicKey:     publicKey,
			PrivateKey:    privateKey,
			MinerID:       temp.MinerID,
			Reputation:    temp.Reputation,
			TotalBlocks:   temp.TotalBlocks,
			ValidBlocks:   temp.ValidBlocks,
			InvalidBlocks: temp.InvalidBlocks,
			StakeAmount:   temp.StakeAmount,
			IsValidator:   temp.IsValidator,
			CreatedAt:     temp.CreatedAt,
			LastActivity:  temp.LastActivity,
		}
		
		mkm.Identities[id] = identity
	}
	
	fmt.Printf("📁 %d identidades de mineradores carregadas\n", len(mkm.Identities))
	return nil
}

// SaveIdentities salva identidades no disco
func (mkm *MinerKeyManager) SaveIdentities() error {
	// Estrutura temporária para serialização
	tempIdentities := make(map[string]map[string]interface{})
	
	for id, identity := range mkm.Identities {
		tempIdentities[id] = map[string]interface{}{
			"public_key":     identity.GetPublicKeyString(),
			"private_key":    identity.GetPrivateKeyString(),
			"miner_id":       identity.MinerID,
			"reputation":     identity.Reputation,
			"total_blocks":   identity.TotalBlocks,
			"valid_blocks":   identity.ValidBlocks,
			"invalid_blocks": identity.InvalidBlocks,
			"stake_amount":   identity.StakeAmount,
			"is_validator":   identity.IsValidator,
			"created_at":     identity.CreatedAt,
			"last_activity":  identity.LastActivity,
		}
	}
	
	data, err := json.MarshalIndent(tempIdentities, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar: %v", err)
	}
	
	filePath := filepath.Join(mkm.DataPath, "miner_identities.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}
	
	return nil
}

// GetIdentity retorna uma identidade por ID
func (mkm *MinerKeyManager) GetIdentity(minerID string) (*MinerIdentity, bool) {
	mkm.mu.RLock()
	defer mkm.mu.RUnlock()
	
	identity, exists := mkm.Identities[minerID]
	return identity, exists
}

// ListIdentities retorna todas as identidades
func (mkm *MinerKeyManager) ListIdentities() []*MinerIdentity {
	mkm.mu.RLock()
	defer mkm.mu.RUnlock()
	
	identities := make([]*MinerIdentity, 0, len(mkm.Identities))
	for _, identity := range mkm.Identities {
		identities = append(identities, identity)
	}
	
	return identities
}

// generateMinerID gera um ID único para o minerador
func generateMinerID(publicKey ed25519.PublicKey) string {
	hash := sha256.Sum256(publicKey)
	return fmt.Sprintf("miner_%s", hex.EncodeToString(hash[:8]))
}

// Funções auxiliares
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
