#!/bin/bash

echo "🧪 Testando Persistência do Sistema Blockchain 2-Layer"
echo "=================================================="

# Limpar dados anteriores
echo "🧹 Limpando dados anteriores..."
rm -rf ./data ./wallets 2>/dev/null

# Teste 1: Primeira execução
echo ""
echo "📋 Teste 1: Primeira execução"
echo "Iniciando GUI..."
./blockchain-gui-mac &
GUI_PID=$!

# Aguardar inicialização
sleep 5

echo "Verificando estado inicial..."
curl -s http://localhost:3000/status | jq '.global_balance, .node.mining_stats.total_blocks' 2>/dev/null || echo "Erro ao acessar status"

# Simular mineração
echo "Simulando mineração..."
curl -s -X POST http://localhost:3000/start-mining > /dev/null
sleep 10

echo "Verificando após mineração..."
curl -s http://localhost:3000/status | jq '.global_balance, .node.mining_stats.total_blocks' 2>/dev/null || echo "Erro ao acessar status"

# Parar GUI
echo "Parando GUI..."
kill $GUI_PID 2>/dev/null
sleep 2

# Verificar arquivos de persistência
echo ""
echo "📁 Verificando arquivos de persistência..."
echo "Ledger Global:"
cat ./data/global_ledger.json 2>/dev/null || echo "Arquivo não encontrado"
echo ""
echo "Estado de Mineração:"
cat ./data/mining_state.json 2>/dev/null || echo "Arquivo não encontrado"

# Teste 2: Segunda execução
echo ""
echo "📋 Teste 2: Segunda execução (deve manter estado)"
echo "Iniciando GUI novamente..."
./blockchain-gui-mac &
GUI_PID=$!

sleep 5

echo "Verificando se estado foi mantido..."
curl -s http://localhost:3000/status | jq '.global_balance, .node.mining_stats.total_blocks' 2>/dev/null || echo "Erro ao acessar status"

# Parar GUI
kill $GUI_PID 2>/dev/null
sleep 2

echo ""
echo "✅ Teste de persistência concluído!"





