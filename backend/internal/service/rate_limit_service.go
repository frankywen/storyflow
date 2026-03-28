package service

import (
	"context"
	"sync"
	"time"
)

// RateLimitService simple in-memory rate limiter
type RateLimitService struct {
	limits       map[string]*limitEntry
	mu           sync.RWMutex
	defaultLimit int
	window       time.Duration
}

type limitEntry struct {
	count     int
	expiresAt time.Time
}

func NewRateLimitService(limit int, window time.Duration) *RateLimitService {
	return &RateLimitService{
		limits:       make(map[string]*limitEntry),
		defaultLimit: limit,
		window:       window,
	}
}

// Check checks if the request is within rate limit
func (s *RateLimitService) Check(ctx context.Context, key string) bool {
	// Lazy cleanup: occasionally clean up expired entries
	if len(s.limits) > 1000 {
		go s.Cleanup()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.limits[key]

	if !exists || entry.expiresAt.Before(now) {
		s.limits[key] = &limitEntry{
			count:     1,
			expiresAt: now.Add(s.window),
		}
		return true
	}

	if entry.count >= s.defaultLimit {
		return false
	}

	entry.count++
	return true
}

// Cleanup cleans up expired entries
func (s *RateLimitService) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.limits {
		if entry.expiresAt.Before(now) {
			delete(s.limits, key)
		}
	}
}