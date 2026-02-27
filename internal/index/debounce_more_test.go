package index

import (
	"context"
	"testing"
	"time"
)

// TestDebounceContextCancelDuringRun verifies that canceling after delay
// prevents results from being emitted.
// This ensures stale searches are discarded.
func TestDebounceContextCancelDuringRun(t *testing.T) {
	exec := DebouncedExecutor[int]{Delay: 10 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	ch := exec.Run(ctx, func(ctx context.Context) (int, error) {
		cancel()
		return 1, nil
	})
	if res, ok := <-ch; ok && res.Value != 0 {
		t.Fatalf("expected no result after cancel")
	}
}
