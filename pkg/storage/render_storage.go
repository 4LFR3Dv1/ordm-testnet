package storage

import (
	"os"
	"path/filepath"
)

// RenderStorage gerencia storage persistente no Render
type RenderStorage struct {
	DataDir string
}

// NewRenderStorage cria um novo storage para Render
func NewRenderStorage() *RenderStorage {
	// Render permite persistência apenas em /opt/render/data
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/render/data"
	}
	
	// Criar diretório se não existir
	os.MkdirAll(dataDir, 0755)
	
	return &RenderStorage{
		DataDir: dataDir,
	}
}

// GetDataPath retorna o caminho completo para um arquivo
func (rs *RenderStorage) GetDataPath(filename string) string {
	return filepath.Join(rs.DataDir, filename)
}

// GetLedgerPath retorna caminho para o ledger
func (rs *RenderStorage) GetLedgerPath() string {
	return rs.GetDataPath("global_ledger.json")
}

// GetWalletsPath retorna caminho para wallets
func (rs *RenderStorage) GetWalletsPath() string {
	return rs.GetDataPath("wallets")
}

// GetMiningStatePath retorna caminho para estado de mineração
func (rs *RenderStorage) GetMiningStatePath() string {
	return rs.GetDataPath("mining_state.json")
}

// GetUsersPath retorna caminho para usuários
func (rs *RenderStorage) GetUsersPath() string {
	return rs.GetDataPath("users.json")
}

// GetBlockchainPath retorna caminho para blockchain
func (rs *RenderStorage) GetBlockchainPath() string {
	return rs.GetDataPath("blockchain")
}

// EnsureDirectories cria todos os diretórios necessários
func (rs *RenderStorage) EnsureDirectories() error {
	dirs := []string{
		rs.DataDir,
		rs.GetWalletsPath(),
		rs.GetBlockchainPath(),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	
	return nil
}

// IsRenderEnvironment verifica se está rodando no Render
func (rs *RenderStorage) IsRenderEnvironment() bool {
	return os.Getenv("PORT") != "" || os.Getenv("RENDER") != ""
}
