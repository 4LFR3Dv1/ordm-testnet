package validator

type Node struct {
    Name      string
    Validator *Validator
}

func NewNode(name string, initialBalances map[string]int) *Node {
    return &Node{
        Name: name,
        Validator: NewValidator(initialBalances),
    }
}

// Envia bloco para toda a rede
func (n *Node) BroadcastBlock(block *Block, network []*Node) {
    for _, node := range network {
        if node.Name != n.Name {
            node.ReceiveBlock(block)
        }
    }
}

// Recebe bloco e valida
func (n *Node) ReceiveBlock(block *Block) error {
    return n.Validator.AddBlock(block)
}
