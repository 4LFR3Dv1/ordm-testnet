#!/bin/bash

echo "🔧 COMPILANDO MINERADOR CLI"
echo "==========================="

# Parar processos existentes
pkill -f "blockchain-gui-mac" 2>/dev/null
sleep 2

# Compilar o minerador CLI
echo "📦 Compilando minerador CLI..."
go build -o ordm-miner-cli miner_cli_simple.go

if [ $? -eq 0 ]; then
    echo "✅ Compilação bem-sucedida!"
    echo "🚀 Executando minerador CLI..."
    echo ""
    echo "📋 INSTRUÇÕES:"
    echo "1. Digite '5' para gerar um PIN"
    echo "2. Digite '6' e insira o PIN para autenticar"
    echo "3. Digite '1' para iniciar mineração"
    echo "4. Digite '3' para ver status"
    echo "5. Digite '4' para ver blocos recentes"
    echo "6. Digite '2' para parar mineração"
    echo "7. Digite 'q' para sair"
    echo ""
    
    # Executar o minerador
    ./ordm-miner-cli
else
    echo "❌ Erro na compilação!"
    exit 1
fi
