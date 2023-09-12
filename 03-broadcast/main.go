package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/stefanobaghino/maelstrom/demo/go"
)

type typeMessage struct {
	Type string `json:"type"`
}

var (
	broadcastOk = typeMessage{Type: "broadcast_ok"}
	topologyOk  = typeMessage{Type: "topology_ok"}
)

type readMessage struct {
	Type     string    `json:"type"`
	Messages []float64 `json:"messages"`
}

type topologyMessage struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

type broadcastMessage struct {
	Type     string    `json:"type"`
	Message  float64   `json:"message"`
	Messages []float64 `json:"messages"`
}

type broadcaster struct {
	config    broadcasterConfig
	node      *maelstrom.Node
	neighbors []string
	data      map[float64]bool
	outbox    chan float64
	topology  chan maelstrom.Message
	broadcast chan maelstrom.Message
	read      chan maelstrom.Message
	errors    chan error
}

type broadcasterConfig struct {
	outboxSize          int
	outboxFlushInterval time.Duration
	gossipTimeout       time.Duration
}

func newBroadcaster(config broadcasterConfig) *broadcaster {
	node := broadcaster{
		config:    config,
		node:      maelstrom.NewNode(),
		neighbors: []string{},
		data:      make(map[float64]bool),
		outbox:    make(chan float64, config.outboxSize),
		topology:  make(chan maelstrom.Message),
		broadcast: make(chan maelstrom.Message),
		read:      make(chan maelstrom.Message),
		errors:    make(chan error),
	}
	return &node
}

func makeHandlerFunc(requests chan<- maelstrom.Message, errors <-chan error) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		requests <- msg
		return <-errors
	}
}

func (b *broadcaster) run() {
	b.node.Handle("topology", makeHandlerFunc(b.topology, b.errors))
	b.node.Handle("broadcast", makeHandlerFunc(b.broadcast, b.errors))
	b.node.Handle("read", makeHandlerFunc(b.read, b.errors))
	go b.startEventLoop()
	if err := b.node.Run(); err != nil {
		log.Fatal(err)
	}
}

func (b *broadcaster) startEventLoop() {
	flushOutbox := time.Tick(b.config.outboxFlushInterval)
	for {
		select {
		case msg := <-b.topology:
			var body topologyMessage
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				b.errors <- err
				continue
			}
			res := b.setupTopology(body)
			b.errors <- b.node.Reply(msg, res)
		case msg := <-b.broadcast:
			var body broadcastMessage
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				b.errors <- err
				continue
			}
			res := b.processBroadcastMessage(body)
			b.errors <- b.node.Reply(msg, res)
		case msg := <-b.read:
			res := b.processReadMessage()
			b.errors <- b.node.Reply(msg, res)
		case <-flushOutbox:
			close(b.outbox)
			messages := make([]float64, 0, len(b.outbox))
			for message := range b.outbox {
				messages = append(messages, message)
			}
			b.outbox = make(chan float64, b.config.outboxSize)
			if len(messages) > 0 {
				for _, neighbor := range b.neighbors {
					go b.gossip(neighbor, broadcastMessage{Type: "broadcast", Messages: messages})
				}
			}
		}
	}
}

// Ignore the provided topology and use a star topology instead.
func (b *broadcaster) setupTopology(_ topologyMessage) typeMessage {
	nodeIDs := b.node.NodeIDs()
	var ownIndex int = -1
	for i, id := range nodeIDs {
		if id == b.node.ID() {
			ownIndex = i
			break
		}
	}
	if ownIndex == -1 {
		log.Fatalf("could not find own index in node IDs")
	}
	if ownIndex == 0 {
		b.neighbors = nodeIDs[1:]
	} else {
		b.neighbors = []string{nodeIDs[0]}
	}
	return topologyOk
}

func (b *broadcaster) processBroadcastMessage(broadcast broadcastMessage) typeMessage {
	if broadcast.Messages != nil {
		for _, message := range broadcast.Messages {
			b.storeMessage(message)
		}
	} else {
		b.storeMessage(broadcast.Message)
	}
	return broadcastOk
}

func (b *broadcaster) storeMessage(message float64) {
	if _, ok := b.data[message]; !ok {
		b.data[message] = true
		b.outbox <- message
	}
}

func (b *broadcaster) processReadMessage() readMessage {
	res := readMessage{Type: "read_ok", Messages: make([]float64, 0, len(b.data))}
	for key := range b.data {
		res.Messages = append(res.Messages, key)
	}
	return res
}

func (b *broadcaster) syncRPCWithTimeout(dest string, req any) (maelstrom.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.gossipTimeout)
	defer cancel()
	return b.node.SyncRPC(ctx, dest, req)
}

func (b *broadcaster) gossip(dest string, broadcast broadcastMessage) {
	msg, err := b.syncRPCWithTimeout(dest, broadcast)
	for err != nil {
		switch err {
		case context.DeadlineExceeded:
			msg, err = b.syncRPCWithTimeout(dest, broadcast)
			continue
		default:
			log.Println("giving up on unexpected error", err)
			return
		}
	}
	var body typeMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		log.Fatalln(err)
	}
	if body.Type != "broadcast_ok" {
		log.Fatalln("unexpected response type", body.Type)
	}
}

func main() {
	newBroadcaster(broadcasterConfig{
		outboxSize:          100,
		outboxFlushInterval: time.Millisecond * 90,
		gossipTimeout:       time.Millisecond * 210,
	}).run()
}
