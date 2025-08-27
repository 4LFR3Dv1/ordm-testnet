# 🏗️ Blockchain 2-Layer - Sistema Completo

## 📋 Visão Geral

Uma **blockchain 2-layer avançada** com arquitetura híbrida PoW/PoS, sistema de autenticação seguro, wallets criptográficas, contratos inteligentes, e interface completa para investidores. O sistema evoluiu de uma simulação básica para uma solução empresarial completa.

## 🎯 Características Principais

### ✅ **Arquitetura 2-Layer**
- **Layer 1 (PoW)**: Mineração com DAG e halving automático
- **Layer 2 (PoS)**: Validação com staking e recompensas dinâmicas
- **Consenso Híbrido**: Combinação de Proof of Work e Proof of Stake

### ✅ **Sistema de Autenticação Seguro**
- **Login 2FA**: PIN único por wallet com validade de 10 segundos
- **Wallets Criptográficas**: Chaves públicas/privadas com BIP-39
- **Autenticação por Node**: RSA 2048-bit + AES-256
- **Sistema de Lockout**: Proteção contra ataques de força bruta

### ✅ **Tokenomics Sustentável**
- **Supply Máximo**: 21M tokens (como Bitcoin)
- **Halving**: A cada 210k blocos
- **Queima de Tokens**: 10% das taxas de transação
- **Stake APY**: 5% base + 2% bônus para validators
- **Recompensa Inicial**: 50 tokens por bloco

### ✅ **Interface Completa**
- **Node Minerador**: Dashboard para controle de mineração
- **Blockchain Explorer**: Interface pública tipo Etherscan
- **Sistema de Wallets**: Gerenciamento completo de carteiras
- **Transferências P2P**: Envio seguro de tokens
- **Monitoramento**: Custos, lucratividade e estatísticas

## 🏛️ Arquitetura do Sistema

### **Componentes Principais**

```
📁 cmd/
├── 🖥️ gui/          # Interface do node minerador (porta 3000)
├── 🔍 explorer/     # Blockchain explorer público (porta 8080)
├── 🗄️ backend/      # Servidor backend global
└── 🔗 node/         # Node básico da rede

📁 pkg/
├── 🔐 auth/         # Sistema de autenticação e 2FA
├── 💰 wallet/       # Gerenciamento de wallets (BIP-39)
├── ⛏️ pow/          # Proof of Work (mineração)
├── 🏆 consensus/    # Consenso híbrido PoW/PoS
├── 🌐 p2p/          # Rede peer-to-peer
├── 📊 ledger/       # Ledger global de transações
├── 💸 economics/    # Tokenomics e recompensas
├── 📝 logger/       # Logs estruturados em JSON
├── 🔧 api/          # API REST pública
├── 🗃️ storage/      # Persistência com BadgerDB
├── 📋 blockchain/   # Estrutura de blocos e DAG
├── 🔒 crypto/       # Criptografia e segurança
├── 📊 audit/        # Sistema de auditoria
└── 📜 contracts/    # Contratos inteligentes básicos
```

### **Fluxo de Operações**

```
1. 🔐 Autenticação
   ├── Geração de wallet com chaves únicas
   ├── Login com PIN 2FA (10s de validade)
   └── Verificação de permissões

2. ⛏️ Mineração (Layer 1)
   ├── Resolução de puzzle PoW
   ├── Criação de blocos com DAG
   ├── Recompensas com halving automático
   └── Registro no ledger global

3. 🏆 Validação (Layer 2)
   ├── Stake mínimo de 1000 tokens
   ├── Validação de transações
   ├── Recompensas adicionais (2%)
   └── Participação na governança

4. 💸 Transações
   ├── Assinatura criptográfica
   ├── Verificação de saldo
   ├── Taxas com queima (10%)
   └── Registro imutável
```

## 🚀 Como Usar

### **1. Pré-requisitos**
```bash
# Go 1.25+ instalado
go version

# Dependências
go mod tidy
```

