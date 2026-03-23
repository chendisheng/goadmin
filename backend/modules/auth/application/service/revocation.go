package service

import (
	"context"
	"sync"
	"time"
)

type MemoryRevocationStore struct {
	mu   sync.RWMutex
	data map[string]time.Time
}

func NewMemoryRevocationStore() *MemoryRevocationStore {
	return &MemoryRevocationStore{data: make(map[string]time.Time)}
}

func (s *MemoryRevocationStore) Revoke(_ context.Context, jti string, expiresAt time.Time) error {
	if s == nil || jti == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[jti] = expiresAt.UTC()
	return nil
}

func (s *MemoryRevocationStore) IsRevoked(_ context.Context, jti string) (bool, error) {
	if s == nil || jti == "" {
		return false, nil
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.data[jti]
	if !ok {
		return false, nil
	}
	if !expiresAt.IsZero() && now.After(expiresAt) {
		delete(s.data, jti)
		return false, nil
	}
	return true, nil
}
