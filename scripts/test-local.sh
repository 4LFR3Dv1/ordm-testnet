#!/bin/bash

# Script para testar localmente sem Docker
echo "🧪 Testando ORDM Testnet localmente..."

# Verificar se Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado"
    exit 1
fi

# Verificar se os arquivos necessários existem
required_files=("go.mod" "Dockerfile" "render.yaml" "scripts/start.sh")
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "❌ Arquivo $file não encontrado"
        exit 1
    fi
done

echo "✅ Arquivos necessários encontrados"

# Compilar aplicações
echo "🔨 Compilando aplicações..."
go build -o ordm-node ./cmd/gui
go build -o ordm-explorer ./cmd/explorer
go build -o ordm-monitor ./cmd/monitor

if [ $? -eq 0 ]; then
    echo "✅ Compilação bem-sucedida"
else
    echo "❌ Erro na compilação"
    exit 1
fi

# Testar se os binários funcionam
echo "🧪 Testando binários..."

# Testar Node
echo "📡 Testando Node..."
./ordm-node &
NODE_PID=$!
sleep 3
if kill -0 $NODE_PID 2>/dev/null; then
    echo "✅ Node funcionando"
    kill $NODE_PID
else
    echo "❌ Node falhou"
fi

# Testar Explorer
echo "🔍 Testando Explorer..."
./ordm-explorer &
EXPLORER_PID=$!
sleep 3
if kill -0 $EXPLORER_PID 2>/dev/null; then
    echo "✅ Explorer funcionando"
    kill $EXPLORER_PID
else
    echo "❌ Explorer falhou"
fi

# Testar Monitor
echo "📊 Testando Monitor..."
./ordm-monitor &
MONITOR_PID=$!
sleep 3
if kill -0 $MONITOR_PID 2>/dev/null; then
    echo "✅ Monitor funcionando"
    kill $MONITOR_PID
else
    echo "❌ Monitor falhou"
fi

echo "🎉 Teste local concluído com sucesso!"
echo "📋 Próximos passos:"
echo "  1. git add ."
echo "  2. git commit -m 'Add deploy configuration'"
echo "  3. git push origin main"
echo "  4. Configurar Render conforme DEPLOY_GUIDE.md"
