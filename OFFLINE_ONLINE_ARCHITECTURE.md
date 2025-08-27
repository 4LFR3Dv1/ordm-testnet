# 🏗️ Arquitetura Offline → Online

## 📋 Visão Geral

Esta arquitetura implementa um sistema blockchain híbrido onde:

- **🏭 Mineradores Offline**: Executam PoW localmente
- **🌐 Testnet Online**: Valida blocos usando PoS
- **🔄 Sincronização**: Conecta mineradores offline à rede online

## 🏭 Minerador Offline

### **Características:**
- ✅ **PoW Independente**: Minera blocos sem depender da rede online
- ✅ **Persistência Local**: Armazena blockchain localmente
- ✅ **Autenticação Criptográfica**: Identidade única com Ed25519
- ✅ **Dashboard Web**: Interface para controle e monitoramento
- ✅ **Sincronização Automática**: Envia blocos para testnet periodicamente

### **Executável:**
```bash
./ordm-offline-miner
```

### **Endpoints:**
- `GET /api/status` - Status do minerador
- `POST /api/start-mining` - Iniciar mineração contínua
- `POST /api/stop-mining` - Parar mineração
- `POST /api/mine-block` - Minerar bloco único
- `GET /api/blocks` - Listar blocos minerados
- `POST /api/sync` - Sincronizar com testnet
- `GET /api/stats` - Estatísticas detalhadas

### **Dashboard:**
- **URL**: `http://localhost:8081`
- **Funcionalidades**:
  - Controle de mineração
  - Estatísticas em tempo real
  - Logs de sincronização
  - Status da conexão

## 🌐 Testnet Online

### **Características:**
- ✅ **Apenas Validação PoS**: Não minera blocos
- ✅ **Recebe Blocos**: Valida blocos de mineradores offline
- ✅ **Explorer Público**: Interface para visualizar blockchain
- ✅ **Sistema de Stake**: Validadores com stake mínimo
- ✅ **Reputação**: Sistema de reputação para validadores

### **Executáveis:**
```bash
./blockchain-gui-mac      # Testnet principal
./blockchain-explorer     # Explorer público
./blockchain-backend      # Backend da rede
```

### **Endpoints de Sincronização:**
- `POST /api/sync/block` - Receber bloco para validação
- `GET /api/sync/status` - Status da sincronização
- `GET /api/validators` - Lista de validadores
- `GET /api/validators/stats` - Estatísticas dos validadores

### **Validação PoS:**
- **Stake Mínimo**: 1000 tokens
- **Reputação**: 0-100 pontos
- **Slashing**: Penalização por comportamento malicioso
- **Ativação**: Validadores devem ter stake suficiente

## 🔄 Sincronização

### **Fluxo de Sincronização:**

1. **🏭 Minerador Offline**
   - Minera bloco com PoW
   - Assina digitalmente o bloco
   - Adiciona à blockchain local

2. **📡 Envio para Testnet**
   - Serializa bloco em JSON
   - Envia via HTTP POST
   - Inclui assinatura e metadados

3. **🌐 Validação Online**
   - Verifica assinatura do minerador
   - Valida PoW do bloco
   - Verifica transações
   - Aplica regras PoS

4. **✅ Confirmação**
   - Adiciona bloco à blockchain online
   - Atualiza ledger global
   - Retorna status de aceitação

### **Segurança:**
- **Assinatura Digital**: Ed25519 para autenticação
- **Verificação PoW**: Valida prova de trabalho
- **Stake Validation**: Verifica stake do minerador
- **Retry Logic**: Reenvio automático em caso de falha

## 🚀 Como Usar

### **1. Iniciar Testnet Online:**
```bash
# Terminal 1 - Testnet
./blockchain-gui-mac

# Terminal 2 - Explorer
./blockchain-explorer

# Terminal 3 - Backend
./blockchain-backend
```

### **2. Iniciar Minerador Offline:**
```bash
# Terminal 4 - Minerador
./ordm-offline-miner
```

### **3. Testar Integração:**
```bash
# Executar teste completo
./scripts/test_offline_online_integration.sh
```

### **4. Acessar Interfaces:**
- **Minerador**: http://localhost:8081
- **Testnet**: http://localhost:3000
- **Explorer**: http://localhost:8080

## 📊 Monitoramento

### **Minerador Offline:**
- Hash rate em tempo real
- Blocos minerados
- Status de sincronização
- Uptime e estatísticas

### **Testnet Online:**
- Validadores ativos
- Total de stake
- Blocos validados
- Reputação dos validadores

## 🔧 Configuração

### **Minerador Offline:**
```json
{
  "sync_interval": "30s",
  "retry_attempts": 3,
  "retry_delay": "5s",
  "testnet_url": "http://localhost:3000"
}
```

### **Testnet Online:**
```json
{
  "min_stake": 1000,
  "validation_timeout": "30s",
  "max_block_size": 1000000
}
```

## 🎯 Benefícios

### **Para Mineradores:**
- ✅ **Independência**: Mineração sem depender da rede online
- ✅ **Eficiência**: Trabalho local sem latência de rede
- ✅ **Controle**: Dashboard completo para monitoramento
- ✅ **Segurança**: Autenticação criptográfica

### **Para Rede:**
- ✅ **Escalabilidade**: Qualquer número de mineradores
- ✅ **Segurança**: Validação PoS robusta
- ✅ **Transparência**: Explorer público
- ✅ **Consistência**: Ledger global atualizado

## 🔮 Próximos Passos

### **Fase 4: Autenticação de Mineradores**
- [ ] Sistema de reputação avançado
- [ ] Delegation de stake
- [ ] Slashing automático

### **Fase 5: Seed Nodes Funcionais**
- [ ] Seed nodes públicos
- [ ] Protocolo gossip
- [ ] Heartbeat e monitoramento

### **Fase 6: Deploy Público**
- [ ] Configuração para produção
- [ ] SSL/TLS
- [ ] Load balancing
- [ ] Monitoramento avançado

---

**🎉 Arquitetura Offline → Online implementada com sucesso!**
