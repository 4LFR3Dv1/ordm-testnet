# 📦 Dependências ORDM Blockchain

## 📋 Visão Geral

Este documento mapeia as **dependências entre componentes** do sistema ORDM Blockchain 2-Layer, identificando relações críticas e pontos de falha.

---

## 🏗️ **Dependências Arquiteturais**

### **Hierarquia de Componentes**
```
┌─────────────────┐
│   Interface     │ (cmd/gui)
│   Principal     │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Backend       │ (cmd/backend)
│   Global        │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Blockchain    │ (pkg/blockchain)
│   Core          │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Storage       │ (pkg/storage)
│   Layer         │
└─────────────────┘
```

### **Dependências Críticas**
```
Interface → Backend → Blockchain → Storage
     │           │           │         │
     ▼           ▼           ▼         ▼
  Auth      Consensus    Crypto    Database
     │           │           │         │
     ▼           ▼           ▼         ▼
  Wallet     Network     P2P      BadgerDB
```

---

## 🔐 **Dependências de Segurança**

### **Cadeia de Autenticação**
```
User Input → 2FA → Session → Wallet → Blockchain
     │         │      │        │          │
     ▼         ▼      ▼        ▼          ▼
  Validation  PIN   Token   Keys     Signature
     │         │      │        │          │
     ▼         ▼      ▼        ▼          ▼
  Rate Limit  TTL   JWT    Ed25519   Verification
```

### **Dependências Criptográficas**
```
Wallet Generation → BIP-39 → Ed25519 → Blockchain
       │              │         │          │
       ▼              ▼         ▼          ▼
   Mnemonics     Entropy    Keys     Signatures
       │              │         │          │
       ▼              ▼         ▼          ▼
   BIP-44       Random     Public    Verification
```

---

## 🌐 **Dependências de Rede**

### **P2P Network Stack**
```
Application → libp2p → Transport → Network
     │           │         │          │
     ▼           ▼         ▼          ▼
  Messages    Protocol   TCP/UDP   Internet
     │           │         │          │
     ▼           ▼         ▼          ▼
  PubSub      Noise     WebRTC    Peers
```

### **Seed Node Dependencies**
```
Seed Node → Discovery → Peers → Validation
    │           │         │         │
    ▼           ▼         ▼         ▼
  libp2p    DNS Seeds  Connect   Consensus
    │           │         │         │
    ▼           ▼         ▼         ▼
  Protocol   Bootstrap  Handshake  PoS
```

---

## 💾 **Dependências de Storage**

### **Data Flow**
```
Application → Storage Interface → BadgerDB → Filesystem
     │               │               │           │
     ▼               ▼               ▼           ▼
  Requests       Abstraction      Engine      OS Layer
     │               │               │           │
     ▼               ▼               ▼           ▼
  Validation     Encryption     Compression   Persistence
```

### **Backup Dependencies**
```
Primary Storage → Backup Service → Cloud Storage
       │               │               │
       ▼               ▼               ▼
   BadgerDB        Compression     S3/Cloud
       │               │               │
       ▼               ▼               ▼
   Encryption       Encryption     Redundancy
```

---

## 🔄 **Dependências de Sincronização**

### **Offline to Online Sync**
```
Offline Miner → Block Queue → Network → Online Validator
      │             │           │            │
      ▼             ▼           ▼            ▼
   PoW Mining    Local DB    libp2p      PoS Validation
      │             │           │            │
      ▼             ▼           ▼            ▼
   Block Creation  Persistence  Transmission  Consensus
```

### **State Synchronization**
```
Local State → Sync Protocol → Global State → Ledger
     │             │              │            │
     ▼             ▼              ▼            ▼
  BadgerDB     HTTP/JSON      Consensus    Database
     │             │              │            │
     ▼             ▼              ▼            ▼
  Encryption   Authentication   Validation   Persistence
```

---

## 💰 **Dependências Econômicas**

### **Reward System**
```
Mining → Block Creation → Reward Calculation → Distribution
  │           │                   │                │
  ▼           ▼                   ▼                ▼
PoW        Validation         Tokenomics       Wallets
  │           │                   │                │
  ▼           ▼                   ▼                ▼
Difficulty  Consensus         Halving         Stake APY
```