### **2. Compilar o Sistema**
```bash
# Compilar node minerador
go build -o blockchain-gui-mac ./cmd/gui

# Compilar explorer
go build -o blockchain-explorer ./cmd/explorer

# Compilar backend (opcional)
go build -o blockchain-backend ./cmd/backend
```

### **3. Executar o Sistema**

#### **Node Minerador (Interface Principal)**
```bash
./blockchain-gui-mac
# Acesse: http://localhost:3000
```

#### **Blockchain Explorer (Público)**
```bash
./blockchain-explorer
# Acesse: http://localhost:8080
```

### **4. Primeiro Acesso**

#### **Criar Wallet**
1. Acesse `http://localhost:3000`
2. Clique em "🔐 Login Avançado"
3. Clique em "💼 Criar Nova Wallet"
4. Anote o **Public Key** e **PIN** gerados
5. Use essas credenciais para fazer login

#### **Fazer Login**
1. Digite o **Public Key** da wallet
2. Digite o **PIN** único (válido por 10s)
3. Clique em "🔐 Acessar Wallet"
4. Dashboard será exibido com todas as funcionalidades

## 💰 Sistema Econômico

### **Recompensas de Mineração**
- **Recompensa Base**: 50 tokens por bloco
- **Halving**: A cada 210k blocos (como Bitcoin)
- **Taxas de Transação**: Adicionadas à recompensa
- **Queima**: 10% das taxas são queimadas

### **Sistema de Stake**
- **Stake Mínimo**: 1000 tokens para ser validator
- **APY Base**: 5% para stakers
- **Bônus Validator**: +2% adicional
- **Total APY**: 7% para validators

### **Exemplo de Recompensa**
```
Wallet com 9000 tokens em stake:
├── Recompensa base: 450 tokens/ano (5%)
├── Bônus validator: 180 tokens/ano (2%)
└── Total: 630 tokens/ano (7% APY)
```

## 🔐 Segurança

### **Autenticação 2FA**
- **PIN Único**: Gerado por wallet (não global)
- **Validade**: 10 segundos
- **Tentativas**: Máximo 3 por wallet
- **Lockout**: 5 minutos após exceder tentativas

### **Wallets Criptográficas**
- **Algoritmo**: ECDSA com curva P-256
- **Endereços**: 40 caracteres hexadecimais
- **Armazenamento**: Criptografado localmente
- **Backup**: Recuperação segura

### **Proteções**
- **Rate Limiting**: 100 requisições/minuto por IP
- **Validação**: Verificação completa de transações
- **Auditoria**: Logs estruturados de todas as operações
- **Imutabilidade**: Dados não podem ser alterados

## 📊 Funcionalidades da Interface

### **Dashboard Principal**
- **Status da Mineração**: Iniciar/parar mineração
- **Estatísticas**: Blocos minerados, recompensas, custos
- **Wallet**: Saldo, transferências, histórico
- **Stake**: Adicionar stake, evoluir para validator
- **Logs**: Histórico completo de operações

### **Sistema de Transferências**
- **Envio P2P**: Transferências entre wallets
- **Validação**: Verificação de saldo e assinatura
- **Taxas**: Cálculo automático com queima
- **Confirmação**: Processamento instantâneo

### **Monitoramento Financeiro**
- **Custos de Energia**: Cálculo automático ($0.12/kWh)
- **Lucratividade**: ROI em tempo real
- **Hash Rate**: Performance de mineração
- **Supply Total**: Tokens em circulação

## 🌐 Blockchain Explorer

### **Funcionalidades Públicas**
- **Blocos**: Lista completa com detalhes
- **Transações**: Histórico de todas as movimentações
- **Wallets**: Saldos e estatísticas públicas
- **Estatísticas**: Métricas da rede em tempo real

### **Acesso**
- **URL**: `http://localhost:8080`
- **Público**: Qualquer pessoa pode acessar
- **Tempo Real**: Atualização automática a cada 5s
- **API**: Endpoints REST para integração

## 🔧 API REST

