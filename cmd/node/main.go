package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"ordm-main/pkg/dag"
	"ordm-main/pkg/network"
	"ordm-main/pkg/pow"
	"ordm-main/pkg/state"
	"ordm-main/pkg/types"
	"ordm-main/pkg/validator"
	"os"
	"strings"
	"sync"
	"time"
)

/* ============ Tipos ============ */

type Tx struct {
	ID     string
	Inputs []string
	Outs   []string
	Fee    int64
}

type MicroHeader struct {
	Parents    []string
	MerkleRoot string
	Timestamp  int64
	Difficulty int
	Nonce      uint64
	Hash       string
	MinerID    string
}

type MicroBlock struct {
	Header MicroHeader
	Txs    []Tx
}

type MacroBlock struct {
	Slot        int
	ParentTips  []string
	AcceptedTxs []Tx
	Hash        string
}

type DAG struct {
	sync.Mutex
	Blocks map[string]MicroBlock
	Edges  map[string][]string
	Tips   map[string]bool
}

func NewDAG() *DAG {
	return &DAG{
		Blocks: map[string]MicroBlock{},
		Edges:  map[string][]string{},
		Tips:   map[string]bool{},
	}
}

func (d *DAG) AddBlock(b MicroBlock) bool {
	d.Lock()
	defer d.Unlock()
	h := b.Header.Hash
	if _, ok := d.Blocks[h]; ok {
		return false
	}
	d.Blocks[h] = b
	d.Edges[h] = append([]string{}, b.Header.Parents...)
	d.Tips[h] = true
	for _, p := range b.Header.Parents {
		delete(d.Tips, p)
	}
	return true
}

func (d *DAG) TipsSlice(max int) []string {
	d.Lock()
	defer d.Unlock()
	out := make([]string, 0, max)
	for h := range d.Tips {
		out = append(out, h)
		if max > 0 && len(out) == max {
			break
		}
	}
	return out
}

func (d *DAG) GetTips(max int) []string {
	return d.TipsSlice(max)
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

/* ============ PoW ============ */

func prefixZeros(hexHash string) int {
	z := 0
	for _, ch := range hexHash {
		if ch == '0' {
			z++
		} else {
			break
		}
	}
	return z
}

func powMine(parents []string, merkle string, diff int, minerID string) MicroHeader {
	var nonce uint64
	for {
		h := MicroHeader{
			Parents:    parents,
			MerkleRoot: merkle,
			Timestamp:  time.Now().Unix(),
			Difficulty: diff,
			Nonce:      nonce,
			MinerID:    minerID,
		}
		bytes := []byte(fmt.Sprintf("%v|%s|%d|%d|%d|%s", parents, merkle, h.Timestamp, diff, nonce, minerID))
		sum := sha256.Sum256(bytes)
		h.Hash = hex.EncodeToString(sum[:])
		if prefixZeros(h.Hash) >= diff {
			return h
		}
		nonce++
	}
}

/* ============ Reconciliação ============ */

func topoSort(d *DAG) []string {
	inDegree := map[string]int{}
	for h := range d.Blocks {
		inDegree[h] = 0
	}
	for h, ps := range d.Edges {
		for range ps {
			inDegree[h]++
		}
	}

	var q []string
	for h, deg := range inDegree {
		if deg == 0 {
			q = append(q, h)
		}
	}

	var order []string
	seen := map[string]bool{}
	for len(q) > 0 {
		x := q[0]
		q = q[1:]
		if seen[x] {
			continue
		}
		seen[x] = true
		order = append(order, x)
		for y, ps := range d.Edges {
			for _, p := range ps {
				if p == x {
					inDegree[y]--
					if inDegree[y] == 0 {
						q = append(q, y)
					}
				}
			}
		}
	}
	return order
}

func reconcileMerge(dags []*DAG) (*DAG, []Tx) {
	global := NewDAG()
	for _, g := range dags {
		g.Lock()
		for h, b := range g.Blocks {
			global.Blocks[h] = b
			global.Edges[h] = append([]string{}, b.Header.Parents...)
		}
		g.Unlock()
	}
	for h := range global.Blocks {
		global.Tips[h] = true
	}
	for _, ps := range global.Edges {
		for _, p := range ps {
			delete(global.Tips, p)
		}
	}

	order := topoSort(global)
	seenTx := map[string]bool{}
	accepted := []Tx{}
	for _, h := range order {
		blk := global.Blocks[h]
		for _, tx := range blk.Txs {
			if seenTx[tx.ID] {
				continue
			}
			seenTx[tx.ID] = true
			accepted = append(accepted, tx)
		}
	}
	return global, accepted
}

/* ============ Rede (gossip TCP + JSON) ============ */

type Message struct {
	Type  string      `json:"type"`
	From  string      `json:"from"`
	Block *MicroBlock `json:"block,omitempty"`
}

func startServer(port string, d *DAG) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	fmt.Println("[net] escutando em", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			dec := json.NewDecoder(c)
			var msg Message
			if err := dec.Decode(&msg); err != nil {
				return
			}
			if msg.Type == "BLOCK" && msg.Block != nil {
				if d.AddBlock(*msg.Block) {
					fmt.Printf("[net] bloco %s de %s | pais=%v | tips=%d | dag=%d\n",
						short(msg.Block.Header.Hash), msg.From, msg.Block.Header.Parents, len(d.Tips), len(d.Blocks))
				}
			}
		}(conn)
	}
}

