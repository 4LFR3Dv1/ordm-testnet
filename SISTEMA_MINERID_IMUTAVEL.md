# 🔐 Sistema de MinerID Imutável por Máquina

## 📋 Visão Geral

O sistema implementa uma **camada física de segurança** onde cada máquina que executa o minerador offline recebe uma **identidade única e imutável** na primeira execução. Esta identidade fica permanentemente vinculada à máquina, criando uma barreira de segurança física.

## 🎯 Características Principais

### 🔒 Identidade Imutável
- **Geração automática** na primeira execução
- **Vinculação física** à máquina através de hardware
- **Persistência permanente** em arquivo local
- **Impossibilidade de alteração** sem acesso físico

### 🖥️ Coleta de Dados da Máquina
O sistema coleta informações únicas do hardware:

```go
// Informações coletadas:
- Hostname da máquina
- Sistema operacional (GOOS)
- Arquitetura (GOARCH)
- Modelo do processador
- Número de CPUs
- Hash SHA-256 combinado
```

### 🔐 Geração do MinerID
```go
// Formato: MINER-XXXX-XXXX-XXXX
// Exemplo: MINER-34F5-8962-4FA5-E295
```

## 🏗️ Arquitetura do Sistema

### 📁 Estrutura de Arquivos
```
data/
├── machine_identity.json    # Identidade imutável da máquina
├── wallets.json            # Wallets do usuário
├── global_ledger.json      # Ledger global
└── mining_state.json       # Estado de mineração
```

### 🔧 Componentes Principais

#### 1. MachineIdentity
```go
type MachineIdentity struct {
    MinerID     string    `json:"miner_id"`      // ID único do minerador
    MachineHash string    `json:"machine_hash"`  // Hash da máquina
    CreatedAt   time.Time `json:"created_at"`    // Data de criação
    LastSeen    time.Time `json:"last_seen"`     // Último acesso
}
```

#### 2. Funções de Identidade
- `generateMachineIdentity()` - Gera identidade única
- `loadOrCreateMachineIdentity()` - Carrega ou cria identidade
- `saveMachineIdentity()` - Salva identidade em arquivo
- `getCurrentMinerID()` - Obtém MinerID atual

#### 3. Autenticação por PIN
- `generateMinerPIN()` - Gera PIN baseado no MinerID
- `validateMinerPIN()` - Valida PIN inserido
- `storeMinerPIN()` - Armazena PIN temporariamente

## 🔄 Fluxo de Autenticação

### 1. Primeira Execução
```
1. Sistema detecta ausência de machine_identity.json
2. Coleta informações únicas do hardware
3. Gera hash SHA-256 da máquina
4. Cria MinerID no formato MINER-XXXX-XXXX-XXXX
5. Salva identidade em data/machine_identity.json
6. Exibe mensagem de confirmação
```

### 2. Execuções Subsequentes
```
1. Sistema carrega identidade existente
2. Atualiza timestamp de último acesso
3. Valida integridade da identidade
4. Disponibiliza MinerID para autenticação
```

### 3. Autenticação por PIN
```
1. Usuário acessa aba "AUTENTICAÇÃO MINER"
2. Sistema carrega MinerID automaticamente
3. Usuário clica em "GERAR PIN NO TERMINAL"
4. PIN é exibido no console do servidor
5. Usuário digita PIN na interface web
6. Sistema valida e autentica o usuário
```

## 🛡️ Segurança

### 🔐 Camadas de Proteção

#### 1. Identidade Física
- **Hardware binding**: Vinculação ao hardware da máquina
- **Imutabilidade**: Identidade não pode ser alterada
- **Unicidade**: Cada máquina tem identidade única

#### 2. Autenticação por PIN
- **Geração baseada em hash**: PIN derivado do MinerID + timestamp
- **Expiração**: PIN válido por 60 segundos
- **Uso único**: PIN é invalidado após uso
- **Exibição no terminal**: PIN só aparece no console do servidor

#### 3. Armazenamento Seguro
- **Arquivo protegido**: machine_identity.json com permissões 600
- **Hash criptográfico**: SHA-256 para integridade
- **Validação de integridade**: Verificação de hash na inicialização

