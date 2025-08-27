# ⛏️ Node Minerador - Blockchain 2-Layer

## 📱 Visão Geral

Interface gráfica para **node minerador individual** da blockchain 2-layer. Desenvolvida para **investidores** que desejam participar da mineração com controle de custos, evolução para validador PoS, sistema de wallet completo e registro global de movimentações.

## 🎯 Propósito

### Para Investidores
- **Mineração Individual**: Cada investidor controla seu próprio node minerador
- **Sistema de Wallet**: Chaves privadas seguras e endereços únicos
- **Transferências**: Envio de tokens entre endereços
- **Controle de Custos**: Monitoramento de gastos com energia e lucratividade
- **Evolução PoS**: Sistema de stake para evoluir de minerador para validador
- **Registro Completo**: Histórico de blocos minerados, recompensas e custos

### Sistema Descentralizado
- **Ente Regulador**: Interface funciona como regulador descentralizado
- **Ledger Global**: Armazenamento global de saldos e movimentações
- **Transparência**: Todos os dados de mineração são registrados
- **Auditoria**: Sistema de logs para auditoria completa

## 🚀 Como Usar

### 1. Executar o Node Minerador

**Linux/Mac:**
```bash
./blockchain-gui-mac
```

**Windows:**
```bash
blockchain-gui.exe
```

### 2. Acessar o Dashboard

Abra seu navegador e acesse:
```
http://localhost:3000
```

## 🎮 Funcionalidades

### Controles de Mineração
- **🚀 Iniciar Node**: Conecta o node à rede blockchain
- **🛑 Parar Node**: Desconecta o node da rede
- **⛏️ Iniciar Mineração**: Começa a minerar blocos
- **⏸️ Parar Mineração**: Para a mineração

### Sistema de Wallet
- **💼 Criar Wallet**: Gera nova wallet com chave privada
- **🔐 Chaves Privadas**: Armazenamento seguro de chaves
- **📊 Saldo Global**: Visualização do saldo no ledger global
- **💸 Transferências**: Envio de tokens para outros endereços

### Sistema de Stake (PoS)
- **🔒 Adicionar Stake**: Investir moedas para evoluir para validador
- **Nível Validator**: Evolução de "Minerador" para "Validador PoS"
- **Stake Mínimo**: 1000 moedas para se tornar validador

### Monitoramento Financeiro
- **💰 Recompensas**: Total de moedas ganhas com mineração
- **⚡ Custo Energia**: Gastos com eletricidade
- **📊 Lucro/Prejuízo**: Cálculo automático de rentabilidade
- **📈 Hash Rate**: Performance de mineração

## 📊 Dashboard

### Seção Wallet
- **Endereço**: Endereço único da wallet
- **Saldo Global**: Saldo atual no ledger global
- **Transferências**: Interface para enviar tokens
- **Histórico**: Movimentações e gerações de tokens

### Estatísticas de Mineração
- **Blocos Minerados**: Contador de blocos encontrados
- **Recompensas**: Total de moedas ganhas (10 por bloco)
- **Custo Energia**: Calculado automaticamente ($0.12/kWh)
- **Lucro/Prejuízo**: Diferença entre recompensas e custos

### Sistema de Stake
- **Stake Atual**: Quantidade de moedas em stake
- **Nível Validator**: Status atual (Minerador/Validador PoS)
- **Evolução**: Interface para adicionar stake

### Logs do Sistema
- **Histórico Completo**: Todas as operações registradas
- **Timestamps**: Horário exato de cada evento
- **Auditoria**: Rastreabilidade completa

## 💰 Modelo Econômico

### Mineração (PoW)
- **Recompensa**: 10 moedas por bloco
- **Custo Energia**: $0.06 por bloco (0.5 kWh × $0.12)
- **Lucro Líquido**: ~$1.14 por bloco

### Validação (PoS)
- **Stake Mínimo**: 1000 moedas
- **Menor Consumo**: Validação mais eficiente
- **Recompensas Adicionais**: Por validação de transações