func broadcast(peers []string, msg Message) {
	payload, _ := json.Marshal(msg)
	for _, addr := range peers {
		go func(a string, body []byte) {
			c, err := net.DialTimeout("tcp", a, 600*time.Millisecond)
			if err != nil {
				return
			}
			defer c.Close()
			c.Write(body)
		}(addr, payload)
	}
}

/* ============ Util ============ */

func short(h string) string {
	if len(h) < 8 {
		return h
	}
	return h[:8]
}

/* ============ Miner & Macro ============ */

func miner(nodeName string, dag *DAG, difficulty int, parentsMax int, peers []string, self string, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			parents := dag.TipsSlice(parentsMax)
			merkle := fmt.Sprintf("merkle-%s-%d", nodeName, time.Now().UnixNano())

			txs := make([]Tx, 1+rand.Intn(5))
			for i := range txs {
				txs[i] = Tx{ID: fmt.Sprintf("%s-%d-%d", nodeName, time.Now().UnixNano(), i), Fee: int64(1 + rand.Intn(3))}
			}

			h := powMine(parents, merkle, difficulty, nodeName)
			block := MicroBlock{Header: h, Txs: txs}

			if dag.AddBlock(block) {
				fmt.Printf("[%s] Bloco %s (%d txs) pais=%v | tips=%d | dag=%d\n",
					nodeName, short(h.Hash), len(txs), parents, len(dag.Tips), len(dag.Blocks))

				broadcast(peers, Message{Type: "BLOCK", From: self, Block: &block})
			}

			time.Sleep(350 * time.Millisecond)
		}
	}
}

func macroBlockCreator(dags []*DAG) {
	slot := 1
	t := time.NewTicker(5 * time.Second)
	for range t.C {
		globalDAG, acceptedTxs := reconcileMerge(dags)
		hashBytes := []byte(fmt.Sprintf("%d|%d|%d", slot, len(globalDAG.Tips), len(acceptedTxs)))
		hash := sha256.Sum256(hashBytes)
		macro := MacroBlock{
			Slot:        slot,
			ParentTips:  keys(globalDAG.Tips),
			AcceptedTxs: acceptedTxs,
			Hash:        hex.EncodeToString(hash[:]),
		}
		shortTips := make([]string, 0, len(macro.ParentTips))
		for _, h := range macro.ParentTips {
			shortTips = append(shortTips, short(h))
		}
		fmt.Printf("\n=== Macro-bloco %d ===\nBlocos aceitos: %d\nTips: %v\nHash: %s\n===\n\n",
			slot, len(acceptedTxs), shortTips, short(macro.Hash))
		slot++
	}
}

/* ============ Persistência do DAG ============ */

