package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	Proxy               string `json:"proxy"`
	ProxyList           string `json:"proxy_list"`
	ProxyCooldown       int    `json:"proxy_cooldown"`
	OutputFile          string `json:"output_file"`
	DefaultPassword     string `json:"default_password"`
	DefaultDomain       string `json:"default_domain"`
	Pacing              string `json:"pacing"`
	ViOTPToken          string `json:"viotp_token"`
	ViOTPServiceID      int    `json:"viotp_service_id"`
	IMAPHost            string `json:"imap_host"`
	IMAPPort            int    `json:"imap_port"`
	IMAPUser            string `json:"imap_user"`
	IMAPPassword        string `json:"imap_password"`
	IMAPUseTLS          bool   `json:"imap_use_tls"`
	CodexEnabled        bool   `json:"codex_enabled"`
	CodexOutput         string `json:"codex_output"`
	CloudflareMailURL   string `json:"cloudflare_mail_url"`
	CloudflareMailToken string `json:"cloudflare_mail_token"`
}

const (
	DefaultProxy          = ""
	DefaultOutputFile     = "results.txt"
	DefaultConfigFilename = "config.json"
	DefaultPassword       = "" // Min 12 characters
	DefaultDomainValue    = ""
	DefaultCodexOutput    = "codex-tokens.json"
)

// DefaultConfigPath returns the default path to the config file.
func DefaultConfigPath() string {
	return DefaultConfigFilename
}

// Load reads the config from a JSON file and applies environment variable overrides.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Proxy:           DefaultProxy,
		OutputFile:      DefaultOutputFile,
		DefaultPassword: DefaultPassword,
		DefaultDomain:   DefaultDomainValue,
		IMAPPort:        993,
		IMAPUseTLS:      true,
		CodexOutput:     DefaultCodexOutput,
	}

	// Try to read the file
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	// Validate password length
	if cfg.DefaultPassword != "" && len(cfg.DefaultPassword) < 12 {
		return nil, fmt.Errorf("default_password must be at least 12 characters (got %d)", len(cfg.DefaultPassword))
	}

	// Environment variable overrides
	if proxy := os.Getenv("PROXY"); proxy != "" {
		cfg.Proxy = proxy
	}
	if proxyList := os.Getenv("PROXY_LIST"); proxyList != "" {
		cfg.ProxyList = proxyList
	}
	if pacing := os.Getenv("PACING"); pacing != "" {
		cfg.Pacing = pacing
	}
	if viOTPToken := os.Getenv("VIOTP_TOKEN"); viOTPToken != "" {
		cfg.ViOTPToken = viOTPToken
	}
	if v := os.Getenv("VIOTP_SERVICE_ID"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			cfg.ViOTPServiceID = id
		}
	}
	if v := os.Getenv("IMAP_HOST"); v != "" {
		cfg.IMAPHost = v
	}
	if v := os.Getenv("IMAP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.IMAPPort = port
		}
	}
	if v := os.Getenv("IMAP_USER"); v != "" {
		cfg.IMAPUser = v
	}
	if v := os.Getenv("IMAP_PASSWORD"); v != "" {
		cfg.IMAPPassword = v
	}
	if v := os.Getenv("IMAP_TLS"); v != "" {
		cfg.IMAPUseTLS = v != "false" && v != "0"
	}
	if v := os.Getenv("PROXY_COOLDOWN"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil {
			cfg.ProxyCooldown = sec
		}
	}
	if v := os.Getenv("CODEX_ENABLED"); v != "" {
		cfg.CodexEnabled = v == "true" || v == "1"
	}
	if v := os.Getenv("CODEX_OUTPUT"); v != "" {
		cfg.CodexOutput = v
	}
	if v := os.Getenv("CLOUDFLARE_MAIL_URL"); v != "" {
		cfg.CloudflareMailURL = v
	}
	if v := os.Getenv("CLOUDFLARE_MAIL_TOKEN"); v != "" {
		cfg.CloudflareMailToken = v
	}

	return cfg, nil
}
