package main

import (
	"testing"
	"time"
)

func TestCooldown(t *testing.T) {
	k := dcKey{DC: 1, IsMedia: false}
	clearCooldown(k)
	if inCooldown(k) {
		t.Fatal("not in cooldown after clear")
	}
	setCooldown(k)
	if !inCooldown(k) {
		t.Fatal("expected in cooldown after set")
	}
	clearCooldown(k)
	if inCooldown(k) {
		t.Fatal("expected not in cooldown after clear")
	}
}

func TestBlacklistTTL(t *testing.T) {
	k := dcKey{DC: 2, IsMedia: true}
	setBlacklisted(k)
	if !isBlacklisted(2, true) {
		t.Fatal("expected blacklisted right after set")
	}
	if isBlacklisted(3, false) {
		t.Fatal("unrelated key must not be blacklisted")
	}

	// Force expiry by rewinding the stored deadline into the past.
	blMu.Lock()
	blacklist[k] = time.Now().Add(-time.Minute)
	blMu.Unlock()

	if isBlacklisted(2, true) {
		t.Fatal("expected expired blacklist entry to report false")
	}
	// Expired entry must be lazily removed.
	blMu.Lock()
	_, ok := blacklist[k]
	blMu.Unlock()
	if ok {
		t.Fatal("expired blacklist entry should be deleted")
	}
}
