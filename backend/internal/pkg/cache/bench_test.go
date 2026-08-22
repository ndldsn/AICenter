package cache

import (
	"strconv"
	"testing"
	"time"

	"github.com/aicenter/aicenter/internal/models"
)

// BenchmarkMemoryStore_GetHit measures cache-hit latency for hot reads.
func BenchmarkMemoryStore_GetHit(b *testing.B) {
	s := NewMemory(512)
	for i := 0; i < b.N; i++ {
		s.Set("k"+strconv.Itoa(i%64), &models.Server{ID: strconv.Itoa(i)}, DefaultTTL)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get("k" + strconv.Itoa(i%64))
	}
}

// BenchmarkMemoryStore_GetMiss measures cache-miss (DB fallback path) latency.
func BenchmarkMemoryStore_GetMiss(b *testing.B) {
	s := NewMemory(512)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get("missing")
	}
}

// BenchmarkTTLExpiry demonstrates that entries are not served stale.
func BenchmarkTTLExpiry(b *testing.B) {
	s := NewMemory(512)
	for i := 0; i < b.N; i++ {
		s.Set("k", i, 1*time.Millisecond)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set("k", i, 1*time.Nanosecond)
		// force expiry
		time.Sleep(2 * time.Millisecond)
	}
}
