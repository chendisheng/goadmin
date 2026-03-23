package watcher

import (
	"fmt"
	"sync"
)

type Watcher interface {
	SetUpdateCallback(func() error) error
	Notify() error
	Close() error
}

type RedisWatcher struct {
	mu       sync.RWMutex
	channel  string
	callback func() error
	closed   bool
}

func NewRedisWatcher(channel string) (*RedisWatcher, error) {
	if channel == "" {
		return nil, fmt.Errorf("watcher channel is required")
	}
	return &RedisWatcher{channel: channel}, nil
}

func (w *RedisWatcher) SetUpdateCallback(cb func() error) error {
	if w == nil {
		return fmt.Errorf("watcher is nil")
	}
	w.mu.Lock()
	w.callback = cb
	w.mu.Unlock()
	return nil
}

func (w *RedisWatcher) Notify() error {
	if w == nil {
		return nil
	}
	w.mu.RLock()
	callback := w.callback
	closed := w.closed
	w.mu.RUnlock()
	if closed || callback == nil {
		return nil
	}
	return callback()
}

func (w *RedisWatcher) Close() error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
	return nil
}
