 # 🚀 ORDM Testnet - Guia de Início Rápido

## 📋 Visão Geral

A **ORDM Testnet** é uma rede de teste pública da ORDM Blockchain que permite aos usuários testar funcionalidades, minerar blocos e interagir com a blockchain antes do lançamento da mainnet.

## 🎯 Características da Testnet

- **Rede Pública**: Qualquer pessoa pode participar
- **Proof of Work**: Algoritmo de consenso PoW
- **P2P Network**: Comunicação peer-to-peer
- **RPC API**: Interface para desenvolvedores
- **Minerador CLI**: Ferramenta de mineração
- **Supply Limitado**: 10 milhões de tokens ORDM

## 🛠️ Pré-requisitos

### Software Necessário
- **Go 1.19+** - [Download](https://golang.org/dl/)
- **Git** - Para clonar o repositório
- **Docker** (opcional) - Para rodar com containers

### Hardware Recomendado
- **CPU**: 2+ cores
- **RAM**: 4GB+
- **Storage**: 10GB+ livre
- **Rede**: Conexão estável com internet

## 🚀 Início Rápido

### 1. Clonar o Repositório
```bash
git clone https://github.com/your-org/ordm-blockchain.git
cd ordm-blockchain
```

### 2. Rodar Node/Minerador Integrado
```bash
# Rodar apenas como node (padrão)
./scripts/run-node.sh

# Rodar apenas como minerador (machineID gerado automaticamente)
./scripts/run-node.sh --mode miner

# Rodar node + minerador simultaneamente
./scripts/run-node.sh --mode both --miner-threads 4

# Rodar com mineração habilitada e nome personalizado
./scripts/run-node.sh --mining --miner-name my-miner
```

### 3. MachineID Automático
Na primeira execução, o sistema gera automaticamente:
- **MachineID**: Identificador único da máquina (criptografado)
- **MinerID**: Derivado do machineID para identificação na rede
- **Arquivo**: `data/testnet/machine_id.json` (persistente)

## 📖 Guias Detalhados

### 🔧 Rodando seu Node da Testnet

#### Opções de Configuração
```bash
./scripts/run-node.sh [opções]

Opções:
  -n, --network NETWORK    Rede (testnet/mainnet) [padrão: testnet]
  -p, --port PORT          Porta HTTP [padrão: 8080]
  --p2p-port PORT          Porta P2P [padrão: 3000]
  --rpc-port PORT          Porta RPC [padrão: 8081]
  -d, --data PATH          Caminho para dados [padrão: ./data/testnet]
  -c, --config FILE        Arquivo de configuração
  -g, --genesis FILE       Arquivo do bloco genesis
  --max-peers NUM          Máximo de peers [padrão: 50]
  --block-time DURATION    Tempo entre blocos (0 = sem mineração automática)
  --difficulty NUM         Dificuldade de mineração [padrão: 4]
  --mode MODE              Modo de operação (node/miner/both) [padrão: node]
  --mining                 Habilitar mineração
  --miner-key KEY          Chave privada do minerador (auto-gerada se não fornecida)
  --miner-threads NUM      Número de threads de mineração [padrão: 1]
  --miner-name NAME        Nome do minerador [padrão: ordm-node]
```

#### Exemplos de Uso
```bash
# Node básico da testnet
./scripts/run-node.sh

# Node com mineração automática (30s entre blocos)
./scripts/run-node.sh --block-time 30s

# Node em portas específicas
./scripts/run-node.sh --port 9090 --p2p-port 4000 --rpc-port 9091

# Node com dificuldade personalizada
./scripts/run-node.sh --difficulty 6

# Apenas minerador (machineID gerado automaticamente)
./scripts/run-node.sh --mode miner --miner-threads 4

# Node + minerador simultaneamente
./scripts/run-node.sh --mode both --miner-name my-miner

# Mineração com nome personalizado
./scripts/run-node.sh --mining --miner-name my-miner --miner-threads 2
```

### ⛏️ Minerando na Testnet

#### MachineID Automático
O sistema gera automaticamente na primeira execução:
- **MachineID**: Identificador único da máquina baseado em hardware
- **MinerID**: Derivado do machineID para identificação na rede
- **Persistência**: Salvo em `data/testnet/machine_id.json`

#### Exemplos de Mineração
```bash
# Mineração básica (machineID gerado automaticamente)
./scripts/run-node.sh --mode miner

# Mineração com múltiplas threads
./scripts/run-node.sh --mode miner --miner-threads 4

# Node + minerador simultaneamente
./scripts/run-node.sh --mode both --miner-name my-miner

# Mineração com nome personalizado
./scripts/run-node.sh --mining --miner-name my-miner --miner-threads 2
```

### 🔌 Enviando Transações via RPC/SDK

#### Usando o SDK Go
```go
package main

import (
    "fmt"
    "log"
    "ordm-main/pkg/sdk"
)

func main() {
    // Criar cliente
    client := sdk.NewORDMClient("http://localhost:8081")
    
    // Enviar transação
    result, err := client.SendTransaction(
        "from_address",
        "to_address", 
        1000, // amount
        1,    // fee
        "data",
        "signature",
    )
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Transação enviada: %v\n", result)
}
```

#### Usando cURL
```bash
# Obter informações da blockchain
curl http://localhost:8081/api/v1/blockchain/info

# Obter transações pendentes
curl http://localhost:8081/api/v1/transactions/pending

# Enviar transação
curl -X POST http://localhost:8081/api/v1/transactions/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": "from_address",
    "to": "to_address",
    "amount": 1000,
    "fee": 1,
    "data": "transaction data",
    "signature": "signature"
  }'
```

## 🐳 Docker Compose

Para rodar múltiplos nodes rapidamente:

```bash
# Subir 3 nodes + 1 minerador
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar todos os serviços
docker-compose down
```

## 📊 Monitoramento

### Endpoints de Status
- `GET /api/v1/blockchain/info` - Informações gerais
- `GET /api/v1/blockchain/status` - Status da rede
- `GET /api/v1/peers` - Lista de peers conectados
- `GET /api/v1/mempool` - Transações pendentes

### Logs
Os logs são exibidos no console. Para salvar em arquivo:
```bash
./scripts/run-node.sh > node.log 2>&1 &
./scripts/run-miner.sh --miner-key abc123 > miner.log 2>&1 &
```

## 🔧 Configuração Avançada

### Arquivo de Configuração
Edite `config/testnet.json` para personalizar:
- Parâmetros de consenso
- Configurações P2P
- Limites de rede
- Configurações de segurança

### Bloco Genesis
O arquivo `genesis/testnet.json` define:
- Supply inicial
- Endereços premine
- Configurações iniciais

## 🚨 Solução de Problemas

### Node não inicia
```bash
# Verificar se Go está instalado
go version

# Verificar se as portas estão livres
netstat -an | grep :8080
netstat -an | grep :3000

# Verificar permissões
ls -la scripts/run-node.sh
```

### Minerador não conecta
```bash
# Verificar se o node está rodando
curl http://localhost:8081/api/v1/blockchain/info

# Verificar conectividade
nc -zv localhost 8081

# Verificar logs do node
tail -f node.log
```

### Problemas de Rede P2P
```bash
# Verificar firewall
sudo ufw status

# Verificar portas P2P
netstat -an | grep :3000

# Adicionar peers manualmente (se necessário)
# Editar config/testnet.json
```

## 📞 Suporte

### Comunidade
- **Discord**: [Link do Discord]
- **Telegram**: [Link do Telegram]
- **GitHub Issues**: [Link do GitHub]

### Recursos
- **Documentação**: [Link da docs]
- **API Reference**: [Link da API]
- **SDK Examples**: [Link dos exemplos]

## 📄 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

---

**🎉 Parabéns! Você está pronto para participar da ORDM Testnet!**
