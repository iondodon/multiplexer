package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue_InitHeadOnPushIfNil(t *testing.T) {
	q := New().(*queue)
	q.head = nil

	q.Push([]byte("data"))

	assert.NotNil(t, q.head)
	assert.Equal(t, []byte("data"), q.head.data)
}

func TestQueue_PushAndGetCursor_EmptyQueue(t *testing.T) {
	q := New().(*queue)
	assert.NotNil(t, q.head, "Head should not be nil after initialization")

	// Initial queue state should have a dummy node
	c := q.GetCursor().(*cursor)
	assert.Empty(t, c.head.data, "Initial cursor data should be empty slice")
	assert.False(t, c.HasNext(), "Empty queue cursor should not have next")
}

func TestQueue_PushAndCursorTraversal(t *testing.T) {
	q := New().(*queue)

	q.Push([]byte("first"))

	c := q.GetCursor().(*cursor)
	assert.NotNil(t, c.Get(), "Cursor should point to the last element")
	assert.False(t, c.HasNext(), "Cursor should not have next element")

	q.Push([]byte("second"))

	assert.NotNil(t, c.Get(), "Cursor should be set to the second last element")
	assert.True(t, c.HasNext(), "Course should have a next element")

	c.Next()

	assert.NotNil(t, c.Get(), "Cursor  is the last element in queue")
	assert.False(t, c.HasNext(), "Course should not have a next element")
}

func TestQueue_HasNext_ReturnsFalseIfNilCursor(t *testing.T) {
	var q queue

	assert.False(t, q.GetCursor().HasNext())
}

func TestQueue_Get_ReturnsNilIfNilCursor(t *testing.T) {
	var q queue

	assert.Nil(t, q.GetCursor().Get())
}
