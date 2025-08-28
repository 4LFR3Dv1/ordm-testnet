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
	DataDir    string
	Persistent bool
	BackupPath string
	LastBackup time.Time
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

// EnsureDirectories cria diretórios necessários
func (rs *RenderStorage) EnsureDirectories() error {
	dirs := []string{
		rs.DataDir,
		rs.BackupPath,
		filepath.Join(rs.DataDir, "wallets"),
		filepath.Join(rs.DataDir, "ledger"),
		filepath.Join(rs.DataDir, "blocks"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório %s: %v", dir, err)
		}
	}

	return nil
}

// GetWalletsPath retorna caminho das wallets
func (rs *RenderStorage) GetWalletsPath() string {
	return filepath.Join(rs.DataDir, "wallets")
}

// GetLedgerPath retorna caminho do ledger
func (rs *RenderStorage) GetLedgerPath() string {
	return filepath.Join(rs.DataDir, "ledger")
}

// IsPersistent verifica se storage é persistente
func (rs *RenderStorage) IsPersistent() bool {
	return rs.Persistent
}
