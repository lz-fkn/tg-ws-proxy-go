package main

import (
	"strings"
	"sync/atomic"
	"testing"
)

func TestStatsSummaryIncludesFrontingConnections(t *testing.T) {
	var s Stats
	atomic.AddInt64(&s.connectionsFront, 3)

	got := s.summary()
	if !strings.Contains(got, "front=3") {
		t.Fatalf("summary %q does not include fronting counter", got)
	}
}
