# 🚀 Teste de Múltiplos Nodes - Blockchain 2-Layer

## 📋 Visão Geral

Este documento explica como testar a comunicação entre diferentes nodes e a validação 2-layer com PoS (Proof of Stake) no seu projeto blockchain.

## 🏗️ Arquitetura do Sistema

### Camada 1: PoW (Proof of Work)
- **Mineração**: Algoritmo SHA-256 com dificuldade configurável
- **DAG**: Estrutura de dados para blocos
- **Rede P2P**: Comunicação entre nodes

### Camada 2: PoS (Proof of Stake)
- **Validadores**: Nodes que validam transações
- **Staking**: Sistema de apostas para validadores
- **Consenso**: Validação de blocos minerados

## 🧪 Como Testar

### 1. Teste Manual - Node Individual

```bash
# Executar um node isolado
go run ./cmd/node/main.go <NODE_ID> <PORTA> -

# Exemplo:
go run ./cmd/node/main.go node1 8080 -
```

### 2. Teste Automático - Múltiplos Nodes

```bash
# Executar o script de teste
./test_multi_nodes.sh
```

Este script irá:
- ✅ Iniciar 3 nodes em portas diferentes
- ✅ Configurar comunicação P2P entre eles
- ✅ Monitorar atividade por 60 segundos
- ✅ Parar todos os nodes automaticamente

### 3. Teste Manual - Múltiplos Nodes

```bash
# Terminal 1 - Node 1
go run ./cmd/node/main.go node1 8080 8081,8082

# Terminal 2 - Node 2  
go run ./cmd/node/main.go node2 8081 8080,8082

# Terminal 3 - Node 3
go run ./cmd/node/main.go node3 8082 8080,8081
```

## 🔍 O que Observar

### Durante a Mineração PoW:
```
[MINER] Iniciando mineração com dificuldade 1
[MINER] Tentativa 10000, nonce: 9999
[MINER] Hash encontrado! Nonce: 12345, Hash: 0abc123... (zeros: 1)
```

### Durante a Comunicação P2P:
```
[node1] Servidor iniciado na porta 8080
[node1] Sincronizando com peers: [8081 8082]
[net] recebido bloco 0abc123 de node2 (pais: []) | DAG=1
```

### Durante a Validação PoS:
```
[node1] Iniciando validação PoS...
[node1] validator1 validou bloco com sucesso
[node1] validator2 validou bloco com sucesso
```

### Durante o Broadcast:
```
[node1] Bloco adicionado ao DAG: true
[node1] Bloco minerado com sucesso!
[node1] Hash: 0abc123...
[node1] DAG size: 1
[node1] Ledger snapshot: map[Alice:90 Bob:55 node1:10]
```

## ⚙️ Configurações

### Dificuldade de Mineração
- **Dificuldade 1**: Muito rápida (para testes)
- **Dificuldade 2**: Rápida (para demonstração)
- **Dificuldade 4+**: Realista (para produção)

### Portas Padrão
- **Node 1**: 8080
- **Node 2**: 8081  
- **Node 3**: 8082

### Saldos Iniciais
- **Alice**: 100 tokens
- **Bob**: 50 tokens
- **Mineradores**: 0 tokens (ganham recompensas)

## 🐛 Troubleshooting

### Node não inicia:
```bash
# Verificar se a porta está livre
lsof -i :8080

# Matar processos anteriores
pkill -f "go run"
```

### Nodes não se comunicam:
```bash
# Verificar se as portas estão corretas
netstat -an | grep 808

# Verificar logs de rede
# Procurar por "[net] listening on" e "[net] recebido bloco"
```

### Mineração muito lenta:
```bash
# Reduzir dificuldade no código
const difficulty = 1  # Em vez de 2 ou 4
```

## 📊 Métricas de Teste

### Esperado após 60 segundos:
- ✅ 3 nodes ativos
- ✅ Comunicação P2P funcionando
- ✅ Blocos sendo minerados
- ✅ Validação PoS ativa
- ✅ DAG sincronizado entre nodes
- ✅ Ledger atualizado

### Logs de Sucesso:
```
[node1] Iniciando mineração...
[node1] Transações pendentes: 2
[node1] Transações aplicadas: 2
[node1] Tips obtidos: []
[node1] Merkle root: abc123...
[node1] Iniciando mineração com dificuldade 1...
[MINER] Hash encontrado! Nonce: 12345, Hash: 0abc123...
[node1] Bloco minerado com sucesso!
[node1] validator1 validou bloco com sucesso
[node1] validator2 validou bloco com sucesso
[net] recebido bloco 0abc123 de node1 (pais: []) | DAG=1
```

## 🎯 Próximos Passos

1. **Testar com mais nodes** (4-5 nodes)
2. **Implementar fork resolution** (quando há conflitos)
3. **Adicionar métricas de performance**
4. **Implementar persistência de dados**
5. **Adicionar interface web** para monitoramento

---

**🎉 Seu blockchain 2-layer está pronto para testes!**






