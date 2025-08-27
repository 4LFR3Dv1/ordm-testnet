package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"ordm-main/pkg/types"
)

func MineBlock(prevHash string, txs []types.Tx, difficulty int, index int) types.MicroBlock {
	nonce := 0
	var hash string
	timestamp := time.Now().Unix()

	for {
		data := fmt.Sprintf("%d-%s-%d", index, prevHash, nonce)
		h := sha256.Sum256([]byte(data))
		hash = hex.EncodeToString(h[:])
		if hash[:difficulty] == string(make([]byte, difficulty)) {
			break
		}
		nonce++
	}

	header := types.MicroHeader{
		Version:    1,
		Parents:    []string{prevHash},
		MerkleRoot: merkleRoot(txs),
		Timestamp:  timestamp,
		Difficulty: difficulty,
		Nonce:      uint64(nonce),
		Hash:       hash,
		MinerID:    "miner",
	}

	return types.MicroBlock{
		Header: header,
		Txs:    txs,
	}
}

// Simula transações aleatórias
func RandomTx() types.Tx {
	users := []string{"Alice", "Bob", "Carol", "Dave"}
	return types.Tx{
		ID:     fmt.Sprintf("tx-%d", rand.Intn(1000)),
		From:   users[rand.Intn(len(users))],
		To:     users[rand.Intn(len(users))],
		Amount: int64(rand.Intn(100)),
		Fee:    int64(rand.Intn(10)),
		Body:   fmt.Sprintf("transfer %d", rand.Intn(100)),
	}
}

// Função auxiliar para calcular merkle root
func merkleRoot(txs []types.Tx) string {
	var hs []string
	for _, tx := range txs {
		hs = append(hs, tx.ID)
	}
	return types.HashStrings(hs...)
}
