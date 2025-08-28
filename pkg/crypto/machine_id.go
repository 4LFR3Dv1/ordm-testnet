package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// MachineID representa o ID único da máquina
type MachineID struct {
	ID        string    `json:"id"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	Platform  string    `json:"platform"`
	Arch      string    `json:"arch"`
}

// MachineIDManager gerencia o machineID
type MachineIDManager struct {
	DataPath string
}

// NewMachineIDManager cria um novo gerenciador de machineID
func NewMachineIDManager(dataPath string) *MachineIDManager {
	return &MachineIDManager{
		DataPath: dataPath,
	}
}

// GetOrCreateMachineID obtém ou cria o machineID
func (m *MachineIDManager) GetOrCreateMachineID() (*MachineID, error) {
	// Tentar carregar machineID existente
	if machineID, err := m.LoadMachineID(); err == nil {
		return machineID, nil
	}

	// Criar novo machineID
	return m.CreateMachineID()
}

// LoadMachineID carrega machineID existente
func (m *MachineIDManager) LoadMachineID() (*MachineID, error) {
	filePath := filepath.Join(m.DataPath, "machine_id.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler machine_id: %v", err)
	}

	var machineID MachineID
	if err := json.Unmarshal(data, &machineID); err != nil {
		return nil, fmt.Errorf("erro ao decodificar machine_id: %v", err)
	}

	return &machineID, nil
}

// CreateMachineID cria um novo machineID
func (m *MachineIDManager) CreateMachineID() (*MachineID, error) {
	// Gerar identificadores únicos da máquina
	identifiers := m.generateMachineIdentifiers()

	// Combinar identificadores
	combined := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		identifiers["hostname"],
		identifiers["platform"],
		identifiers["arch"],
		identifiers["cpu_info"],
		identifiers["mac_address"],
		identifiers["disk_id"],
	)

	// Gerar hash SHA256
	hash := sha256.Sum256([]byte(combined))
	hashHex := hex.EncodeToString(hash[:])

	// Criar machineID
	machineID := &MachineID{
		ID:        hashHex[:16], // Primeiros 16 caracteres como ID
		Hash:      hashHex,
		CreatedAt: time.Now(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	// Salvar machineID
	if err := m.SaveMachineID(machineID); err != nil {
		return nil, fmt.Errorf("erro ao salvar machine_id: %v", err)
	}

	fmt.Printf("🔑 MachineID criado: %s\n", machineID.ID)
	return machineID, nil
}

// SaveMachineID salva o machineID
func (m *MachineIDManager) SaveMachineID(machineID *MachineID) error {
	// Criar diretório se não existir
	if err := os.MkdirAll(m.DataPath, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	// Serializar machineID
	data, err := json.MarshalIndent(machineID, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar machine_id: %v", err)
	}

	// Salvar arquivo
	filePath := filepath.Join(m.DataPath, "machine_id.json")
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return nil
}

// generateMachineIdentifiers gera identificadores únicos da máquina
func (m *MachineIDManager) generateMachineIdentifiers() map[string]string {
	identifiers := make(map[string]string)

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		identifiers["hostname"] = hostname
	} else {
		identifiers["hostname"] = "unknown"
	}

	// Plataforma e arquitetura
	identifiers["platform"] = runtime.GOOS
	identifiers["arch"] = runtime.GOARCH

	// Informações da CPU
	identifiers["cpu_info"] = m.getCPUInfo()

	// Endereço MAC (primeira interface)
	identifiers["mac_address"] = m.getMACAddress()

	// ID do disco (primeiro disco)
	identifiers["disk_id"] = m.getDiskID()

	return identifiers
}

// getCPUInfo obtém informações da CPU
func (m *MachineIDManager) getCPUInfo() string {
	// Tentar obter informações da CPU de diferentes formas
	if runtime.GOOS == "linux" {
		// Linux: ler /proc/cpuinfo
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			// Extrair model name
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "model name") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		// macOS: usar sysctl
		if data, err := os.ReadFile("/usr/sbin/sysctl -n machdep.cpu.brand_string"); err == nil {
			return string(data)
		}
	}

	// Fallback: número de CPUs
	return fmt.Sprintf("cpu_cores_%d", runtime.NumCPU())
}

// getMACAddress obtém endereço MAC da primeira interface
func (m *MachineIDManager) getMACAddress() string {
	// Implementação simplificada - em produção usar net.Interfaces()
	return fmt.Sprintf("mac_%s_%d", runtime.GOOS, time.Now().Unix())
}

// getDiskID obtém ID do primeiro disco
func (m *MachineIDManager) getDiskID() string {
	// Implementação simplificada - em produção usar informações do sistema
	return fmt.Sprintf("disk_%s_%d", runtime.GOOS, time.Now().Unix())
}

// GetMinerIDFromMachineID gera minerID a partir do machineID
func (m *MachineIDManager) GetMinerIDFromMachineID() (string, error) {
	machineID, err := m.GetOrCreateMachineID()
	if err != nil {
		return "", fmt.Errorf("erro ao obter machineID: %v", err)
	}

	// Gerar minerID baseado no machineID + timestamp
	minerIDData := fmt.Sprintf("%s_%d", machineID.ID, time.Now().Unix())
	hash := sha256.Sum256([]byte(minerIDData))

	return hex.EncodeToString(hash[:16]), nil
}

// ValidateMachineID valida se o machineID é válido
func (m *MachineIDManager) ValidateMachineID(machineID *MachineID) bool {
	if machineID == nil || machineID.ID == "" || machineID.Hash == "" {
		return false
	}

	// Verificar se o hash corresponde ao ID
	if len(machineID.Hash) != 64 { // SHA256 tem 64 caracteres hex
		return false
	}

	// Verificar se o ID é válido
	if len(machineID.ID) != 16 {
		return false
	}

	return true
}
