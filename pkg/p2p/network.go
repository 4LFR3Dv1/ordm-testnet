package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Connected bool      `json:"connected"`
}

// PeerExchangeMessage representa uma mensagem de troca de peers
type PeerExchangeMessage struct {
	Peers     []PeerInfo `json:"peers"`
	From      string     `json:"from"`
	Timestamp int64      `json:"timestamp"`
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
	topics          map[string]interface{}
	subscriptions   map[string]interface{}
	logger          func(string, ...interface{})
	port            int
	bootstrapConfig struct {
		BootstrapPeers []string
		PeerExchange   bool
		ReconnectDelay time.Duration
	}
	discoveryCtx       context.Context
	reconnectTicker    *time.Ticker
	knownPeers         map[string]time.Time
	knownPeersMutex    sync.RWMutex
	peerExchangeTicker *time.Ticker
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
		topics:          make(map[string]interface{}),
		subscriptions:   make(map[string]interface{}),
		logger:          logger,
		port:            port,
		bootstrapConfig: struct {
			BootstrapPeers []string
			PeerExchange   bool
			ReconnectDelay time.Duration
		}{
			BootstrapPeers: []string{},
			PeerExchange:   false,
			ReconnectDelay: 30 * time.Second,
		},
		discoveryCtx: ctx,
		knownPeers:   make(map[string]time.Time),
	}

	// Configurar handlers de rede
	host.SetStreamHandler("/blockchain/1.0.0", network.handleStream)
	host.Network().Notify(&networkNotifee{network: network})

	// Registrar handlers padrão
	network.registerDefaultHandlers()

	// Iniciar conectividade automática
	network.startAutoConnectivity()

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
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	// Publicar mensagem via libp2p pubsub
	if topicPub, ok := topic.(*pubsub.Topic); ok {
		if err := topicPub.Publish(n.ctx, messageData); err != nil {
			return fmt.Errorf("erro ao publicar mensagem: %v", err)
		}
	} else {
		return fmt.Errorf("tipo de tópico inválido")
	}

	n.logger("📤 Publicando mensagem no tópico: %s", topicName)
	return nil
}

// BroadcastBlock transmite um novo bloco
func (n *P2PNetwork) BroadcastBlock(block BlockMessage) error {
	message := Message{
		Type: "new_block",
		Data: block,
	}

	return n.Publish("ordm/blocks", message)
}

