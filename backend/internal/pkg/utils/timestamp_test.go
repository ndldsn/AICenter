package utils

import (
	"testing"
	"time"
)

func TestParseTimestamp_SQLiteFormat(t *testing.T) {
	tt, err := ParseTimestamp("2006-01-02 15:04:05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tt.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Fatalf("got %v", tt)
	}
}

func TestParseTimestamp_RFC3339(t *testing.T) {
	tt, err := ParseTimestamp("2006-01-02T15:04:05Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tt.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Fatalf("got %v", tt)
	}
}

func TestParseTimestamp_EveningTime(t *testing.T) {
	// Regression guard: the old "2006-01-02 15:04:05" helper returned
	// 23:59:59 for "2026-08-22 23:59:59" when only RFC3339Nano +
	// RFC3339 + 2006-01-02T15:04:05 were tried; the space-separated
	// layout must stay reachable.
	tt, err := ParseTimestamp("2026-08-22 23:59:59")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tt.Equal(time.Date(2026, 8, 22, 23, 59, 59, 0, time.UTC)) {
		t.Fatalf("got %v (want 2026-08-22 23:59:59 UTC)", tt)
	}
}

func TestParseTimestamp_Empty(t *testing.T) {
	tt, err := ParseTimestamp("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tt.IsZero() {
		t.Fatalf("expected zero time, got %v", tt)
	}
}