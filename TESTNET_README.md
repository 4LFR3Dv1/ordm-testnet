# 🌐 ORDM Testnet - Guia Completo

## 📋 Visão Geral

A **ORDM Testnet** é a rede de testes pública da blockchain 2-layer ORDM. Esta rede permite que desenvolvedores testem funcionalidades, validem contratos e participem da mineração antes do lançamento da mainnet.

## 🚀 Características da Testnet

### **✅ Funcionalidades Disponíveis**
- **Mineração PoW**: Teste de mineração com dificuldade reduzida
- **Staking PoS**: Sistema de validação com stake mínimo baixo
- **Faucet**: Distribuição gratuita de tokens de teste
- **Explorer**: Interface pública para visualizar a blockchain
- **API REST**: Endpoints para integração
- **Seed Nodes**: Nós de entrada públicos e estáveis

### **💰 Tokenomics da Testnet**
- **Supply**: Ilimitado (apenas para testes)
- **Recompensa**: 50 tokens por bloco
- **Stake Mínimo**: 100 tokens (reduzido para testes)
- **Faucet**: 50 tokens por requisição (1x por hora por IP)

## 🔧 Como Participar

### **1. Pré-requisitos**

#### **Sistema Operacional**
- **Linux**: Ubuntu 20.04+ (recomendado)
- **macOS**: 10.15+ 
- **Windows**: 10+ (com WSL recomendado)

#### **Software Necessário**
```bash
# Go 1.25+
go version

# Git
git --version

# Make (opcional)
make --version
```

### **2. Instalação**

#### **Opção 1: Binário Pré-compilado**
```bash
# Baixar binários
wget https://github.com/seu-usuario/ordm-main/releases/download/v1.0.0-testnet/ordm-node-linux
wget https://github.com/seu-usuario/ordm-main/releases/download/v1.0.0-testnet/ordm-explorer-linux

# Tornar executáveis
chmod +x ordm-node-linux ordm-explorer-linux
```

#### **Opção 2: Compilar do Código Fonte**
```bash
# Clonar repositório
git clone https://github.com/seu-usuario/ordm-main.git
cd ordm-main

# Baixar dependências
go mod tidy

# Compilar
go build -o ordm-node ./cmd/node
go build -o ordm-explorer ./cmd/explorer
```

### **3. Configuração**

#### **Criar Arquivo de Configuração**
```bash
mkdir -p ~/.ordm-testnet
```

```json
# ~/.ordm-testnet/config.json
{
    "network": "testnet",
    "node": {
        "port": 3001,
        "api_port": 8080,
        "max_peers": 50,
        "heartbeat": 30
    },
    "seed_nodes": [
        "/ip4/18.188.123.45/tcp/3001/p2p/QmSeedNode1",
        "/ip4/52.15.67.89/tcp/3001/p2p/QmSeedNode2",
        "/ip4/34.201.234.56/tcp/3001/p2p/QmSeedNode3"
    ],
    "faucet": {
        "enabled": true,
        "max_amount": 50,
        "daily_limit": 100
    },
    "mining": {
        "enabled": true,
        "difficulty": 2,
        "reward": 50
    }
}
```

### **4. Executar o Node**

#### **Iniciar Node Básico**
```bash
./ordm-node --config ~/.ordm-testnet/config.json
```

#### **Iniciar com Interface Web**
```bash
./ordm-node --config ~/.ordm-testnet/config.json --web --port 3000
```

#### **Iniciar Explorer**
```bash
./ordm-explorer --port 8080
```

## 💰 Usando o Faucet

### **Obter Tokens de Teste**

#### **Via API REST**
```bash
curl -X POST https://testnet.ordm.com/api/testnet/faucet \
  -H "Content-Type: application/json" \
  -d '{
    "address": "sua_wallet_address_aqui",
    "amount": 50
  }'
```

#### **Via Interface Web**
1. Acesse: `https://testnet.ordm.com/faucet`
2. Digite seu endereço de wallet
3. Clique em "Request Tokens"
4. Aguarde a confirmação

