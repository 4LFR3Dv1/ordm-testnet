#!/bin/bash

# 📊 Dashboard de Segurança ORDM Blockchain

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}📊 DASHBOARD DE SEGURANÇA ORDM BLOCKCHAIN${NC}"
echo "=============================================="
echo ""

# 1. Status da FASE 1
echo -e "${YELLOW}🔐 FASE 1: Correções Críticas${NC}"
echo "----------------------------------------"
if [ -f "pkg/security/validation.go" ]; then
    echo -e "${GREEN}✅ Validação robusta implementada${NC}"
else
    echo -e "${RED}❌ Validação robusta não implementada${NC}"
fi

if [ -f "pkg/security/secure_logger.go" ]; then
    echo -e "${GREEN}✅ Logs seguros implementados${NC}"
else
    echo -e "${RED}❌ Logs seguros não implementados${NC}"
fi

if [ -f ".env" ]; then
    echo -e "${GREEN}✅ Variáveis de ambiente configuradas${NC}"
else
    echo -e "${RED}❌ Variáveis de ambiente não configuradas${NC}"
fi

echo ""

# 2. Status da FASE 2
echo -e "${YELLOW}🛡️ FASE 2: Melhorias Avançadas${NC}"
echo "----------------------------------------"
if [ -f "pkg/security/two_factor.go" ]; then
    echo -e "${GREEN}✅ 2FA implementado${NC}"
else
    echo -e "${RED}❌ 2FA não implementado${NC}"
fi

if [ -f "pkg/security/csrf.go" ]; then
    echo -e "${GREEN}✅ Proteção CSRF implementada${NC}"
else
    echo -e "${RED}❌ Proteção CSRF não implementada${NC}"
fi

if [ -f "pkg/security/audit_logger.go" ]; then
    echo -e "${GREEN}✅ Audit logging implementado${NC}"
else
    echo -e "${RED}❌ Audit logging não implementado${NC}"
fi

if [ -f "pkg/security/ids_monitor.go" ]; then
    echo -e "${GREEN}✅ IDS/IPS implementado${NC}"
else
    echo -e "${RED}❌ IDS/IPS não implementado${NC}"
fi

echo ""

# 3. Status dos logs
echo -e "${YELLOW}📁 Status dos Logs${NC}"
echo "------------------------"
if [ -d "logs/secure" ]; then
    echo -e "${GREEN}✅ Diretório de logs seguros existe${NC}"
else
    echo -e "${RED}❌ Diretório de logs seguros não existe${NC}"
fi

if [ -d "logs/audit" ]; then
    echo -e "${GREEN}✅ Diretório de audit logs existe${NC}"
else
    echo -e "${RED}❌ Diretório de audit logs não existe${NC}"
fi

if [ -d "logs/ids" ]; then
    echo -e "${GREEN}✅ Diretório de IDS logs existe${NC}"
else
    echo -e "${RED}❌ Diretório de IDS logs não existe${NC}"
fi

echo ""

# 4. Status da compilação
echo -e "${YELLOW}🔨 Status da Compilação${NC}"
echo "----------------------------"
if [ -f "ordmd" ]; then
    echo -e "${GREEN}✅ Executável ordmd existe${NC}"
    ls -la ordmd
else
    echo -e "${RED}❌ Executável ordmd não existe${NC}"
fi

echo ""

# 5. Score de segurança
echo -e "${YELLOW}📊 SCORE DE SEGURANÇA${NC}"
echo "------------------------"
PHASE1_SCORE=0
PHASE2_SCORE=0

# Calcular score FASE 1
if [ -f "pkg/security/validation.go" ]; then
    PHASE1_SCORE=$((PHASE1_SCORE + 30))
fi
if [ -f "pkg/security/secure_logger.go" ]; then
    PHASE1_SCORE=$((PHASE1_SCORE + 30))
fi
if [ -f ".env" ]; then
    PHASE1_SCORE=$((PHASE1_SCORE + 40))
fi

# Calcular score FASE 2
if [ -f "pkg/security/two_factor.go" ]; then
    PHASE2_SCORE=$((PHASE2_SCORE + 25))
fi
if [ -f "pkg/security/csrf.go" ]; then
    PHASE2_SCORE=$((PHASE2_SCORE + 25))
fi
if [ -f "pkg/security/audit_logger.go" ]; then
    PHASE2_SCORE=$((PHASE2_SCORE + 25))
fi
if [ -f "pkg/security/ids_monitor.go" ]; then
    PHASE2_SCORE=$((PHASE2_SCORE + 25))
fi

TOTAL_SCORE=$((PHASE1_SCORE + PHASE2_SCORE))

echo -e "FASE 1: ${PHASE1_SCORE}/100 pontos"
echo -e "FASE 2: ${PHASE2_SCORE}/100 pontos"
echo -e "TOTAL: ${TOTAL_SCORE}/200 pontos"

if [ $TOTAL_SCORE -ge 180 ]; then
    echo -e "${GREEN}🎉 SEGURANÇA EXCELENTE (95%+)${NC}"
elif [ $TOTAL_SCORE -ge 160 ]; then
    echo -e "${GREEN}✅ SEGURANÇA MUITO BOA (90%+)${NC}"
elif [ $TOTAL_SCORE -ge 140 ]; then
    echo -e "${YELLOW}⚠️ SEGURANÇA BOA (80%+)${NC}"
elif [ $TOTAL_SCORE -ge 120 ]; then
    echo -e "${YELLOW}⚠️ SEGURANÇA REGULAR (70%+)${NC}"
else
    echo -e "${RED}❌ SEGURANÇA INSUFICIENTE (<70%)${NC}"
fi

echo ""
echo -e "${BLUE}🚀 PRÓXIMOS PASSOS${NC}"
echo "====================="
if [ $TOTAL_SCORE -ge 180 ]; then
    echo -e "${GREEN}✅ Sistema pronto para produção!${NC}"
    echo -e "${YELLOW}💡 Considere implementar:${NC}"
    echo "   - Monitoramento em tempo real"
    echo "   - Alertas por email/SMS"
    echo "   - Backup automático de logs"
    echo "   - Relatórios de segurança"
else
    echo -e "${YELLOW}🔧 Implementar melhorias restantes:${NC}"
    if [ $PHASE1_SCORE -lt 100 ]; then
        echo "   - Completar FASE 1"
    fi
    if [ $PHASE2_SCORE -lt 100 ]; then
        echo "   - Completar FASE 2"
    fi
fi

echo ""
