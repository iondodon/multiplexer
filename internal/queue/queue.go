package queue

import (
	"log/slog"
	"sync"
)

var (
	head  *Node      = nil
	mutex sync.Mutex = sync.Mutex{}
	cond  *sync.Cond = sync.NewCond(&mutex)
)

type Node struct {
	data []byte
	next *Node
}

func GetDedicatedReader() *Node {
	mutex.Lock()
	defer mutex.Unlock()
	for head == nil {
		cond.Wait()
	}
	return head
}

func Push(data []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	if head == nil {
		head = &Node{data: data, next: nil}
	} else {
		head.next = &Node{data: data, next: nil}
		head = head.next
	}
	cond.Broadcast()
	slog.Info("pushed", "data", data)
}

func (n *Node) Read() ([]byte, *Node) {
	mutex.Lock()
	defer mutex.Unlock()

	for n == head {
		cond.Wait()
	}

	data := n.data
	slog.Info("read", "data", data)
	return data, n.next
}
