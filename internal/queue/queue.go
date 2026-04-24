package queue

import "sync"

var (
	instance *queue
	once     *sync.Once  = &sync.Once{}
	mutex    *sync.Mutex = &sync.Mutex{}
)

type queue struct {
	head *node
}

type node struct {
	data string
	next *node
}

func GetInstance() *queue {
	once.Do(func() {
		instance = &queue{head: nil}
	})
	return instance
}

func (q *queue) Push(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	if q.head == nil {
		q.head = &node{data: data, next: nil}
	} else {
		q.head.next = &node{data: data, next: nil}
		q.head = q.head.next
	}
}

func (q *queue) ReadNext() (string, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	if q.head != nil {
		data := q.head.data
		q.head = q.head.next
		return data, true
	} else {
		return "", false
	}
}
