#!/bin/bash

# 🧹 Script de Limpeza do Repositório ORDM
# Remove arquivos duplicados e desnecessários

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 Iniciando limpeza do repositório ORDM...${NC}"
echo ""

# 1. Remover executáveis antigos
echo -e "${YELLOW}📦 Removendo executáveis antigos...${NC}"
rm -f offline_miner
rm -f ordm-offline-miner
rm -f test-build
rm -f web
rm -f test-resilience
echo -e "${GREEN}✅ Executáveis antigos removidos${NC}"

# 2. Remover logs antigos
echo -e "${YELLOW}📋 Removendo logs antigos...${NC}"
rm -f *.log
echo -e "${GREEN}✅ Logs antigos removidos${NC}"

# 3. Remover arquivos de teste antigos
echo -e "${YELLOW}🧪 Removendo arquivos de teste antigos...${NC}"
rm -f test_*.sh
rm -f test_*.go
rm -f test_*.json
echo -e "${GREEN}✅ Arquivos de teste antigos removidos${NC}"

# 4. Remover arquivos temporários
echo -e "${YELLOW}🗑️ Removendo arquivos temporários...${NC}"
rm -f .DS_Store
rm -f *.tmp
rm -f *.bak
echo -e "${GREEN}✅ Arquivos temporários removidos${NC}"

# 5. Verificar executável integrado
echo -e "${YELLOW}🔍 Verificando executável integrado...${NC}"
if [ -f "./ordmd" ]; then
    echo -e "${GREEN}✅ Executável integrado 'ordmd' encontrado${NC}"
    ls -la ./ordmd
else
    echo -e "${RED}❌ Executável integrado 'ordmd' não encontrado${NC}"
    echo -e "${YELLOW}📦 Compilando executável integrado...${NC}"
    go build -o ordmd ./cmd/ordmd
    echo -e "${GREEN}✅ Executável integrado compilado${NC}"
fi

# 6. Verificar arquivos essenciais
echo -e "${YELLOW}📋 Verificando arquivos essenciais...${NC}"
ESSENTIAL_FILES=(
    "cmd/ordmd/main.go"
    "config/testnet.json"
    "genesis/testnet.json"
    "scripts/run-node.sh"
    "TESTNET_README.md"
    "Dockerfile"
    "render.yaml"
)

for file in "${ESSENTIAL_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $file${NC}"
    else
        echo -e "${RED}❌ $file (FALTANDO)${NC}"
    fi
done

# 7. Verificar estrutura de diretórios
echo -e "${YELLOW}📁 Verificando estrutura de diretórios...${NC}"
if [ -d "cmd/ordmd" ]; then
    echo -e "${GREEN}✅ cmd/ordmd/ (executável integrado)${NC}"
else
    echo -e "${RED}❌ cmd/ordmd/ (FALTANDO)${NC}"
fi

if [ -d "pkg/crypto" ]; then
    echo -e "${GREEN}✅ pkg/crypto/ (machineID)${NC}"
else
    echo -e "${RED}❌ pkg/crypto/ (FALTANDO)${NC}"
fi

if [ -d "config" ]; then
    echo -e "${GREEN}✅ config/ (configurações)${NC}"
else
    echo -e "${RED}❌ config/ (FALTANDO)${NC}"
fi

if [ -d "genesis" ]; then
    echo -e "${GREEN}✅ genesis/ (bloco genesis)${NC}"
else
    echo -e "${RED}❌ genesis/ (FALTANDO)${NC}"
fi

# 8. Testar executável integrado
echo -e "${YELLOW}🧪 Testando executável integrado...${NC}"
if [ -f "./ordmd" ]; then
    echo -e "${BLUE}Testando help do executável...${NC}"
    ./ordmd --help | head -10
    echo -e "${GREEN}✅ Executável integrado funcionando${NC}"
else
    echo -e "${RED}❌ Executável integrado não encontrado${NC}"
fi

# 9. Resumo final
echo ""
echo -e "${BLUE}📊 RESUMO DA LIMPEZA${NC}"
echo "================================"
echo -e "${GREEN}✅ Executáveis antigos removidos${NC}"
echo -e "${GREEN}✅ Logs antigos removidos${NC}"
echo -e "${GREEN}✅ Arquivos de teste antigos removidos${NC}"
echo -e "${GREEN}✅ Arquivos temporários removidos${NC}"
echo -e "${GREEN}✅ Executável integrado verificado${NC}"
echo -e "${GREEN}✅ Arquivos essenciais verificados${NC}"
echo -e "${GREEN}✅ Estrutura de diretórios verificada${NC}"
echo ""

# 10. Próximos passos
echo -e "${BLUE}🚀 PRÓXIMOS PASSOS${NC}"
echo "====================="
echo -e "${YELLOW}1. Testar executável integrado:${NC}"
echo "   ./ordmd --mode both --miner-threads 2"
echo ""
echo -e "${YELLOW}2. Fazer deploy no Render:${NC}"
echo "   - Commit das mudanças"
echo "   - Push para o repositório"
echo "   - Render fará deploy automático"
echo ""
echo -e "${YELLOW}3. Verificar deploy:${NC}"
echo "   curl https://ordm-testnet.onrender.com/api/v1/blockchain/info"
echo ""

echo -e "${GREEN}🎉 Limpeza do repositório concluída com sucesso!${NC}"