### Transferências
- **Taxa de Transação**: 1 token por transferência
- **Confirmação**: Processamento instantâneo
- **Segurança**: Assinatura criptográfica

## 🔧 Sistema de Wallet

### Geração de Chaves
- **Algoritmo**: ECDSA com curva P-256
- **Endereços**: 40 caracteres hexadecimais
- **Segurança**: Chaves privadas armazenadas localmente

### Armazenamento
- **Formato**: JSON criptografado
- **Localização**: `./wallets/`
- **Permissões**: 600 (apenas proprietário)

### Transações
- **Assinatura**: ECDSA com chave privada
- **Verificação**: Validação criptográfica
- **Nonce**: Prevenção de replay attacks

## 📊 Ledger Global

### Armazenamento
- **Formato**: JSON estruturado
- **Localização**: `./data/global_ledger.json`
- **Sincronização**: Tempo real

### Registros
- **Movimentações**: Todas as transferências
- **Gerações**: Tokens criados por mineração/stake
- **Timestamps**: Horário exato de cada operação
- **Block Hash**: Referência ao bloco da blockchain

### Estatísticas
- **Supply Total**: Total de tokens em circulação
- **Endereços Ativos**: Número de wallets
- **Movimentações**: Histórico completo
- **Top Holders**: Maiores detentores

## 🔧 Configuração

### Parâmetros de Mineração
- **Dificuldade**: Configurável (0-10)
- **Preço Energia**: Ajustável por região
- **Hash Rate**: Monitorado automaticamente

### Conexão com Rede
- **Porta Padrão**: 8080
- **Peers**: Conecta automaticamente com outros nodes
- **Sincronização**: Mantém blockchain atualizada

### Wallet
- **Diretório**: `./wallets/`
- **Backup**: Recomendado backup das chaves
- **Segurança**: Não compartilhar chaves privadas

## 📈 Análise de Rentabilidade

### Fatores Considerados
- **Preço da Moeda**: Valor de mercado das recompensas
- **Custo Energia**: Tarifa local de eletricidade
- **Hash Rate**: Performance do hardware
- **Dificuldade**: Competição na rede
- **Taxas de Transação**: Receita adicional

### Indicadores
- **ROI**: Retorno sobre investimento
- **Break-even**: Ponto de equilíbrio
- **Payback**: Tempo para recuperar investimento
- **Cash Flow**: Fluxo de caixa em tempo real

## 🛠️ Instalação

### Pré-requisitos
- Go 1.25+ instalado
- Acesso à internet
- Hardware adequado para mineração
- 1GB de espaço em disco

### Compilar
```bash
./build_gui.sh
```

### Instalar
```bash
./install.sh
```

## 🔧 Correções Implementadas

### ✅ **Mineração Contínua**
- **Problema**: Mineração parava após o primeiro bloco
- **Solução**: Implementada mineração contínua que roda a cada 2 segundos
- **Resultado**: Blocos minerados continuamente com logs detalhados

### ✅ **Atualizações em Tempo Real**
- **Problema**: Interface não atualizava em tempo real
- **Solução**: JavaScript atualiza dados a cada 2 segundos via `/status`
- **Resultado**: Saldo e estatísticas atualizados automaticamente

### ✅ **Persistência de Saldos**
- **Problema**: Saldos não eram mantidos após reinicialização
- **Solução**: Ledger global salvo automaticamente após cada operação
- **Resultado**: Saldos persistem corretamente no disco

### ✅ **Sistema de Stake Funcional**
- **Problema**: Stake não deduzia tokens do saldo
- **Solução**: Implementada função `ProcessStake()` que deduz tokens
- **Resultado**: Stake agora afeta o saldo global corretamente

### ✅ **Logs Detalhados**
- **Problema**: Logs não mostravam hash e timestamp dos blocos
- **Solução**: Logs agora incluem hash único e timestamp para cada operação
- **Resultado**: 
  - Blocos: `⛏️ Bloco #X minerado! Hash: xxx | Timestamp: xxx | +50 tokens | Saldo: xxx`
  - Transferências: `💸 Transferência processada! Hash: xxx | De: xxx | Para: xxx | Valor: xxx | Timestamp: xxx`

