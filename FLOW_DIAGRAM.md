# 🔄 Diagrama de Fluxo ORDM Blockchain

## 📋 Visão Geral

Este documento apresenta os **diagramas de fluxo consolidados** do sistema ORDM Blockchain 2-Layer, mostrando o fluxo de dados e operações principais.

---

## 🏗️ **Fluxo Principal do Sistema**

### **Arquitetura 2-Layer**
```mermaid
graph TB
    subgraph "Layer 1: Mineração Offline"
        A[Minerador] --> B[PoW Mining]
        B --> C[Block Creation]
        C --> D[Local Storage]
    end
    
    subgraph "Layer 2: Validação Online"
        E[Seed Nodes] --> F[PoS Validation]
        F --> G[Consensus]
        G --> H[Global Ledger]
    end
    
    D --> E
    H --> I[Explorer]
    H --> J[API]
    
    style A fill:#ff9999
    style E fill:#99ccff
    style H fill:#99ff99
```

---

## 🔐 **Fluxo de Autenticação**

### **Processo de Login 2FA**
```mermaid
sequenceDiagram
    participant U as User
    participant I as Interface
    participant A as Auth System
    participant W as Wallet
    participant B as Blockchain
    
    U->>I: Acessar Interface
    I->>A: Solicitar PIN
    A->>A: Gerar PIN (60s)
    A->>I: Retornar PIN
    I->>U: Mostrar PIN
    U->>I: Inserir PIN
    I->>A: Validar PIN
    A->>A: Verificar TTL
    A->>W: Autenticar Wallet
    W->>B: Verificar Saldo
    B->>W: Retornar Saldo
    W->>A: Confirmar Autenticação
    A->>I: Criar Sessão
    I->>U: Dashboard Acessível
```

---

## ⛏️ **Fluxo de Mineração**

### **Processo de Mineração PoW**
```mermaid
flowchart TD
    A[Iniciar Mineração] --> B[Carregar Estado]
    B --> C[Verificar Dificuldade]
    C --> D[Resolver Puzzle PoW]
    D --> E{Hash Válido?}
    E -->|Não| D
    E -->|Sim| F[Criar Bloco]
    F --> G[Assinar Bloco]
    G --> H[Adicionar Transações]
    H --> I[Validar Bloco]
    I --> J[Salvar Localmente]
    J --> K[Adicionar à Fila de Sync]
    K --> L[Continuar Mineração]
    L --> D
    
    style A fill:#ff9999
    style F fill:#99ff99
    style J fill:#99ccff
```

---

## 🔄 **Fluxo de Sincronização**

### **Offline para Online**
```mermaid
sequenceDiagram
    participant M as Minerador
    participant Q as Sync Queue
    participant S as Seed Node
    participant V as Validador
    participant L as Ledger
    
    M->>Q: Adicionar Bloco
    Q->>Q: Validar Assinatura
    Q->>S: Enviar Pacote
    S->>S: Verificar MinerID
    S->>V: Encaminhar para Validação
    V->>V: Verificar PoW
    V->>V: Validar Transações
    V->>L: Votar Aceitação
    L->>L: Consolidar Votos
    L->>L: Atualizar Estado
    L->>M: Confirmar Aceitação
    M->>M: Remover da Fila
```

---

## 💰 **Fluxo de Transações**

### **Processo de Transferência**
```mermaid
flowchart TD
    A[Iniciar Transferência] --> B[Verificar Saldo]
    B --> C{Saldo Suficiente?}
    C -->|Não| D[Erro: Saldo Insuficiente]
    C -->|Sim| E[Criar Transação]
    E --> F[Assinar Transação]
    F --> G[Calcular Taxas]
    G --> H[Validar Destino]
    H --> I[Adicionar ao Pool]
    I --> J[Aguardar Mineração]
    J --> K[Incluir no Bloco]
    K --> L[Confirmar Transação]
    L --> M[Atualizar Saldos]
    
    style A fill:#ff9999
    style L fill:#99ff99
    style M fill:#99ccff
```

---

## 🏆 **Fluxo de Validação PoS**

### **Processo de Stake e Validação**
```mermaid
flowchart TD
    A[Stake Tokens] --> B[Verificar Mínimo]
    B --> C{1000+ Tokens?}
    C -->|Não| D[Erro: Stake Insuficiente]
    C -->|Sim| E[Ativar Validador]
    E --> F[Aguardar Blocos]
    F --> G[Receber Bloco]
    G --> H[Verificar PoW]
    H --> I[Validar Transações]
    I --> J[Votar Aceitação]
    J --> K[Participar Consenso]
    K --> L[Receber Recompensas]
    L --> M[Atualizar Stake]
    
    style A fill:#ff9999
    style E fill:#99ff99
    style L fill:#99ccff
```

