#!/bin/bash

# 📦 Script para PARTE 4A: Auditoria de Dependências
# Subparte 4.1 do PLANO_ATUALIZACOES.md

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

log "🔄 Iniciando PARTE 4A: Auditoria de Dependências"

# 4.1.1 Analisar dependências críticas
log "4.1.1 - Analisando dependências críticas..."

# Verificar dependências atuais
log "Dependências atuais:"
go list -m all | wc -l

# Listar dependências principais
log "Dependências principais:"
go list -m all | grep -E "(libp2p|badger|crypto|auth)" || true

# 4.1.2 Resolver conflitos de versão
log "4.1.2 - Resolvendo conflitos de versão..."

# Verificar se Badger v3 está presente
if go list -m all | grep -q "badger/v3"; then
    log "Removendo Badger v3..."
    go mod edit -droprequire github.com/dgraph-io/badger/v3
    log "✅ Badger v3 removido"
else
    log "Badger v3 não encontrado"
fi

# Verificar múltiplas versões de libp2p
log "Verificando versões de libp2p..."
go list -m all | grep libp2p || true

# 4.1.3 Limpar dependências
log "4.1.3 - Limpando dependências..."
go mod tidy

# Verificar dependências após limpeza
log "Dependências após limpeza:"
go list -m all | wc -l

# Criar relatório de dependências
log "Criando relatório de dependências..."
cat > DEPENDENCIES_REPORT.md << 'EOF'
# 📦 Relatório de Dependências ORDM

## Dependências Críticas (Manter)

### Rede P2P
- `github.com/libp2p/go-libp2p` - Rede peer-to-peer
- `github.com/libp2p/go-libp2p-pubsub` - Sistema de pub/sub

### Database
- `github.com/dgraph-io/badger/v4` - Database key-value

### Criptografia
- `golang.org/x/crypto` - Funções criptográficas
- `github.com/tyler-smith/go-bip39` - Geração de mnemônicos

### Autenticação
- `github.com/golang-jwt/jwt` - Tokens JWT

## Dependências Opcionais (Avaliar)

### Monitoramento
- `github.com/prometheus/client_golang` - Métricas Prometheus

### Logging
- `go.uber.org/zap` - Logger estruturado

### Testes
- `github.com/stretchr/testify` - Framework de testes

## Dependências Removidas

### Conflitos Resolvidos
- `github.com/dgraph-io/badger/v3` - Substituído por v4

## Recomendações

1. **Manter apenas dependências essenciais**
2. **Usar versões estáveis**
3. **Auditar regularmente**
4. **Implementar vendoring para produção**

## Métricas

- **Total de dependências**: [será preenchido automaticamente]
- **Dependências diretas**: [será preenchido automaticamente]
- **Dependências transitivas**: [será preenchido automaticamente]
EOF

# Contar dependências
TOTAL_DEPS=$(go list -m all | wc -l)
DIRECT_DEPS=$(go list -m -u all | grep -v "indirect" | wc -l)
TRANSITIVE_DEPS=$((TOTAL_DEPS - DIRECT_DEPS))

# Atualizar relatório com métricas
sed -i.bak "s/\[será preenchido automaticamente\]/$TOTAL_DEPS/g" DEPENDENCIES_REPORT.md
sed -i.bak "s/\[será preenchido automaticamente\]/$DIRECT_DEPS/g" DEPENDENCIES_REPORT.md
sed -i.bak "s/\[será preenchido automaticamente\]/$TRANSITIVE_DEPS/g" DEPENDENCIES_REPORT.md

log "✅ PARTE 4A: Auditoria de Dependências concluída!"
log "📋 Arquivos criados:"
log "   - DEPENDENCIES_REPORT.md"
log "📊 Métricas:"
log "   - Total de dependências: $TOTAL_DEPS"
log "   - Dependências diretas: $DIRECT_DEPS"
log "   - Dependências transitivas: $TRANSITIVE_DEPS"

