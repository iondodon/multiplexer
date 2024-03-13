# Multiplexer

This multiplexer server receives messages via SSH from a remote sender and forwards them to client subscribers. The multiplexer takes each message from the remote sender and places them in a queue. Each subscribed client has a dedicated goroutine (called a client goroutine) that iterates through the queue and sends each message to the client.

## How to run

Create a `.env` file from the `.env.example` file and add the necessary configuration values.

- Being in the `multiplexer` directory, use `go run main.go` or `timeout 40s go run main.go` to run the multiplexer.
- At startup, it will try to connect to the defined `SSH_ADDRESS` using SSH and will start a server that accepts client/subscriber connections.
- Use `telnet localhost 8080` to connect as a client to the multiplexer.
- To cancel the multiplexer, use Ctrl+C.

## Used libraries

- golang.org/x/crypto v0.20.0
- github.com/joho/godotenv v1.5.1
- github.com/stretchr/testify v1.9.0
- go.uber.org/mock v0.4.0

## How it works

The multiplexer uses a queue, implemented as a linked list, where it is possible only to push new messages into this queue. Additionally, it is possible to obtain a Cursor for this queue. When a new Cursor is obtained, it points to the head of the queue, which is the most recently added element. The Cursor can be thought of as a representation of the tail of the queue. There can be multiple Cursors, one for each subscriber. The client goroutine will move its Cursor forward, retrieve the message at each position, and send that message to the client. Cursors move forward independently and stop when they reach the head of the queue, waiting for new nodes/messages to become available in the queue.

<img src="architecture.png" alt="Architecture" style="display:block; margin: auto;">

When the last Cursor leaves its current node and moves to the next one, the previous node should become eligible for garbage collection. The methods of using this Queue heavily rely on garbage collection which, as observed during testing, is not necessarily memory efficient due to the significant pressure on garbage collection. If it does not deallocate resources quickly enough, memory usage can consistently increase.

## Assumptions

When a new client connects, a new Cursor is created for it, pointing to the head of the Queue. For performance reasons, the Queue implementation is not synchronized.

- Since the head of the Queue is accessed concurrently by the SourceListener and the new ClientFeeder, the Cursor obtained by the ClientFeeder might not be the most recent message from the queue, but this is not critical.
- Also, the ClientFeeder attempts to concurrently access the next new element in the queue by checking `Cursor.HasNext()`. I assume there is no race condition since the SourceListener is the only one who adds a new element, and `.HasNext()` will return true at some point. Meanwhile, the ClientFeeder will just use `.HasNext()` and then read the message.

## What problems it has

Since this approach relies on the garbage collector to deallocate messages from the queue, memory usage increases. After conducting a test to see how far the Cursors are from the head of the queue when memory increases, it was noticed that they are usually at the head of the queue, waiting for new messages. This indicates that the messages quickly become eligible for garbage collection. A Rust implementation of this method might be more efficient in terms of addressing this problem.

## What's to do next

- More unit tests are needed, especially for the negative scenarios.
- Do benchmarking using `pprof`.
