package network

import (
	"encoding/json"
	"fmt"
	"net"
	"ordm-main/pkg/dag"
	"ordm-main/pkg/types"
	"time"
)

type Message struct {
	Type   string              `json:"type"`
	Block  *types.MicroBlock   `json:"block,omitempty"`
	From   string              `json:"from"`
	Hashes []string            `json:"hashes,omitempty"`
}

func StartServer(port string, d *dag.DAG) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	fmt.Println("[net] listening on", port)

	for {
		conn, err := ln.Accept()
		if err != nil { continue }
		go handleConn(conn, d)
	}
}


func handleConn(c net.Conn, d *dag.DAG) {
	defer c.Close()
	dec := json.NewDecoder(c)
	var msg Message
	if err := dec.Decode(&msg); err != nil { return }

	switch msg.Type {
	case "BLOCK":
		if msg.Block != nil && d.AddBlock(msg.Block) {
			fmt.Printf("[net] recebido bloco %s de %s (pais: %s) | DAG=%d\n",
				types.Short(msg.Block.Header.Hash), msg.From, types.FormatParents(msg.Block.Header.Parents), d.Size())
		}

	case "SYNC_REQUEST":
		// envia lista de hashes de blocos que possui
		hashes := d.AllBlockHashes()
		response := Message{
			Type:   "SYNC_RESPONSE",
			Hashes: hashes,
			From:   "meuNo",
		}
		json.NewEncoder(c).Encode(response)

	case "SYNC_RESPONSE":
		// verifica quais blocos faltam
		missing := d.MissingBlocks(msg.Hashes)
		for _, b := range missing {
			block := d.GetBlock(b)
			if block != nil {
				Broadcast([]string{msg.From}, Message{
					Type:  "BLOCK",
					Block: block,
					From:  "meuNo",
				})
			}
		}
	}
}

func SyncNode(peers []string, d *dag.DAG) {
	for _, p := range peers {
		go func(addr string) {
			c, err := net.DialTimeout("tcp", addr, 1*time.Second)
			if err != nil { return }
			defer c.Close()
			json.NewEncoder(c).Encode(Message{
				Type: "SYNC_REQUEST",
				From: "meuNo",
			})
		}(p)
	}
}
func Broadcast(peers []string, msg Message) {
	for _, p := range peers {
		go func(addr string) {
			c, err := net.DialTimeout("tcp", addr, 600*time.Millisecond)
			if err != nil { return }
			defer c.Close()
			json.NewEncoder(c).Encode(msg)
		}(p)
	}
}
