package queue

import "sync"

var (
	instance *queue
	once     *sync.Once    = &sync.Once{}
	rwMutex  *sync.RWMutex = &sync.RWMutex{}
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
	rwMutex.Lock()
	defer rwMutex.Unlock()

	if q.head == nil {
		q.head = &node{data: data, next: nil}
	} else {
		q.head.next = &node{data: data, next: nil}
		q.head = q.head.next
	}
}

func (q *queue) HasNext() bool {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	return q.head != nil
}

// HasNext must be called first. This can be "improved". But, should we?
func (q *queue) Read() string {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	data := q.head.data
	q.head = q.head.next
	return data
}
