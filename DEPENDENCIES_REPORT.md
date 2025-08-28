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

- **Total de dependências**:      272
- **Dependências diretas**:      272
- **Dependências transitivas**:      272
