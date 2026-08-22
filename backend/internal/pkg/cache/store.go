// Package cache provides a tiny, dependency-free TTL + LRU in-process cache.
// It is the default cache layer in development / small deployments; a Redis
// adapter can implement the same Store interface without touching call sites.
package cache

import (
	"container/list"
	"sync"
	"time"
)

// Entry is a cached value with an expiry.
type Entry struct {
	Value     any
	ExpiresAt time.Time
}

// Store is the cache abstraction.
type Store interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
	Clear()
	Stats() Stats
}

// Stats reports hit/miss counters.
type Stats struct {
	Hits   int
	Misses int
	Keys   int
}

// MemoryStore is a single-process TTL + LRU store.
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*list.Element
	lru  *list.List // front = most recent
	cap  int
	now  func() time.Time
	hits int
	miss int
}

type item struct {
	key   string
	entry Entry
}

// NewMemory returns a cache holding at most `capacity` live entries (LRU).
func NewMemory(capacity int) *MemoryStore {
	if capacity <= 0 {
		capacity = 1024
	}
	return &MemoryStore{data: map[string]*list.Element{}, lru: list.New(), cap: capacity, now: time.Now}
}

func (m *MemoryStore) Get(key string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if el, ok := m.data[key]; ok {
		it := el.Value.(*item)
		if m.now().After(it.entry.ExpiresAt) {
			m.removeElementLocked(el)
			m.miss++
			return nil, false
		}
		m.lru.MoveToFront(el)
		m.hits++
		return it.entry.Value, true
	}
	m.miss++
	return nil, false
}

func (m *MemoryStore) Set(key string, value any, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if el, ok := m.data[key]; ok {
		it := el.Value.(*item)
		it.entry = Entry{Value: value, ExpiresAt: m.now().Add(ttl)}
		m.lru.MoveToFront(el)
		return
	}
	if len(m.data) >= m.cap {
		m.removeOldestLocked()
	}
	it := &item{key: key, entry: Entry{Value: value, ExpiresAt: m.now().Add(ttl)}}
	el := m.lru.PushFront(it)
	m.data[key] = el
}

func (m *MemoryStore) removeOldestLocked() {
	oldest := m.lru.Back()
	if oldest != nil {
		it := oldest.Value.(*item)
		delete(m.data, it.key)
		m.lru.Remove(oldest)
	}
}

func (m *MemoryStore) removeElementLocked(el *list.Element) {
	it := el.Value.(*item)
	delete(m.data, it.key)
	m.lru.Remove(el)
}

func (m *MemoryStore) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if el, ok := m.data[key]; ok {
		m.removeElementLocked(el)
	}
}

func (m *MemoryStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = map[string]*list.Element{}
	m.lru.Init()
}

func (m *MemoryStore) Stats() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Stats{Hits: m.hits, Misses: m.miss, Keys: len(m.data)}
}

// DefaultTTL is the TTL applied to hot read caches.
const DefaultTTL = 30 * time.Second

// Key helpers for common cache keys.
func ServerListKey() string            { return "servers:list" }
func ServerKey(id string) string       { return "servers:get:" + id }
func AlertRulesKey() string            { return "alert:rules" }
func MetricsLatestKey(serverID string) string { return "metrics:latest:" + serverID }
func ModelsKey() string                { return "ai:models" }
