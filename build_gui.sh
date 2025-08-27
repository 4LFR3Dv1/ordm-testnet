#!/bin/bash

echo "🔨 Compilando Blockchain 2-Layer GUI..."

# Verificar se o Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado. Por favor, instale o Go primeiro."
    exit 1
fi

# Limpar builds anteriores
echo "🧹 Limpando builds anteriores..."
rm -f blockchain-gui
rm -f blockchain-gui.exe

# Compilar para diferentes plataformas
echo "📦 Compilando para diferentes plataformas..."

# Linux/Mac
echo "🐧 Compilando para Linux/Mac..."
GOOS=linux GOARCH=amd64 go build -o blockchain-gui-linux ./cmd/gui/main.go
GOOS=darwin GOARCH=amd64 go build -o blockchain-gui-mac ./cmd/gui/main.go

# Windows
echo "🪟 Compilando para Windows..."
GOOS=windows GOARCH=amd64 go build -o blockchain-gui.exe ./cmd/gui/main.go

# Verificar se a compilação foi bem-sucedida
if [ -f "blockchain-gui-linux" ]; then
    echo "✅ Executável Linux criado: blockchain-gui-linux"
fi

if [ -f "blockchain-gui-mac" ]; then
    echo "✅ Executável Mac criado: blockchain-gui-mac"
fi

if [ -f "blockchain-gui.exe" ]; then
    echo "✅ Executável Windows criado: blockchain-gui.exe"
fi

echo ""
echo "🎉 Compilação concluída!"
echo ""
echo "📋 Como usar:"
echo "   Linux/Mac: ./blockchain-gui-linux ou ./blockchain-gui-mac"
echo "   Windows: blockchain-gui.exe"
echo ""
echo "🌐 A interface estará disponível em: http://localhost:3000"
echo ""





