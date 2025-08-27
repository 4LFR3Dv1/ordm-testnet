package dag

import (
	"ordm-main/pkg/types"
	"sync"
)

type DAG struct {
	mu     sync.RWMutex
	Blocks map[string]*types.MicroBlock
	Edges  map[string][]string // filho -> pais
	Tips   map[string]bool
}

func New() *DAG {
	return &DAG{
		Blocks: make(map[string]*types.MicroBlock),
		Edges:  make(map[string][]string),
		Tips:   make(map[string]bool),
	}
}

func (d *DAG) AddBlock(b *types.MicroBlock) bool {
	h := b.Header.Hash

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.Blocks[h]; ok {
		return false
	}
	d.Blocks[h] = b
	d.Edges[h] = append([]string{}, b.Header.Parents...)

	// assume-se que pais podem ser desconhecidos (chegam depois). Marca tips.
	d.Tips[h] = true
	for _, p := range b.Header.Parents {
		delete(d.Tips, p)
	}
	return true
}

func (d *DAG) GetTips(max int) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]string, 0, max)
	for h := range d.Tips {
		out = append(out, h)
		if len(out) == max { break }
	}
	return out
}

func (d *DAG) Size() int {
	d.mu.RLock(); defer d.mu.RUnlock()
	return len(d.Blocks)
}

func (d *DAG) SnapshotTips() []string {
	d.mu.RLock(); defer d.mu.RUnlock()
	out := make([]string, 0, len(d.Tips))
	for h := range d.Tips { out = append(out, h) }
	return out
}

func (d *DAG) AllBlockHashes() []string {
	hashes := []string{}
	for h := range d.Blocks {
		hashes = append(hashes, h)
	}
	return hashes
}

func (d *DAG) MissingBlocks(hashes []string) []string {
	missing := []string{}
	for _, h := range hashes {
		if _, ok := d.Blocks[h]; !ok {
			missing = append(missing, h)
		}
	}
	return missing
}

func (d *DAG) GetBlock(hash string) *types.MicroBlock {
	return d.Blocks[hash]
}
