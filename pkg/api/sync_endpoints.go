package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"ordm-main/pkg/blockchain"
	"ordm-main/pkg/validation"
)

// SyncEndpoints registra os endpoints de sincronização
type SyncEndpoints struct {
	PoSValidator *validation.PoSValidator
}

// NewSyncEndpoints cria novos endpoints de sincronização
func NewSyncEndpoints(posValidator *validation.PoSValidator) *SyncEndpoints {
	return &SyncEndpoints{
		PoSValidator: posValidator,
	}
}

// RegisterSyncRoutes registra as rotas de sincronização
func (se *SyncEndpoints) RegisterSyncRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/sync/block", se.handleSyncBlock)
	mux.HandleFunc("/api/sync/status", se.handleSyncStatus)
	mux.HandleFunc("/api/validators", se.handleValidators)
	mux.HandleFunc("/api/validators/stats", se.handleValidatorStats)
}

// handleSyncBlock processa sincronização de blocos
func (se *SyncEndpoints) handleSyncBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parsear requisição de sincronização
	var syncReq struct {
		BlockData  []byte `json:"block_data"`
		BlockHash  string `json:"block_hash"`
		MinerID    string `json:"miner_id"`
		Signature  []byte `json:"signature"`
		Timestamp  int64  `json:"timestamp"`
		Difficulty uint64 `json:"difficulty"`
		Nonce      uint64 `json:"nonce"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&syncReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Deserializar bloco
	var block blockchain.RealBlock
	if err := block.FromJSON(syncReq.BlockData); err != nil {
		se.sendSyncResponse(w, false, "erro ao deserializar bloco: "+err.Error(), nil)
		return
	}
	
	// Validar bloco usando PoS
	validation, err := se.PoSValidator.ValidateBlock(&block, syncReq.Signature)
	if err != nil {
		se.sendSyncResponse(w, false, "erro na validação: "+err.Error(), nil)
		return
	}
	
	if !validation.IsValid {
		se.sendSyncResponse(w, false, validation.Reason, map[string]interface{}{
			"rejection_reason": validation.Reason,
		})
		return
	}
	
	// Bloco válido - adicionar à blockchain
	// Aqui seria implementada a lógica para adicionar o bloco à blockchain online
	
	se.sendSyncResponse(w, true, "bloco aceito com sucesso", map[string]interface{}{
		"block_number": block.Header.Number,
		"block_hash":   block.GetBlockHashString(),
		"validator_id": validation.ValidatorID,
		"stake_used":   validation.StakeUsed,
	})
	
	fmt.Printf("✅ Bloco sincronizado: #%d %s\n", block.Header.Number, block.GetBlockHashString()[:16])
}

// handleSyncStatus retorna o status da sincronização
func (se *SyncEndpoints) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	stats := se.PoSValidator.GetValidatorStats()
	
	response := map[string]interface{}{
		"status":              "online",
		"timestamp":           time.Now().Unix(),
		"validator_stats":     stats,
		"sync_endpoint":       "/api/sync/block",
		"min_stake_required":  se.PoSValidator.MinStake,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidators retorna lista de validadores
func (se *SyncEndpoints) handleValidators(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	validators := se.PoSValidator.GetTopValidators(100)
	
	response := map[string]interface{}{
		"validators": validators,
		"total":      len(validators),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidatorStats retorna estatísticas dos validadores
func (se *SyncEndpoints) handleValidatorStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	stats := se.PoSValidator.GetValidatorStats()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// sendSyncResponse envia resposta de sincronização
func (se *SyncEndpoints) sendSyncResponse(w http.ResponseWriter, success bool, message string, data map[string]interface{}) {
	response := map[string]interface{}{
		"success":   success,
		"message":   message,
		"timestamp": time.Now().Unix(),
		"accepted":  success,
	}
	
	if data != nil {
		for key, value := range data {
			response[key] = value
		}
	}
	
	if !success && data != nil {
		if reason, ok := data["rejection_reason"]; ok {
			response["rejection_reason"] = reason
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
