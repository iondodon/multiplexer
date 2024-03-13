package queue

type Queue interface {
	Push(data []byte)
	GetCursor() Cursor
}

type Cursor interface {
	HasNext() bool
	Get() []byte
	Next()
}

type node struct {
	data []byte
	next *node
}

type queue struct {
	head *node
}

type cursor struct {
	head *node
}

func (q *queue) GetCursor() Cursor {
	return &cursor{q.head}
}

func (q *queue) Push(data []byte) {
	if q.head == nil {
		q.head = &node{data: data}
		return
	}
	q.head.next = &node{data: data}
	q.head = q.head.next
}

func New() Queue {
	return &queue{
		head: &node{data: nil},
	}
}

func (c *cursor) HasNext() bool {
	if c.head == nil {
		return false
	}
	return c.head.next != nil
}

func (c *cursor) Get() []byte {
	if c.head == nil {
		return nil
	}
	return c.head.data
}

func (c *cursor) Next() {
	if c.head == nil {
		return
	}
	c.head = c.head.next
}
