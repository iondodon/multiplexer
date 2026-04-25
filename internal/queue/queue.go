package queue

import (
	"log/slog"
	"sync"
)

var (
	statingNode   *Node       = &Node{}
	head          *Node       = statingNode
	startingPoint *Node       = statingNode
	mutex         *sync.Mutex = &sync.Mutex{}
)

type Node struct {
	data string
	next *Node
}

func GetReader() *Node {
	mutex.Lock()
	defer mutex.Unlock()
	return startingPoint
}

func Push(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	slog.Info("Pushed", "data", data)
	head.next = &Node{data: data, next: nil}
	startingPoint = head
	head = head.next
}

func (n *Node) ReadNext() (string, bool, *Node) {
	mutex.Lock()
	defer mutex.Unlock()

	if n != head {
		data := n.data
		slog.Info("Read", "data", data)
		return data, true, n.next
	} else {
		slog.Info("Should wait")
		return "", false, nil
	}
}
