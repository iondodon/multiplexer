package source

const ChunkSize = 64

func Fill0(message []byte, chunkSize int) []byte {
	msgLen := len(message)

	padSize := chunkSize - (msgLen-1)%chunkSize - 1

	return append(message, make([]byte, padSize)...)
}
