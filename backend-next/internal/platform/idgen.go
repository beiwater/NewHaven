package platform

import (
	"fmt"
	"sync/atomic"
	"time"
)

// IDGen generates unique IDs. Thread-safe.
type IDGen struct {
	counter atomic.Int64
}

func NewIDGen() *IDGen {
	return &IDGen{}
}

// Next returns a globally unique ID based on nanosecond timestamp + sequence.
func (g *IDGen) Next(prefix string) string {
	n := g.counter.Add(1)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), n)
}

// NanoID returns a bare nanosecond-based ID without prefix.
func (g *IDGen) NanoID() string {
	n := g.counter.Add(1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), n)
}
