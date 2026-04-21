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
		instance = &q{
			head: &node{
				// Intentionally added the first node with empty "data"
				// because we do not want to check if the head is nil
				// everytime when we push into the queue.
				data: "",
				next: nil,
			},
		}
	})
	return instance
}

func (q q) Push(data string) {
	q.head.next = &node{data: data}
	q.head = q.head.next
}
