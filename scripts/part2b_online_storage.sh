#!/bin/bash

# 💾 Script para PARTE 2B: Storage Online
# Subparte 2.2 do PLANO_ATUALIZACOES.md

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

log "🔄 Iniciando PARTE 2B: Storage Online"

# 2.2.1 Corrigir storage no Render
log "2.2.1 - Criando pkg/storage/render_storage.go..."
cat > pkg/storage/render_storage.go << 'EOF'
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RenderStorage gerencia storage persistente no Render
type RenderStorage struct {
	DataDir      string
	Persistent   bool
	BackupPath   string
	LastBackup   time.Time
}

// NewRenderStorage cria novo storage para Render
func NewRenderStorage() *RenderStorage {
	dataDir := "/opt/render/data"
	backupPath := "/opt/render/backup"

	// Em desenvolvimento, usar diretório local
	if os.Getenv("NODE_ENV") != "production" {
		dataDir = "./data"
		backupPath = "./backup"
	}

	// Criar diretórios se não existirem
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(backupPath, 0755)

	return &RenderStorage{
		DataDir:    dataDir,
		Persistent: true,
		BackupPath: backupPath,
		LastBackup: time.Now(),
	}
}

// SaveData salva dados persistentes
func (rs *RenderStorage) SaveData(filename string, data interface{}) error {
	filePath := filepath.Join(rs.DataDir, filename)
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dados: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	// Backup automático a cada hora
	if time.Since(rs.LastBackup) > time.Hour {
		rs.createBackup(filename, jsonData)
	}

	return nil
}

// LoadData carrega dados persistentes
func (rs *RenderStorage) LoadData(filename string, data interface{}) error {
	filePath := filepath.Join(rs.DataDir, filename)
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", filePath)
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("erro ao deserializar dados: %v", err)
	}

	return nil
}

// createBackup cria backup dos dados
func (rs *RenderStorage) createBackup(filename string, data []byte) error {
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupFile := fmt.Sprintf("%s.%s.backup", filename, timestamp)
	backupPath := filepath.Join(rs.BackupPath, backupFile)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("erro ao criar backup: %v", err)
	}

	rs.LastBackup = time.Now()
	return nil
}

// GetDataPath retorna caminho dos dados
func (rs *RenderStorage) GetDataPath() string {
	return rs.DataDir
}

// IsPersistent verifica se storage é persistente
func (rs *RenderStorage) IsPersistent() bool {
	return rs.Persistent
}
EOF

log "✅ PARTE 2B: Storage Online concluída!"
log "📋 Arquivo criado: pkg/storage/render_storage.go"

