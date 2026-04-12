package storage

import (
	"sync"
	"time"
)

// URLData stores the original URL and its expiration time
type URLData struct {
	LongURL   string
	ExpiresAt time.Time
}

// URLmap stores short ID to original URL mapping
var URLmap = make(map[string]URLData)

// Reversemap stores userID:originalURL to short ID mapping for deduplication
var Reversemap = make(map[string]string)

var Mu sync.Mutex
