#!/bin/bash

echo "🔧 INICIANDO MINERADOR ORDM"
echo "============================"

# Parar qualquer processo existente
echo "🛑 Parando processos existentes..."
pkill -f ordm-miner-cli 2>/dev/null
sleep 2

# Verificar se ainda há processos rodando
if pgrep -f ordm-miner-cli > /dev/null; then
    echo "❌ Ainda há processos rodando. Aguardando..."
    sleep 3
    pkill -9 -f ordm-miner-cli 2>/dev/null
fi

# Verificar estado atual
echo "📊 Verificando estado atual..."
if [ -f "data/mining_state.json" ]; then
    TOTAL_BLOCKS=$(cat data/mining_state.json | grep -o '"total_blocks":[0-9]*' | cut -d':' -f2)
    TOTAL_REWARDS=$(cat data/mining_state.json | grep -o '"total_rewards":[0-9]*' | cut -d':' -f2)
    echo "   📦 Total de blocos: $TOTAL_BLOCKS"
    echo "   💰 Total de recompensas: $TOTAL_REWARDS tokens"
fi

if [ -d "data/blocks" ]; then
    BLOCK_FILES=$(ls data/blocks/block_*.json 2>/dev/null | wc -l)
    echo "   📁 Arquivos de blocos: $BLOCK_FILES"
fi

echo ""
echo "🚀 Iniciando minerador..."
echo "📋 INSTRUÇÕES:"
echo "1. Digite '6' para autenticar"
echo "2. Digite o PIN quando solicitado"
echo "3. Digite '1' para iniciar mineração"
echo "4. Digite '3' para ver status"
echo "5. Digite '4' para ver blocos"
echo "6. Digite '2' para parar mineração"
echo "7. Digite 'q' para sair"
echo ""

# Executar o minerador
./ordm-miner-cli
