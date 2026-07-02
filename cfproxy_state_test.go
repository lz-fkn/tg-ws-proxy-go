package main

import "testing"

func TestSetAndTryCFProxyDomains(t *testing.T) {
	cfg := &Config{}
	cfg.setCFProxyDomains([]string{"a.tld", "b.tld", "a.tld"})

	if !cfg.hasCFProxyDomains() {
		t.Fatal("expected domains present")
	}
	if cfg.cfproxyDomainPoolSize() != 2 {
		t.Fatalf("pool size = %d, want 2 (dedup)", cfg.cfproxyDomainPoolSize())
	}
	if cfg.cfproxyActiveDomain() == "" {
		t.Fatal("active domain must be set")
	}

	order := cfg.cfproxyDomainsForTry(2)
	if len(order) != 2 {
		t.Fatalf("cfproxyDomainsForTry = %v, want 2 entries", order)
	}
	wantFirst := normalizeCFProxyDomain(cfg.FallbackCFProxyPerDCActive[2])
	if wantFirst == "" {
		wantFirst = cfg.cfproxyActiveDomain()
	}
	if order[0] != wantFirst {
		t.Errorf("per-DC active must be first: order=%v wantFirst=%q", order, wantFirst)
	}
	seen := map[string]bool{order[0]: true, order[1]: true}
	if !seen["a.tld"] || !seen["b.tld"] {
		t.Errorf("both domains must appear: %v", order)
	}
}

func TestSetCFProxyDomainsEmptyFallsBackToDefault(t *testing.T) {
	cfg := &Config{}
	cfg.setCFProxyDomains(nil)
	if !cfg.hasCFProxyDomains() {
		t.Fatal("empty input should fall back to default pool")
	}
}

func TestPromoteCFProxyDomain(t *testing.T) {
	cfg := &Config{}
	cfg.setCFProxyDomains([]string{"a.tld", "b.tld"})
	cfg.promoteCFProxyDomain(2, "b.tld")
	if cfg.cfproxyActiveDomain() != "b.tld" {
		t.Errorf("active = %q, want b.tld after promote", cfg.cfproxyActiveDomain())
	}
	// promoting an unknown domain must not change anything
	cfg.promoteCFProxyDomain(2, "zzz.tld")
	if cfg.cfproxyActiveDomain() != "b.tld" {
		t.Error("unknown domain must not be promoted")
	}
}

func TestCFProxyWorkerDomains(t *testing.T) {
	cfg := &Config{}
	if cfg.hasCFProxyWorkerDomains() {
		t.Fatal("no worker domains by default")
	}
	cfg.FallbackCFProxyWorkerDomains = []string{"w1.workers.dev", "w2.workers.dev"}
	if !cfg.hasCFProxyWorkerDomains() {
		t.Fatal("expected worker domains present")
	}
	got := cfg.cfproxyWorkerDomainsForTry()
	if len(got) != 2 {
		t.Fatalf("worker domains = %v, want 2", got)
	}
}
