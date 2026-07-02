package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func parseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("tg-ws-proxy-go", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	host := fs.String("host", "127.0.0.1", "Listen host")
	port := fs.Int("port", 1443, "Listen port")
	secret := fs.String("secret", "", "MTProto secret (32 hex chars)")
	genSecret := fs.Bool("gen-secret", false, "Generate random secret and print it")
	printLink := fs.Bool("print-link", false, "Print the tg:// connect link and exit")
	logLevel := fs.Int("loglevel", 3, "Log Level (0=FATAL, 1=ERROR, 2=WARN, 3=INFO, 4=VERBOSE)")
	useTimestamps := fs.Bool("timestamps", false, "Show timestamps in logs")
	bufKB := fs.Int("buf-kb", 256, "Socket buffer size in KB")
	poolSize := fs.Int("pool-size", 4, "WS pool size per DC")
	fakeTLSDomain := fs.String("fake-tls-domain", "", "Enable Fake TLS (ee-secret) with masking domain (https://github.com/Flowseal/tg-ws-proxy/blob/main/docs/FakeTlsNginx.md)")
	cfproxyDomain := fs.String("cfproxy-domain", defaultCFProxyDomain, "Cloudflare-proxied domain for WS fallback")
	cfproxyDomains := fs.String("cfproxy-domains", "", "Comma-separated Cloudflare proxy domain pool for WS fallback")
	cfproxyWorkerDomains := fs.String("cfproxy-worker-domain", "", "Comma-separated Cloudflare Worker domain(s) for WS fallback (e.g. name-1234.user.workers.dev); tried first when set")
	noCfproxy := fs.Bool("no-cfproxy", false, "Disable Cloudflare proxy fallback")
	cfproxyPriority := fs.Bool("cfproxy-priority", true, "Try cfproxy before TCP fallback")
	noCfproxyDomainRefresh := fs.Bool("no-cfproxy-domain-refresh", false, "Disable periodic CF proxy domain refresh from URL")
	cfproxyDomainsURL := fs.String("cfproxy-domains-url", defaultCFProxyDomainsURL, "URL to fetch CF proxy domain list from")
	maxConns := fs.Int("max-conns", defaultMaxConns, "Max concurrent client sessions")
	dcIPDefault := fs.String("dc-ip-default", "149.154.167.220", "Default WS target IP for all implicit DCs when --dc-ip is not provided")
	dcIPDefaultPool := fs.String("dc-ip-default-pool", "", "Default WS target IP pool for implicit DCs, comma-separated")
	credits := fs.Bool("credits", false, "Show credits and exit")

	var dcIPs multiFlag
	var dcIPPools multiFlag
	fs.Var(&dcIPs, "dc-ip", "Target DC IP as DC:IP; repeatable")
	fs.Var(&dcIPPools, "dc-ip-pool", "Target pool as DC:IP1,IP2,...; repeatable")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })

	LogLevel = *logLevel
	UseTimestamps = *useTimestamps

	if *printLink && *secret == "" {
		return nil, errors.New("--print-link requires --secret")
	}

	if *secret == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		*secret = hex.EncodeToString(b)
		if !*genSecret {
			Info("Generated secret: %s", *secret)
		}
	}
	if len(*secret) != 32 {
		return nil, errors.New("secret must be exactly 32 hex chars")
	}
	if _, err := hex.DecodeString(*secret); err != nil {
		return nil, errors.New("secret must be valid hex")
	}

	defaultTargetIP := strings.TrimSpace(*dcIPDefault)
	if net.ParseIP(defaultTargetIP) == nil {
		return nil, fmt.Errorf("invalid --dc-ip-default: %s", defaultTargetIP)
	}

	defaultPool := []string{defaultTargetIP}
	if strings.TrimSpace(*dcIPDefaultPool) != "" {
		poolIPs, err := parseIPCSV(*dcIPDefaultPool)
		if err != nil {
			return nil, fmt.Errorf("invalid --dc-ip-default-pool: %w", err)
		}
		defaultPool = poolIPs
	}

	dcMap := map[int]string{}
	dcPool := map[int][]string{}
	for _, dc := range []int{2, 4} {
		dcPool[dc] = append([]string(nil), defaultPool...)
		dcMap[dc] = defaultPool[0]
	}

	for _, item := range dcIPs {
		parts := strings.SplitN(item, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --dc-ip: %s", item)
		}
		dc, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid dc: %s", parts[0])
		}
		if net.ParseIP(parts[1]) == nil {
			return nil, fmt.Errorf("invalid ip: %s", parts[1])
		}
		ip := parts[1]
		dcPool[dc] = []string{ip}
		dcMap[dc] = ip
	}

	for _, item := range dcIPPools {
		parts := strings.SplitN(item, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --dc-ip-pool: %s", item)
		}
		dc, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid dc in --dc-ip-pool: %s", parts[0])
		}
		poolIPs, err := parseIPCSV(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid --dc-ip-pool for dc %d: %w", dc, err)
		}
		dcPool[dc] = append([]string(nil), poolIPs...)
		dcMap[dc] = dcPool[dc][0]
	}

	userDomainProvided := provided["cfproxy-domain"]
	userPoolProvided := strings.TrimSpace(*cfproxyDomains) != ""
	userDomain := normalizeCFProxyDomain(*cfproxyDomain)
	userFixedDomain := userDomainProvided && userDomain != ""

	domainPool := defaultCFProxyDomains()
	if userPoolProvided {
		if userDomainProvided {
			return nil, errors.New("use only one of --cfproxy-domain or --cfproxy-domains")
		}
		parsedDomains, err := parseCFProxyDomainCSV(*cfproxyDomains)
		if err != nil {
			return nil, fmt.Errorf("invalid --cfproxy-domains: %w", err)
		}
		domainPool = parsedDomains
	}

	if userDomainProvided {
		if strings.Contains(userDomain, ",") {
			return nil, errors.New("invalid --cfproxy-domain: multiple domains are not allowed; use --cfproxy-domains for comma-separated pool")
		}
		if userDomain == "" {
			return nil, errors.New("invalid --cfproxy-domain: empty domain")
		}
		domainPool = []string{userDomain}
	} else if !userPoolProvided && userDomain != "" {
		domainPool = appendUniqueDomains(domainPool, userDomain)
	}

	normalizedFakeTLSDomain := normalizeCFProxyDomain(*fakeTLSDomain)
	if normalizedFakeTLSDomain != "" && !isLikelyDomain(normalizedFakeTLSDomain) {
		return nil, fmt.Errorf("invalid --fake-tls-domain: %s", *fakeTLSDomain)
	}

	var workerDomains []string
	if strings.TrimSpace(*cfproxyWorkerDomains) != "" {
		wd, err := parseCFProxyDomainCSV(*cfproxyWorkerDomains)
		if err != nil {
			return nil, fmt.Errorf("invalid --cfproxy-worker-domain: %w", err)
		}
		workerDomains = wd
	}

	cfg := &Config{
		Host:                         *host,
		Port:                         *port,
		SecretHex:                    *secret,
		GenSecret:                    *genSecret,
		PrintLink:                    *printLink,
		FakeTLSDomain:                normalizedFakeTLSDomain,
		DCMap:                        dcMap,
		DCPool:                       dcPool,
		FallbackCFProxy:              !*noCfproxy,
		FallbackCFProxyPriority:      *cfproxyPriority,
		FallbackCFProxyDomain:        "",
		FallbackCFProxyWorkerDomains: workerDomains,
		FallbackCFProxyUserDomain:    userFixedDomain || userPoolProvided,
		FallbackCFProxyRefresh:       !*noCfproxyDomainRefresh,
		FallbackCFProxyDomainsURL:    strings.TrimSpace(*cfproxyDomainsURL),
		FallbackCFProxyDomains:       nil,
		FallbackCFProxyActive:        "",
		FallbackCFProxyPerDCActive:   make(map[int]string),
		BufKB:                        maxInt(*bufKB, 4),
		PoolSize:                     maxInt(*poolSize, 0),
		MaxConns:                     maxInt(*maxConns, 1),
		Credits:                      *credits,
	}

	cfg.setCFProxyDomains(domainPool)
	return cfg, nil
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseIPCSV(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		ip := strings.TrimSpace(p)
		if ip == "" {
			continue
		}
		if net.ParseIP(ip) == nil {
			return nil, fmt.Errorf("invalid ip: %s", ip)
		}
		out = appendUniqueIP(out, ip)
	}
	if len(out) == 0 {
		return nil, errors.New("empty ip pool")
	}
	return out, nil
}

func appendUniqueIP(dst []string, ip string) []string {
	for _, v := range dst {
		if v == ip {
			return dst
		}
	}
	return append(dst, ip)
}
