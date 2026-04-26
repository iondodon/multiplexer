package queue

import (
	"log/slog"
	"sync"
)

var (
	statingNode *Node = &Node{}
	head        *Node = statingNode
	// The starting point should always be one step behind the head.
	startingPoint *Node         = statingNode
	mutex         *sync.RWMutex = &sync.RWMutex{}
)

type Node struct {
	data string
	next *Node
}

func GetReader() *Node {
	mutex.RLock()
	defer mutex.RUnlock()
	return startingPoint
}

func Push(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	slog.Info("pushed", "data", data)
	head.next = &Node{data: data, next: nil}
	startingPoint = head
	head = head.next
}

func (n *Node) ReadNext() (string, *Node) {
	mutex.Lock()
	defer mutex.Unlock()

	if n != head {
		data := n.data
		slog.Info("read", "data", data)
		return data, n.next
	} else {
		slog.Info("should wait")
		return "", nil
	}
}