#### **Limites do Faucet**
- **Por Requisição**: 50 tokens
- **Por Hora**: 1 requisição por IP
- **Por Dia**: 100 tokens por IP
- **Validação**: Endereço deve ser válido (26-42 caracteres hex)

### **Verificar Saldo**
```bash
curl https://testnet.ordm.com/api/testnet/balances/sua_wallet_address_aqui
```

## ⛏️ Mineração na Testnet

### **Iniciar Mineração**
```bash
# Via linha de comando
./ordm-node --mining --difficulty 2

# Via interface web
# Acesse http://localhost:3000 e clique em "Start Mining"
```

### **Configurações de Mineração**
- **Dificuldade**: 2 (reduzida para testes)
- **Recompensa**: 50 tokens por bloco
- **Tempo por Bloco**: ~10 segundos
- **Stake Mínimo**: 100 tokens

### **Monitorar Mineração**
```bash
# Ver logs em tempo real
tail -f ~/.ordm-testnet/logs/mining.log

# Ver estatísticas via API
curl http://localhost:8080/api/testnet/stats
```

## 🏆 Sistema de Staking

### **Fazer Stake**
```bash
# Via API
curl -X POST http://localhost:8080/api/testnet/stake \
  -H "Content-Type: application/json" \
  -d '{
    "address": "sua_wallet_address",
    "amount": 100
  }'
```

### **Benefícios do Staking**
- **APY**: 5% base + 2% bônus para validators
- **Validação**: Participar na validação de blocos
- **Governança**: Votar em propostas da rede
- **Recompensas**: Tokens adicionais por validação

## 🌐 Explorer da Testnet

### **Acessar Explorer**
- **URL**: `https://testnet.ordm.com`
- **Funcionalidades**:
  - Visualizar blocos em tempo real
  - Consultar transações
  - Ver saldos de wallets
  - Estatísticas da rede
  - Histórico de staking

### **Endpoints Públicos**
```bash
# Status da rede
curl https://testnet.ordm.com/api/testnet/status

# Lista de blocos
curl https://testnet.ordm.com/api/testnet/blocks

# Transações recentes
curl https://testnet.ordm.com/api/testnet/transactions

# Seed nodes
curl https://testnet.ordm.com/api/testnet/seed-nodes
```

## 🔗 Conectar à Rede

### **Seed Nodes Disponíveis**
```
/ip4/18.188.123.45/tcp/3001/p2p/QmSeedNode1  (US East)
/ip4/52.15.67.89/tcp/3001/p2p/QmSeedNode2    (US West)
/ip4/34.201.234.56/tcp/3001/p2p/QmSeedNode3  (EU West)
```

### **Configurar Conexão**
```bash
# Adicionar seed nodes manualmente
./ordm-node --peers "/ip4/18.188.123.45/tcp/3001/p2p/QmSeedNode1"

# Ou usar configuração automática
./ordm-node --auto-discover
```

## 📊 Monitoramento

### **Logs do Sistema**
```bash
# Logs do node
tail -f ~/.ordm-testnet/logs/node.log

# Logs de mineração
tail -f ~/.ordm-testnet/logs/mining.log

# Logs de rede
tail -f ~/.ordm-testnet/logs/network.log
```

### **Métricas via API**
```bash
# Status geral
curl http://localhost:8080/api/testnet/status

# Estatísticas do faucet
curl http://localhost:8080/api/testnet/faucet/stats

# Informações da rede
curl http://localhost:8080/api/testnet/network
```

## 🛠️ Desenvolvimento

### **API para Desenvolvedores**

#### **Endpoints Principais**
```bash
# Saúde da rede
GET /api/testnet/status

# Faucet
POST /api/testnet/faucet
GET /api/testnet/faucet/stats
GET /api/testnet/faucet/history

# Seed nodes
GET /api/testnet/seed-nodes
GET /api/testnet/peers

# Rede
GET /api/testnet/network
```

#### **Exemplo de Integração**
```javascript
// JavaScript/Node.js
const axios = require('axios');

// Obter tokens do faucet
async function getTestTokens(address) {
    try {
        const response = await axios.post('https://testnet.ordm.com/api/testnet/faucet', {
            address: address,
            amount: 50
        });
        return response.data;
    } catch (error) {
        console.error('Erro no faucet:', error.response.data);
    }
}

// Verificar status da rede
async function getNetworkStatus() {
    const response = await axios.get('https://testnet.ordm.com/api/testnet/status');
    return response.data;
}
```