// BroadcastTransaction transmite uma nova transação
func (n *P2PNetwork) BroadcastTransaction(tx TransactionMessage) error {
	message := Message{
		Type: "new_transaction",
		Data: tx,
	}

	return n.Publish("ordm/transactions", message)
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

	// Handler de troca de peers
	n.RegisterHandler("peer_exchange", func(msg Message) error {
		return n.handlePeerExchange(msg)
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

// ===== CONECTIVIDADE AUTOMÁTICA =====

// startAutoConnectivity inicia a conectividade automática
func (n *P2PNetwork) startAutoConnectivity() {
	n.logger("🔄 Iniciando conectividade automática...")

	// Conectar aos peers bootstrap
	go n.connectToBootstrapPeers()

	// Iniciar reconexão automática
	go n.startAutoReconnect()

	// Iniciar peer exchange
	if n.bootstrapConfig.PeerExchange {
		go n.startPeerExchange()
	}
}

// startAutoReconnect inicia reconexão automática
func (n *P2PNetwork) startAutoReconnect() {
	n.reconnectTicker = time.NewTicker(n.bootstrapConfig.ReconnectDelay)
	defer n.reconnectTicker.Stop()

	for {
		select {
		case <-n.discoveryCtx.Done():
			return
		case <-n.reconnectTicker.C:
			n.reconnectToKnownPeers()
		}
	}
}

// startPeerExchange inicia troca de peers
func (n *P2PNetwork) startPeerExchange() {
	n.peerExchangeTicker = time.NewTicker(2 * time.Minute)
	defer n.peerExchangeTicker.Stop()

	for {
		select {
		case <-n.discoveryCtx.Done():
			return
		case <-n.peerExchangeTicker.C:
			n.broadcastPeerList()
		}
	}
}

// connectToBootstrapPeers conecta aos peers bootstrap
func (n *P2PNetwork) connectToBootstrapPeers() {
	n.logger("🔗 Conectando aos peers bootstrap...")

	// Conectar aos peers bootstrap configurados
	for _, peerAddr := range n.bootstrapConfig.BootstrapPeers {
		go func(addr string) {
			if err := n.Connect(addr); err != nil {
				n.logger("⚠️ Falha ao conectar ao peer bootstrap %s: %v", addr, err)
				// Adicionar à lista de peers conhecidos para tentar reconectar
				n.addKnownPeer(addr)
			}
		}(peerAddr)
	}

	// Tentar conectar aos peers locais conhecidos
	go n.connectToLocalPeers()
}

// reconnectToKnownPeers tenta reconectar aos peers conhecidos
func (n *P2PNetwork) reconnectToKnownPeers() {
	n.knownPeersMutex.Lock()
	defer n.knownPeersMutex.Unlock()

	now := time.Now()
	for peerAddr, lastSeen := range n.knownPeers {
		// Tentar reconectar se passou mais de 5 minutos
		if now.Sub(lastSeen) > 5*time.Minute {
			go func(addr string) {
				if err := n.Connect(addr); err != nil {
					n.logger("⚠️ Falha na reconexão ao peer %s: %v", addr, err)
				} else {
					n.logger("✅ Reconectado ao peer: %s", addr)
					// Remover da lista de conhecidos se conectou com sucesso
					n.knownPeersMutex.Lock()
					delete(n.knownPeers, addr)
					n.knownPeersMutex.Unlock()
				}
			}(peerAddr)
		}
	}
}

// addKnownPeer adiciona um peer à lista de conhecidos
func (n *P2PNetwork) addKnownPeer(peerAddr string) {
	n.knownPeersMutex.Lock()
	defer n.knownPeersMutex.Unlock()

	n.knownPeers[peerAddr] = time.Now()
	n.logger("📝 Peer adicionado à lista de conhecidos: %s", peerAddr)
}

// broadcastPeerList envia lista de peers para outros nodes
func (n *P2PNetwork) broadcastPeerList() {
	n.peersMutex.RLock()
	peerList := make([]PeerInfo, 0, len(n.peers))
	for _, peer := range n.peers {
		peerList = append(peerList, *peer)
	}
	n.peersMutex.RUnlock()

	// Adicionar peers conhecidos
	n.knownPeersMutex.RLock()
	for peerAddr := range n.knownPeers {
		peerList = append(peerList, PeerInfo{
			ID:        peerAddr,
			Connected: false,
			LastSeen:  n.knownPeers[peerAddr],
		})
	}
	n.knownPeersMutex.RUnlock()

	message := Message{
		Type: "peer_exchange",
		Data: PeerExchangeMessage{
			Peers:     peerList,
			From:      n.host.ID().String(),
			Timestamp: time.Now().Unix(),
		},
		From:      n.host.ID().String(),
		Timestamp: time.Now().Unix(),
	}

	// Enviar para todos os tópicos
	for topicName := range n.topics {
		n.Publish(topicName, message)
	}

	n.logger("📤 Lista de %d peers enviada", len(peerList))
}

// handlePeerExchange processa mensagens de troca de peers
func (n *P2PNetwork) handlePeerExchange(msg Message) error {
	var peerExchange PeerExchangeMessage
	if data, ok := msg.Data.(map[string]interface{}); ok {
		exchangeData, _ := json.Marshal(data)
		json.Unmarshal(exchangeData, &peerExchange)
	}

	n.logger("📥 Recebida lista de %d peers de %s", len(peerExchange.Peers), msg.From)

	// Tentar conectar aos novos peers
	for _, peerInfo := range peerExchange.Peers {
		if peerInfo.ID != n.host.ID().String() {
			// Tentar conectar se não estiver conectado
			if !peerInfo.Connected {
				go func(peerID string) {
					// Tentar conectar usando o ID do peer
					if err := n.connectToPeerByID(peerID); err != nil {
						n.logger("⚠️ Falha ao conectar ao peer %s: %v", peerID, err)
						n.addKnownPeer(peerID)
					}
				}(peerInfo.ID)
			}
		}
	}

	return nil
}

// connectToLocalPeers tenta conectar aos peers locais conhecidos
func (n *P2PNetwork) connectToLocalPeers() {
	n.logger("🔍 Tentando conectar aos peers locais...")

	// Tentar conectar aos peers locais baseado na porta atual
	// Se estamos na porta 3003, tentar conectar às portas 3004 e 3005
	// Se estamos na porta 3004, tentar conectar às portas 3003 e 3005
	// Se estamos na porta 3005, tentar conectar às portas 3003 e 3004

	var targetPorts []int
	switch n.port {
	case 3003:
		targetPorts = []int{3004, 3005}
	case 3004:
		targetPorts = []int{3003, 3005}
	case 3005:
		targetPorts = []int{3003, 3004}
	default:
		// Para outras portas, tentar as portas padrão
		targetPorts = []int{3003, 3004, 3005}
	}

	// Tentar conectar a cada porta alvo
	for _, targetPort := range targetPorts {
		if targetPort != n.port { // Não conectar a si mesmo
			go func(port int) {
				// Aguardar um pouco antes de tentar conectar
				time.Sleep(2 * time.Second)

				// Tentar conectar usando o método melhorado
				if err := n.ConnectToLocalPeer(port); err != nil {
					n.logger("⚠️ Falha ao conectar ao peer na porta %d: %v", port, err)
				} else {
					n.logger("✅ Conectado ao peer na porta %d", port)
				}
			}(targetPort)
		}
	}
}

// connectToPeerByID tenta conectar a um peer pelo ID
func (n *P2PNetwork) connectToPeerByID(peerID string) error {
	// Para peers locais, tentar conectar usando endereços padrão
	localAddresses := []string{
		fmt.Sprintf("/ip4/127.0.0.1/tcp/3003/p2p/%s", peerID),
		fmt.Sprintf("/ip4/127.0.0.1/tcp/3004/p2p/%s", peerID),
		fmt.Sprintf("/ip4/127.0.0.1/tcp/3005/p2p/%s", peerID),
	}

	for _, addr := range localAddresses {
		if err := n.Connect(addr); err == nil {
			return nil
		}
	}

	return fmt.Errorf("não foi possível conectar ao peer %s", peerID)
}

// GetConnectedPeers retorna lista de peers conectados
func (n *P2PNetwork) GetConnectedPeers() []PeerInfo {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()

	peers := make([]PeerInfo, 0, len(n.peers))
	for _, peer := range n.peers {
		peer.Connected = true
		peers = append(peers, *peer)
	}

	return peers
}

// GetKnownPeers retorna lista de peers conhecidos
func (n *P2PNetwork) GetKnownPeers() []string {
	n.knownPeersMutex.RLock()
	defer n.knownPeersMutex.RUnlock()

	peers := make([]string, 0, len(n.knownPeers))
	for peer := range n.knownPeers {
		peers = append(peers, peer)
	}

	return peers
}

// ConnectToPeer conecta a um peer específico
func (n *P2PNetwork) ConnectToPeer(peerAddr string) error {
	n.logger("🔗 Conectando manualmente ao peer: %s", peerAddr)
	return n.Connect(peerAddr)
}

// ConnectToLocalPeer conecta a um peer local pela porta
func (n *P2PNetwork) ConnectToLocalPeer(port int) error {
	// Para conectar a um peer local, precisamos descobrir seu ID primeiro
	// Vamos tentar conectar usando o endereço básico e ver se conseguimos descobrir o peer

	n.logger("🔗 Tentando conectar ao peer local na porta %d", port)

	// Primeiro, tentar conectar usando o endereço básico
	peerAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)

	// Tentar conectar
	if err := n.Connect(peerAddr); err != nil {
		n.logger("⚠️ Falha ao conectar usando endereço básico: %v", err)

		// Se falhar, tentar descobrir o peer usando mDNS ou outros métodos
		// Por enquanto, vamos tentar conectar usando um ID conhecido se disponível
		return n.tryConnectWithKnownPeerID(port)
	}

	n.logger("✅ Conectado ao peer local na porta %d", port)
	return nil
}

// tryConnectWithKnownPeerID tenta conectar usando IDs de peers conhecidos
func (n *P2PNetwork) tryConnectWithKnownPeerID(port int) error {
	// Primeiro, tentar descobrir o ID do peer dinamicamente
	if peerID, err := n.discoverPeerID(port); err == nil {
		// Tentar conectar usando o ID descoberto
		peerAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, peerID)
		n.logger("🔗 Tentando conectar usando ID descoberto: %s", peerAddr)

		if err := n.Connect(peerAddr); err != nil {
			n.logger("⚠️ Falha ao conectar usando ID descoberto: %v", err)
			return err
		}

		n.logger("✅ Conectado ao peer usando ID descoberto na porta %d", port)
		return nil
	}

	// Se não conseguir descobrir, tentar com IDs conhecidos
	knownPeerIDs := map[int]string{
		3003: "12D3KooWFrCqre68gmYp3nQyPRuWgh7tT8STQkChfYg56kcLu9BU",
		3004: "12D3KooWPMVcKNKkZKQ714vpUq8FcsPcL49vuJmozEuS4LPTRh1P",
		3005: "12D3KooWQU5YR6bmHydwN7mggKmJLVDweLHZvf8f9HeWHuHEGEjg",
	}

	if peerID, exists := knownPeerIDs[port]; exists {
		// Tentar conectar usando o ID conhecido
		peerAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, peerID)
		n.logger("🔗 Tentando conectar usando ID conhecido: %s", peerAddr)

		if err := n.Connect(peerAddr); err != nil {
			n.logger("⚠️ Falha ao conectar usando ID conhecido: %v", err)
			return err
		}

		n.logger("✅ Conectado ao peer usando ID conhecido na porta %d", port)
		return nil
	}

	return fmt.Errorf("não foi possível conectar ao peer na porta %d - ID desconhecido", port)
}

