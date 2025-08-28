# 📦 Plano para Meta de <50 Dependências

## 🎯 Situação Atual
- **Dependências atuais**:      273
- **Meta**: <50 dependências
- **Ainda precisa remover**: 223 dependências

## 📋 Estratégias para Redução Adicional

### 1. Analisar Dependências Transitivas (Prioridade Alta)
- **libp2p** traz ~200 dependências
- **Avaliar**: Substituir por implementação mais simples
- **Alternativa**: Usar apenas TCP/UDP básico

### 2. Simplificar Multiaddr (Prioridade Média)
- **Atual**: Suporte completo a multiaddr
- **Proposta**: Usar apenas IP:porta
- **Redução esperada**: ~50 dependências

### 3. Substituir btcec (Prioridade Baixa)
- **Atual**: Bitcoin crypto library
- **Proposta**: Usar crypto padrão do Go
- **Redução esperada**: ~20 dependências

### 4. Otimizar BadgerDB (Prioridade Baixa)
- **Atual**: BadgerDB v4 completo
- **Proposta**: Usar apenas funcionalidades básicas
- **Redução esperada**: ~30 dependências

## 🚨 Riscos e Considerações

### Alto Risco
- **Substituir libp2p**: Pode quebrar funcionalidade P2P
- **Simplificar multiaddr**: Pode afetar conectividade

### Médio Risco
- **Substituir btcec**: Pode afetar compatibilidade Bitcoin
- **Otimizar BadgerDB**: Pode afetar performance

### Baixo Risco
- **Remover dependências não usadas**: Já feito
- **Pin versões**: Melhora estabilidade

## 📊 Cronograma Sugerido

### Semana 1: Análise
- [ ] Analisar uso real de libp2p
- [ ] Avaliar alternativas mais simples
- [ ] Testar funcionalidades P2P

### Semana 2: Implementação
- [ ] Implementar alternativa a libp2p
- [ ] Testar conectividade
- [ ] Verificar performance

### Semana 3: Otimização
- [ ] Simplificar multiaddr
- [ ] Otimizar BadgerDB
- [ ] Testes finais

## 🎯 Métricas de Sucesso

- [ ] <50 dependências totais
- [ ] Todas as funcionalidades mantidas
- [ ] Performance não degradada
- [ ] Conectividade P2P funcionando

