package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type PINGenerator struct {
	length     int
	useLetters bool
	useNumbers bool
	useSymbols bool
}

func NewPINGenerator() *PINGenerator {
	return &PINGenerator{
		length:     8, // Aumentado para 8 dígitos
		useNumbers: true,
	}
}

func (pg *PINGenerator) GeneratePIN() (string, error) {
	const numbers = "0123456789"

	pin := make([]byte, pg.length)
	for i := range pin {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(numbers))))
		if err != nil {
			return "", fmt.Errorf("erro ao gerar número aleatório: %v", err)
		}
		pin[i] = numbers[num.Int64()]
	}

	return string(pin), nil
}

func (pg *PINGenerator) ValidatePIN(pin string) error {
	if len(pin) != pg.length {
		return fmt.Errorf("PIN deve ter %d dígitos", pg.length)
	}

	for _, char := range pin {
		if char < '0' || char > '9' {
			return fmt.Errorf("PIN deve conter apenas números")
		}
	}

	if pg.isSequential(pin) {
		return fmt.Errorf("PIN não pode ser sequencial")
	}

	if pg.isRepetitive(pin) {
		return fmt.Errorf("PIN não pode ser repetitivo")
	}

	return nil
}

func (pg *PINGenerator) isSequential(pin string) bool {
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

func (pg *PINGenerator) isRepetitive(pin string) bool {
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
