package security

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// InputValidator valida entradas de usuário
type InputValidator struct {
	MaxAddressLength int
	MinAddressLength int
	MaxAmount        *big.Int
	MinAmount        *big.Int
}

// NewInputValidator cria um novo validador de inputs
func NewInputValidator() *InputValidator {
	maxAmount, _ := new(big.Int).SetString("1000000000000000000000000", 10) // 1M tokens
	minAmount, _ := new(big.Int).SetString("1", 10)                         // 1 token

	return &InputValidator{
		MaxAddressLength: 42,
		MinAddressLength: 26,
		MaxAmount:        maxAmount,
		MinAmount:        minAmount,
	}
}

// ValidateInput valida input baseado no tipo
func (iv *InputValidator) ValidateInput(input string, inputType string) error {
	switch inputType {
	case "address":
		return iv.validateAddress(input)
	case "amount":
		return iv.validateAmount(input)
	case "public_key":
		return iv.validatePublicKey(input)
	case "username":
		return iv.validateUsername(input)
	case "password":
		return iv.validatePassword(input)
	case "email":
		return iv.validateEmail(input)
	default:
		return fmt.Errorf("tipo de input não suportado: %s", inputType)
	}
}

// validateAddress valida endereço de wallet
func (iv *InputValidator) validateAddress(address string) error {
	if address == "" {
		return fmt.Errorf("endereço não pode ser vazio")
	}

	if len(address) < iv.MinAddressLength || len(address) > iv.MaxAddressLength {
		return fmt.Errorf("endereço deve ter entre %d e %d caracteres", iv.MinAddressLength, iv.MaxAddressLength)
	}

	// Validar formato: apenas letras maiúsculas, minúsculas e números
	addressRegex := regexp.MustCompile(`^[A-Za-z0-9]+$`)
	if !addressRegex.MatchString(address) {
		return fmt.Errorf("endereço contém caracteres inválidos")
	}

	// Validar checksum básico (implementação simplificada)
	if !iv.validateChecksum(address) {
		return fmt.Errorf("checksum do endereço inválido")
	}

	return nil
}

// validateAmount valida quantidade de tokens
func (iv *InputValidator) validateAmount(amountStr string) error {
	if amountStr == "" {
		return fmt.Errorf("quantidade não pode ser vazia")
	}

	// Validar formato numérico
	amountRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	if !amountRegex.MatchString(amountStr) {
		return fmt.Errorf("quantidade deve ser um número válido")
	}

	// Converter para big.Int para precisão
	amount, ok := new(big.Int).SetString(strings.Replace(amountStr, ".", "", -1), 10)
	if !ok {
		return fmt.Errorf("quantidade inválida")
	}

	// Validar limites
	if amount.Cmp(iv.MinAmount) < 0 {
		return fmt.Errorf("quantidade deve ser pelo menos %s", iv.MinAmount.String())
	}

	if amount.Cmp(iv.MaxAmount) > 0 {
		return fmt.Errorf("quantidade não pode exceder %s", iv.MaxAmount.String())
	}

	return nil
}

// validatePublicKey valida chave pública
func (iv *InputValidator) validatePublicKey(publicKey string) error {
	if publicKey == "" {
		return fmt.Errorf("chave pública não pode ser vazia")
	}

	// Validar formato base64
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`)
	if !base64Regex.MatchString(publicKey) {
		return fmt.Errorf("chave pública deve estar em formato base64 válido")
	}

	// Validar comprimento mínimo
	if len(publicKey) < 32 {
		return fmt.Errorf("chave pública muito curta")
	}

	return nil
}

// validateUsername valida nome de usuário
func (iv *InputValidator) validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("nome de usuário não pode ser vazio")
	}

	if len(username) < 3 || len(username) > 20 {
		return fmt.Errorf("nome de usuário deve ter entre 3 e 20 caracteres")
	}

	// Validar caracteres permitidos
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("nome de usuário contém caracteres inválidos")
	}

	// Verificar palavras reservadas
	reservedWords := []string{"admin", "root", "system", "test", "guest"}
	for _, word := range reservedWords {
		if strings.ToLower(username) == word {
			return fmt.Errorf("nome de usuário não pode ser uma palavra reservada")
		}
	}

	return nil
}

// validatePassword valida senha
func (iv *InputValidator) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("senha não pode ser vazia")
	}

	if len(password) < 8 {
		return fmt.Errorf("senha deve ter pelo menos 8 caracteres")
	}

	if len(password) > 128 {
		return fmt.Errorf("senha muito longa")
	}

	// Verificar complexidade
	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	)

	complexity := 0
	if hasUpper {
		complexity++
	}
	if hasLower {
		complexity++
	}
	if hasNumber {
		complexity++
	}
	if hasSpecial {
		complexity++
	}

	if complexity < 3 {
		return fmt.Errorf("senha deve conter pelo menos 3 tipos de caracteres (maiúsculas, minúsculas, números, especiais)")
	}

	return nil
}

// validateEmail valida email
func (iv *InputValidator) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email não pode ser vazio")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("formato de email inválido")
	}

	if len(email) > 254 {
		return fmt.Errorf("email muito longo")
	}

	return nil
}

// validateChecksum valida checksum do endereço (implementação simplificada)
func (iv *InputValidator) validateChecksum(address string) bool {
	// Implementação simplificada - em produção usar algoritmo real
	if len(address) < 4 {
		return false
	}

	// Verificar se termina com caracteres válidos
	lastChars := address[len(address)-4:]
	validEndings := []string{"0000", "1111", "2222", "3333", "4444", "5555", "6666", "7777", "8888", "9999"}

	for _, ending := range validEndings {
		if lastChars == ending {
			return true
		}
	}

	return false
}

// SanitizeInput remove caracteres perigosos
func (iv *InputValidator) SanitizeInput(input string) string {
	// Remover caracteres de controle
	sanitized := regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(input, "")

	// Remover scripts
	sanitized = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(sanitized, "")
	sanitized = regexp.MustCompile(`(?i)javascript:`).ReplaceAllString(sanitized, "")

	// Remover tags HTML
	sanitized = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(sanitized, "")

	return strings.TrimSpace(sanitized)
}

// ValidateTransactionInput valida entrada de transação
func (iv *InputValidator) ValidateTransactionInput(from, to, amount string) error {
	if err := iv.validateAddress(from); err != nil {
		return fmt.Errorf("endereço de origem inválido: %v", err)
	}

	if err := iv.validateAddress(to); err != nil {
		return fmt.Errorf("endereço de destino inválido: %v", err)
	}

	if err := iv.validateAmount(amount); err != nil {
		return fmt.Errorf("quantidade inválida: %v", err)
	}

	if from == to {
		return fmt.Errorf("endereços de origem e destino não podem ser iguais")
	}

	return nil
}
