package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type InputValidator struct{}

func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

func (v *InputValidator) ValidateWalletAddress(address string) error {
	if len(address) == 0 {
		return fmt.Errorf("endereço da wallet não pode estar vazio")
	}

	if len(address) < 10 {
		return fmt.Errorf("endereço da wallet muito curto")
	}

	if len(address) > 100 {
		return fmt.Errorf("endereço da wallet muito longo")
	}

	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(address) {
		return fmt.Errorf("endereço da wallet contém caracteres inválidos")
	}

	return nil
}

func (v *InputValidator) ValidatePIN(pin string) error {
	if len(pin) != 8 {
		return fmt.Errorf("PIN deve ter exatamente 8 dígitos")
	}

	for _, char := range pin {
		if !unicode.IsDigit(char) {
			return fmt.Errorf("PIN deve conter apenas números")
		}
	}

	if v.isSequential(pin) {
		return fmt.Errorf("PIN não pode ser sequencial")
	}

	if v.isRepetitive(pin) {
		return fmt.Errorf("PIN não pode ser repetitivo")
	}

	return nil
}

func (v *InputValidator) ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("nome de usuário deve ter pelo menos 3 caracteres")
	}

	if len(username) > 50 {
		return fmt.Errorf("nome de usuário deve ter no máximo 50 caracteres")
	}

	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(username) {
		return fmt.Errorf("nome de usuário contém caracteres inválidos")
	}

	return nil
}

func (v *InputValidator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("senha deve ter pelo menos 8 caracteres")
	}

	if len(password) > 128 {
		return fmt.Errorf("senha deve ter no máximo 128 caracteres")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("senha deve conter pelo menos uma letra maiúscula")
	}
	if !hasLower {
		return fmt.Errorf("senha deve conter pelo menos uma letra minúscula")
	}
	if !hasDigit {
		return fmt.Errorf("senha deve conter pelo menos um número")
	}
	if !hasSpecial {
		return fmt.Errorf("senha deve conter pelo menos um caractere especial")
	}

	return nil
}

func (v *InputValidator) ValidateAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("quantidade deve ser maior que zero")
	}

	if amount > 1000000000 {
		return fmt.Errorf("quantidade muito alta")
	}

	return nil
}

func (v *InputValidator) SanitizeInput(input string) string {
	dangerous := []string{"<script>", "</script>", "javascript:", "onload=", "onerror="}
	sanitized := input

	for _, danger := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, danger, "")
	}

	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000]
	}

	return sanitized
}

func (v *InputValidator) isSequential(pin string) bool {
	if len(pin) < 3 {
		return false
	}

	ascending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			ascending = false
			break
		}
	}

	descending := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]-1 {
			descending = false
			break
		}
	}

	return ascending || descending
}

func (v *InputValidator) isRepetitive(pin string) bool {
	if len(pin) < 3 {
		return false
	}

	first := rune(pin[0])
	for _, char := range pin {
		if char != first {
			return false
		}
	}

	return true
}
