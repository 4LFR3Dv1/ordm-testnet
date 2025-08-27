package validator

import "fmt"

type Validator struct {
    Blocks    map[string]*Block
    DAG       map[string][]*Block
    MainChain []*Block
    State     *State
}

// Cria um validador vazio
func NewValidator(initialBalances map[string]int) *Validator {
    return &Validator{
        Blocks: make(map[string]*Block),
        DAG:    make(map[string][]*Block),
        MainChain: []*Block{},
        State: &State{Balances: initialBalances},
    }
}

// Adiciona bloco com validação
func (v *Validator) AddBlock(block *Block) error {
    if _, exists := v.Blocks[block.ID]; exists {
        return fmt.Errorf("bloco %s já existe", block.ID)
    }

    // Validar todas as transações
    for _, tx := range block.Transactions {
        balance := v.State.Balances[tx.From]
        if balance < tx.Amount {
            return fmt.Errorf("saldo insuficiente para %s enviar %d", tx.From, tx.Amount)
        }
    }

    // Adiciona no DAG e blocks
    v.Blocks[block.ID] = block
    v.DAG[block.ParentID] = append(v.DAG[block.ParentID], block)

    // Reorganiza main chain automaticamente
    v.reorganizeMainChain()
    return nil
}

// Escolhe a cadeia mais longa como main chain
func (v *Validator) reorganizeMainChain() {
    var longest []*Block
    var dfs func(b *Block, chain []*Block)
    dfs = func(b *Block, chain []*Block) {
        chain = append(chain, b)
        children := v.DAG[b.ID]
        if len(children) == 0 {
            if len(chain) > len(longest) {
                longest = append([]*Block{}, chain...)
            }
        }
        for _, c := range children {
            dfs(c, chain)
        }
    }

    // encontrar blocos genesis (ParentID vazio)
    for _, b := range v.DAG[""] {
        dfs(b, []*Block{})
    }

    v.MainChain = longest
    v.State = RebuildState(v.MainChain)
}

// Retorna o estado atual da main chain
func (v *Validator) GetMainChain() []*Block {
    return v.MainChain
}