### **Stake Management**
```
Validator → Stake Verification → Consensus → Rewards
    │              │                │           │
    ▼              ▼                ▼           ▼
  Wallet       Minimum Stake    PoS Voting   APY Calc
    │              │                │           │
    ▼              ▼                ▼           ▼
  Balance      1000 Tokens     2/3 Majority   5% + 2%
```

---

## 🔧 **Dependências de Implementação**

### **Build Dependencies**
```
Source Code → Go Modules → Dependencies → Binary
     │            │            │           │
     ▼            ▼            ▼           ▼
  .go files   go.mod/go.sum   External   Executable
     │            │            │           │
     ▼            ▼            ▼           ▼
  Compilation  Versioning   Vendoring   Distribution
```

### **Runtime Dependencies**
```
Binary → Runtime → OS → Hardware
   │        │      │      │
   ▼        ▼      ▼      ▼
Go Runtime  libc  Kernel  CPU/RAM
   │        │      │      │
   ▼        ▼      ▼      ▼
Garbage    System  Drivers  Resources
Collection  Calls
```

---

## 📊 **Dependências de Monitoramento**

### **Metrics Collection**
```
Application → Metrics → Prometheus → Dashboard
     │           │           │           │
     ▼           ▼           ▼           ▼
  Events      Counters    Exposition   Visualization
     │           │           │           │
     ▼           ▼           ▼           ▼
  Logging     Gauges      HTTP/Text   Grafana
```

### **Health Checks**
```
Service → Health Check → Load Balancer → Client
   │           │               │           │
   ▼           ▼               ▼           ▼
Status     Endpoint        Routing      Response
   │           │               │           │
   ▼           ▼               ▼           ▼
Database   /health         Failover    Success/Error
```

---

## ⚠️ **Pontos de Falha Críticos**

### **Single Points of Failure**
1. **Database**: BadgerDB é único ponto de falha para dados
2. **Network**: libp2p dependency pode falhar
3. **Crypto**: Ed25519 library é crítica para segurança
4. **Storage**: Filesystem dependency para persistência

### **Cascade Failures**
```
Database Failure → Storage → Blockchain → Backend → Interface
     │                │           │           │         │
     ▼                ▼           ▼           ▼         ▼
  Data Loss      No Persistence  No Blocks  No API   No UI
```

---

## 🔄 **Dependências Circulares**

### **Detected Circular Dependencies**
```
Auth ↔ Wallet ↔ Blockchain ↔ Storage
 │       │         │           │
 └───────┴─────────┴───────────┘
```

### **Resolution Strategy**
- **Interface Segregation**: Separar interfaces
- **Dependency Injection**: Inverter dependências
- **Event-Driven**: Usar eventos para comunicação

---

## 📈 **Dependências Futuras**

### **Planned Dependencies**
1. **Layer 2**: Rollups e sidechains
2. **Cross-chain**: Bridges para outras blockchains
3. **Smart Contracts**: EVM ou WASM
4. **Zero-Knowledge**: zk-SNARKs/zk-STARKs

### **Dependency Management**
- **Version Pinning**: Fixar versões críticas
- **Vendoring**: Incluir dependências no repo
- **Audit Regular**: Verificar vulnerabilidades
- **Minimal Dependencies**: Manter apenas essenciais

---

## 🛠️ **Como Gerenciar Dependências**

### **Comandos Úteis**
```bash
# Ver dependências
go mod graph

# Atualizar dependências
go get -u all

# Verificar vulnerabilidades
go list -json -deps ./... | nancy sleuth

# Vendor dependências
go mod vendor

# Limpar cache
go clean -modcache
```

### **Boas Práticas**
1. **Pin Versions**: Usar versões específicas
2. **Regular Updates**: Atualizar mensalmente
3. **Security Audits**: Verificar vulnerabilidades
4. **Documentation**: Manter este arquivo atualizado

---

**📝 Este documento deve ser atualizado sempre que novas dependências forem adicionadas ou removidas.**
