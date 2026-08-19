package main

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestIdleWatchdogKeepsOneWayTrafficAlive(t *testing.T) {
	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())
	stop := make(chan struct{})
	fired := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		runIdleWatchdog(&lastActivity, stop, 30*time.Millisecond, 2*time.Millisecond, func() { fired <- struct{}{} })
		close(done)
	}()

	deadline := time.Now().Add(80 * time.Millisecond)
	for time.Now().Before(deadline) {
		lastActivity.Store(time.Now().UnixNano())
		time.Sleep(5 * time.Millisecond)
	}
	select {
	case <-fired:
		t.Fatal("watchdog fired despite continuous traffic in one direction")
	default:
	}

	close(stop)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchdog did not stop")
	}
}

func TestIdleWatchdogClosesIdleSession(t *testing.T) {
	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())
	stop := make(chan struct{})
	fired := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		runIdleWatchdog(&lastActivity, stop, 20*time.Millisecond, 2*time.Millisecond, func() { fired <- struct{}{} })
		close(done)
	}()

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("watchdog did not close an idle session")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchdog did not finish after closing idle session")
	}
}
