#!/bin/bash

# Script para configurar GitHub e fazer deploy
echo "🚀 Configurando GitHub para ORDM Testnet..."

# Verificar se Git está configurado
if ! git config --global user.email > /dev/null 2>&1; then
    echo "❌ Git não está configurado. Configure primeiro:"
    echo "   git config --global user.email 'seu-email@gmail.com'"
    echo "   git config --global user.name 'Seu Nome'"
    exit 1
fi

# Verificar se repositório existe
if [ -z "$1" ]; then
    echo "❌ Uso: $0 <nome-do-repositorio>"
    echo "   Exemplo: $0 ordm-testnet"
    exit 1
fi

REPO_NAME=$1
GITHUB_USER="4LFR3Dv1"

echo "📋 Configurando repositório: $GITHUB_USER/$REPO_NAME"

# Adicionar remote
echo "🔗 Adicionando remote..."
git remote add origin https://github.com/$GITHUB_USER/$REPO_NAME.git

# Tentar push
echo "📤 Fazendo push para GitHub..."
if git push -u origin main; then
    echo "✅ Push realizado com sucesso!"
    echo ""
    echo "🎉 Repositório configurado:"
    echo "   https://github.com/$GITHUB_USER/$REPO_NAME"
    echo ""
    echo "📋 Próximos passos:"
    echo "   1. Acesse: https://render.com"
    echo "   2. Login com GitHub"
    echo "   3. 'New +' → 'Web Service'"
    echo "   4. Selecione o repositório: $REPO_NAME"
    echo "   5. Configure conforme DEPLOY_GUIDE.md"
else
    echo "❌ Erro no push. Tentando alternativas..."
    
    # Tentar com SSH
    echo "🔄 Tentando com SSH..."
    git remote set-url origin git@github.com:$GITHUB_USER/$REPO_NAME.git
    
    if git push -u origin main; then
        echo "✅ Push realizado com SSH!"
    else
        echo "❌ Falha no push. Verifique:"
        echo "   1. Repositório existe no GitHub?"
        echo "   2. Tem permissões de push?"
        echo "   3. SSH key configurada?"
        echo ""
        echo "🔧 Soluções:"
        echo "   1. Crie o repositório manualmente no GitHub"
        echo "   2. Configure SSH key: https://docs.github.com/en/authentication/connecting-to-github-with-ssh"
        echo "   3. Ou use HTTPS com token: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token"
    fi
fi
