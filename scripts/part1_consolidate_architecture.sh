#!/bin/bash

# 🎯 Script para PARTE 1: Consolidação Arquitetural
# Baseado no PLANO_ATUALIZACOES.md

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

log "🔄 Iniciando PARTE 1: Consolidação Arquitetural"

# 1.1.1 Remover documentações conflitantes
log "1.1.1 - Removendo documentações conflitantes..."

if [ -f "REAL_ARCHITECTURE.md" ]; then
    mv REAL_ARCHITECTURE.md REAL_ARCHITECTURE.md.backup
    log "✅ REAL_ARCHITECTURE.md movido para backup"
fi

if [ -f "NEW_ARCHITECTURE.md" ]; then
    mv NEW_ARCHITECTURE.md NEW_ARCHITECTURE.md.backup
    log "✅ NEW_ARCHITECTURE.md movido para backup"
fi

# 1.1.2 Criar documentação de decisões
log "1.1.2 - Criando DECISIONS.md..."
cat > DECISIONS.md << 'EOF'
# 📋 Decisões Arquiteturais ORDM

## Histórico de Decisões Técnicas

### 2024-01-XX: Consolidação Arquitetural
- **Problema**: Múltiplas arquiteturas documentadas causavam confusão
- **Decisão**: Manter apenas OFFLINE_ONLINE_ARCHITECTURE.md como base
- **Justificativa**: Eliminar inconsistências e criar fonte única de verdade
- **Impacto**: Desenvolvedores terão referência clara

### 2024-01-XX: Separação de Storage
- **Problema**: Dados perdidos em deploys do Render
- **Decisão**: Implementar storage persistente em /opt/render/data
- **Justificativa**: Garantir persistência em ambiente de produção
- **Impacto**: Dados sobrevivem a reinicializações

### 2024-01-XX: Segurança 2FA
- **Problema**: PIN de 10 segundos muito curto
- **Decisão**: Aumentar para 60 segundos
- **Justificativa**: Melhor experiência do usuário
- **Impacto**: Menos falhas de login legítimas
EOF

# 1.1.3 Documentar dependências
log "1.1.3 - Criando DEPENDENCIES.md..."
cat > DEPENDENCIES.md << 'EOF'
# 🔗 Dependências entre Componentes ORDM

## Dependências Principais

### Camada Offline → Online
- **Minerador Offline** → **Seed Nodes** → **Validadores** → **Ledger Global**
- **Dependência**: Sincronização assíncrona de blocos

### Autenticação
- **Wallet Manager** → **2FA System** → **Session Manager**
- **Dependência**: Autenticação sequencial

### Storage
- **Offline Storage** → **Crypto Module** → **Keystore**
- **Dependência**: Criptografia de dados

## Dependências de Build

### Core Dependencies
- `libp2p` - Rede P2P
- `badger` - Database
- `crypto` - Criptografia
- `auth` - Autenticação

### Optional Dependencies
- `monitoring` - Métricas
- `api` - REST endpoints
- `explorer` - Interface web
EOF

# 1.2.1 Criar diagrama de fluxo
log "1.2.1 - Criando diagrama de fluxo..."
cat > FLOW_DIAGRAM.md << 'EOF'
# 🔄 Diagrama de Fluxo ORDM

## Fluxo Principal
```
Minerador Offline (PoW) → Seed Nodes (P2P) → Validadores (PoS) → Ledger Global
```

## Fluxo Detalhado
1. **Mineração**: Minerador resolve PoW localmente
2. **Assinatura**: Bloco é assinado digitalmente
3. **Envio**: Pacote de blocos enviado para seed nodes
4. **Validação**: Validadores verificam PoW e transações
5. **Consenso**: Votação PoS para aceitar/rejeitar
6. **Confirmação**: Bloco adicionado ao ledger global
7. **Recompensa**: Minerador e validadores recebem tokens

## Interfaces por Papel
- **Minerador**: Apenas mineração e envio
- **Validador**: Stake, validação, recompensas
- **Usuário**: Transações, explorer, dashboard
EOF

# 1.2.2 Documentar APIs
log "1.2.2 - Criando API_CONTRACTS.md..."
cat > API_CONTRACTS.md << 'EOF'
# 📡 Contratos de API ORDM

