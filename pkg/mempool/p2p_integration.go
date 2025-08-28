package mempool

import (
	"encoding/json"
	"fmt"

	"ordm-main/pkg/p2p"
)

// P2PMempoolIntegration integra mempool com rede P2P
type P2PMempoolIntegration struct {
	mempool *DistributedMempool
	network *p2p.P2PNetwork
	logger  func(string, ...interface{})
}

// NewP2PMempoolIntegration cria nova integração P2P-mempool
func NewP2PMempoolIntegration(mempool *DistributedMempool, network *p2p.P2PNetwork, logger func(string, ...interface{})) *P2PMempoolIntegration {
	pi := &P2PMempoolIntegration{
		mempool: mempool,
		network: network,
		logger:  logger,
	}

	// Configurar handlers P2P
	pi.setupP2PHandlers()

	return pi
}

// setupP2PHandlers configura handlers para mensagens P2P
func (pi *P2PMempoolIntegration) setupP2PHandlers() {
	// Handler para novas transações recebidas
	pi.network.RegisterHandler("new_transaction", func(msg p2p.Message) error {
		pi.logger("📥 Transação recebida via P2P de %s", msg.From)

		// Deserializar mensagem para Transaction
		var txMsg p2p.TransactionMessage
		if msgData, ok := msg.Data.(map[string]interface{}); ok {
			// Converter map para JSON e depois para TransactionMessage
			jsonData, err := json.Marshal(msgData)
			if err != nil {
				return fmt.Errorf("erro ao serializar dados da mensagem: %v", err)
			}

			if err := json.Unmarshal(jsonData, &txMsg); err != nil {
				return fmt.Errorf("erro ao deserializar TransactionMessage: %v", err)
			}

			// Converter para Transaction do mempool
			tx := &Transaction{
				ID:        txMsg.TxHash,
				From:      txMsg.From,
				To:        txMsg.To,
				Amount:    txMsg.Amount,
				Fee:       txMsg.Fee,
				Timestamp: txMsg.Timestamp,
				Status:    "pending",
			}

			// Adicionar ao mempool
			if err := pi.mempool.AddTransaction(tx); err != nil {
				pi.logger("❌ Erro ao adicionar transação ao mempool: %v", err)
				return err
			}

			pi.logger("✅ Transação %s processada e adicionada ao mempool", tx.ID)
		} else {
			pi.logger("⚠️ Formato de dados inesperado para transação")
		}

		return nil
	})

	// Handler para solicitação de transações pendentes
	pi.network.RegisterHandler("request_pending_txs", func(msg p2p.Message) error {
		pi.logger("📋 Solicitação de transações pendentes recebida de %s", msg.From)

		// Obter transações pendentes
		pendingTxs := pi.mempool.GetPendingTransactions(50) // Limite de 50 transações

		// Converter para formato P2P
		var txMessages []p2p.TransactionMessage
		for _, tx := range pendingTxs {
			txMsg := p2p.TransactionMessage{
				TxHash:    tx.ID,
				From:      tx.From,
				To:        tx.To,
				Amount:    tx.Amount,
				Fee:       tx.Fee,
				Timestamp: tx.Timestamp,
			}
			txMessages = append(txMessages, txMsg)
		}

		// TODO: Implementar resposta direta ao peer
		pi.logger("📤 Enviando %d transações pendentes", len(txMessages))

		return nil
	})
}

// BroadcastTransaction faz broadcast de uma transação para a rede P2P
func (pi *P2PMempoolIntegration) BroadcastTransaction(tx *Transaction) error {
	if pi.network == nil {
		return fmt.Errorf("rede P2P não está disponível")
	}

	// Converter para formato P2P
	txMsg := p2p.TransactionMessage{
		TxHash:    tx.ID,
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Timestamp: tx.Timestamp,
	}

	// Fazer broadcast
	if err := pi.network.BroadcastTransaction(txMsg); err != nil {
		return fmt.Errorf("erro ao fazer broadcast da transação: %v", err)
	}

	pi.logger("📡 Broadcast da transação %s enviado", tx.ID)
	return nil
}

// RequestPendingTransactions solicita transações pendentes de peers
func (pi *P2PMempoolIntegration) RequestPendingTransactions() error {
	if pi.network == nil {
		return fmt.Errorf("rede P2P não está disponível")
	}

	// TODO: Implementar envio para peers específicos
	pi.logger("📤 Solicitando transações pendentes de peers")

	return nil
}

// SyncMempool sincroniza mempool com peers
func (pi *P2PMempoolIntegration) SyncMempool() error {
	pi.logger("🔄 Iniciando sincronização do mempool com peers")

	// Solicitar transações pendentes de peers
	if err := pi.RequestPendingTransactions(); err != nil {
		return fmt.Errorf("erro ao solicitar transações: %v", err)
	}

	// Broadcast de transações locais pendentes
	pendingTxs := pi.mempool.GetPendingTransactions(0) // Todas as pendentes
	for _, tx := range pendingTxs {
		if err := pi.BroadcastTransaction(tx); err != nil {
			pi.logger("⚠️ Erro ao fazer broadcast da transação %s: %v", tx.ID, err)
		}
	}

	pi.logger("✅ Sincronização do mempool concluída")
	return nil
}

// GetMempoolStats retorna estatísticas do mempool
func (pi *P2PMempoolIntegration) GetMempoolStats() map[string]interface{} {
	stats := pi.mempool.GetMempoolStats()

	// Adicionar informações P2P
	if pi.network != nil {
		networkInfo := pi.network.GetNetworkInfo()
		stats["p2p_peers"] = networkInfo["peer_count"]
		stats["p2p_connected"] = pi.network.GetPeerCount() > 0
	} else {
		stats["p2p_peers"] = 0
		stats["p2p_connected"] = false
	}

	return stats
}
