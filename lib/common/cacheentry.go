package common

import "time"

type CacheEntry struct {
	Id        uint
	Key       string
	Value     []byte
	Timestamp time.Time
}