// discoverPeerID tenta descobrir o ID do peer fazendo uma requisição HTTP
func (n *P2PNetwork) discoverPeerID(port int) (string, error) {
	// Calcular a porta HTTP correspondente corretamente
	// Mapeamento: P2P 3003 -> HTTP 8081, P2P 3004 -> HTTP 8082, P2P 3005 -> HTTP 8083
	var httpPort int
	switch port {
	case 3003:
		httpPort = 8081
	case 3004:
		httpPort = 8082
	case 3005:
		httpPort = 8083
	default:
		// Para outras portas, tentar o cálculo anterior
		httpPort = port + 8078
	}

	// Tentar fazer uma requisição HTTP para obter o status P2P
	url := fmt.Sprintf("http://localhost:%d/api/p2p-status", httpPort)

	// Fazer requisição HTTP com timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("falha ao conectar ao peer HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status HTTP inválido: %d", resp.StatusCode)
	}

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("falha ao ler resposta: %v", err)
	}

	// Parsear JSON para extrair o node_id
	var p2pStatus map[string]interface{}
	if err := json.Unmarshal(body, &p2pStatus); err != nil {
		return "", fmt.Errorf("falha ao parsear JSON: %v", err)
	}

	// Extrair node_id do network_info
	if networkInfo, ok := p2pStatus["network_info"].(map[string]interface{}); ok {
		if nodeID, ok := networkInfo["node_id"].(string); ok {
			n.logger("🔍 ID do peer descoberto na porta %d: %s", port, nodeID)
			return nodeID, nil
		}
	}

	return "", fmt.Errorf("não foi possível extrair node_id da resposta")
}

