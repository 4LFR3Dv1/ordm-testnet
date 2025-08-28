#!/bin/bash

# 🔐 Script de Configuração de Segurança ORDM Blockchain
# Configura variáveis de ambiente seguras para produção

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔐 Configurando Segurança ORDM Blockchain...${NC}"
echo ""

# 1. Gerar senha segura para admin
echo -e "${YELLOW}🔑 Gerando senha segura para admin...${NC}"
ADMIN_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
echo -e "${GREEN}✅ Senha admin gerada: ${ADMIN_PASSWORD}${NC}"

# 2. Gerar hash da senha
echo -e "${YELLOW}🔐 Gerando hash da senha...${NC}"
ADMIN_PASSWORD_HASH=$(echo -n "$ADMIN_PASSWORD" | openssl dgst -sha256 -binary | base64)
echo -e "${GREEN}✅ Hash da senha gerado${NC}"

# 3. Gerar chave JWT secreta
echo -e "${YELLOW}🔑 Gerando chave JWT secreta...${NC}"
JWT_SECRET=$(openssl rand -base64 64)
echo -e "${GREEN}✅ Chave JWT gerada${NC}"

# 4. Gerar chave de criptografia
echo -e "${YELLOW}🔐 Gerando chave de criptografia...${NC}"
ENCRYPTION_KEY=$(openssl rand -base64 32)
echo -e "${GREEN}✅ Chave de criptografia gerada${NC}"

# 5. Criar arquivo .env
echo -e "${YELLOW}📝 Criando arquivo .env...${NC}"
cat > .env << EOF
# 🔐 ORDM Blockchain - Variáveis de Segurança
# ⚠️  NUNCA commitar este arquivo no git!

# Credenciais de Administração
ADMIN_PASSWORD=$ADMIN_PASSWORD
ADMIN_PASSWORD_HASH=$ADMIN_PASSWORD_HASH

# Chaves de Criptografia
JWT_SECRET=$JWT_SECRET
ENCRYPTION_KEY=$ENCRYPTION_KEY

# Configurações de Segurança
SECURE_LOGGING=true
MASK_SENSITIVE_DATA=true
RATE_LIMIT_ATTEMPTS=3
RATE_LIMIT_WINDOW=5m

# Configurações de Rede
ORDM_NETWORK=testnet
P2P_PORT=3000
RPC_PORT=8081

# Configurações de Log
LOG_LEVEL=INFO
LOG_ENCRYPTION=true
LOG_ROTATION_SIZE=100MB
LOG_ROTATION_AGE=7d
EOF

echo -e "${GREEN}✅ Arquivo .env criado${NC}"

# 6. Configurar permissões
echo -e "${YELLOW}🔒 Configurando permissões...${NC}"
chmod 600 .env
echo -e "${GREEN}✅ Permissões configuradas (600)${NC}"

# 7. Criar diretório de logs seguro
echo -e "${YELLOW}📁 Criando diretório de logs seguro...${NC}"
mkdir -p logs/secure
chmod 700 logs/secure
echo -e "${GREEN}✅ Diretório de logs seguro criado${NC}"

# 8. Configurar .gitignore
echo -e "${YELLOW}🚫 Configurando .gitignore...${NC}"
if ! grep -q ".env" .gitignore 2>/dev/null; then
    echo "" >> .gitignore
    echo "# 🔐 Arquivos de Segurança" >> .gitignore
    echo ".env" >> .gitignore
    echo "*.key" >> .gitignore
    echo "*.pem" >> .gitignore
    echo "logs/secure/*" >> .gitignore
    echo "machine_id.json" >> .gitignore
fi
echo -e "${GREEN}✅ .gitignore configurado${NC}"

# 9. Criar script de backup de segurança
echo -e "${YELLOW}💾 Criando script de backup de segurança...${NC}"
cat > scripts/backup_security.sh << 'EOF'
#!/bin/bash

# 💾 Script de Backup de Segurança ORDM
# Faz backup das configurações de segurança

set -e

BACKUP_DIR="./backup/security/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup do arquivo .env
if [ -f ".env" ]; then
    cp .env "$BACKUP_DIR/"
    echo "✅ Backup do .env criado"
fi

# Backup de chaves
if [ -d "wallets" ]; then
    cp -r wallets "$BACKUP_DIR/"
    echo "✅ Backup das wallets criado"
fi

# Backup de logs seguros
if [ -d "logs/secure" ]; then
    cp -r logs/secure "$BACKUP_DIR/"
    echo "✅ Backup dos logs seguros criado"
fi

# Comprimir backup
tar -czf "$BACKUP_DIR.tar.gz" -C "$(dirname "$BACKUP_DIR")" "$(basename "$BACKUP_DIR")"
rm -rf "$BACKUP_DIR"

echo "✅ Backup de segurança criado: $BACKUP_DIR.tar.gz"
echo "🔐 Mova este arquivo para local seguro!"
EOF

chmod +x scripts/backup_security.sh
echo -e "${GREEN}✅ Script de backup criado${NC}"

# 10. Criar script de validação de segurança
echo -e "${YELLOW}🧪 Criando script de validação de segurança...${NC}"
cat > scripts/validate_security.sh << 'EOF'
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
EOF

chmod +x scripts/validate_security.sh
echo -e "${GREEN}✅ Script de validação criado${NC}"

# 11. Resumo final
echo ""
echo -e "${BLUE}📊 RESUMO DA CONFIGURAÇÃO DE SEGURANÇA${NC}"
echo "=============================================="
echo -e "${GREEN}✅ Senha admin segura gerada${NC}"
echo -e "${GREEN}✅ Hash da senha gerado${NC}"
echo -e "${GREEN}✅ Chave JWT secreta gerada${NC}"
echo -e "${GREEN}✅ Chave de criptografia gerada${NC}"
echo -e "${GREEN}✅ Arquivo .env criado com permissões 600${NC}"
echo -e "${GREEN}✅ Diretório de logs seguro criado${NC}"
echo -e "${GREEN}✅ .gitignore configurado${NC}"
echo -e "${GREEN}✅ Script de backup criado${NC}"
echo -e "${GREEN}✅ Script de validação criado${NC}"
echo ""

# 12. Instruções importantes
echo -e "${BLUE}🚨 INSTRUÇÕES IMPORTANTES${NC}"
echo "=============================="
echo -e "${YELLOW}1. Guarde a senha admin em local seguro:${NC}"
echo "   $ADMIN_PASSWORD"
echo ""
echo -e "${YELLOW}2. NUNCA commite o arquivo .env no git${NC}"
echo ""
echo -e "${YELLOW}3. Faça backup das configurações:${NC}"
echo "   ./scripts/backup_security.sh"
echo ""
echo -e "${YELLOW}4. Valide as configurações:${NC}"
echo "   ./scripts/validate_security.sh"
echo ""
echo -e "${YELLOW}5. Para usar em produção, carregue as variáveis:${NC}"
echo "   source .env"
echo ""

echo -e "${GREEN}🎉 Configuração de segurança concluída com sucesso!${NC}"
echo -e "${GREEN}🔐 ORDM Blockchain está pronto para produção segura!${NC}"
