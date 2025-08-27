package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// NodeIdentity representa a identidade única de um node
type NodeIdentity struct {
	NodeID        string    `json:"node_id"`
	PublicKey     string    `json:"public_key"`
	PrivateKeyEnc string    `json:"private_key_enc"` // Chave privada criptografada
	CreatedAt     time.Time `json:"created_at"`
	LastAccess    time.Time `json:"last_access"`
	IsActive      bool      `json:"is_active"`
	Permissions   []string  `json:"permissions"`
}

// NodeAuthManager gerencia autenticação de nodes
type NodeAuthManager struct {
	NodeID          string
	PrivateKey      *rsa.PrivateKey
	PublicKey       *rsa.PublicKey
	MasterKey       []byte // Chave mestra para criptografar private keys
	IdentityPath    string
	IsAuthenticated bool
}

// NewNodeAuthManager cria novo gerenciador de autenticação
func NewNodeAuthManager(nodeID, identityPath string) (*NodeAuthManager, error) {
	manager := &NodeAuthManager{
		NodeID:       nodeID,
		IdentityPath: identityPath,
	}

	// Gerar ou carregar identidade do node
	err := manager.loadOrCreateIdentity()
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// loadOrCreateIdentity carrega ou cria identidade do node
func (nam *NodeAuthManager) loadOrCreateIdentity() error {
	identityFile := filepath.Join(nam.IdentityPath, nam.NodeID+".json")

	if _, err := os.Stat(identityFile); os.IsNotExist(err) {
		// Criar nova identidade
		return nam.createNewIdentity(identityFile)
	}

	// Carregar identidade existente
	return nam.loadExistingIdentity(identityFile)
}

// createNewIdentity cria nova identidade para o node
func (nam *NodeAuthManager) createNewIdentity(identityFile string) error {
	// Gerar par de chaves RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("erro ao gerar chave privada: %v", err)
	}

	nam.PrivateKey = privateKey
	nam.PublicKey = &privateKey.PublicKey

	// Gerar chave mestra
	nam.MasterKey = make([]byte, 32)
	_, err = rand.Read(nam.MasterKey)
	if err != nil {
		return fmt.Errorf("erro ao gerar chave mestra: %v", err)
	}

	// Criptografar chave privada
	encryptedPrivateKey, err := nam.encryptPrivateKey()
	if err != nil {
		return fmt.Errorf("erro ao criptografar chave privada: %v", err)
	}

	// Criar identidade
	identity := NodeIdentity{
		NodeID:        nam.NodeID,
		PublicKey:     nam.publicKeyToString(),
		PrivateKeyEnc: base64.StdEncoding.EncodeToString(encryptedPrivateKey),
		CreatedAt:     time.Now(),
		LastAccess:    time.Now(),
		IsActive:      true,
		Permissions:   []string{"mining", "validation", "wallet_access"},
	}

	// Salvar identidade
	return nam.saveIdentity(identity, identityFile)
}

// loadExistingIdentity carrega identidade existente
func (nam *NodeAuthManager) loadExistingIdentity(identityFile string) error {
	data, err := os.ReadFile(identityFile)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de identidade: %v", err)
	}

	var identity NodeIdentity
	err = json.Unmarshal(data, &identity)
	if err != nil {
		return fmt.Errorf("erro ao deserializar identidade: %v", err)
	}

	// Decodificar chave pública
	nam.PublicKey, err = nam.stringToPublicKey(identity.PublicKey)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave pública: %v", err)
	}

	// Decodificar chave privada criptografada
	encryptedData, err := base64.StdEncoding.DecodeString(identity.PrivateKeyEnc)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave privada: %v", err)
	}

	// Descriptografar chave privada
	nam.PrivateKey, err = nam.decryptPrivateKey(encryptedData)
	if err != nil {
		return fmt.Errorf("erro ao descriptografar chave privada: %v", err)
	}

	// Atualizar último acesso
	identity.LastAccess = time.Now()
	nam.saveIdentity(identity, identityFile)

	return nil
}

// encryptPrivateKey criptografa a chave privada
func (nam *NodeAuthManager) encryptPrivateKey() ([]byte, error) {
	// Serializar chave privada
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(nam.PrivateKey)

	// Criar cipher AES
	block, err := aes.NewCipher(nam.MasterKey)
	if err != nil {
		return nil, err
	}

	// Gerar IV
	iv := make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	// Criptografar
	ciphertext := make([]byte, len(privateKeyBytes))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, privateKeyBytes)

	// Concatenar IV + dados criptografados
	result := make([]byte, 0, len(iv)+len(ciphertext))
	result = append(result, iv...)
	result = append(result, ciphertext...)

	return result, nil
}

// decryptPrivateKey descriptografa a chave privada
func (nam *NodeAuthManager) decryptPrivateKey(encryptedData []byte) (*rsa.PrivateKey, error) {
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("dados criptografados muito pequenos")
	}

	// Separar IV e dados
	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	// Criar cipher AES
	block, err := aes.NewCipher(nam.MasterKey)
	if err != nil {
		return nil, err
	}

	// Descriptografar
	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	// Deserializar chave privada
	privateKey, err := x509.ParsePKCS1PrivateKey(plaintext)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// publicKeyToString converte chave pública para string
func (nam *NodeAuthManager) publicKeyToString() string {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(nam.PublicKey)
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

// stringToPublicKey converte string para chave pública
func (nam *NodeAuthManager) stringToPublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PublicKey(publicKeyBytes)
}

// saveIdentity salva identidade em arquivo
func (nam *NodeAuthManager) saveIdentity(identity NodeIdentity, identityFile string) error {
	os.MkdirAll(filepath.Dir(identityFile), 0755)

	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(identityFile, data, 0600)
}

// AuthenticateNode autentica o node com o backend
func (nam *NodeAuthManager) AuthenticateNode(backendURL string) error {
	// TODO: Implementar comunicação com backend
	// Por enquanto, simular autenticação bem-sucedida
	fmt.Printf("🔐 Node %s autenticado com sucesso\n", nam.NodeID)
	nam.IsAuthenticated = true

	return nil
}

// createSignature cria assinatura digital
func (nam *NodeAuthManager) createSignature() (string, error) {
	// Criar dados para assinatura
	data := fmt.Sprintf("%s:%d", nam.NodeID, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))

	// Assinar com chave privada
	signature, err := rsa.SignPKCS1v15(nil, nam.PrivateKey, 0, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// IsNodeAuthenticated verifica se node está autenticado
func (nam *NodeAuthManager) IsNodeAuthenticated() bool {
	return nam.IsAuthenticated
}

// GetNodeID retorna ID do node
func (nam *NodeAuthManager) GetNodeID() string {
	return nam.NodeID
}
