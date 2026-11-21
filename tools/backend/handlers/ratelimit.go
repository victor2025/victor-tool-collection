package handlers

import (
	"sync"
	"time"
)

// rateLimiter 是进程内、按 key 的固定窗口限流器。
// 单机单进程部署（systemd 一个实例）下足够用，不引入 Redis 等额外依赖。
type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	window  time.Duration
	max     int
}

type rateEntry struct {
	count int
	start time.Time
}

func newRateLimiter(window time.Duration, max int) *rateLimiter {
	return &rateLimiter{
		entries: make(map[string]*rateEntry),
		window:  window,
		max:     max,
	}
}

// Allow 返回是否放行；被拒时同时返回建议的重试等待时长。
func (l *rateLimiter) Allow(key string) (bool, time.Duration) {
	if key == "" {
		return true, 0
	}
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// 顺手清理过期的 key，避免 map 无限增长
	if len(l.entries) > 4096 {
		for k, e := range l.entries {
			if now.Sub(e.start) > l.window {
				delete(l.entries, k)
			}
		}
	}

	e, ok := l.entries[key]
	if !ok || now.Sub(e.start) >= l.window {
		l.entries[key] = &rateEntry{count: 1, start: now}
		return true, 0
	}
	if e.count >= l.max {
		return false, l.window - now.Sub(e.start)
	}
	e.count++
	return true, 0
}
