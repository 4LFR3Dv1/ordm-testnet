package state

import "sync"
import "ordm-main/pkg/types"

type Ledger struct {
	mu       sync.Mutex
	Balances map[string]int64
}

func NewLedger(initial map[string]int64) *Ledger {
	return &Ledger{Balances: initial}
}
func (l *Ledger) ApplyTxs(txs []types.Tx) []types.Tx {
	l.mu.Lock()
	defer l.mu.Unlock()

	var applied []types.Tx
	for _, tx := range txs {
		if tx.Amount > 0 && l.Balances[tx.From] >= tx.Amount {
			l.Balances[tx.From] -= tx.Amount
			l.Balances[tx.To] += tx.Amount
			applied = append(applied, tx)
		}
	}
	return applied
}
func (l *Ledger) ApplyTx(from, to string, amount int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if amount <= 0 || l.Balances[from] < amount {
		return false
	}

	l.Balances[from] -= amount
	l.Balances[to] += amount
	return true
}

func (l *Ledger) Reward(miner string, reward int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Balances[miner] += reward
}

func (l *Ledger) Snapshot() map[string]int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	copy := make(map[string]int64, len(l.Balances))
	for k, v := range l.Balances {
		copy[k] = v
	}
	return copy
}