// discoverPeerIDRealTime descobre o ID do peer em tempo real e conecta
func (n *P2PNetwork) discoverPeerIDRealTime(port int) error {
	n.logger("🔍 Descobrindo ID do peer em tempo real na porta %d...", port)

	// Tentar descobrir o ID do peer
	peerID, err := n.discoverPeerID(port)
	if err != nil {
		n.logger("⚠️ Falha ao descobrir ID do peer na porta %d: %v", port, err)
		return err
	}

	// Tentar conectar usando o ID descoberto
	peerAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, peerID)
	n.logger("🔗 Conectando usando ID descoberto em tempo real: %s", peerAddr)

	if err := n.Connect(peerAddr); err != nil {
		n.logger("⚠️ Falha ao conectar usando ID descoberto: %v", err)
		return err
	}

	n.logger("✅ Conectado ao peer usando ID descoberto em tempo real na porta %d", port)
	return nil
}

// ConnectToPeerByPort conecta a um peer pela porta usando descoberta dinâmica
func (n *P2PNetwork) ConnectToPeerByPort(port int) error {
	n.logger("🔗 Conectando ao peer na porta %d usando descoberta dinâmica...", port)

	// Primeiro tentar descoberta em tempo real
	if err := n.discoverPeerIDRealTime(port); err == nil {
		return nil
	}

	// Se falhar, tentar métodos alternativos
	n.logger("⚠️ Descoberta em tempo real falhou, tentando métodos alternativos...")

	// Tentar conectar usando o método melhorado
	return n.ConnectToLocalPeer(port)
}

// GetConnectionStatus retorna o status das conexões
func (n *P2PNetwork) GetConnectionStatus() map[string]interface{} {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()

	status := map[string]interface{}{
		"connected_peers": len(n.peers),
		"known_peers":     len(n.knownPeers),
		"port":            n.port,
		"node_id":         n.host.ID().String(),
	}

	// Adicionar lista de peers conectados
	connectedPeers := make([]string, 0, len(n.peers))
	for peerID := range n.peers {
		connectedPeers = append(connectedPeers, peerID.String())
	}
	status["peer_list"] = connectedPeers

	return status
}