---

## 🔐 **Fluxo de Segurança**

### **Proteção de Chaves**
```mermaid
flowchart TD
    A[Gerar Chaves] --> B[Criar Mnemonic]
    B --> C[Derivar Chaves]
    C --> D[Criptografar com AES-256]
    D --> E[Salvar no Keystore]
    E --> F[Backup Automático]
    F --> G[Monitorar Acesso]
    G --> H{Detectar Intrusão?}
    H -->|Sim| I[Bloquear Acesso]
    H -->|Não| J[Continuar Operação]
    
    style A fill:#ff9999
    style D fill:#99ff99
    style I fill:#ffcc99
```

---

## 📊 **Fluxo de Monitoramento**

### **Coleta de Métricas**
```mermaid
flowchart TD
    A[Evento do Sistema] --> B[Logger]
    B --> C[Structured Logs]
    C --> D[Metrics Collector]
    D --> E[Prometheus]
    E --> F[Grafana Dashboard]
    F --> G[Alertas]
    G --> H{Threshold Exceeded?}
    H -->|Sim| I[Enviar Alerta]
    H -->|Não| J[Continuar Monitoramento]
    
    style A fill:#ff9999
    style E fill:#99ff99
    style I fill:#ffcc99
```

---

## 🌐 **Fluxo de Rede P2P**

### **Descoberta de Peers**
```mermaid
sequenceDiagram
    participant N as Node
    participant S as Seed Node
    participant P as Peer
    participant V as Validator
    
    N->>S: Conectar
    S->>S: Verificar Identidade
    S->>N: Lista de Peers
    N->>P: Handshake
    P->>P: Validar Node
    P->>N: Aceitar Conexão
    N->>V: Sincronizar Estado
    V->>N: Enviar Blockchain
    N->>N: Atualizar Estado Local
```

---

## 💾 **Fluxo de Storage**

### **Persistência de Dados**
```mermaid
flowchart TD
    A[Dados do Sistema] --> B[Validação]
    B --> C[Criptografia]
    C --> D[BadgerDB]
    D --> E[Compressão]
    E --> F[Backup Local]
    F --> G[Backup Remoto]
    G --> H[Verificação de Integridade]
    H --> I{Integridade OK?}
    I -->|Não| J[Restaurar Backup]
    I -->|Sim| K[Dados Persistidos]
    
    style A fill:#ff9999
    style D fill:#99ff99
    style K fill:#99ccff
```

---

## 🔄 **Fluxo de Recuperação**

### **Processo de Failover**
```mermaid
flowchart TD
    A[Detectar Falha] --> B[Isolar Componente]
    B --> C[Ativar Backup]
    C --> D[Verificar Integridade]
    D --> E{Backup Válido?}
    E -->|Não| F[Restaurar de Snapshot]
    E -->|Sim| G[Restaurar Backup]
    F --> H[Verificar Sistema]
    G --> H
    H --> I{Sistema OK?}
    I -->|Não| J[Alertar Administrador]
    I -->|Sim| K[Sistema Recuperado]
    
    style A fill:#ff9999
    style C fill:#99ff99
    style K fill:#99ccff
```

---

## 📈 **Fluxo de Performance**

### **Otimização de Recursos**
```mermaid
flowchart TD
    A[Monitorar Performance] --> B[Coletar Métricas]
    B --> C[Analisar Tendências]
    C --> D{Performance OK?}
    D -->|Não| E[Identificar Gargalo]
    D -->|Sim| F[Continuar Monitoramento]
    E --> G[Otimizar Recursos]
    G --> H[Testar Melhorias]
    H --> I{Melhoria Efetiva?}
    I -->|Não| G
    I -->|Sim| J[Implementar Produção]
    
    style A fill:#ff9999
    style E fill:#99ff99
    style J fill:#99ccff
```

---

## 🎯 **Resumo dos Fluxos**

### **Fluxos Principais**
1. **Autenticação**: Login 2FA com PIN temporal
2. **Mineração**: PoW offline com sincronização
3. **Validação**: PoS online com stake
4. **Transações**: Transferências P2P seguras
5. **Storage**: Persistência criptografada
6. **Monitoramento**: Métricas em tempo real

### **Pontos de Controle**
- **Segurança**: Autenticação em cada etapa
- **Performance**: Monitoramento contínuo
- **Confiabilidade**: Backup e recuperação
- **Escalabilidade**: Arquitetura distribuída

---

**🔄 Estes diagramas representam o fluxo consolidado e otimizado do sistema ORDM Blockchain 2-Layer.**
