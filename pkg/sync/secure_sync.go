package sync

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	
	"ordm-main/pkg/blockchain"
	"ordm-main/pkg/crypto"
)

// SyncManager gerencia a sincronização entre minerador offline e testnet online
type SyncManager struct {
	TestnetURL     string
	MinerIdentity  *crypto.MinerIdentity
	SyncInterval   time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
	mu             sync.RWMutex
	lastSyncTime   time.Time
	syncStatus     string
	syncedBlocks   int64
	failedBlocks   int64
	httpClient     *http.Client
}

// SyncRequest representa uma requisição de sincronização
type SyncRequest struct {
	BlockData     []byte `json:"block_data"`
	BlockHash     string `json:"block_hash"`
	MinerID       string `json:"miner_id"`
	Signature     []byte `json:"signature"`
	Timestamp     int64  `json:"timestamp"`
	Difficulty    uint64 `json:"difficulty"`
	Nonce         uint64 `json:"nonce"`
}

// SyncResponse representa a resposta da testnet
type SyncResponse struct {
	Success       bool   `json:"success"`
	BlockNumber   int64  `json:"block_number,omitempty"`
	Message       string `json:"message"`
	Error         string `json:"error,omitempty"`
	Accepted      bool   `json:"accepted"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// NewSyncManager cria um novo gerenciador de sincronização
func NewSyncManager(testnetURL string, minerIdentity *crypto.MinerIdentity) *SyncManager {
	return &SyncManager{
		TestnetURL:    testnetURL,
		MinerIdentity: minerIdentity,
		SyncInterval:  30 * time.Second, // Sincronizar a cada 30 segundos
		RetryAttempts: 3,
		RetryDelay:    5 * time.Second,
		syncStatus:    "disconnected",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartSync inicia a sincronização automática
func (sm *SyncManager) StartSync() {
	go func() {
		ticker := time.NewTicker(sm.SyncInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				sm.performSync()
			}
		}
	}()
	
	fmt.Printf("🔄 Sincronização automática iniciada (intervalo: %v)\n", sm.SyncInterval)
}

// performSync executa uma sincronização
func (sm *SyncManager) performSync() {
	sm.mu.Lock()
	sm.syncStatus = "syncing"
	sm.mu.Unlock()
	
	// Aqui seria implementada a lógica para buscar blocos não sincronizados
	// Por enquanto, apenas simular
	
	fmt.Printf("📡 Sincronizando com testnet: %s\n", sm.TestnetURL)
	
	sm.mu.Lock()
	sm.lastSyncTime = time.Now()
	sm.syncStatus = "connected"
	sm.mu.Unlock()
}

// SyncBlock sincroniza um bloco específico com a testnet
func (sm *SyncManager) SyncBlock(block *blockchain.RealBlock) (*SyncResponse, error) {
	// Assinar o bloco
	blockData, err := block.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar bloco: %v", err)
	}
	
	signature, err := sm.MinerIdentity.SignBlock(blockData)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar bloco: %v", err)
	}
	
	// Criar requisição de sincronização
	syncReq := &SyncRequest{
		BlockData:  blockData,
		BlockHash:  block.GetBlockHashString(),
		MinerID:    sm.MinerIdentity.MinerID,
		Signature:  signature,
		Timestamp:  time.Now().Unix(),
		Difficulty: block.Header.Difficulty,
		Nonce:      block.Header.Nonce,
	}
	
	// Enviar para testnet
	return sm.sendSyncRequest(syncReq)
}

// sendSyncRequest envia requisição de sincronização para a testnet
func (sm *SyncManager) sendSyncRequest(syncReq *SyncRequest) (*SyncResponse, error) {
	// Serializar requisição
	reqData, err := json.Marshal(syncReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}
	
	// Criar requisição HTTP
	req, err := http.NewRequest("POST", sm.TestnetURL+"/api/sync/block", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Miner-ID", sm.MinerIdentity.MinerID)
	
	// Enviar com retry
	var lastErr error
	for attempt := 0; attempt < sm.RetryAttempts; attempt++ {
		resp, err := sm.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("tentativa %d falhou: %v", attempt+1, err)
			if attempt < sm.RetryAttempts-1 {
				time.Sleep(sm.RetryDelay)
				continue
			}
			return nil, lastErr
		}
		defer resp.Body.Close()
		
		// Ler resposta
		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("erro ao ler resposta: %v", err)
			continue
		}
		
		// Parsear resposta
		var syncResp SyncResponse
		if err := json.Unmarshal(respData, &syncResp); err != nil {
			lastErr = fmt.Errorf("erro ao parsear resposta: %v", err)
			continue
		}
		
		// Atualizar estatísticas
		sm.mu.Lock()
		if syncResp.Accepted {
			sm.syncedBlocks++
		} else {
			sm.failedBlocks++
		}
		sm.mu.Unlock()
		
		return &syncResp, nil
	}
	
	return nil, lastErr
}

// GetSyncStatus retorna o status da sincronização
func (sm *SyncManager) GetSyncStatus() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	return map[string]interface{}{
		"status":         sm.syncStatus,
		"last_sync":      sm.lastSyncTime,
		"synced_blocks":  sm.syncedBlocks,
		"failed_blocks":  sm.failedBlocks,
		"testnet_url":    sm.TestnetURL,
		"miner_id":       sm.MinerIdentity.MinerID,
	}
}

// ValidateBlockSignature valida a assinatura de um bloco
func ValidateBlockSignature(blockData []byte, signature []byte, publicKey ed25519.PublicKey) bool {
	return ed25519.Verify(publicKey, blockData, signature)
}

// GenerateSyncID gera um ID único para sincronização
func GenerateSyncID() string {
	data := make([]byte, 16)
	rand.Read(data)
	return hex.EncodeToString(data)
}
