#!/bin/sh

# Script de inicialização para ORDM Testnet no Render
echo "🚀 Iniciando ORDM Testnet..."

# Definir variáveis de ambiente
export PORT=${PORT:-3000}
export EXPLORER_PORT=${EXPLORER_PORT:-8080}
export MONITOR_PORT=${MONITOR_PORT:-9090}

# Em produção (Render), usar apenas a porta principal
if [ "$NODE_ENV" = "production" ]; then
    echo "🏭 Modo produção detectado - usando porta única: $PORT"
    export EXPLORER_PORT=$PORT
    export MONITOR_PORT=$PORT
fi

# Criar diretórios se não existirem
mkdir -p /app/data /app/logs /app/backups /app/wallets

# Função para iniciar serviço em background
start_service() {
    local service_name=$1
    local command=$2
    local port=$3
    
    echo "📡 Iniciando $service_name na porta $port..."
    $command &
    local pid=$!
    echo "$service_name iniciado com PID: $pid"
    
    # Aguardar um pouco para o serviço inicializar
    sleep 3
    
    # Verificar se o serviço está rodando
    if kill -0 $pid 2>/dev/null; then
        echo "✅ $service_name está rodando"
    else
        echo "❌ $service_name falhou ao iniciar"
        return 1
    fi
}

# Iniciar serviços
echo "🔧 Iniciando serviços da testnet..."

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
echo "  Node:     http://localhost:$PORT"
echo "  Explorer: http://localhost:$EXPLORER_PORT"
echo "  Monitor:  http://localhost:$MONITOR_PORT"

# Manter o container rodando
echo "🔄 Mantendo serviços ativos..."
while true; do
    sleep 30
    echo "💓 Heartbeat - Serviços ativos"
done