### **Endpoints Principais**
```bash
# Saúde do sistema
GET /api/health

# Blocos
GET /api/blocks?page=1&limit=20

# Transações
GET /api/transactions?limit=50

# Wallets
GET /api/wallets

# Estatísticas
GET /api/stats

# Explorer
GET /api/explorer
```

### **Exemplo de Uso**
```bash
# Verificar saúde
curl http://localhost:3000/api/health

# Listar blocos recentes
curl http://localhost:3000/api/blocks?limit=10

# Ver transações
curl http://localhost:3000/api/transactions?limit=20
```

## 📈 Métricas Atuais

### **Dados do Sistema**
- **Blocos Minerados**: 678+ blocos
- **Supply Total**: ~33,900 tokens
- **Wallets Ativas**: 20+ wallets
- **Transações**: 440+ movimentações
- **Stake Total**: 9,000 tokens (validator)

### **Performance**
- **Tempo por Bloco**: ~10 segundos
- **Hash Rate**: 338+ blocos/hora
- **Uptime**: Sistema estável
- **Latência**: < 100ms para APIs

## 🛠️ Configuração Avançada

### **Parâmetros de Mineração**
```json
{
  "difficulty": 2,
  "energy_cost": 0.12,
  "stake_minimum": 1000,
  "validator_bonus": 0.02
}
```

### **Tokenomics**
```json
{
  "initial_reward": 50,
  "halving_interval": 210000,
  "max_supply": 21000000,
  "burn_rate": 0.1,
  "stake_apy": 0.05
}
```

## 🔍 Troubleshooting

### **Problemas Comuns**

#### **Login não funciona**
- Verificar se o PIN está correto
- Aguardar geração de novo PIN
- Verificar se a wallet existe

#### **Mineração para**
- Verificar se o node está ativo
- Reiniciar o sistema
- Verificar logs de erro

#### **Transferência falha**
- Verificar saldo suficiente
- Confirmar endereço de destino
- Verificar se a wallet está ativa

### **Logs e Debug**
```bash
# Ver logs em tempo real
tail -f logs/blockchain.json

# Filtrar por nível
jq 'select(.level == "ERROR")' logs/blockchain.json

# Ver estado da mineração
cat data/mining_state.json
```

## 🚀 Próximas Melhorias

### **Fase 3 - Funcionalidades Avançadas**
- [ ] **Smart Contracts**: Contratos Turing-complete
- [ ] **DeFi Protocols**: Lending, staking pools
- [ ] **Cross-chain Bridges**: Integração com outras blockchains
- [ ] **Zero-Knowledge Proofs**: Privacidade avançada
- [ ] **Governança Descentralizada**: DAO para decisões

### **Melhorias de Performance**
- [ ] **Sharding**: Escalabilidade horizontal
- [ ] **Layer 2 Solutions**: Rollups e sidechains
- [ ] **Optimizations**: Melhorias de throughput
- [ ] **Mobile App**: Interface mobile nativa

## 📞 Suporte

### **Recursos**
- **Documentação**: Este README
- **Logs**: Sistema de logs detalhado
- **Explorer**: Interface pública para debug
- **API**: Endpoints para integração

### **Contato**
- **Issues**: GitHub Issues
- **Documentação**: Arquivos .md no projeto
- **Comunidade**: Fórum de usuários

## 📄 Licença

Este projeto é desenvolvido como uma **prova de conceito** de blockchain 2-layer. Use para fins educacionais e de desenvolvimento.

---

## 🎉 Resumo

Esta **blockchain 2-layer** evoluiu de uma simulação básica para uma **solução empresarial completa** com:

✅ **Arquitetura híbrida PoW/PoS** real e funcional  
✅ **Sistema de autenticação 2FA** com PIN único por wallet  
✅ **Tokenomics sustentável** com halving e queima de tokens  
✅ **Interface completa** para investidores e usuários  
✅ **Blockchain explorer** público tipo Etherscan  
✅ **API REST** para integração externa  
✅ **Sistema de auditoria** completo  
✅ **Segurança de nível bancário**  

**🚀 Transforme-se de minerador em validator e participe da governança descentralizada!**
