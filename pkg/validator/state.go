package validator

type State struct {
    Balances map[string]int
}

// RebuildState recalcula os saldos a partir de uma cadeia de blocos
func RebuildState(chain []*Block) *State {
    state := &State{Balances: make(map[string]int)}
    for _, block := range chain {
        for _, tx := range block.Transactions {
            if state.Balances[tx.From] < tx.Amount {
                // Saldo insuficiente, ignorar transação
                continue
            }
            state.Balances[tx.From] -= tx.Amount
            state.Balances[tx.To] += tx.Amount
        }
    }
    return state
}
