package queue

import "sync"

var (
	instance *q
	once     sync.Once
)

type q struct {
	head *node
}

type node struct {
	data string
	next *node
}

func GetInstance() *q {
	once.Do(func() {
		instance = &q{}
	})
	return instance
}

func (q q) Push(data string) {
	if q.head == nil {
		q.head = &node{data: data}
		return
	}

	q.head.next = &node{data: data}
	q.head = q.head.next
}