### ✅ **Correções de Concorrência**
- **Problema**: Deadlocks no ledger durante operações
- **Solução**: Removido `defer gl.Mutex.Unlock()` e implementado unlock manual
- **Resultado**: Sistema mais estável sem travamentos

## 🔍 Troubleshooting

### Problemas Comuns

**1. Node não conecta**
- Verificar conexão com internet
- Confirmar se porta 8080 está livre
- Verificar logs do sistema

**2. Mineração não lucrativa**
- Ajustar preço da energia
- Verificar dificuldade da rede
- Considerar upgrade de hardware

**3. Stake não evolui**
- Confirmar quantidade mínima (1000 moedas)
- Verificar se node está sincronizado
- Aguardar confirmação da rede

**4. Transferência falha**
- Verificar saldo suficiente
- Confirmar endereço de destino
- Verificar taxa de transação

**5. Wallet não carrega**
- Verificar permissões do diretório
- Confirmar integridade dos arquivos
- Fazer backup e recriar se necessário

## 📊 Métricas de Performance

### Indicadores Técnicos
- **Uptime**: Tempo de operação contínua
- **Hash Rate**: Blocos por segundo
- **Latência**: Tempo de resposta da rede
- **Sincronização**: Status da blockchain

### Indicadores Financeiros
- **ROI Diário**: Retorno sobre investimento
- **Custo por Bloco**: Gastos por bloco minerado
- **Lucro Líquido**: Receita menos custos
- **Taxa de Transação**: Receita por transferências

### Indicadores de Rede
- **Supply Total**: Tokens em circulação
- **Endereços Ativos**: Wallets ativas
- **Movimentações**: Volume de transações
- **Gerações**: Novos tokens criados

## 🔐 Segurança

### Boas Práticas
- **Backup Regular**: Manter cópias das wallets
- **Monitoramento**: Acompanhar logs constantemente
- **Atualizações**: Manter software atualizado
- **Firewall**: Proteger contra ataques
- **Chaves Privadas**: Nunca compartilhar

### Auditoria
- **Logs Completos**: Rastreabilidade total
- **Transparência**: Dados abertos para verificação
- **Compliance**: Conformidade regulatória
- **Ledger Global**: Registro imutável

## 🚀 Próximos Passos

### Melhorias Planejadas
- [ ] Gráficos de performance em tempo real
- [ ] Alertas de lucratividade
- [ ] Integração com exchanges
- [ ] Dashboard mobile
- [ ] API para integração externa
- [ ] Relatórios fiscais automáticos
- [ ] Multi-signature wallets
- [ ] Smart contracts

### Evolução do Sistema
- [ ] Suporte a múltiplas criptomoedas
- [ ] Pool de mineração
- [ ] Staking delegado
- [ ] Governança descentralizada
- [ ] Cross-chain transfers
- [ ] DeFi integrations

## 📞 Suporte

### Recursos de Ajuda
- **Documentação**: Este README
- **Logs**: Sistema de logs detalhado
- **Comunidade**: Fórum de usuários
- **Suporte Técnico**: Equipe de desenvolvimento

### Contato
- **Email**: suporte@blockchain-2layer.com
- **Telegram**: @blockchain2layer
- **Discord**: Blockchain 2-Layer Community

---

## 🎉 Resumo

O **Node Minerador** é uma solução completa para investidores que desejam:

✅ **Participar da mineração** de forma individual e controlada  
✅ **Gerenciar wallets** com chaves privadas seguras  
✅ **Realizar transferências** entre endereços  
✅ **Monitorar custos** e lucratividade em tempo real  
✅ **Evoluir para validador** através do sistema de stake  
✅ **Auditar operações** com logs completos  
✅ **Registrar movimentações** no ledger global  
✅ **Maximizar retornos** com controle total  

**🎯 Transforme-se de minerador em validador e participe da governança descentralizada!**
