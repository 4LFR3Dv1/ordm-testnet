# 📦 Estratégia de Redução Gradual de Dependências

## 🎯 Objetivo
Reduzir dependências de 273 para <50 sem perder funcionalidades

## 📋 Fase 1: Remoção Segura (Imediata)
- ✅ Remover BadgerDB v3 (mantendo v4)
- ✅ Executar go mod tidy
- ✅ Verificar compilação

## 📋 Fase 2: Análise de Uso (Próxima)
- 🔍 Verificar uso de Prometheus
- 🔍 Verificar uso de Zap
- 🔍 Verificar uso de Gorilla Mux
- 🔍 Verificar uso de WebSocket

## 📋 Fase 3: Substituição (Futura)
- 🔄 Substituir Zap por log padrão
- 🔄 Substituir Prometheus por métricas simples
- 🔄 Simplificar multiaddr se possível
- 🔄 Avaliar necessidade de btcec

## 📋 Fase 4: Otimização (Futura)
- ⚡ Implementar vendoring
- ⚡ Pin versões específicas
- ⚡ Remover dependências transitivas desnecessárias

## 🚨 Critérios de Segurança
1. **Nunca remover** dependências core
2. **Sempre testar** após cada remoção
3. **Manter backup** do go.mod
4. **Verificar funcionalidades** críticas

## 📊 Métricas de Sucesso
- [ ] <50 dependências totais
- [ ] Todas as funcionalidades mantidas
- [ ] Tempo de build reduzido
- [ ] Tamanho do binário reduzido
- [ ] Sem vulnerabilidades críticas

## 🔄 Processo de Teste
1. Remover dependência
2. go mod tidy
3. go build ./...
4. Executar testes básicos
5. Verificar funcionalidades críticas
6. Documentar mudanças

