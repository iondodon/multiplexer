package queue

import "github.com/stretchr/testify/mock"

type MockQueue struct {
	mock.Mock
}

func (mq *MockQueue) Push(data []byte) {
	mq.Called(data)
}

func (mq *MockQueue) GetCursor() Cursor {
	args := mq.Called()
	return args.Get(0).(Cursor)
}

type MockCursor struct {
	mock.Mock
}

func (mc *MockCursor) HasNext() bool {
	args := mc.Called()
	return args.Bool(0)
}

func (mc *MockCursor) Get() []byte {
	args := mc.Called()
	return args.Get(0).([]byte)
}

func (mc *MockCursor) Next() {
	mc.Called()
}
