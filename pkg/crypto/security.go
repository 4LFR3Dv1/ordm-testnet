package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"
)

// NodeIdentity representa a identidade criptográfica de um node
type NodeIdentity struct {
	NodeID        string    `json:"node_id"`
	PublicKey     string    `json:"public_key"`
	PrivateKeyEnc string    `json:"private_key_enc"` // Chave privada criptografada
	CreatedAt     time.Time `json:"created_at"`
	LastAccess    time.Time `json:"last_access"`
	IsActive      bool      `json:"is_active"`
	Permissions   []string  `json:"permissions"`
	StakeAmount   int64     `json:"stake_amount"`
	ValidatorRank int       `json:"validator_rank"`
}

// SecurityManager gerencia segurança e autenticação
type SecurityManager struct {
	NodeID          string
	PrivateKey      *rsa.PrivateKey
	PublicKey       *rsa.PublicKey
	MasterKey       []byte // Chave mestra para criptografar private keys
	SessionToken    string
	IsAuthenticated bool
}

// NewSecurityManager cria um novo gerenciador de segurança
func NewSecurityManager(nodeID string) (*SecurityManager, error) {
	// Gerar par de chaves RSA 2048-bit
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar chaves RSA: %v", err)
	}

	// Gerar chave mestra AES-256
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave mestra: %v", err)
	}

	return &SecurityManager{
		NodeID:          nodeID,
		PrivateKey:      privateKey,
		PublicKey:       &privateKey.PublicKey,
		MasterKey:       masterKey,
		IsAuthenticated: false,
	}, nil
}

// EncryptPrivateKey criptografa a chave privada com AES-256
func (sm *SecurityManager) EncryptPrivateKey() (string, error) {
	// Serializar chave privada
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(sm.PrivateKey)

	// Criar cipher AES
	block, err := aes.NewCipher(sm.MasterKey)
	if err != nil {
		return "", err
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Gerar nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Criptografar
	ciphertext := gcm.Seal(nonce, nonce, privateKeyBytes, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey descriptografa a chave privada
func (sm *SecurityManager) DecryptPrivateKey(encryptedKey string) error {
	// Decodificar base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return err
	}

	// Criar cipher AES
	block, err := aes.NewCipher(sm.MasterKey)
	if err != nil {
		return err
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Extrair nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext muito curto")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Descriptografar
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	// Parsear chave privada
	privateKey, err := x509.ParsePKCS1PrivateKey(plaintext)
	if err != nil {
		return err
	}

	sm.PrivateKey = privateKey
	sm.PublicKey = &privateKey.PublicKey
	return nil
}

// CreateSignature cria assinatura digital
func (sm *SecurityManager) CreateSignature(data []byte) (string, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(nil, sm.PrivateKey, 0, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifica assinatura digital
func (sm *SecurityManager) VerifySignature(data []byte, signature string, publicKey *rsa.PublicKey) error {
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(publicKey, 0, hash[:], sigBytes)
}

// GenerateSessionToken gera token de sessão
func (sm *SecurityManager) GenerateSessionToken() string {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	sm.SessionToken = base64.StdEncoding.EncodeToString(tokenBytes)
	return sm.SessionToken
}

// ValidateSessionToken valida token de sessão
func (sm *SecurityManager) ValidateSessionToken(token string) bool {
	return token == sm.SessionToken
}

// PublicKeyToString converte chave pública para string
func (sm *SecurityManager) PublicKeyToString() string {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(sm.PublicKey)
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

// StringToPublicKey converte string para chave pública
func StringToPublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PublicKey(publicKeyBytes)
}

// HashData cria hash SHA-256 de dados
func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateNonce gera nonce único
func GenerateNonce() string {
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	return base64.StdEncoding.EncodeToString(nonceBytes)
}