### **SDK para Desenvolvedores**
```bash
# Instalar SDK (quando disponível)
go get github.com/seu-usuario/ordm-sdk

# Exemplo de uso
package main

import (
    "github.com/seu-usuario/ordm-sdk/client"
)

func main() {
    // Conectar à testnet
    client := client.NewTestnetClient()
    
    // Obter saldo
    balance, err := client.GetBalance("sua_wallet_address")
    
    // Enviar transação
    tx, err := client.SendTransaction("from", "to", 100)
}
```

## 🔐 Segurança

### **Boas Práticas**
- **Backup**: Faça backup regular das suas wallets
- **Testes**: Use apenas tokens de teste
- **Monitoramento**: Monitore logs e métricas
- **Atualizações**: Mantenha o software atualizado

### **Rate Limiting**
- **API**: 100 requisições/minuto por IP
- **Faucet**: 1 requisição/hora por IP
- **P2P**: 1000 mensagens/minuto por peer

## 🐛 Troubleshooting

### **Problemas Comuns**

#### **Node não conecta**
```bash
# Verificar conectividade
ping 18.188.123.45

# Verificar portas
telnet 18.188.123.45 3001

# Verificar logs
tail -f ~/.ordm-testnet/logs/node.log
```

#### **Faucet não funciona**
```bash
# Verificar rate limit
curl https://testnet.ordm.com/api/testnet/faucet/stats

# Verificar endereço
# Deve ter entre 26-42 caracteres hex
```

#### **Mineração não lucrativa**
```bash
# Verificar dificuldade
curl http://localhost:8080/api/testnet/stats

# Verificar hash rate
# Deve ser > 1 H/s para ser lucrativo
```

### **Logs de Erro**
```bash
# Erro comum: "connection refused"
# Solução: Verificar se seed nodes estão online

# Erro comum: "insufficient balance"
# Solução: Usar faucet para obter tokens

# Erro comum: "rate limit exceeded"
# Solução: Aguardar 1 hora entre requisições
```

## 📞 Suporte

### **Recursos de Ajuda**
- **Documentação**: Este README
- **Explorer**: `https://testnet.ordm.com`
- **API Docs**: `https://testnet.ordm.com/api/docs`
- **Logs**: Sistema de logs detalhado

### **Canais de Comunidade**
- **Discord**: `https://discord.gg/ordm-testnet`
- **Telegram**: `@ordm_testnet`
- **GitHub**: Issues no repositório
- **Email**: `testnet@ordm.com`

### **Reportar Bugs**
```bash
# Incluir informações:
# - Versão do software
# - Sistema operacional
# - Logs de erro
# - Passos para reproduzir
# - Comportamento esperado vs atual
```

## 🎯 Próximos Passos

### **Roadmap da Testnet**
- [ ] **Fase 1**: Rede básica (✅ Concluído)
- [ ] **Fase 2**: Faucet e explorer (✅ Concluído)
- [ ] **Fase 3**: Smart contracts básicos
- [ ] **Fase 4**: DeFi protocols
- [ ] **Fase 5**: Governança descentralizada

### **Migração para Mainnet**
- **Data Estimada**: Q2 2024
- **Processo**: Snapshot da testnet
- **Tokens**: 1:1 para mainnet
- **Stake**: Migração automática

---

## 🎉 Conclusão

A **ORDM Testnet** oferece um ambiente completo para testar e desenvolver na blockchain 2-layer. Com faucet, explorer público e seed nodes estáveis, você pode:

✅ **Testar funcionalidades** antes da mainnet  
✅ **Desenvolver aplicações** com tokens gratuitos  
✅ **Participar da mineração** com dificuldade reduzida  
✅ **Validar contratos** em ambiente seguro  
✅ **Contribuir para a rede** como validator  

**🚀 Junte-se à comunidade e ajude a construir o futuro da blockchain!**
