#!/bin/bash

echo "🚀 Iniciando ORDM Testnet..."
echo "🔧 Iniciando serviços da testnet..."

# Configurar variáveis de ambiente
export PORT=${PORT:-3000}
export EXPLORER_PORT=${EXPLORER_PORT:-8080}
export MONITOR_PORT=${MONITOR_PORT:-9090}
export DATA_DIR=${DATA_DIR:-/tmp/ordm-data}
export RENDER_EXTERNAL_URL=${RENDER_EXTERNAL_URL:-https://ordm-testnet-1.onrender.com}

# Em produção (Render), usar apenas a porta principal
if [ "$NODE_ENV" = "production" ]; then
    echo "🏭 Modo produção detectado - usando porta única: $PORT"
    export EXPLORER_PORT=$PORT
    export MONITOR_PORT=$PORT
fi

echo "📊 Configuração:"
echo "  - Porta principal: $PORT"
echo "  - Porta Explorer: $EXPLORER_PORT"
echo "  - Porta Monitor: $MONITOR_PORT"
echo "  - Data Directory: $DATA_DIR"
echo "  - External URL: $RENDER_EXTERNAL_URL"

# Criar diretórios necessários
echo "📁 Criando diretórios..."
mkdir -p $DATA_DIR
mkdir -p $DATA_DIR/wallets
mkdir -p $DATA_DIR/blockchain

# Função para iniciar serviço
start_service() {
    local service_name=$1
    local binary_path=$2
    local port=$3
    
    echo "📡 Iniciando $service_name na porta $port..."
    
    # Verificar se o binário existe
    if [ ! -f "$binary_path" ]; then
        echo "❌ Binário não encontrado: $binary_path"
        return 1
    fi
    
    # Tornar executável
    chmod +x "$binary_path"
    
    # Iniciar em background
    $binary_path &
    local pid=$!
    
    # Aguardar um pouco para verificar se iniciou
    sleep 2
    
    # Verificar se o processo ainda está rodando
    if kill -0 $pid 2>/dev/null; then
        echo "$service_name iniciado com PID: $pid"
        return 0
    else
        echo "❌ $service_name falhou ao iniciar"
        return 1
    fi
}

# Iniciar Node (porta principal)
if ! start_service "Node" "./ordm-node" $PORT; then
    echo "❌ Falha ao iniciar Node"
    exit 1
fi

# Em produção, integrar Explorer e Monitor no Node principal
if [ "$NODE_ENV" = "production" ]; then
    echo "🏭 Modo produção - Explorer e Monitor integrados no Node"
else
    # Em desenvolvimento, iniciar serviços separados
    if ! start_service "Explorer" "./ordm-explorer" $EXPLORER_PORT; then
        echo "⚠️ Explorer falhou, mas continuando..."
    fi

    if ! start_service "Monitor" "./ordm-monitor" $MONITOR_PORT; then
        echo "⚠️ Monitor falhou, mas continuando..."
    fi
fi

echo "🎉 ORDM Testnet iniciada com sucesso!"
echo "📊 URLs disponíveis:"
echo "  Node: http://localhost:$PORT"
echo "  Explorer: http://localhost:$EXPLORER_PORT"
echo "  Monitor: http://localhost:$MONITOR_PORT"

# Manter o script rodando
echo "🔄 Mantendo serviços ativos..."
while true; do
    sleep 30
    echo "💓 Heartbeat - Serviços ativos"
done
