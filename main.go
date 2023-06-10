package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func idGen(prefix func() string) func() string {
	var i uint64 = 0
	return func() string {
		id := fmt.Sprintf("%s-%d", prefix(), i)
		i++
		return id
	}
}

func handleGenerateRequest(n *maelstrom.Node) maelstrom.HandlerFunc {
	genId := idGen(n.ID)
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "generate_ok"
		body["id"] = genId()
		return n.Reply(msg, body)
	}
}

func main() {
	n := maelstrom.NewNode()
	n.Handle("generate", handleGenerateRequest(n))
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
