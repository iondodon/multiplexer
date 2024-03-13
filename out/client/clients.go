package client

import (
	"sync"
	"sync/atomic"
)

type ClientsCounter struct {
	NClients *atomic.Uint64
	Cond     *sync.Cond
}
