#!/bin/bash

# 🧪 Script de Validação de Segurança ORDM
# Valida se as configurações de segurança estão corretas

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 Validando Configurações de Segurança...${NC}"
echo ""

# 1. Verificar arquivo .env
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ Arquivo .env encontrado${NC}"
    
    # Verificar permissões
    PERMS=$(stat -c %a .env 2>/dev/null || stat -f %Lp .env 2>/dev/null)
    if [ "$PERMS" = "600" ]; then
        echo -e "${GREEN}✅ Permissões do .env corretas (600)${NC}"
    else
        echo -e "${RED}❌ Permissões do .env incorretas: $PERMS (deve ser 600)${NC}"
    fi
else
    echo -e "${RED}❌ Arquivo .env não encontrado${NC}"
fi

# 2. Verificar variáveis de ambiente
if [ -n "$ADMIN_PASSWORD" ]; then
    echo -e "${GREEN}✅ ADMIN_PASSWORD configurada${NC}"
else
    echo -e "${YELLOW}⚠️  ADMIN_PASSWORD não configurada${NC}"
fi

if [ -n "$JWT_SECRET" ]; then
    echo -e "${GREEN}✅ JWT_SECRET configurada${NC}"
else
    echo -e "${YELLOW}⚠️  JWT_SECRET não configurada${NC}"
fi

# 3. Verificar diretório de logs
if [ -d "logs/secure" ]; then
    echo -e "${GREEN}✅ Diretório de logs seguro existe${NC}"
    
    PERMS=$(stat -c %a logs/secure 2>/dev/null || stat -f %Lp logs/secure 2>/dev/null)
    if [ "$PERMS" = "700" ]; then
        echo -e "${GREEN}✅ Permissões do diretório de logs corretas (700)${NC}"
    else
        echo -e "${RED}❌ Permissões do diretório de logs incorretas: $PERMS (deve ser 700)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Diretório de logs seguro não existe${NC}"
fi

# 4. Verificar .gitignore
if grep -q ".env" .gitignore; then
    echo -e "${GREEN}✅ .env está no .gitignore${NC}"
else
    echo -e "${RED}❌ .env não está no .gitignore${NC}"
fi

# 5. Testar compilação com segurança
echo -e "${YELLOW}🔨 Testando compilação com segurança...${NC}"
if go build -o ordmd ./cmd/ordmd; then
    echo -e "${GREEN}✅ Compilação bem-sucedida${NC}"
else
    echo -e "${RED}❌ Erro na compilação${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Validação de segurança concluída!${NC}"
