package common

import "time"

type CacheEntry struct {
	Id        uint
	Value     []byte
	Timestamp time.Time
}
