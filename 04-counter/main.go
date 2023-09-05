package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type typeMessage struct {
	Type string `json:"type"`
}

var addOk typeMessage = typeMessage{Type: "add_ok"}

type addMessage struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
	ID    string `json:"id"`
}

type readOkMessage struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func readOk(value int) readOkMessage {
	return readOkMessage{Type: "read_ok", Value: value}
}

func idGen(prefix func() string) func() string {
	var i uint64 = 0
	return func() string {
		id := fmt.Sprintf("%s-%d", prefix(), i)
		i++
		return id
	}
}

func makeHandlerFunc(requests chan<- maelstrom.Message, errors <-chan error) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		requests <- msg
		return <-errors
	}
}

func syncRPCWithTimeout(node *maelstrom.Node, dest string, req any) (maelstrom.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	return node.SyncRPC(ctx, dest, req)
}

func gossip(node *maelstrom.Node, dest string, add addMessage) {
	msg, err := syncRPCWithTimeout(node, dest, add)
	for err != nil {
		switch err {
		case context.DeadlineExceeded:
			msg, err = syncRPCWithTimeout(node, dest, add)
			continue
		default:
			log.Println("giving up on unexpected error", err)
			return
		}
	}
	var body typeMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Fatalln("error decoding add_ok message", err)
	}
	if body.Type != "add_ok" {
		log.Fatalln("unexpected message type (expected add_ok)", addOk.Type)
	}
}

func main() {
	node := maelstrom.NewNode()
	genId := idGen(node.ID)
	events := make(map[string]int)
	adds := make(chan maelstrom.Message)
	reads := make(chan maelstrom.Message)
	errors := make(chan error)
	node.Handle("add", makeHandlerFunc(adds, errors))
	node.Handle("read", makeHandlerFunc(reads, errors))
	go func() {
		for {
			select {
			case add := <-adds:
				var msg addMessage
				if err := json.Unmarshal(add.Body, &msg); err != nil {
					log.Fatalln("error decoding add message", err)
				}
				if msg.ID == "" {
					msg.ID = genId()
					for _, dest := range node.NodeIDs() {
						if dest == node.ID() {
							continue
						}
						go gossip(node, dest, msg)
					}
				}
				events[msg.ID] = msg.Delta
				errors <- node.Reply(add, addOk)
			case read := <-reads:
				sum := 0
				for _, delta := range events {
					sum += delta
				}
				errors <- node.Reply(read, readOk(sum))
			}
		}
	}()
	if err := node.Run(); err != nil {
		log.Fatalln("error running the node", err)
	}
}
