package proxy

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Pool provides proxy URL rotation with health tracking.
type Pool interface {
	// Next returns the next healthy proxy URL. Blocks if all proxies are cooling down.
	Next(ctx context.Context) (string, error)
	// Report records the outcome of using a proxy.
	Report(proxyURL string, success bool)
	// Stats returns a snapshot of per-proxy health statistics.
	Stats() map[string]ProxyStats
}

// ProxyStats holds health statistics for a single proxy.
type ProxyStats struct {
	URL       string    `json:"url"`
	Success   int64     `json:"success"`
	Failures  int64     `json:"failures"`
	CoolUntil time.Time `json:"cool_until,omitempty"`
	Health    float64   `json:"health"`
}

type proxyEntry struct {
	url              string
	success          atomic.Int64
	failures         atomic.Int64
	consecutiveFails atomic.Int64
	mu               sync.RWMutex
	coolUtil         time.Time
}

func (e *proxyEntry) isCooling(now time.Time) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return now.Before(e.coolUtil)
}

func (e *proxyEntry) setCooldown(until time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.coolUtil = until
}

func (e *proxyEntry) coolUntil() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.coolUtil
}

func (e *proxyEntry) stats() ProxyStats {
	s := e.success.Load()
	f := e.failures.Load()
	health := float64(s) / float64(s+f+1)
	return ProxyStats{
		URL:       e.url,
		Success:   s,
		Failures:  f,
		CoolUntil: e.coolUntil(),
		Health:    health,
	}
}

// RoundRobinPool rotates proxies in order, skipping those in cooldown.
type RoundRobinPool struct {
	entries  []*proxyEntry
	index    atomic.Int64
	cooldown time.Duration
}

// NewRoundRobinPool creates a pool from a list of proxy URLs.
func NewRoundRobinPool(proxies []string, cooldown time.Duration) (*RoundRobinPool, error) {
	if len(proxies) == 0 {
		return nil, fmt.Errorf("proxy pool: no proxies provided")
	}
	entries := make([]*proxyEntry, len(proxies))
	for i, p := range proxies {
		entries[i] = &proxyEntry{url: p}
	}
	return &RoundRobinPool{entries: entries, cooldown: cooldown}, nil
}

// NewSinglePool wraps a single proxy URL into a pool for backward compatibility.
func NewSinglePool(proxy string) *RoundRobinPool {
	pool, _ := NewRoundRobinPool([]string{proxy}, 0)
	return pool
}

// Next returns the next healthy proxy. If all are cooling, waits for the earliest recovery.
func (p *RoundRobinPool) Next(ctx context.Context) (string, error) {
	n := int64(len(p.entries))
	for {
		now := time.Now()
		var earliestCool time.Time
		// Try all proxies starting from current index
		for i := int64(0); i < n; i++ {
			idx := p.index.Add(1) % n
			entry := p.entries[idx]
			if !entry.isCooling(now) {
				return entry.url, nil
			}
			cool := entry.coolUntil()
			if earliestCool.IsZero() || cool.Before(earliestCool) {
				earliestCool = cool
			}
		}

		// All proxies cooling — wait for earliest recovery
		waitDur := time.Until(earliestCool)
		if waitDur <= 0 {
			continue // Already recovered, retry
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(waitDur):
			// Retry after cooldown
		}
	}
}

// Report records the outcome of a proxy request.
func (p *RoundRobinPool) Report(proxyURL string, success bool) {
	for _, e := range p.entries {
		if e.url == proxyURL {
			if success {
				e.success.Add(1)
				e.consecutiveFails.Store(0)
			} else {
				e.failures.Add(1)
				cf := e.consecutiveFails.Add(1)
				// Exponential backoff: base * 2^(cf-1), capped at 10 * base
				factor := int64(1)
				if cf > 1 {
					factor = int64(1) << (cf - 1)
				}
				const maxFactor = 10
				if factor > maxFactor {
					factor = maxFactor
				}
				e.setCooldown(time.Now().Add(p.cooldown * time.Duration(factor)))
			}
			return
		}
	}
}

// Stats returns a snapshot of all proxy health statistics.
func (p *RoundRobinPool) Stats() map[string]ProxyStats {
	result := make(map[string]ProxyStats, len(p.entries))
	for _, e := range p.entries {
		result[e.url] = e.stats()
	}
	return result
}

// LoadProxies reads proxy URLs from a file (one per line, # comments, blank lines ignored).
func LoadProxies(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("proxy pool: failed to open %s: %w", path, err)
	}
	defer f.Close()

	var proxies []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		proxies = append(proxies, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("proxy pool: error reading %s: %w", path, err)
	}
	if len(proxies) == 0 {
		return nil, fmt.Errorf("proxy pool: no proxies found in %s", path)
	}
	return proxies, nil
}
