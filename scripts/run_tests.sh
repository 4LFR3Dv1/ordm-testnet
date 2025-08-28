#!/bin/bash

# Script para executar todos os testes

set -e

echo "🧪 Executando testes unitários..."

# Executar testes de blockchain
echo "📦 Testes de blockchain..."
go test ./pkg/blockchain -v

# Executar testes de wallet
echo "💰 Testes de wallet..."
go test ./pkg/wallet -v

# Executar testes de autenticação
echo "🔐 Testes de autenticação..."
go test ./pkg/auth -v

# Executar todos os testes
echo "🎯 Executando todos os testes..."
go test ./... -v

echo "✅ Todos os testes concluídos!"
