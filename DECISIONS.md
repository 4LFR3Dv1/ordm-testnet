# 🎯 Decisões Arquiteturais ORDM

## 📋 Visão Geral

Este documento registra as **decisões arquiteturais importantes** tomadas durante o desenvolvimento do ORDM Blockchain 2-Layer, incluindo justificativas e alternativas consideradas.

---

## 🏗️ **Decisões de Arquitetura**

### **1. Separação Offline/Online**
- **Decisão**: Mineração offline, validação online
- **Data**: Janeiro 2025
- **Justificativa**: Permite mineração sem dependência de rede
- **Alternativas Consideradas**:
  - Mineração totalmente online (rejeitado: dependência de rede)
  - Mineração híbrida (rejeitado: complexidade excessiva)
- **Benefícios**: Escalabilidade, independência, eficiência
- **Impacto**: Alto - define a arquitetura principal

### **2. Consenso Híbrido PoW/PoS**
- **Decisão**: PoW para mineração, PoS para validação
- **Data**: Janeiro 2025
- **Justificativa**: Combina segurança do PoW com eficiência do PoS
- **Alternativas Consideradas**:
  - Apenas PoW (rejeitado: ineficiente)
  - Apenas PoS (rejeitado: menos seguro)
  - DPoS (rejeitado: centralização)
- **Benefícios**: Segurança, eficiência, descentralização
- **Impacto**: Alto - define o modelo de consenso

### **3. Storage Local Criptografado**
- **Decisão**: BadgerDB local com criptografia AES-256
- **Data**: Janeiro 2025
- **Justificativa**: Performance e segurança para dados sensíveis
- **Alternativas Consideradas**:
  - SQLite (rejeitado: performance inferior)
  - PostgreSQL (rejeitado: complexidade)
  - JSON files (rejeitado: não escalável)
- **Benefícios**: Velocidade, segurança, privacidade
- **Impacto**: Médio - afeta performance e segurança

### **4. Autenticação 2FA**
- **Decisão**: PIN único por wallet com validade temporal
- **Data**: Janeiro 2025
- **Justificativa**: Segurança sem complexidade excessiva
- **Alternativas Consideradas**:
  - TOTP (rejeitado: complexidade para usuários)
  - SMS (rejeitado: dependência externa)
  - Hardware tokens (rejeitado: custo)
- **Benefícios**: Segurança, usabilidade, controle
- **Impacto**: Médio - afeta experiência do usuário

---

## 🔐 **Decisões de Segurança**

### **5. Criptografia Ed25519**
- **Decisão**: Usar Ed25519 para assinaturas digitais
- **Data**: Janeiro 2025
- **Justificativa**: Performance superior e segurança comprovada
- **Alternativas Consideradas**:
  - ECDSA (rejeitado: performance inferior)
  - RSA (rejeitado: chaves grandes)
- **Benefícios**: Performance, segurança, tamanho de chave
- **Impacto**: Alto - afeta toda a segurança

### **6. Wallets BIP-39**
- **Decisão**: Implementar padrão BIP-39 para wallets
- **Data**: Janeiro 2025
- **Justificativa**: Padrão industrial e compatibilidade
- **Alternativas Consideradas**:
  - Wallets customizadas (rejeitado: incompatibilidade)
  - BIP-32 apenas (rejeitado: menos funcional)
- **Benefícios**: Compatibilidade, padrão, segurança
- **Impacto**: Médio - afeta interoperabilidade

### **7. Rate Limiting**
- **Decisão**: 100 requisições/minuto por IP
- **Data**: Janeiro 2025
- **Justificativa**: Proteção contra ataques DDoS
- **Alternativas Consideradas**:
  - Sem limite (rejeitado: vulnerável)
  - Limite por wallet (rejeitado: complexo)
- **Benefícios**: Proteção, estabilidade
- **Impacto**: Baixo - afeta apenas segurança

---

## 🌐 **Decisões de Rede**

### **8. Protocolo P2P libp2p**
- **Decisão**: Usar libp2p para comunicação P2P
- **Data**: Janeiro 2025
- **Justificativa**: Biblioteca madura e funcionalidades avançadas
- **Alternativas Consideradas**:
  - WebRTC (rejeitado: complexidade)
  - TCP direto (rejeitado: funcionalidades limitadas)
- **Benefícios**: Funcionalidades, maturidade, comunidade
- **Impacto**: Alto - define comunicação de rede

### **9. Seed Nodes**
- **Decisão**: Seed nodes para descoberta de peers
- **Data**: Janeiro 2025
- **Justificativa**: Descoberta automática de peers
- **Alternativas Consideradas**:
  - DNS seeds (rejeitado: dependência externa)
  - Lista estática (rejeitado: não escalável)
- **Benefícios**: Descoberta automática, escalabilidade
- **Impacto**: Médio - afeta conectividade

### **10. Sincronização Assíncrona**
- **Decisão**: Sincronização assíncrona de blocos
- **Data**: Janeiro 2025
- **Justificativa**: Performance e tolerância a falhas
- **Alternativas Consideradas**:
  - Síncrona (rejeitado: performance)
  - Batch (rejeitado: complexidade)
