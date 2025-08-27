# 📊 Cálculo de Blocos Minerados e Recompensas

## 🌱 **Bloco Genesis**

O bloco genesis é o primeiro bloco da blockchain com as seguintes características:
- **Hash**: `0000000000000000000000000000000000000000000000000000000000000000`
- **Número**: `0`
- **Recompensa**: `0` (não há mineração no genesis)
- **Timestamp**: Momento de criação da blockchain

## ⛏️ **Mineração de Blocos**

### **Número de Blocos**
- **Fórmula**: `blockNumber = blockHeight + 1`
- **Exemplo**: Se a altura atual é 100, o próximo bloco será #101
- **Contagem**: Começa em 1 (após o genesis)

### **Hash do Bloco**
```go
data := fmt.Sprintf("%s|%d|%d|%s|%d|%d|%d|%d",
    block.ParentHash,    // Hash do bloco pai
    block.Number,        // Número do bloco
    block.Timestamp,     // Timestamp
    block.MinerID,       // ID do minerador
    block.Difficulty,    // Dificuldade PoW
    block.Nonce,         // Nonce para PoW
    block.TotalFees,     // Taxas totais
    block.BurnedTokens,  // Tokens queimados
)
hash := sha256.Sum256([]byte(data))
```

## 💰 **Cálculo de Recompensas**

### **Recompensa Base (Halving)**
```go
// Parâmetros
initialReward := 50        // 50 tokens por bloco inicial
halvingInterval := 210000  // Halving a cada 210k blocos
maxSupply := 21000000      // 21M tokens máximo

// Fórmula de halving
epoch := blockNumber / halvingInterval
reward := initialReward
for i := 0; i < epoch; i++ {
    reward = reward / 2
    if reward <= 0 {
        reward = 1 // Mínimo de 1 token
        break
    }
}
```

### **Tabela de Halving**
| Época | Blocos | Recompensa | Acumulado |
|-------|--------|------------|-----------|
| 0     | 1-210,000 | 50 tokens | 10,500,000 |
| 1     | 210,001-420,000 | 25 tokens | 15,750,000 |
| 2     | 420,001-630,000 | 12.5 tokens | 18,375,000 |
| 3     | 630,001-840,000 | 6.25 tokens | 19,687,500 |
| 4     | 840,001-1,050,000 | 3.125 tokens | 20,343,750 |
| ...   | ... | ... | ... |
| ∞     | >21M | 0 tokens | 21,000,000 |

### **Taxas de Transação**
- **Cálculo**: Soma de todas as taxas das transações no bloco
- **Exemplo**: 5 transações com taxas de 1, 2, 1, 3, 1 = 8 tokens

### **Queima de Tokens (Token Burning)**
```go
burnRate := 0.1 // 10% das taxas são queimadas
burnedTokens := int64(float64(totalFees) * burnRate)
```

### **Recompensa Total**
```go
totalReward := miningReward + totalFees - burnedTokens
```

## 📈 **Supply Total**

### **Fórmula de Atualização**
```go
// Após cada bloco minerado
totalSupply += miningReward
totalSupply -= burnedTokens
```

### **Limite Máximo**
- **Supply Máximo**: 21,000,000 tokens
- **Quando atingido**: Mineração para, apenas taxas continuam
- **Inflação**: Para quando o supply máximo é atingido

## 🔍 **Exemplo Prático**

### **Bloco #1**
```
Minerador: "node_001"
Dificuldade: 2
Transações: 3 (taxas: 1, 2, 1)
Recompensa base: 50 tokens
Taxas totais: 4 tokens
Tokens queimados: 0.4 tokens (10% de 4)
Recompensa total: 50 + 4 - 0.4 = 53.6 tokens
Supply total: 53.6 tokens
```

### **Bloco #210,001 (Primeiro Halving)**
```
Minerador: "node_002"
Dificuldade: 3
Transações: 5 (taxas: 2, 1, 3, 1, 2)
Recompensa base: 25 tokens (halving aplicado)
Taxas totais: 9 tokens
Tokens queimados: 0.9 tokens
Recompensa total: 25 + 9 - 0.9 = 33.1 tokens
Supply total: ~10,500,000 + 33.1 tokens
```

## 📊 **Estatísticas de Mineração**

### **Por Minerador**
```go
minerStats := make(map[string]int64)
for _, block := range blocks {
    if block.Number > 0 { // Excluir genesis
        minerStats[block.MinerID] += block.Reward
    }
}
```

### **Taxa de Inflação**
```go
blocksPerYear := 365 * 24 * 60 * 60 / 10 // 10s por bloco
annualRewards := currentReward * blocksPerYear
inflationRate := (annualRewards / totalSupply) * 100
```

## 🎯 **Validação de Integridade**

### **Verificações**
1. **Hash do Bloco**: Deve corresponder aos dados
2. **Bloco Pai**: Deve existir na blockchain
3. **Número Sequencial**: Cada bloco deve ter número = pai + 1
4. **Supply Total**: Não pode exceder 21M tokens
5. **Recompensa**: Deve seguir a fórmula de halving

### **Exemplo de Validação**
```go
// Verificar bloco #100
block := GetBlockByNumber(100)
parentBlock := GetBlockByNumber(99)

// Validações
assert(block.Number == parentBlock.Number + 1)
assert(block.ParentHash == parentBlock.Hash)
assert(block.Reward == calculateMiningReward(100))
assert(totalSupply <= maxSupply)
```

## 🔧 **Configurações Atuais**

| Parâmetro | Valor | Descrição |
|-----------|-------|-----------|
| Recompensa Inicial | 50 tokens | Por bloco no início |
| Intervalo Halving | 210,000 blocos | Como Bitcoin |
| Supply Máximo | 21,000,000 tokens | Limite total |
| Taxa de Queima | 10% | Das taxas de transação |
| Dificuldade Inicial | 2 | Zeros no início do hash |
| Tempo por Bloco | ~10 segundos | Estimativa |

## 📋 **Resumo**

1. **Blocos**: Numerados sequencialmente a partir de 1
2. **Recompensas**: 50 tokens inicial, halving a cada 210k blocos
3. **Taxas**: Adicionadas à recompensa do minerador
4. **Queima**: 10% das taxas são queimadas (deflação)
5. **Supply**: Máximo de 21M tokens
6. **Validação**: Integridade verificada a cada bloco

Este sistema garante uma economia sustentável com inflação controlada e deflação através da queima de tokens.
