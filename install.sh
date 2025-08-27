#!/bin/bash

echo "🔗 Blockchain 2-Layer - Instalação"
echo "=================================="

# Verificar sistema operacional
OS=$(uname -s)
case "$OS" in
    Linux*)     PLATFORM="linux";;
    Darwin*)    PLATFORM="mac";;
    CYGWIN*)    PLATFORM="windows";;
    MINGW*)     PLATFORM="windows";;
    *)          PLATFORM="unknown";;
esac

echo "🖥️  Sistema detectado: $PLATFORM"

# Verificar se o Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado."
    echo "📥 Por favor, instale o Go primeiro: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go encontrado: $(go version)"

# Compilar a interface
echo "🔨 Compilando interface gráfica..."
./build_gui.sh

# Criar diretório de instalação
INSTALL_DIR="$HOME/.blockchain-2layer"
mkdir -p "$INSTALL_DIR"

# Copiar executável apropriado
case "$PLATFORM" in
    "linux")
        if [ -f "blockchain-gui-linux" ]; then
            cp blockchain-gui-linux "$INSTALL_DIR/blockchain-gui"
            chmod +x "$INSTALL_DIR/blockchain-gui"
            echo "✅ Executável Linux instalado em: $INSTALL_DIR/blockchain-gui"
        fi
        ;;
    "mac")
        if [ -f "blockchain-gui-mac" ]; then
            cp blockchain-gui-mac "$INSTALL_DIR/blockchain-gui"
            chmod +x "$INSTALL_DIR/blockchain-gui"
            echo "✅ Executável Mac instalado em: $INSTALL_DIR/blockchain-gui"
        fi
        ;;
    "windows")
        if [ -f "blockchain-gui.exe" ]; then
            cp blockchain-gui.exe "$INSTALL_DIR/"
            echo "✅ Executável Windows instalado em: $INSTALL_DIR/blockchain-gui.exe"
        fi
        ;;
esac

# Criar script de atalho
case "$PLATFORM" in
    "linux"|"mac")
        cat > "$HOME/.local/bin/blockchain-gui" << 'EOF'
#!/bin/bash
$HOME/.blockchain-2layer/blockchain-gui
EOF
        chmod +x "$HOME/.local/bin/blockchain-gui"
        echo "✅ Atalho criado: $HOME/.local/bin/blockchain-gui"
        ;;
    "windows")
        echo "📝 Para Windows, execute: $INSTALL_DIR/blockchain-gui.exe"
        ;;
esac

# Criar arquivo de configuração
cat > "$INSTALL_DIR/config.json" << 'EOF'
{
    "nodes": [
        {"name": "Node1", "port": 8080, "peers": "8081,8082"},
        {"name": "Node2", "port": 8081, "peers": "8080,8082"},
        {"name": "Node3", "port": 8082, "peers": "8080,8081"}
    ],
    "gui_port": 3000,
    "auto_start": false,
    "log_level": "info"
}
EOF

echo "✅ Configuração criada em: $INSTALL_DIR/config.json"

# Criar script de desinstalação
cat > "$INSTALL_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
echo "🗑️  Desinstalando Blockchain 2-Layer GUI..."

# Parar processos em execução
pkill -f "blockchain-gui" 2>/dev/null || true

# Remover atalhos
rm -f "$HOME/.local/bin/blockchain-gui"

# Remover diretório de instalação
rm -rf "$HOME/.blockchain-2layer"

echo "✅ Desinstalação concluída!"
EOF

chmod +x "$INSTALL_DIR/uninstall.sh"

echo ""
echo "🎉 Instalação concluída!"
echo ""
echo "📋 Como usar:"
case "$PLATFORM" in
    "linux"|"mac")
        echo "   blockchain-gui"
        echo "   ou"
        echo "   $INSTALL_DIR/blockchain-gui"
        ;;
    "windows")
        echo "   $INSTALL_DIR/blockchain-gui.exe"
        ;;
esac
echo ""
echo "🌐 Interface disponível em: http://localhost:3000"
echo ""
echo "🗑️  Para desinstalar: $INSTALL_DIR/uninstall.sh"
echo ""





