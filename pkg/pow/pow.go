package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"ordm-main/pkg/types"
)

// Mine cria um novo header de bloco usando Proof of Work
func Mine(parents []string, merkleRoot string, difficulty int, minerID string) types.MicroHeader {
	nonce := uint64(0)
	var hash string
	timestamp := time.Now().Unix()

	fmt.Printf("[MINER] Iniciando mineração com dificuldade %d\n", difficulty)
	attempts := 0

	for {
		attempts++
		if attempts%10000 == 0 {
			fmt.Printf("[MINER] Tentativa %d, nonce: %d\n", attempts, nonce)
		}

		// Cria dados para mineração
		data := fmt.Sprintf("%v-%s-%d-%s", parents, merkleRoot, timestamp, minerID)
		h := sha256.Sum256([]byte(data))
		hash = hex.EncodeToString(h[:])

		// Verifica se o hash atende à dificuldade (deve começar com 'difficulty' zeros)
		zeros := 0
		for _, ch := range hash {
			if ch == '0' {
				zeros++
			} else {
				break
			}
		}
		if zeros >= difficulty {
			fmt.Printf("[MINER] Hash encontrado! Nonce: %d, Hash: %s (zeros: %d)\n", nonce, hash, zeros)
			break
		}

		// Para dificuldade 0, aceita qualquer hash
		if difficulty == 0 {
			fmt.Printf("[MINER] Hash aceito (dificuldade 0)! Nonce: %d, Hash: %s\n", nonce, hash)
			break
		}
		nonce++
	}

	return types.MicroHeader{
		Version:    1,
		Parents:    parents,
		MerkleRoot: merkleRoot,
		Timestamp:  timestamp,
		Difficulty: difficulty,
		Nonce:      nonce,
		Hash:       hash,
		MinerID:    minerID,
	}
}
