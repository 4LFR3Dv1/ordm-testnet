package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// SimpleP2PNode representa um node P2P simplificado
type SimpleP2PNode struct {
	nodeID       string
	port         int
	peers        map[string]*PeerInfo
	mu           sync.RWMutex
	listener     net.Listener
	messageQueue chan *Message
	running      bool
	stopChan     chan struct{}
}

// PeerInfo informações sobre um peer
type PeerInfo struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	LastSeen  time.Time `json:"last_seen"`
	Latency   int64     `json:"latency_ms"`
	Stake     int64     `json:"stake"`
	Validator bool      `json:"validator"`
	Connected bool      `json:"connected"`
}

// Message representa uma mensagem P2P
type Message struct {
	Type      string          `json:"type"`
	From      string          `json:"from"`
	To        string          `json:"to,omitempty"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	Signature string          `json:"signature,omitempty"`
	TTL       int             `json:"ttl"`
}

// NewSimpleP2PNode cria um novo node P2P simplificado
func NewSimpleP2PNode(nodeID string, port int) *SimpleP2PNode {
	return &SimpleP2PNode{
		nodeID:       nodeID,
		port:         port,
		peers:        make(map[string]*PeerInfo),
		messageQueue: make(chan *Message, 1000),
		stopChan:     make(chan struct{}),
	}
}

// Start inicia o node P2P
func (sp *SimpleP2PNode) Start() error {
	// Iniciar listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", sp.port))
	if err != nil {
		return fmt.Errorf("erro ao iniciar listener: %v", err)
	}
	sp.listener = listener
	sp.running = true

	// Aceitar conexões
	go sp.acceptConnections()

	// Processar mensagens
	go sp.processMessages()

	// Heartbeat
	go sp.heartbeat()

	fmt.Printf("🌐 Node P2P iniciado: %s (porta %d)\n", sp.nodeID, sp.port)
	return nil
}

// Stop para o node P2P
func (sp *SimpleP2PNode) Stop() error {
	sp.running = false
	close(sp.stopChan)

	if sp.listener != nil {
		sp.listener.Close()
	}

	return nil
}

// acceptConnections aceita conexões de entrada
func (sp *SimpleP2PNode) acceptConnections() {
	for sp.running {
		conn, err := sp.listener.Accept()
		if err != nil {
			if sp.running {
				fmt.Printf("❌ Erro ao aceitar conexão: %v\n", err)
			}
			continue
		}

		go sp.handleConnection(conn)
	}
}

// handleConnection processa uma conexão
func (sp *SimpleP2PNode) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Ler mensagem
	var message Message
	if err := json.NewDecoder(conn).Decode(&message); err != nil {
		return
	}

	// Adicionar à fila de mensagens
	select {
	case sp.messageQueue <- &message:
	default:
		// Fila cheia, descartar mensagem
	}
}

// processMessages processa mensagens da fila
func (sp *SimpleP2PNode) processMessages() {
	for {
		select {
		case message := <-sp.messageQueue:
			sp.handleMessage(message)
		case <-sp.stopChan:
			return
		}
	}
}

// handleMessage processa uma mensagem
func (sp *SimpleP2PNode) handleMessage(message *Message) {
	switch message.Type {
	case "BLOCK":
		sp.handleBlockMessage(message)
	case "TRANSACTION":
		sp.handleTransactionMessage(message)
	case "PEER_INFO":
		sp.handlePeerInfoMessage(message)
	case "HEARTBEAT":
		sp.handleHeartbeatMessage(message)
	case "SYNC_REQUEST":
		sp.handleSyncRequest(message)
	case "SYNC_RESPONSE":
		sp.handleSyncResponse(message)
	default:
		fmt.Printf("📨 Mensagem desconhecida: %s\n", message.Type)
	}
}

// handleBlockMessage processa mensagem de bloco
func (sp *SimpleP2PNode) handleBlockMessage(message *Message) {
	fmt.Printf("📦 Bloco recebido de %s\n", message.From)
	// Implementar lógica de processamento de bloco
}

// handleTransactionMessage processa mensagem de transação
func (sp *SimpleP2PNode) handleTransactionMessage(message *Message) {
	fmt.Printf("💸 Transação recebida de %s\n", message.From)
	// Implementar lógica de processamento de transação
}

// handlePeerInfoMessage processa mensagem de info do peer
func (sp *SimpleP2PNode) handlePeerInfoMessage(message *Message) {
	var peerInfo PeerInfo
	if err := json.Unmarshal(message.Data, &peerInfo); err != nil {
		return
	}

	sp.mu.Lock()
	sp.peers[peerInfo.ID] = &peerInfo
	sp.mu.Unlock()
}

// handleHeartbeatMessage processa heartbeat
func (sp *SimpleP2PNode) handleHeartbeatMessage(message *Message) {
	sp.mu.Lock()
	if peer, exists := sp.peers[message.From]; exists {
		peer.LastSeen = time.Now()
	}
	sp.mu.Unlock()
}

// handleSyncRequest processa requisição de sincronização
func (sp *SimpleP2PNode) handleSyncRequest(message *Message) {
	// Implementar sincronização
}

// handleSyncResponse processa resposta de sincronização
func (sp *SimpleP2PNode) handleSyncResponse(message *Message) {
	// Implementar processamento de sincronização
}

// ConnectToPeer conecta a um peer
func (sp *SimpleP2PNode) ConnectToPeer(address string) error {
	// Adicionar peer à lista
	sp.mu.Lock()
	sp.peers[address] = &PeerInfo{
		ID:        address,
		Address:   address,
		LastSeen:  time.Now(),
		Connected: true,
	}
	sp.mu.Unlock()

	fmt.Printf("🔗 Conectado ao peer: %s\n", address)
	return nil
}

// SendMessage envia mensagem para um peer
func (sp *SimpleP2PNode) SendMessage(address string, message *Message) error {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("erro ao conectar: %v", err)
	}
	defer conn.Close()

	// Serializar mensagem
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	// Enviar
	_, err = conn.Write(data)
	return err
}

// Broadcast envia mensagem para todos os peers
func (sp *SimpleP2PNode) Broadcast(message *Message) error {
	sp.mu.RLock()
	peers := make([]string, 0, len(sp.peers))
	for peerID := range sp.peers {
		peers = append(peers, peerID)
	}
	sp.mu.RUnlock()

	for _, peerID := range peers {
		go sp.SendMessage(peerID, message)
	}

	return nil
}

// heartbeat envia heartbeats periódicos
func (sp *SimpleP2PNode) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sp.sendHeartbeat()
		case <-sp.stopChan:
			return
		}
	}
}

// sendHeartbeat envia heartbeat para todos os peers
func (sp *SimpleP2PNode) sendHeartbeat() {
	message := &Message{
		Type:      "HEARTBEAT",
		From:      sp.nodeID,
		Timestamp: time.Now(),
		TTL:       1,
	}

	sp.Broadcast(message)
}

// GetPeers retorna lista de peers
func (sp *SimpleP2PNode) GetPeers() []*PeerInfo {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	peers := make([]*PeerInfo, 0, len(sp.peers))
	for _, peer := range sp.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetPeerCount retorna número de peers
func (sp *SimpleP2PNode) GetPeerCount() int {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return len(sp.peers)
}

// GetNodeID retorna o ID do node
func (sp *SimpleP2PNode) GetNodeID() string {
	return sp.nodeID
}

// IsRunning verifica se o node está rodando
func (sp *SimpleP2PNode) IsRunning() bool {
	return sp.running
}
