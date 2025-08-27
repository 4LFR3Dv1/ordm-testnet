#!/bin/bash

echo "=== TESTE DE MÚLTIPLOS NODES - BLOCKCHAIN 2-LAYER ==="
echo ""

# Matar processos anteriores
pkill -f "go run" 2>/dev/null
sleep 2

echo "🚀 Iniciando Node 1 (Porta 8080)..."
go run ./cmd/node/main.go node1 8080 8081,8082 &
NODE1_PID=$!
sleep 3

echo "🚀 Iniciando Node 2 (Porta 8081)..."
go run ./cmd/node/main.go node2 8081 8080,8082 &
NODE2_PID=$!
sleep 3

echo "🚀 Iniciando Node 3 (Porta 8082)..."
go run ./cmd/node/main.go node3 8082 8080,8081 &
NODE3_PID=$!
sleep 3

echo ""
echo "✅ Todos os nodes iniciados!"
echo "📊 Monitorando atividade dos nodes..."
echo ""

# Monitorar por 60 segundos
for i in {1..60}; do
    echo "⏱️  Tempo: ${i}s/60s"
    sleep 1
done

echo ""
echo "🛑 Parando todos os nodes..."
kill $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
pkill -f "go run" 2>/dev/null

echo "✅ Teste concluído!"