### 🚨 Proteções contra Ataques

#### 1. Ataques de Força Bruta
- **Rate limiting**: Limitação de tentativas de PIN
- **Expiração rápida**: PIN válido por apenas 60 segundos
- **Uso único**: PIN invalidado após primeira tentativa

#### 2. Ataques de Replay
- **Timestamp**: PIN inclui timestamp para evitar replay
- **Validação de tempo**: Verificação de expiração
- **Limpeza automática**: PINs expirados são removidos

#### 3. Ataques de Manipulação
- **Hash de integridade**: Verificação de hash da máquina
- **Arquivo protegido**: Permissões restritas no arquivo
- **Validação de formato**: Verificação de estrutura do JSON

## 💻 Interface de Usuário

### 🎨 Aba "AUTENTICAÇÃO MINER"
- **MinerID automático**: Campo preenchido automaticamente
- **Campo readonly**: MinerID não pode ser alterado
- **Informação visual**: Indicação de identidade imutável
- **Botão de geração**: Gera PIN no terminal do servidor

### 📱 Experiência do Usuário
```
1. Acessa aba "AUTENTICAÇÃO MINER"
2. Vê MinerID da máquina carregado automaticamente
3. Clica em "GERAR PIN NO TERMINAL"
4. Verifica PIN no console onde o servidor roda
5. Digita PIN na interface web
6. É autenticado e redirecionado para o dashboard
```

## 🔧 Implementação Técnica

### 📦 Dependências
```go
import (
    "crypto/sha256"
    "encoding/hex"
    "os/exec"
    "runtime"
    "sync"
    "time"
)
```

### 🗂️ Estrutura de Dados
```go
// Storage temporário de PINs
var minerPINs = make(map[string]struct {
    pin       string
    timestamp int64
})
var minerPINsMutex sync.RWMutex
```

### 🔄 Handlers HTTP
- `GET /get-machine-minerid` - Obtém MinerID da máquina
- `POST /authenticate-miner` - Gera PIN no terminal
- `POST /validate-miner-pin` - Valida PIN inserido

## 🎯 Benefícios

### 🔒 Segurança
- **Identidade única**: Cada máquina tem identidade distinta
- **Vinculação física**: Impossível transferir identidade
- **Autenticação robusta**: PIN baseado em criptografia
- **Proteção contra ataques**: Múltiplas camadas de segurança

### 🚀 Usabilidade
- **Configuração automática**: Sem necessidade de configuração manual
- **Interface intuitiva**: Processo simples e claro
- **Feedback visual**: Informações claras sobre o processo
- **Persistência**: Configuração mantida entre execuções

### 🔧 Manutenibilidade
- **Código modular**: Funções bem definidas e separadas
- **Documentação clara**: Comentários explicativos
- **Tratamento de erros**: Validações e mensagens de erro
- **Logs informativos**: Rastreamento de operações

## 🚀 Próximos Passos

### 🔮 Melhorias Futuras
1. **Criptografia adicional**: Criptografar arquivo de identidade
2. **Backup de identidade**: Sistema de backup da identidade
3. **Validação de hardware**: Verificação periódica de integridade
4. **Auditoria**: Logs detalhados de autenticação
5. **Integração com blockchain**: Registrar identidades na blockchain

### 🔧 Otimizações
1. **Cache de PINs**: Melhorar performance de validação
2. **Compressão**: Comprimir dados de identidade
3. **Validação assíncrona**: Validação em background
4. **Interface melhorada**: Mais feedback visual

## 📝 Conclusão

O sistema de MinerID imutável por máquina implementa uma **camada física de segurança robusta** que:

- ✅ **Garante unicidade** de cada minerador
- ✅ **Previne transferência** de identidades
- ✅ **Fornece autenticação segura** via PIN
- ✅ **Mantém simplicidade** para o usuário
- ✅ **Oferece persistência** entre execuções

Esta implementação cria uma **base sólida** para a distribuição segura do minerador offline, garantindo que cada máquina tenha uma identidade única e imutável que pode ser usada para autenticação e rastreamento na rede blockchain.

