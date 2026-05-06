package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("missing config file uses defaults", func(t *testing.T) {
		proxyBefore := os.Getenv("PROXY")
		t.Cleanup(func() {
			_ = os.Setenv("PROXY", proxyBefore)
		})
		_ = os.Unsetenv("PROXY")

		cfg, err := Load(filepath.Join(t.TempDir(), "missing.json"))
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Proxy != DefaultProxy {
			t.Fatalf("Proxy = %q, want %q", cfg.Proxy, DefaultProxy)
		}
		if cfg.OutputFile != DefaultOutputFile {
			t.Fatalf("OutputFile = %q, want %q", cfg.OutputFile, DefaultOutputFile)
		}
		if cfg.DefaultPassword != DefaultPassword {
			t.Fatalf("DefaultPassword = %q, want %q", cfg.DefaultPassword, DefaultPassword)
		}
		if cfg.DefaultDomain != DefaultDomainValue {
			t.Fatalf("DefaultDomain = %q, want %q", cfg.DefaultDomain, DefaultDomainValue)
		}
	})

	t.Run("json values are loaded", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.json")
		content := []byte(`{"proxy":"http://localhost:8080","output_file":"out.txt","default_password":"longpassword12","default_domain":"example.com"}`)
		if err := os.WriteFile(configPath, content, 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Proxy != "http://localhost:8080" {
			t.Fatalf("Proxy = %q, want %q", cfg.Proxy, "http://localhost:8080")
		}
		if cfg.OutputFile != "out.txt" {
			t.Fatalf("OutputFile = %q, want %q", cfg.OutputFile, "out.txt")
		}
		if cfg.DefaultPassword != "longpassword12" {
			t.Fatalf("DefaultPassword = %q, want %q", cfg.DefaultPassword, "longpassword12")
		}
		if cfg.DefaultDomain != "example.com" {
			t.Fatalf("DefaultDomain = %q, want %q", cfg.DefaultDomain, "example.com")
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(configPath, []byte("{"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if _, err := Load(configPath); err == nil {
			t.Fatal("Load() error = nil, want non-nil")
		}
	})

	t.Run("short default password returns error", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.json")
		content := []byte(`{"default_password":"short"}`)
		if err := os.WriteFile(configPath, content, 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if _, err := Load(configPath); err == nil {
			t.Fatal("Load() error = nil, want non-nil")
		}
	})

	t.Run("PROXY env overrides config proxy", func(t *testing.T) {
		proxyBefore := os.Getenv("PROXY")
		t.Cleanup(func() {
			_ = os.Setenv("PROXY", proxyBefore)
		})

		configPath := filepath.Join(t.TempDir(), "config.json")
		content := []byte(`{"proxy":"http://from-config:8080","default_password":"longpassword12"}`)
		if err := os.WriteFile(configPath, content, 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if err := os.Setenv("PROXY", "http://from-env:8080"); err != nil {
			t.Fatalf("Setenv() error = %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Proxy != "http://from-env:8080" {
			t.Fatalf("Proxy = %q, want %q", cfg.Proxy, "http://from-env:8080")
		}
	})
}
