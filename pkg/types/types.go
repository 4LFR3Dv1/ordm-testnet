package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Tx struct {
	ID     string
	From   string
	To     string
	Amount int64
	Fee    int64
	Body   string
}

type MicroHeader struct {
	Version     int
	Parents     []string
	MerkleRoot  string
	Timestamp   int64
	Difficulty  int
	Nonce       uint64
	Hash        string
	MinerID     string
}

type MicroBlock struct {
	Header MicroHeader
	Txs    []Tx
}

type MacroBlock struct {
	Slot    int64
	Tips    []string
	Count   int
	Root    string
	Hash    string
}

func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func HashStrings(ss ...string) string {
	h := sha256.New()
	for _, s := range ss {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func Short(h string) string {
	if len(h) < 8 { return h }
	return h[:8]
}

func FormatParents(ps []string) string {
	if len(ps) == 0 { return "GENESIS" }
	if len(ps) == 1 { return Short(ps[0]) }
	return fmt.Sprintf("%s,%s", Short(ps[0]), Short(ps[1]))
}
