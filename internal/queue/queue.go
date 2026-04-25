package queue

import (
	"log/slog"
	"sync"
)

var (
	statingNode *node  = &node{}
	instance    *queue = &queue{
		head: statingNode,
		tail: statingNode,
	}
	mutex *sync.Mutex = &sync.Mutex{}
)

type queue struct {
	head *node
	tail *node
}

type node struct {
	data string
	next *node
}

func GetInstance() *queue {
	return instance
}

func (q *queue) Push(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	slog.Info("Pushed", "data", data)
	q.head.next = &node{data: data, next: nil}
	q.head = q.head.next
}

func (q *queue) ReadNext() (string, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	if q.tail != q.head {
		data := q.tail.data
		slog.Info("Read", "data", data)
		q.tail = q.tail.next
		return data, true
	} else {
		slog.Info("Should wait")
		return "", false
	}
}