## Endpoints de Sincronização

### POST /api/sync/block
```json
{
  "blocks": [
    {
      "hash": "block_hash",
      "parent_hash": "parent_hash",
      "number": 1234,
      "miner_id": "miner_address",
      "transactions": [],
      "pow_proof": "proof_data",
      "signature": "miner_signature"
    }
  ],
  "miner_id": "miner_address",
  "batch_id": "unique_batch_id",
  "timestamp": 1640995200
}
```

### GET /api/sync/status
```json
{
  "status": "syncing",
  "last_sync": "2024-01-01T00:00:00Z",
  "pending_blocks": 5,
  "synced_blocks": 1234
}
```

## Endpoints de Validação

### POST /api/validator/vote
```json
{
  "block_hash": "block_hash",
  "validator_id": "validator_address",
  "vote": true,
  "stake_amount": 1000,
  "signature": "validator_signature"
}
```

### GET /api/validator/stats
```json
{
  "validator_id": "validator_address",
  "stake_amount": 1000,
  "rewards_earned": 50,
  "blocks_validated": 100,
  "reputation": 95.5
}
```
EOF

# 1.3.1 Definir interfaces
log "1.3.1 - Criando interfaces específicas..."
mkdir -p cmd/gui/interfaces

# Interface de Minerador
cat > cmd/gui/interfaces/miner_interface.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Minerador Offline ORDM</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .active { background-color: #d4edda; color: #155724; }
        .inactive { background-color: #f8d7da; color: #721c24; }
        .stats { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .control-panel { margin: 20px 0; }
        button { padding: 10px 20px; margin: 5px; border: none; border-radius: 5px; cursor: pointer; }
        .start { background-color: #28a745; color: white; }
        .stop { background-color: #dc3545; color: white; }
        .sync { background-color: #007bff; color: white; }
    </style>
</head>
<body>
    <h1>⛏️ Minerador Offline ORDM</h1>
    
    <div class="status active" id="status">
        Status: Mineração Ativa
    </div>
    
    <div class="stats">
        <div>
            <h3>Estatísticas de Mineração</h3>
            <p>Blocos Minerados: <span id="blocks-mined">0</span></p>
            <p>Hash Rate: <span id="hash-rate">0</span> H/s</p>
            <p>Pacotes Pendentes: <span id="pending-packages">0</span></p>
            <p>Último Bloco: <span id="last-block">-</span></p>
        </div>
        
        <div>
            <h3>Configurações</h3>
            <p>Dificuldade: <span id="difficulty">2</span></p>
            <p>Energia: <span id="energy-cost">$0.12/kWh</span></p>
            <p>Sincronização: <span id="sync-interval">30s</span></p>
        </div>
    </div>
    
    <div class="control-panel">
        <button class="start" onclick="startMining()">Iniciar Mineração</button>
        <button class="stop" onclick="stopMining()">Parar Mineração</button>
        <button class="sync" onclick="syncBlocks()">Sincronizar</button>
    </div>
    
    <div>
        <h3>Últimos Blocos Minerados</h3>
        <div id="recent-blocks">
            <p>Nenhum bloco minerado ainda</p>
        </div>
    </div>
    
    <script>
        function startMining() {
            fetch('/api/mining/start', {method: 'POST'})
                .then(response => response.json())
                .then(data => {
                    document.getElementById('status').textContent = 'Status: Mineração Ativa';
                    document.getElementById('status').className = 'status active';
                });
        }
        
        function stopMining() {
            fetch('/api/mining/stop', {method: 'POST'})
                .then(response => response.json())
                .then(data => {
                    document.getElementById('status').textContent = 'Status: Mineração Parada';
                    document.getElementById('status').className = 'status inactive';
                });
        }
        
        function syncBlocks() {
            fetch('/api/sync', {method: 'POST'})
                .then(response => response.json())
                .then(data => {
                    alert('Sincronização iniciada');
                });
        }
        
        // Atualizar estatísticas a cada 5 segundos
        setInterval(() => {
            fetch('/api/mining/stats')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('blocks-mined').textContent = data.total_blocks;
                    document.getElementById('hash-rate').textContent = data.hash_rate;
                    document.getElementById('pending-packages').textContent = data.pending_packages;
                });
        }, 5000);
    </script>
</body>
</html>
EOF

# Interface de Validador
cat > cmd/gui/interfaces/validator_interface.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Validador Online ORDM</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .active { background-color: #d4edda; color: #155724; }
        .inactive { background-color: #f8d7da; color: #721c24; }
        .stats { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .control-panel { margin: 20px 0; }
        button { padding: 10px 20px; margin: 5px; border: none; border-radius: 5px; cursor: pointer; }
        .stake { background-color: #28a745; color: white; }
        .unstake { background-color: #dc3545; color: white; }
        .rewards { background-color: #ffc107; color: black; }
    </style>
</head>
<body>
    <h1>🏆 Validador Online ORDM</h1>
    
    <div class="status active" id="status">
        Status: Validador Ativo
    </div>
    
    <div class="stats">
        <div>
            <h3>Estatísticas de Stake</h3>
            <p>Stake Atual: <span id="stake-amount">0</span> tokens</p>
            <p>APY: <span id="apy">7%</span> (5% + 2% bônus)</p>
            <p>Recompensas Ganhas: <span id="rewards-earned">0</span> tokens</p>
            <p>Blocos Validados: <span id="blocks-validated">0</span></p>
        </div>
        
        <div>
            <h3>Votações</h3>
            <p>Aceitos: <span id="votes-accepted">0</span></p>
            <p>Rejeitados: <span id="votes-rejected">0</span></p>
            <p>Taxa de Aceitação: <span id="acceptance-rate">0%</span></p>
        </div>
    </div>
    
    <div class="control-panel">
        <button class="stake" onclick="addStake()">Adicionar Stake</button>
        <button class="unstake" onclick="removeStake()">Remover Stake</button>
        <button class="rewards" onclick="claimRewards()">Coletar Recompensas</button>
    </div>
    
    <div>
        <h3>Blocos Recebidos Recentemente</h3>
        <div id="recent-blocks">
            <p>Nenhum bloco recebido ainda</p>
        </div>
    </div>
    
    <script>
        function addStake() {
            const amount = prompt('Quantidade de tokens para stake:');
            if (amount) {
                fetch('/api/validator/stake', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({amount: parseInt(amount)})
                })
                .then(response => response.json())
                .then(data => {
                    alert('Stake adicionado com sucesso');
                    updateStats();
                });
            }
        }
        
        function removeStake() {
            const amount = prompt('Quantidade de tokens para remover do stake:');
            if (amount) {
                fetch('/api/validator/unstake', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({amount: parseInt(amount)})
                })
                .then(response => response.json())
                .then(data => {
                    alert('Stake removido com sucesso');
                    updateStats();
                });
            }
        }
        
        function claimRewards() {
            fetch('/api/validator/claim-rewards', {method: 'POST'})
                .then(response => response.json())
                .then(data => {
                    alert('Recompensas coletadas: ' + data.amount + ' tokens');
                    updateStats();
                });
        }
        
        function updateStats() {
            fetch('/api/validator/stats')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('stake-amount').textContent = data.stake_amount;
                    document.getElementById('rewards-earned').textContent = data.rewards_earned;
                    document.getElementById('blocks-validated').textContent = data.blocks_validated;
                    document.getElementById('votes-accepted').textContent = data.votes_accepted;
                    document.getElementById('votes-rejected').textContent = data.votes_rejected;
                    
                    const total = data.votes_accepted + data.votes_rejected;
                    const rate = total > 0 ? Math.round((data.votes_accepted / total) * 100) : 0;
                    document.getElementById('acceptance-rate').textContent = rate + '%';
                });
        }
        
        // Atualizar estatísticas a cada 10 segundos
        setInterval(updateStats, 10000);
        updateStats();
    </script>
</body>
</html>
EOF

log "✅ PARTE 1: Consolidação Arquitetural concluída!"
log "📋 Arquivos criados:"
log "   - DECISIONS.md"
log "   - DEPENDENCIES.md"
log "   - FLOW_DIAGRAM.md"
log "   - API_CONTRACTS.md"
log "   - cmd/gui/interfaces/miner_interface.html"
log "   - cmd/gui/interfaces/validator_interface.html"

