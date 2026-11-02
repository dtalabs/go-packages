package limiter

import (
	"sync"
	"time"
)

type attempt struct {
	count int
	start time.Time
}

type Limiter struct {
	mu     sync.Mutex
	values map[string]attempt
	limit  int
	window time.Duration
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{values: map[string]attempt{}, limit: limit, window: window}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	a := l.values[key]
	if a.start.IsZero() || now.Sub(a.start) >= l.window {
		l.values[key] = attempt{1, now}
		return true
	}
	if a.count >= l.limit {
		return false
	}
	a.count++
	l.values[key] = a
	return true
}
