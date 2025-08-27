package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

// Message representa uma mensagem P2P
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	From      string      `json:"from"`
	Timestamp int64       `json:"timestamp"`
	Signature string      `json:"signature,omitempty"`
}

// BlockMessage representa uma mensagem de bloco
type BlockMessage struct {
	BlockHash   string `json:"block_hash"`
	BlockNumber int64  `json:"block_number"`
	Miner       string `json:"miner"`
	Timestamp   int64  `json:"timestamp"`
	Data        []byte `json:"data"`
}

// TransactionMessage representa uma mensagem de transação
type TransactionMessage struct {
	TxHash    string `json:"tx_hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

// PeerInfo representa informações de um peer
type PeerInfo struct {
	ID        string    `json:"id"`
	Addresses []string  `json:"addresses"`
	Latency   int64     `json:"latency_ms"`
	LastSeen  time.Time `json:"last_seen"`
	Version   string    `json:"version"`
	Services  []string  `json:"services"`
}

// P2PNetwork implementa a rede peer-to-peer
type P2PNetwork struct {
	host            host.Host
	pubsub          *pubsub.PubSub
	ctx             context.Context
	cancel          context.CancelFunc
	peers           map[peer.ID]*PeerInfo
	peersMutex      sync.RWMutex
	messageHandlers map[string]func(Message) error
	topics          map[string]*pubsub.Topic
	subscriptions   map[string]*pubsub.Subscription
	logger          func(string, ...interface{})
	port            int
}

// NewP2PNetwork cria uma nova rede P2P
func NewP2PNetwork(port int, logger func(string, ...interface{})) (*P2PNetwork, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Gerar chave privada
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("erro ao gerar chave: %v", err)
	}

	// Criar multiaddr
	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("erro ao criar multiaddr: %v", err)
	}

	// Criar host libp2p
	host, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(priv),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("erro ao criar host: %v", err)
	}

	// Criar pubsub
	pubsub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("erro ao criar pubsub: %v", err)
	}

	network := &P2PNetwork{
		host:            host,
		pubsub:          pubsub,
		ctx:             ctx,
		cancel:          cancel,
		peers:           make(map[peer.ID]*PeerInfo),
		messageHandlers: make(map[string]func(Message) error),
		topics:          make(map[string]*pubsub.Topic),
		subscriptions:   make(map[string]*pubsub.Subscription),
		logger:          logger,
		port:            port,
	}

	// Configurar handlers de rede
	host.SetStreamHandler("/blockchain/1.0.0", network.handleStream)
	host.Network().Notify(&networkNotifee{network: network})

	// Registrar handlers padrão
	network.registerDefaultHandlers()

	return network, nil
}

// Start inicia a rede P2P
func (n *P2PNetwork) Start() error {
	n.logger("🚀 Rede P2P iniciada na porta %d", n.port)
	n.logger("📡 ID do node: %s", n.host.ID().String())

	// Mostrar endereços de escuta
	for _, addr := range n.host.Addrs() {
		n.logger("📍 Escutando em: %s/p2p/%s", addr, n.host.ID())
	}

	// Iniciar monitoramento de peers
	go n.monitorPeers()

	// Iniciar heartbeat
	go n.heartbeat()

	return nil
}

// Stop para a rede P2P
func (n *P2PNetwork) Stop() error {
	n.cancel()
	return n.host.Close()
}

// Connect conecta a um peer
func (n *P2PNetwork) Connect(peerAddr string) error {
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("erro ao parsear endereço: %v", err)
	}

	peer, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("erro ao obter info do peer: %v", err)
	}

	if err := n.host.Connect(n.ctx, *peer); err != nil {
		return fmt.Errorf("erro ao conectar: %v", err)
	}

	n.logger("✅ Conectado ao peer: %s", peer.ID)
	return nil
}

// Subscribe inscreve em um tópico
func (n *P2PNetwork) Subscribe(topicName string) error {
	topic, err := n.pubsub.Join(topicName)
	if err != nil {
		return fmt.Errorf("erro ao entrar no tópico: %v", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("erro ao se inscrever: %v", err)
	}

	n.topics[topicName] = topic
	n.subscriptions[topicName] = sub

	// Iniciar handler de mensagens
	go n.handleMessages(topicName, sub)

	n.logger("📡 Inscrito no tópico: %s", topicName)
	return nil
}

// Publish publica uma mensagem em um tópico
func (n *P2PNetwork) Publish(topicName string, message Message) error {
	topic, exists := n.topics[topicName]
	if !exists {
		return fmt.Errorf("tópico não encontrado: %s", topicName)
	}

	// Adicionar metadados
	message.From = n.host.ID().String()
	message.Timestamp = time.Now().Unix()

	// Serializar mensagem
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	// Publicar
	if err := topic.Publish(n.ctx, data); err != nil {
		return fmt.Errorf("erro ao publicar: %v", err)
	}

	n.logger("📤 Mensagem publicada no tópico %s: %s", topicName, message.Type)
	return nil
}

// BroadcastBlock transmite um novo bloco
func (n *P2PNetwork) BroadcastBlock(block BlockMessage) error {
	message := Message{
		Type: "new_block",
		Data: block,
	}

	return n.Publish("blocks", message)
}

// BroadcastTransaction transmite uma nova transação
func (n *P2PNetwork) BroadcastTransaction(tx TransactionMessage) error {
	message := Message{
		Type: "new_transaction",
		Data: tx,
	}

	return n.Publish("transactions", message)
}

// GetPeers retorna lista de peers conectados
func (n *P2PNetwork) GetPeers() []*PeerInfo {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()

	peers := make([]*PeerInfo, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}

	return peers
}

// GetPeerCount retorna número de peers
func (n *P2PNetwork) GetPeerCount() int {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()
	return len(n.peers)
}

// GetNetworkInfo retorna informações da rede
func (n *P2PNetwork) GetNetworkInfo() map[string]interface{} {
	peers := n.GetPeers()

	// Calcular latência média
	var totalLatency int64
	for _, p := range peers {
		totalLatency += p.Latency
	}

	avgLatency := int64(0)
	if len(peers) > 0 {
		avgLatency = totalLatency / int64(len(peers))
	}

	return map[string]interface{}{
		"node_id":         n.host.ID().String(),
		"peer_count":      len(peers),
		"avg_latency":     avgLatency,
		"topics":          len(n.topics),
		"listening_addrs": n.host.Addrs(),
		"peers":           peers,
	}
}

// RegisterHandler registra um handler de mensagem
func (n *P2PNetwork) RegisterHandler(messageType string, handler func(Message) error) {
	n.messageHandlers[messageType] = handler
}

// handleStream processa streams de conexão
func (n *P2PNetwork) handleStream(stream network.Stream) {
	defer stream.Close()

	var message Message
	if err := json.NewDecoder(stream).Decode(&message); err != nil {
		n.logger("❌ Erro ao decodificar mensagem: %v", err)
		return
	}

	// Processar mensagem
	if handler, exists := n.messageHandlers[message.Type]; exists {
		if err := handler(message); err != nil {
			n.logger("❌ Erro ao processar mensagem %s: %v", message.Type, err)
		}
	} else {
		n.logger("⚠️ Handler não encontrado para mensagem: %s", message.Type)
	}
}

// handleMessages processa mensagens de pubsub
func (n *P2PNetwork) handleMessages(topicName string, sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			if n.ctx.Err() != nil {
				return // Contexto cancelado
			}
			n.logger("❌ Erro ao receber mensagem: %v", err)
			continue
		}

		// Ignorar mensagens próprias
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			n.logger("❌ Erro ao deserializar mensagem: %v", err)
			continue
		}

		n.logger("📥 Mensagem recebida do tópico %s: %s de %s",
			topicName, message.Type, msg.ReceivedFrom)

		// Processar mensagem
		if handler, exists := n.messageHandlers[message.Type]; exists {
			if err := handler(message); err != nil {
				n.logger("❌ Erro ao processar mensagem %s: %v", message.Type, err)
			}
		}
	}
}

// monitorPeers monitora peers conectados
func (n *P2PNetwork) monitorPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.peersMutex.Lock()

			// Limpar peers desconectados
			for peerID, peerInfo := range n.peers {
				if time.Since(peerInfo.LastSeen) > 2*time.Minute {
					delete(n.peers, peerID)
					n.logger("👋 Peer desconectado: %s", peerID)
				}
			}

			n.peersMutex.Unlock()
		}
	}
}

// heartbeat envia heartbeat para peers
func (n *P2PNetwork) heartbeat() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			message := Message{
				Type: "heartbeat",
				Data: map[string]interface{}{
					"timestamp": time.Now().Unix(),
					"version":   "1.0.0",
				},
			}

			// Enviar para todos os tópicos
			for topicName := range n.topics {
				n.Publish(topicName, message)
			}
		}
	}
}

// registerDefaultHandlers registra handlers padrão
func (n *P2PNetwork) registerDefaultHandlers() {
	// Handler de heartbeat
	n.RegisterHandler("heartbeat", func(msg Message) error {
		n.logger("💓 Heartbeat recebido de: %s", msg.From)
		return nil
	})

	// Handler de novo bloco
	n.RegisterHandler("new_block", func(msg Message) error {
		var block BlockMessage
		if data, ok := msg.Data.(map[string]interface{}); ok {
			blockData, _ := json.Marshal(data)
			json.Unmarshal(blockData, &block)
		}

		n.logger("🆕 Novo bloco recebido: #%d de %s", block.BlockNumber, msg.From)
		return nil
	})

	// Handler de nova transação
	n.RegisterHandler("new_transaction", func(msg Message) error {
		var tx TransactionMessage
		if data, ok := msg.Data.(map[string]interface{}); ok {
			txData, _ := json.Marshal(data)
			json.Unmarshal(txData, &tx)
		}

		n.logger("💸 Nova transação recebida: %s de %s", tx.TxHash, msg.From)
		return nil
	})
}

// networkNotifee implementa notificações de rede
type networkNotifee struct {
	network *P2PNetwork
}

func (nn *networkNotifee) Listen(network.Network, multiaddr.Multiaddr) {
	nn.network.logger("🔊 Escutando em nova interface")
}

func (nn *networkNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {
	nn.network.logger("🔇 Parou de escutar em interface")
}

func (nn *networkNotifee) Connected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()

	nn.network.peersMutex.Lock()
	nn.network.peers[peerID] = &PeerInfo{
		ID:       peerID.String(),
		LastSeen: time.Now(),
		Version:  "1.0.0",
		Services: []string{"blockchain"},
	}
	nn.network.peersMutex.Unlock()

	nn.network.logger("✅ Peer conectado: %s", peerID)
}

func (nn *networkNotifee) Disconnected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()

	nn.network.peersMutex.Lock()
	delete(nn.network.peers, peerID)
	nn.network.peersMutex.Unlock()

	nn.network.logger("❌ Peer desconectado: %s", peerID)
}

func (nn *networkNotifee) OpenedStream(network.Network, network.Stream) {}
func (nn *networkNotifee) ClosedStream(network.Network, network.Stream) {}
