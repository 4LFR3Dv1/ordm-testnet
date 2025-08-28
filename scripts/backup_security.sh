#!/bin/bash

# 💾 Script de Backup de Segurança ORDM
# Faz backup das configurações de segurança

set -e

BACKUP_DIR="./backup/security/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup do arquivo .env
if [ -f ".env" ]; then
    cp .env "$BACKUP_DIR/"
    echo "✅ Backup do .env criado"
fi

# Backup de chaves
if [ -d "wallets" ]; then
    cp -r wallets "$BACKUP_DIR/"
    echo "✅ Backup das wallets criado"
fi

# Backup de logs seguros
if [ -d "logs/secure" ]; then
    cp -r logs/secure "$BACKUP_DIR/"
    echo "✅ Backup dos logs seguros criado"
fi

# Comprimir backup
tar -czf "$BACKUP_DIR.tar.gz" -C "$(dirname "$BACKUP_DIR")" "$(basename "$BACKUP_DIR")"
rm -rf "$BACKUP_DIR"

echo "✅ Backup de segurança criado: $BACKUP_DIR.tar.gz"
echo "🔐 Mova este arquivo para local seguro!"
