// go.mod: module l2off

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	"math/rand"
)

type Tx struct {
	ID     string
	Inputs []string // simplificado
	Outs   []string
	Fee    int64
}

type MicroHeader struct {
	Parents     []string
	MerkleRoot  string
	Timestamp   int64
	Difficulty  int // zeros no prefixo
	Nonce       uint64
	Hash        string
}

type MicroBlock struct {
	Header MicroHeader
	Txs    []Tx
}

type DAG struct {
	Blocks map[string]MicroBlock
	Edges  map[string][]string // hash -> pais
	Tips   map[string]bool
}

func NewDAG() *DAG {
	return &DAG{
		Blocks: map[string]MicroBlock{},
		Edges:  map[string][]string{},
		Tips:   map[string]bool{},
	}
}

func (d *DAG) AddBlock(b MicroBlock) {
	h := b.Header.Hash
	if _, ok := d.Blocks[h]; ok {
		return
	}
	d.Blocks[h] = b
	d.Edges[h] = b.Header.Parents
	d.Tips[h] = true
	for _, p := range b.Header.Parents {
		delete(d.Tips, p)
	}
}

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

func powMine(parents []string, merkle string, diff int) MicroHeader {
	var nonce uint64
	for {
		h := MicroHeader{
			Parents:    parents,
			MerkleRoot: merkle,
			Timestamp:  time.Now().Unix(),
			Difficulty: diff,
			Nonce:      nonce,
		}
		bytes := []byte(fmt.Sprintf("%v|%s|%d|%d|%d", parents, merkle, h.Timestamp, diff, nonce))
		sum := sha256.Sum256(bytes)
		h.Hash = hex.EncodeToString(sum[:])
		if prefixZeros(h.Hash) >= diff {
			return h
		}
		nonce++
	}
}

func topoSort(d *DAG) []string {
	// Kahn simplificado
	inDegree := map[string]int{}
	for h := range d.Blocks {
		inDegree[h] = 0
	}
	for h, ps := range d.Edges {
		for _, p := range ps {
			inDegree[h]++
			_ = p
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

func workValue(diff int) int64 {
	// proxy do trabalho acumulado
	if diff <= 0 { return 1 }
	v := int64(1)
	for i := 0; i < diff; i++ { v *= 2 }
	return v
}

func reconcileMerge(dags []*DAG) (*DAG, []string) {
	// merge simples: une blocos e ordena, resolvendo conflitos por "trabalho acumulado"
	global := NewDAG()
	for _, g := range dags {
		for h, b := range g.Blocks {
			global.Blocks[h] = b
			global.Edges[h] = b.Header.Parents
		}
	}
	// recomputa tips
	for h := range global.Blocks {
		global.Tips[h] = true
	}
	for h, ps := range global.Edges {
		for _, p := range ps {
			delete(global.Tips, p)
			_ = h
		}
	}
	order := topoSort(global)

	// conflito de tx (super simples): mantém a primeira aparição no caminho de maior trabalho
	seenTx := map[string]bool{}
	accWork := map[string]int64{}
	accepted := []string{}

	for _, h := range order {
		blk := global.Blocks[h]
		// calcula trabalho acumulado deste cabeçalho
		w := workValue(blk.Header.Difficulty)
		// (num protótipo real some trabalho de ancestrais; aqui mantemos simples)
		accWork[h] = w

		validTxs := []Tx{}
		conflict := false
		for _, tx := range blk.Txs {
			if seenTx[tx.ID] {
				conflict = true
				continue
			}
			validTxs = append(validTxs, tx)
		}
		if conflict && len(validTxs) == 0 {
			// se só conflitos, aceita vazio (ou descarta)
		}
		// marca txs aceitas
		for _, tx := range validTxs {
			seenTx[tx.ID] = true
		}
		accepted = append(accepted, h)
	}
	return global, accepted
}

func main() {
	// Simula: dois nós offline minerando em paralelo
	d1 := NewDAG()
	d2 := NewDAG()

	parents := []string{}
	for i := 0; i < 5; i++ {
		// nó 1
		tx1 := Tx{ID: fmt.Sprintf("A-%d", i), Fee: 1}
		h1 := powMine(parents, fmt.Sprintf("merkleA-%d", i), 4)
		b1 := MicroBlock{Header: h1, Txs: []Tx{tx1}}
		d1.AddBlock(b1)

		// nó 2
		tx2 := Tx{ID: fmt.Sprintf("B-%d", i), Fee: 1}
		h2 := powMine(parents, fmt.Sprintf("merkleB-%d", i), 4)
		b2 := MicroBlock{Header: h2, Txs: []Tx{tx2}}
		d2.AddBlock(b2)

		// Às vezes cria arestas cruzadas (DAG paralelo)
		if i%2 == 0 {
			parents = []string{h1.Hash, h2.Hash}
		} else {
			parents = []string{h1.Hash}
		}
	}

	merged, order := reconcileMerge([]*DAG{d1, d2})
	fmt.Printf("Blocos no DAG global: %d\n", len(merged.Blocks))
	fmt.Printf("Ordem topológica (aceitos): %d\n", len(order))
	fmt.Printf("Tips: %d\n", len(merged.Tips))
	// Aqui você empacotaria em um macro-bloco e submeteria ao consenso online.
	_ = rand.Int() // apenas para usar rand e evitar warnings
}