- **Benefícios**: Performance, tolerância a falhas
- **Impacto**: Médio - afeta performance

---

## 💰 **Decisões Econômicas**

### **11. Tokenomics Bitcoin-like**
- **Decisão**: Supply máximo de 21M tokens com halving
- **Data**: Janeiro 2025
- **Justificativa**: Modelo comprovado e deflacionário
- **Alternativas Consideradas**:
  - Supply infinito (rejeitado: inflação)
  - Supply fixo (rejeitado: sem incentivos)
- **Benefícios**: Escassez, incentivos, previsibilidade
- **Impacto**: Alto - define economia

### **12. Stake Mínimo 1000 Tokens**
- **Decisão**: Stake mínimo de 1000 tokens para validadores
- **Data**: Janeiro 2025
- **Justificativa**: Balancear acessibilidade e segurança
- **Alternativas Consideradas**:
  - 100 tokens (rejeitado: muito baixo)
  - 10000 tokens (rejeitado: muito alto)
- **Benefícios**: Acessibilidade, segurança
- **Impacto**: Médio - afeta participação

### **13. APY 5% + 2% Bônus**
- **Decisão**: 5% APY base + 2% bônus para validadores
- **Data**: Janeiro 2025
- **Justificativa**: Incentivos atrativos sem inflação excessiva
- **Alternativas Consideradas**:
  - 10% total (rejeitado: inflação alta)
  - 3% total (rejeitado: pouco atrativo)
- **Benefícios**: Incentivos, sustentabilidade
- **Impacto**: Médio - afeta adoção

---

## 📊 **Decisões de Performance**

### **14. BadgerDB para Storage**
- **Decisão**: BadgerDB como database principal
- **Data**: Janeiro 2025
- **Justificativa**: Performance superior para workloads de blockchain
- **Alternativas Consideradas**:
  - LevelDB (rejeitado: performance inferior)
  - RocksDB (rejeitado: complexidade)
- **Benefícios**: Performance, simplicidade
- **Impacto**: Alto - afeta performance geral

### **15. Compressão de Dados**
- **Decisão**: Compressão automática de dados históricos
- **Data**: Janeiro 2025
- **Justificativa**: Reduzir uso de storage
- **Alternativas Consideradas**:
  - Sem compressão (rejeitado: uso excessivo de storage)
  - Compressão manual (rejeitado: complexidade)
- **Benefícios**: Eficiência de storage
- **Impacto**: Baixo - afeta apenas storage

---

## 🔧 **Decisões de Implementação**

### **16. Linguagem Go**
- **Decisão**: Implementar em Go
- **Data**: Janeiro 2025
- **Justificativa**: Performance, concorrência, comunidade
- **Alternativas Consideradas**:
  - Rust (rejeitado: curva de aprendizado)
  - C++ (rejeitado: complexidade)
  - Node.js (rejeitado: performance)
- **Benefícios**: Performance, simplicidade, comunidade
- **Impacto**: Alto - define tecnologia base

### **17. API REST**
- **Decisão**: API REST para integração
- **Data**: Janeiro 2025
- **Justificativa**: Simplicidade e compatibilidade
- **Alternativas Consideradas**:
  - GraphQL (rejeitado: complexidade)
  - gRPC (rejeitado: complexidade)
- **Benefícios**: Simplicidade, compatibilidade
- **Impacto**: Médio - afeta integração

### **18. Docker para Deploy**
- **Decisão**: Usar Docker para deploy
- **Data**: Janeiro 2025
- **Justificativa**: Portabilidade e consistência
- **Alternativas Consideradas**:
  - Deploy direto (rejeitado: inconsistência)
  - VMs (rejeitado: overhead)
- **Benefícios**: Portabilidade, consistência
- **Impacto**: Baixo - afeta apenas deploy

---

## 📈 **Decisões Futuras (Pendentes)**

### **19. Layer 2 Solutions**
- **Status**: Pendente
- **Alternativas**: Rollups, Sidechains, State Channels
- **Critérios**: Escalabilidade, segurança, compatibilidade

### **20. Cross-chain Bridges**
- **Status**: Pendente
- **Alternativas**: Atomic swaps, Relays, Validators
- **Critérios**: Segurança, eficiência, compatibilidade

### **21. Smart Contracts Turing-complete**
- **Status**: Pendente
- **Alternativas**: EVM, WASM, Custom VM
- **Critérios**: Performance, segurança, flexibilidade

---

## 📋 **Processo de Decisão**

### **Critérios de Avaliação**
1. **Segurança**: Impacto na segurança do sistema
2. **Performance**: Impacto na performance
3. **Escalabilidade**: Capacidade de crescer
4. **Usabilidade**: Facilidade de uso
5. **Manutenibilidade**: Facilidade de manutenção

### **Processo**
1. **Identificação** do problema/oportunidade
2. **Análise** de alternativas
3. **Avaliação** contra critérios
4. **Decisão** e documentação
5. **Implementação** e validação

---

**📝 Este documento deve ser atualizado sempre que uma nova decisão arquitetural importante for tomada.**
