#!/bin/bash

# 🧪 Script para Testar Múltiplos Nodes P2P
# Testa comunicação entre 3 nodes em portas diferentes

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')] INFO: $1${NC}"
}

# Função para limpar processos
cleanup() {
    log "🧹 Limpando processos..."
    pkill -f "test-build" || true
    pkill -f "go run cmd/web" || true
    sleep 2
}

# Função para testar endpoint
test_endpoint() {
    local url=$1
    local description=$2
    local max_retries=10
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            log "✅ $description: $url"
            return 0
        else
            retry_count=$((retry_count + 1))
            sleep 1
        fi
    done
    
    error "❌ $description: $url (falhou após $max_retries tentativas)"
    return 1
}

# Função para testar mineração
test_mining() {
    local port=$1
    local node_id=$2
    
    log "⛏️ Testando mineração no $node_id..."
    
    response=$(curl -s -X POST "http://localhost:$port/api/test/mine-block")
    if echo "$response" | grep -q "success.*true"; then
        block_number=$(echo "$response" | grep -o '"block_number":[0-9]*' | cut -d':' -f2)
        log "✅ Bloco #$block_number minerado no $node_id"
        return 0
    else
        error "❌ Falha na mineração no $node_id"
        return 1
    fi
}

# Função para testar broadcast
test_broadcast() {
    local port=$1
    local node_id=$2
    
    log "💸 Testando broadcast no $node_id..."
    
    response=$(curl -s -X POST "http://localhost:$port/api/test/broadcast")
    if echo "$response" | grep -q "success.*true"; then
        tx_hash=$(echo "$response" | grep -o '"tx_hash":"[^"]*"' | cut -d'"' -f4)
        log "✅ Transação $tx_hash broadcastada no $node_id"
        return 0
    else
        error "❌ Falha no broadcast no $node_id"
        return 1
    fi
}

# Função para verificar peers
check_peers() {
    local port=$1
    local node_id=$2
    
    log "👥 Verificando peers do $node_id..."
    
    response=$(curl -s "http://localhost:$port/api/p2p/peers")
    peer_count=$(echo "$response" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    
    if [ "$peer_count" -gt 0 ]; then
        log "✅ $node_id tem $peer_count peers conectados"
        return 0
    else
        warn "⚠️ $node_id não tem peers conectados"
        return 1
    fi
}

# Função para conectar nodes
connect_nodes() {
    log "🔗 Conectando nodes..."
    
    # Obter endereços P2P dos nodes
    node1_p2p=$(curl -s "http://localhost:3000/api/p2p/status" | grep -o '"listening_addrs":\[[^]]*\]' | grep -o '/ip4/[^"]*')
    node2_p2p=$(curl -s "http://localhost:3001/api/p2p/status" | grep -o '"listening_addrs":\[[^]]*\]' | grep -o '/ip4/[^"]*')
    node3_p2p=$(curl -s "http://localhost:3002/api/p2p/status" | grep -o '"listening_addrs":\[[^]]*\]' | grep -o '/ip4/[^"]*')
    
    log "📡 Endereços P2P:"
    log "  Node 1: $node1_p2p"
    log "  Node 2: $node2_p2p"
    log "  Node 3: $node3_p2p"
    
    # Aqui seria implementada a conexão P2P real
    # Por enquanto, apenas log
    warn "⚠️ Conexão P2P manual necessária (implementar)"
}

# Main
main() {
    log "🚀 Iniciando Teste de Múltiplos Nodes P2P"
    
    # Limpar processos anteriores
    cleanup
    
    # Iniciar Node 1 (porta 3000, P2P 3002)
    log "🌐 Iniciando Node 1 (Web: 3000, P2P: 3002)..."
    ./test-build -port 3000 -p2p-port 3002 > node1.log 2>&1 &
    NODE1_PID=$!
    
    # Aguardar Node 1 inicializar
    sleep 5
    
    # Iniciar Node 2 (porta 3001, P2P 3003)
    log "🌐 Iniciando Node 2 (Web: 3001, P2P: 3003)..."
    ./test-build -port 3001 -p2p-port 3003 > node2.log 2>&1 &
    NODE2_PID=$!
    
    # Aguardar Node 2 inicializar
    sleep 5
    
    # Iniciar Node 3 (porta 3002, P2P 3004)
    log "🌐 Iniciando Node 3 (Web: 3002, P2P: 3004)..."
    ./test-build -port 3002 -p2p-port 3004 > node3.log 2>&1 &
    NODE3_PID=$!
    
    # Aguardar Node 3 inicializar
    sleep 5
    
    log "📊 Testando conectividade dos nodes..."
    
    # Testar endpoints dos 3 nodes
    test_endpoint "http://localhost:3000/health" "Node 1 Health"
    test_endpoint "http://localhost:3001/health" "Node 2 Health"
    test_endpoint "http://localhost:3002/health" "Node 3 Health"
    
    test_endpoint "http://localhost:3000/api/p2p/status" "Node 1 P2P Status"
    test_endpoint "http://localhost:3001/api/p2p/status" "Node 2 P2P Status"
    test_endpoint "http://localhost:3002/api/p2p/status" "Node 3 P2P Status"
    
    # Verificar peers (inicialmente devem estar vazios)
    log "👥 Verificando peers iniciais..."
    check_peers 3000 "Node 1"
    check_peers 3001 "Node 2"
    check_peers 3002 "Node 3"
    
    # Tentar conectar nodes
    connect_nodes
    
    # Testar mineração em cada node
    log "⛏️ Testando mineração em todos os nodes..."
    test_mining 3000 "Node 1"
    test_mining 3001 "Node 2"
    test_mining 3002 "Node 3"
    
    # Testar broadcast em cada node
    log "💸 Testando broadcast em todos os nodes..."
    test_broadcast 3000 "Node 1"
    test_broadcast 3001 "Node 2"
    test_broadcast 3002 "Node 3"
    
    # Aguardar um pouco para ver se há comunicação
    log "⏳ Aguardando comunicação entre nodes..."
    sleep 10
    
    # Verificar peers novamente
    log "👥 Verificando peers após testes..."
    check_peers 3000 "Node 1"
    check_peers 3001 "Node 2"
    check_peers 3002 "Node 3"
    
    # Mostrar estatísticas finais
    log "📊 Estatísticas finais:"
    echo "Node 1: $(curl -s http://localhost:3000/api/stats | grep -o '"total_blocks":[0-9]*' | cut -d':' -f2) blocos"
    echo "Node 2: $(curl -s http://localhost:3001/api/stats | grep -o '"total_blocks":[0-9]*' | cut -d':' -f2) blocos"
    echo "Node 3: $(curl -s http://localhost:3002/api/stats | grep -o '"total_blocks":[0-9]*' | cut -d':' -f2) blocos"
    
    log "✅ Teste de múltiplos nodes concluído!"
    log "📋 URLs dos nodes:"
    log "  Node 1: http://localhost:3000"
    log "  Node 2: http://localhost:3001"
    log "  Node 3: http://localhost:3002"
    
    # Manter nodes rodando por um tempo para inspeção manual
    log "🔄 Nodes continuarão rodando por 60 segundos para inspeção manual..."
    sleep 60
    
    # Limpar
    cleanup
    log "🧹 Teste finalizado!"
}

# Executar main
main "$@"






