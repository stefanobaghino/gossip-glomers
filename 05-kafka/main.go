package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type kv struct {
	k string
	v int
}

type sendRequest struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	Msg  int    `json:"msg"`
}

type sendResponse struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
}

func sendResponseMsg(offset int) sendResponse {
	return sendResponse{Type: "send_ok", Offset: offset}
}

type pollRequest struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type pollResponse struct {
	Type string              `json:"type"`
	Msgs map[string][][2]int `json:"msgs"`
}

func pollResponseMsg(msgs map[string][][2]int) pollResponse {
	return pollResponse{Type: "poll_ok", Msgs: msgs}
}

type commitOffsetsRequest struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type commitOffsetsResponse struct {
	Type string `json:"type"`
}

var commitOffsetsResponseMsg = commitOffsetsResponse{Type: "commit_offsets_ok"}

type listCommittedOffsetsRequest struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

type listCommittedOffsetsResponse struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

func listCommittedOffsetsResponseMsg(offsets map[string]int) listCommittedOffsetsResponse {
	return listCommittedOffsetsResponse{Type: "list_committed_offsets_ok", Offsets: offsets}
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

func retryOnTimeout(f func() (maelstrom.Message, error)) (maelstrom.Message, error) {
	for {
		msg, err := f()
		if err == context.DeadlineExceeded {
			continue
		}
		return msg, err
	}
}

func main() {
	node := maelstrom.NewNode()
	journal := make([]kv, 0)
	offset := 0
	committed := make(map[string]int)
	sendRequests := make(chan maelstrom.Message)
	pollRequests := make(chan maelstrom.Message)
	commitOffsetsRequests := make(chan maelstrom.Message)
	listCommittedOffsetsRequests := make(chan maelstrom.Message)
	errors := make(chan error)
	node.Handle("send", makeHandlerFunc(sendRequests, errors))
	node.Handle("poll", makeHandlerFunc(pollRequests, errors))
	node.Handle("commit_offsets", makeHandlerFunc(commitOffsetsRequests, errors))
	node.Handle("list_committed_offsets", makeHandlerFunc(listCommittedOffsetsRequests, errors))
	go func() {
		for {
			select {
			case msg := <-sendRequests:
				var send sendRequest
				if err := json.Unmarshal(msg.Body, &send); err != nil {
					log.Fatalln(err)
				}
				primary := node.NodeIDs()[0]
				if node.ID() != primary {
					fromPrimary, _ := retryOnTimeout(func() (maelstrom.Message, error) {
						return syncRPCWithTimeout(node, primary, send)
					})
					errors <- node.Reply(msg, fromPrimary.Body)
					continue
				}
				journal = append(journal, kv{k: send.Key, v: send.Msg})
				errors <- node.Reply(msg, sendResponseMsg(offset))
				offset += 1
			case msg := <-pollRequests:
				var poll pollRequest
				if err := json.Unmarshal(msg.Body, &poll); err != nil {
					log.Fatalln(err)
				}
				primary := node.NodeIDs()[0]
				if node.ID() != primary {
					fromPrimary, _ := retryOnTimeout(func() (maelstrom.Message, error) {
						return syncRPCWithTimeout(node, primary, poll)
					})
					errors <- node.Reply(msg, fromPrimary.Body)
					continue
				}
				msgs := make(map[string][][2]int)
				for offset, kv := range journal {
					if _, ok := poll.Offsets[kv.k]; ok && offset >= poll.Offsets[kv.k] {
						msgs[kv.k] = append(msgs[kv.k], [2]int{offset, kv.v})
					}
				}
				errors <- node.Reply(msg, pollResponseMsg(msgs))
			case msg := <-commitOffsetsRequests:
				var commitOffsets commitOffsetsRequest
				if err := json.Unmarshal(msg.Body, &commitOffsets); err != nil {
					log.Fatalln(err)
				}
				primary := node.NodeIDs()[0]
				if node.ID() != primary {
					fromPrimary, _ := retryOnTimeout(func() (maelstrom.Message, error) {
						return syncRPCWithTimeout(node, primary, commitOffsets)
					})
					errors <- node.Reply(msg, fromPrimary.Body)
					continue
				}
				for key, offset := range commitOffsets.Offsets {
					if committed[key] < offset {
						committed[key] = offset
					}
				}
				errors <- node.Reply(msg, commitOffsetsResponseMsg)
			case msg := <-listCommittedOffsetsRequests:
				var listCommittedOffsets listCommittedOffsetsRequest
				if err := json.Unmarshal(msg.Body, &listCommittedOffsets); err != nil {
					log.Fatalln(err)
				}
				primary := node.NodeIDs()[0]
				if node.ID() != primary {
					fromPrimary, _ := retryOnTimeout(func() (maelstrom.Message, error) {
						return syncRPCWithTimeout(node, primary, listCommittedOffsets)
					})
					errors <- node.Reply(msg, fromPrimary.Body)
					continue
				}
				errors <- node.Reply(msg, listCommittedOffsetsResponseMsg(committed))
			}
		}
	}()
	if err := node.Run(); err != nil {
		log.Fatalln("error running the node", err)
	}
}
