package httpx

import (
	"sync"
	"time"
)

type ipLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newIPLimiter() *ipLimiter {
	return &ipLimiter{hits: make(map[string][]time.Time)}
}

func (l *ipLimiter) allow(ip string, n int, windowSec int) bool {
	now := time.Now()
	cut := now.Add(-time.Duration(windowSec) * time.Second)

	l.mu.Lock()
	defer l.mu.Unlock()

	kept := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if cap(kept) == 0 {
		kept = []time.Time{}
	}
	if len(kept) >= n {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}
