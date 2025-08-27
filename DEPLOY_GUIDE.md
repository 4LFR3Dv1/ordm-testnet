# 🚀 Guia de Deploy: GitHub + Render

Este guia te ajudará a fazer o deploy da ORDM Testnet no GitHub e Render.

## 📋 Pré-requisitos

- Conta no GitHub
- Conta no Render (gratuita)
- Git instalado localmente
- Docker instalado (opcional, para testes locais)

## 🔧 Passo 1: Preparar o Repositório Local

### 1.1 Inicializar Git (se ainda não feito)
```bash
git init
git add .
git commit -m "Initial commit: ORDM Testnet"
```

### 1.2 Verificar arquivos importantes
Certifique-se de que estes arquivos estão presentes:
- `Dockerfile`
- `render.yaml`
- `scripts/start.sh`
- `.github/workflows/deploy.yml`
- `go.mod` e `go.sum`

## 🌐 Passo 2: Criar Repositório no GitHub

### 2.1 Criar novo repositório
1. Acesse [github.com](https://github.com)
2. Clique em "New repository"
3. Configure:
   - **Repository name**: `ordm-testnet`
   - **Description**: `ORDM Blockchain 2-Layer Testnet`
   - **Visibility**: Public (recomendado)
   - **Initialize with**: Não marque nada

### 2.2 Conectar repositório local
```bash
git remote add origin https://github.com/SEU_USUARIO/ordm-testnet.git
git branch -M main
git push -u origin main
```

## 🚀 Passo 3: Configurar Render

### 3.1 Criar conta no Render
1. Acesse [render.com](https://render.com)
2. Faça login com GitHub
3. Clique em "New +" → "Web Service"

### 3.2 Conectar repositório
1. Selecione o repositório `ordm-testnet`
2. Configure o serviço:
   - **Name**: `ordm-testnet`
   - **Environment**: `Docker`
   - **Region**: `Oregon` (ou mais próxima)
   - **Branch**: `main`
   - **Root Directory**: `/` (deixe vazio)

### 3.3 Configurar variáveis de ambiente
Adicione estas variáveis:
```
PORT=3000
EXPLORER_PORT=8080
MONITOR_PORT=9090
NODE_ENV=production
```

### 3.4 Configurar build e start
- **Build Command**: `docker build -t ordm-testnet .`
- **Start Command**: `docker run -p $PORT:3000 -p 8080:8080 -p 9090:9090 ordm-testnet`

### 3.5 Deploy
Clique em "Create Web Service" e aguarde o deploy.

## 🔐 Passo 4: Configurar CI/CD (Opcional)

### 4.1 Obter tokens do Render
1. No Render, vá em "Account" → "API Keys"
2. Crie uma nova API key
3. Copie o token

### 4.2 Configurar secrets no GitHub
1. No seu repositório GitHub, vá em "Settings" → "Secrets and variables" → "Actions"
2. Adicione estes secrets:
   - `RENDER_TOKEN`: Token da API do Render
   - `RENDER_SERVICE_ID`: ID do serviço (encontrado na URL do Render)

### 4.3 Testar CI/CD
Faça um push para a branch main:
```bash
git add .
git commit -m "Add CI/CD configuration"
git push
```

## 📊 Passo 5: Verificar Deploy

### 5.1 URLs da aplicação
Após o deploy, você terá URLs como:
- **Main App**: `https://ordm-testnet.onrender.com`
- **Explorer**: `https://ordm-testnet.onrender.com:8080`
- **Monitor**: `https://ordm-testnet.onrender.com:9090`

### 5.2 Testar endpoints
```bash
# Testar API principal
curl https://ordm-testnet.onrender.com/status

# Testar faucet
curl https://ordm-testnet.onrender.com/api/testnet/faucet/stats

# Testar seed nodes
curl https://ordm-testnet.onrender.com/api/testnet/seed-nodes
```

## 🔧 Passo 6: Configurações Avançadas

### 6.1 Domínio personalizado (Opcional)
1. No Render, vá em "Settings" → "Custom Domains"
2. Adicione seu domínio
3. Configure DNS conforme instruções

### 6.2 Monitoramento
- **Logs**: Acesse "Logs" no Render
- **Métricas**: Use o dashboard interno em `/monitor`
- **Alertas**: Configure no Render Dashboard

### 6.3 Backup automático
O sistema já inclui backup automático configurado.

## 🛠️ Solução de Problemas

### Problema: Build falha
```bash
# Verificar logs
docker build -t ordm-testnet . --progress=plain

# Verificar dependências
go mod tidy
go mod verify
```

### Problema: Aplicação não inicia
```bash
# Verificar logs no Render
# Verificar variáveis de ambiente
# Verificar portas
```

### Problema: CI/CD não funciona
1. Verificar secrets no GitHub
2. Verificar permissões do token do Render
3. Verificar workflow YAML

## 📈 Monitoramento e Manutenção

### 6.1 Logs em tempo real
```bash
# No Render Dashboard
# Vá em "Logs" → "Live"
```

### 6.2 Métricas de performance
- Acesse o dashboard de monitoramento
- Verifique uso de CPU/memória
- Monitore taxa de erro

### 6.3 Atualizações
```bash
# Fazer alterações
git add .
git commit -m "Update: descrição das mudanças"
git push origin main
# Deploy automático acontecerá
```

## 🎯 Próximos Passos

### 6.1 Melhorias sugeridas
- [ ] Configurar domínio personalizado
- [ ] Implementar CDN para assets
- [ ] Configurar banco de dados persistente
- [ ] Implementar autenticação OAuth
- [ ] Adicionar métricas avançadas

### 6.2 Escalabilidade
- [ ] Configurar load balancer
- [ ] Implementar cache Redis
- [ ] Otimizar queries de banco
- [ ] Configurar auto-scaling

## 📞 Suporte

- **GitHub Issues**: Para bugs e feature requests
- **Render Support**: Para problemas de infraestrutura
- **Documentação**: Consulte `README.md` e `TESTNET_README.md`

---

**🎉 Parabéns! Sua ORDM Testnet está no ar!**

Agora você pode compartilhar as URLs com desenvolvedores e investidores para testar a blockchain.
