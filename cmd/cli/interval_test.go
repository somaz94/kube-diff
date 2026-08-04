package cli

import (
	"testing"
	"time"
)

func TestNextRunDelay(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		interval time.Duration
		lastRun  time.Time
		want     time.Duration
	}{
		{
			name:     "no interval runs immediately",
			interval: 0,
			lastRun:  now.Add(-time.Second),
			want:     0,
		},
		{
			name:     "negative interval runs immediately",
			interval: -time.Second,
			lastRun:  now.Add(-time.Second),
			want:     0,
		},
		{
			name:     "interval already elapsed runs immediately",
			interval: 10 * time.Second,
			lastRun:  now.Add(-30 * time.Second),
			want:     0,
		},
		{
			name:     "exactly at the boundary runs immediately",
			interval: 10 * time.Second,
			lastRun:  now.Add(-10 * time.Second),
			want:     0,
		},
		{
			name:     "inside the interval waits out the remainder",
			interval: 10 * time.Second,
			lastRun:  now.Add(-4 * time.Second),
			want:     6 * time.Second,
		},
		{
			name:     "a change right after a run waits the whole interval",
			interval: time.Minute,
			lastRun:  now,
			want:     time.Minute,
		},
		{
			name:     "the zero lastRun of a fresh watcher runs immediately",
			interval: time.Minute,
			lastRun:  time.Time{},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextRunDelay(tt.interval, tt.lastRun, now); got != tt.want {
				t.Errorf("nextRunDelay(%v, lastRun=%v) = %v, want %v", tt.interval, tt.lastRun, got, tt.want)
			}
		})
	}
}

// A change landing inside the interval must be deferred, never dropped: the
// delay is always positive so the caller can reschedule instead of discarding.
func TestNextRunDelay_NeverDropsAPendingChange(t *testing.T) {
	interval := 30 * time.Second
	lastRun := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	// Walk the whole throttled window; every point must yield a runnable moment.
	for elapsed := time.Duration(0); elapsed < interval; elapsed += time.Second {
		now := lastRun.Add(elapsed)
		delay := nextRunDelay(interval, lastRun, now)
		if delay <= 0 {
			t.Fatalf("elapsed=%v: delay = %v, want a positive deferral", elapsed, delay)
		}
		// Deferring by exactly this delay must land on a runnable moment.
		if after := nextRunDelay(interval, lastRun, now.Add(delay)); after != 0 {
			t.Errorf("elapsed=%v: after waiting %v the change still defers by %v, want 0", elapsed, delay, after)
		}
	}
}
