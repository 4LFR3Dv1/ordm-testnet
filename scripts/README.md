# 🚀 Scripts de Atualização ORDM

Este diretório contém scripts para implementar as atualizações identificadas na análise crítica do projeto ORDM Blockchain 2-Layer.

## 📋 Visão Geral

Os scripts estão divididos em **partes menores** para evitar timeouts e perda de arquivos durante a execução. Cada script é focado em uma funcionalidade específica.

## 🎯 Scripts Disponíveis

### Scripts Individuais

| Script | Descrição | Funcionalidade |
|--------|-----------|----------------|
| `part1_consolidate_architecture.sh` | Consolidação Arquitetural | Documentação unificada e interfaces |
| `part2a_offline_storage.sh` | Storage Offline | Persistência local para mineradores |
| `part2b_online_storage.sh` | Storage Online | Persistência no Render |
| `part2c_sync_protocol.sh` | Protocolo de Sincronização | Pacotes de blocos e validação |
| `part3_security.sh` | Segurança | 2FA, rate limiting, keystore |
| `part4a_dependencies.sh` | Auditoria de Dependências | Limpeza e otimização |
| `part5a_unit_tests.sh` | Testes Unitários | Testes para componentes críticos |

### Script Principal

| Script | Descrição |
|--------|-----------|
| `run_all_updates.sh` | **Script principal** com menu interativo |

## 🚀 Como Usar

### Opção 1: Script Principal (Recomendado)

```bash
# Executar o script principal
./scripts/run_all_updates.sh
```

O script principal oferece um menu interativo com as seguintes opções:

1. **PARTE 1**: Consolidação Arquitetural
2. **PARTE 2A**: Storage Offline
3. **PARTE 2B**: Storage Online
4. **PARTE 2C**: Protocolo de Sincronização
5. **PARTE 3**: Segurança
6. **PARTE 4A**: Auditoria de Dependências
7. **PARTE 5A**: Testes Unitários
8. **Executar TODAS as partes (sequencial)**
9. **Executar TODAS as partes (paralelo)**
10. **Verificar status das atualizações**
11. **Sair**

### Opção 2: Scripts Individuais

```bash
# Executar partes específicas
./scripts/part1_consolidate_architecture.sh
./scripts/part2a_offline_storage.sh
./scripts/part3_security.sh
# ... etc
```

### Opção 3: Execução Sequencial Completa

```bash
# Executar todas as partes em sequência
./scripts/run_all_updates.sh
# Escolher opção 8
```

## 📋 Pré-requisitos

Antes de executar os scripts, certifique-se de ter:

- **Go 1.25+** instalado
- **Git** instalado
- **Bash** disponível
- **Permissões de escrita** no diretório do projeto

## 🔍 Verificação de Status

Para verificar quais atualizações já foram aplicadas:

```bash
./scripts/run_all_updates.sh
# Escolher opção 10
```

## 📊 Arquivos Criados

Após a execução bem-sucedida, os seguintes arquivos serão criados:

### Documentação
- `DECISIONS.md` - Histórico de decisões técnicas
- `DEPENDENCIES.md` - Dependências entre componentes
- `FLOW_DIAGRAM.md` - Diagrama de fluxo
- `API_CONTRACTS.md` - Contratos de API

### Código
- `pkg/storage/offline_storage.go` - Storage offline criptografado
- `pkg/storage/render_storage.go` - Storage online persistente
- `pkg/sync/protocol.go` - Protocolo de sincronização
- `pkg/auth/rate_limiter.go` - Rate limiting
- `pkg/auth/pin_generator.go` - Geração de PIN seguro
- `pkg/crypto/keystore.go` - Keystore seguro
- `pkg/logger/secure_logger.go` - Logger seguro

### Testes
- `pkg/blockchain/real_block_test.go` - Testes de blockchain
- `pkg/wallet/secure_wallet_test.go` - Testes de wallet
- `pkg/auth/user_manager_test.go` - Testes de autenticação
- `scripts/run_tests.sh` - Script para executar testes

### Interfaces
- `cmd/gui/interfaces/miner_interface.html` - Interface de minerador
- `cmd/gui/interfaces/validator_interface.html` - Interface de validador

## ⚠️ Importante

### Backup
Antes de executar os scripts, faça backup do projeto:

```bash
# Criar backup
cp -r . ../ordm-backup-$(date +%Y%m%d-%H%M%S)
```

### Ordem de Execução
Para melhor resultado, execute os scripts na seguinte ordem:

1. **PARTE 1** (Consolidação Arquitetural)
2. **PARTE 2A, 2B, 2C** (Storage e Sincronização)
3. **PARTE 3** (Segurança)
4. **PARTE 4A** (Dependências)
5. **PARTE 5A** (Testes)

### Tratamento de Erros
Se um script falhar:

1. Verifique os logs de erro
2. Corrija o problema identificado
3. Execute o script novamente
4. Use a opção 10 para verificar o status

## 🧪 Executando Testes

Após aplicar as atualizações, execute os testes:

```bash
# Executar todos os testes
./scripts/run_tests.sh

# Ou executar testes específicos
go test ./pkg/blockchain -v
go test ./pkg/wallet -v
go test ./pkg/auth -v
```

## 📈 Métricas de Sucesso

Após a execução bem-sucedida, você deve ver:

- **Cobertura de testes**: >80%
- **Dependências**: <50 (redução de 60%)
- **Tempo de build**: <5 minutos
- **Tamanho de binário**: <50MB

## 🔧 Troubleshooting

### Problema: Script não executa
```bash
# Verificar permissões
chmod +x scripts/*.sh

# Verificar se bash está disponível
which bash
```

### Problema: Go não encontrado
```bash
# Verificar instalação do Go
go version

# Instalar Go se necessário
# https://golang.org/doc/install
```

### Problema: Dependências conflitantes
```bash
# Limpar módulos
go clean -modcache

# Atualizar dependências
go mod tidy
```

## 📞 Suporte

Se encontrar problemas:

1. Verifique os logs de erro
2. Consulte a documentação do projeto
3. Verifique se todos os pré-requisitos estão atendidos
4. Execute a verificação de status (opção 10)

---

**🎉 Boa sorte com as atualizações do ORDM Blockchain 2-Layer!**