func saveDAG(filename string, dag *DAG) error {
	dag.Lock()
	defer dag.Unlock()
	data, err := json.MarshalIndent(dag, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func loadDAG(filename string) (*DAG, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return NewDAG(), nil
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var dag DAG
	if err := json.Unmarshal(data, &dag); err != nil {
		return nil, err
	}
	return &dag, nil
}

/* ============ main ============ */

func main() {
	fmt.Println("=== INICIANDO NODE ===")
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 4 {
		fmt.Println("uso: go run ./cmd/node <NODE_ID> <PORTA> <peer1,peer2,...|->")
		return
	}

	nodeID := os.Args[1]
	port := os.Args[2]
	rawPeers := os.Args[3]

	var peers []string
	if rawPeers != "-" {
		peers = strings.Split(rawPeers, ",")
	}
	_ = peers // evita "declared and not used" por enquanto
	_ = port  // evita "declared and not used"

	// --- Carregar DAG ---
	d := dag.New()

	// --- Iniciar servidor de rede ---
	go network.StartServer(port, d)
	fmt.Printf("[%s] Servidor iniciado na porta %s\n", nodeID, port)

	// --- Sincronizar com peers ---
	if len(peers) > 0 {
		fmt.Printf("[%s] Sincronizando com peers: %v\n", nodeID, peers)
		network.SyncNode(peers, d)
	}

	// --- Ledger inicial ---
	initialBalances := map[string]int64{
		"Alice": 100,
		"Bob":   50,
	}
	ledger := state.NewLedger(initialBalances)

	// --- Converte Snapshot para int (validator) ---
	snapshot := make(map[string]int)
	for k, v := range ledger.Snapshot() {
		snapshot[k] = int(v)
	}

	// --- Validators ---
	validator1 := validator.NewNode("validator1", snapshot)
	validator2 := validator.NewNode("validator2", snapshot)
	validators := []*validator.Node{validator1, validator2}

	// --- Loop de mineração (PoW) ---
	const difficulty = 0 // Dificuldade 0 para mineração instantânea (para testes)
	const parentsMax = 2

	go func() {
		fmt.Printf("[%s] Iniciando mineração...\n", nodeID)
		for {
			fmt.Printf("[%s] Iniciando novo ciclo de mineração...\n", nodeID)

			// Simula txs pendentes
			pendingTxs := []types.Tx{
				{ID: types.HashStrings(fmt.Sprintf("tx-%d", time.Now().UnixNano())), From: "Alice", To: "Bob", Amount: 10, Fee: 1, Body: "Alice->Bob:10"},
				{ID: types.HashStrings(fmt.Sprintf("tx-%d", time.Now().UnixNano())), From: "Bob", To: "Alice", Amount: 5, Fee: 1, Body: "Bob->Alice:5"},
			}

			fmt.Printf("[%s] Transações pendentes: %d\n", nodeID, len(pendingTxs))

			// Aplica transações no ledger
			appliedTxs := ledger.ApplyTxs(pendingTxs)
			fmt.Printf("[%s] Transações aplicadas: %d\n", nodeID, len(appliedTxs))

			// Obter tips do DAG
			tips := d.GetTips(parentsMax)
			fmt.Printf("[%s] Tips obtidos: %v\n", nodeID, tips)

			// Calcular merkle root
			merkle := merkleRoot(appliedTxs)
			fmt.Printf("[%s] Merkle root: %s\n", nodeID, merkle)

			// Minerar bloco
			fmt.Printf("[%s] Iniciando mineração com dificuldade %d...\n", nodeID, difficulty)
			blockHeader := pow.Mine(tips, merkle, difficulty, nodeID)
			fmt.Printf("[%s] Bloco minerado! Hash: %s\n", nodeID, blockHeader.Hash)

			// Criar MicroBlock
			block := &types.MicroBlock{
				Header: blockHeader,
				Txs:    appliedTxs,
			}

			fmt.Printf("[%s] MicroBlock criado\n", nodeID)

			// Recompensa minerador
			ledger.Reward(nodeID, 10)
			fmt.Printf("[%s] Recompensa de 10 adicionada\n", nodeID)

			// Adiciona bloco ao DAG
			added := d.AddBlock(block)

			// Broadcast do bloco para outros nodes
			if len(peers) > 0 {
				network.Broadcast(peers, network.Message{
					Type:  "BLOCK",
					Block: block,
					From:  nodeID,
				})
			}
			fmt.Printf("[%s] Bloco adicionado ao DAG: %t\n", nodeID, added)

			fmt.Printf("[%s] Bloco minerado com sucesso!\n", nodeID)
			fmt.Printf("[%s] Hash: %s\n", nodeID, block.Header.Hash)
			fmt.Printf("[%s] Nonce: %d\n", nodeID, block.Header.Nonce)
			fmt.Printf("[%s] DAG size: %d\n", nodeID, len(d.Blocks))
			fmt.Printf("[%s] Ledger snapshot: %v\n", nodeID, ledger.Snapshot())

			// --- PoS validação ---
			fmt.Printf("[%s] Iniciando validação PoS...\n", nodeID)
			for _, v := range validators {
				// Converter MicroBlock para Block do validator
				validatorBlock := &validator.Block{
					ID:           block.Header.Hash,
					ParentID:     "",
					Transactions: []validator.Transaction{},
				}

				// Converter transações
				for _, tx := range block.Txs {
					validatorTx := validator.Transaction{
						From:   tx.From,
						To:     tx.To,
						Amount: int(tx.Amount),
					}
					validatorBlock.Transactions = append(validatorBlock.Transactions, validatorTx)
				}

				if err := v.ReceiveBlock(validatorBlock); err != nil {
					fmt.Printf("[%s] %s rejeitou bloco: %v\n", nodeID, v.Name, err)
				} else {
					fmt.Printf("[%s] %s validou bloco com sucesso\n", nodeID, v.Name)
				}
			}

			fmt.Printf("[%s] Ciclo de mineração concluído. Aguardando 3 segundos...\n", nodeID)
			time.Sleep(3 * time.Second)
			fmt.Printf("[%s] Acordou do sleep, iniciando próximo ciclo...\n", nodeID)
		}
	}()

	// --- Macroblocks (se aplicável) ---
	// go macroBlockCreator([]*dag.DAG{d}) // habilite se tiver a função

	fmt.Printf("[%s] Node configurado. Iniciando loop principal...\n", nodeID)
	select {} // mantém o node rodando
}

// --- Funções auxiliares ---
func merkleRoot(txs []types.Tx) string {
	var hs []string
	for _, tx := range txs {
		hs = append(hs, tx.ID)
	}
	return types.HashStrings(hs...)
}
