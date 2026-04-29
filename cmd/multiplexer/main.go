package main

import "github.com/iondodon/multiplexer/internal/multiplexer"

func main() {
	var m = multiplexer.Get()
	m.Start()
}
