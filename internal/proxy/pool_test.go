package proxy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewRoundRobinPool_Empty(t *testing.T) {
	_, err := NewRoundRobinPool(nil, 0)
	if err == nil {
		t.Fatal("expected error for empty proxy list")
	}
}

func TestNewSinglePool(t *testing.T) {
	pool := NewSinglePool("http://proxy:8080")
	url, err := pool.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if url != "http://proxy:8080" {
		t.Fatalf("url = %q, want http://proxy:8080", url)
	}
}

func TestRoundRobinRotation(t *testing.T) {
	proxies := []string{"http://a:1", "http://b:2", "http://c:3"}
	pool, err := NewRoundRobinPool(proxies, 0)
	if err != nil {
		t.Fatalf("NewRoundRobinPool() error = %v", err)
	}

	seen := make(map[string]int)
	for i := 0; i < 9; i++ {
		url, err := pool.Next(context.Background())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		seen[url]++
	}

	for _, p := range proxies {
		if seen[p] != 3 {
			t.Errorf("proxy %q seen %d times, want 3", p, seen[p])
		}
	}
}

func TestCooldownSkipsProxy(t *testing.T) {
	proxies := []string{"http://a:1", "http://b:2"}
	pool, err := NewRoundRobinPool(proxies, 5*time.Second)
	if err != nil {
		t.Fatalf("NewRoundRobinPool() error = %v", err)
	}

	// Report failure on proxy a → enters cooldown
	pool.Report("http://a:1", false)

	// Next should skip a and return b
	for i := 0; i < 5; i++ {
		url, err := pool.Next(context.Background())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if url == "http://a:1" {
			t.Fatal("expected proxy a to be in cooldown")
		}
		if url != "http://b:2" {
			t.Fatalf("url = %q, want http://b:2", url)
		}
	}
}

func TestAllCoolingWaitsAndRecovers(t *testing.T) {
	proxies := []string{"http://a:1"}
	cooldown := 100 * time.Millisecond
	pool, err := NewRoundRobinPool(proxies, cooldown)
	if err != nil {
		t.Fatalf("NewRoundRobinPool() error = %v", err)
	}

	pool.Report("http://a:1", false)

	start := time.Now()
	url, err := pool.Next(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if url != "http://a:1" {
		t.Fatalf("url = %q", url)
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("expected wait ~%v, got %v", cooldown, elapsed)
	}
}

func TestContextCancellationDuringCooldown(t *testing.T) {
	proxies := []string{"http://a:1"}
	pool, err := NewRoundRobinPool(proxies, 10*time.Second)
	if err != nil {
		t.Fatalf("NewRoundRobinPool() error = %v", err)
	}

	pool.Report("http://a:1", false)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = pool.Next(ctx)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestReportSuccess(t *testing.T) {
	pool := NewSinglePool("http://a:1")
	pool.Report("http://a:1", true)
	pool.Report("http://a:1", true)
	pool.Report("http://a:1", false)

	stats := pool.Stats()
	s := stats["http://a:1"]
	if s.Success != 2 {
		t.Errorf("success = %d, want 2", s.Success)
	}
	if s.Failures != 1 {
		t.Errorf("failures = %d, want 1", s.Failures)
	}
}

func TestConcurrentAccess(t *testing.T) {
	proxies := []string{"http://a:1", "http://b:2", "http://c:3"}
	pool, err := NewRoundRobinPool(proxies, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("NewRoundRobinPool() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			url, err := pool.Next(context.Background())
			if err != nil {
				t.Errorf("Next() error = %v", err)
				return
			}
			pool.Report(url, true)
		}()
	}
	wg.Wait()

	stats := pool.Stats()
	total := int64(0)
	for _, s := range stats {
		total += s.Success
	}
	if total != 20 {
		t.Errorf("total success = %d, want 20", total)
	}
}

func TestLoadProxies(t *testing.T) {
	content := strings.Join([]string{
		"# This is a comment",
		"http://proxy1:8080",
		"",
		"  http://proxy2:8080  ",
		"# Another comment",
		"http://proxy3:8080",
	}, "\n")

	dir := t.TempDir()
	path := filepath.Join(dir, "proxies.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	proxies, err := LoadProxies(path)
	if err != nil {
		t.Fatalf("LoadProxies() error = %v", err)
	}
	if len(proxies) != 3 {
		t.Fatalf("len(proxies) = %d, want 3", len(proxies))
	}
	if proxies[0] != "http://proxy1:8080" {
		t.Errorf("proxies[0] = %q", proxies[0])
	}
	if proxies[1] != "http://proxy2:8080" {
		t.Errorf("proxies[1] = %q", proxies[1])
	}
}

func TestLoadProxies_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte("# only comments\n\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := LoadProxies(path)
	if err == nil {
		t.Fatal("expected error for empty proxy file")
	}
}

func TestLoadProxies_MissingFile(t *testing.T) {
	_, err := LoadProxies("/tmp/nonexistent-proxy-file-12345.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
