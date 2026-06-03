package service

import (
	"fmt"
	"math/rand"
	"time"
)

func uniqueMarketID(prefix string, parts ...int) string {
	id := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	for _, part := range parts {
		id += fmt.Sprintf("-%d", part)
	}
	return fmt.Sprintf("%s-%d", id, rand.Int63())
}
