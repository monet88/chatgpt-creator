package email

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// IMAPConfig holds IMAP connection settings.
type IMAPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	UseTLS   bool
}

// IMAPPooler provides OTP retrieval from an IMAP catch-all mailbox.
// It maintains a single persistent connection serialized with a mutex.
type IMAPPooler struct {
	config       IMAPConfig
	conn         net.Conn
	reader       *bufio.Reader
	mu           sync.Mutex
	tag          int
	pollInterval time.Duration
}

// NewIMAPPooler creates and connects an IMAP pooler.
func NewIMAPPooler(cfg IMAPConfig) (*IMAPPooler, error) {
	p := &IMAPPooler{
		config:       cfg,
		pollInterval: 3 * time.Second,
	}
	if err := p.connect(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *IMAPPooler) connect() error {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	var conn net.Conn
	var err error
	if p.config.UseTLS {
		conn, err = tls.Dial("tcp", addr, &tls.Config{ServerName: p.config.Host})
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("imap: failed to connect to %s: %w", addr, err)
	}

	p.conn = conn
	p.reader = bufio.NewReader(conn)
	p.tag = 0

	// Read greeting
	if _, err := p.readLine(); err != nil {
		conn.Close()
		return fmt.Errorf("imap: failed to read greeting: %w", err)
	}

	// LOGIN
	resp, err := p.command(fmt.Sprintf("LOGIN %s %s", p.config.User, p.config.Password))
	if err != nil {
		conn.Close()
		return fmt.Errorf("imap: login failed: %w", err)
	}
	if !strings.Contains(resp, "OK") {
		conn.Close()
		return fmt.Errorf("imap: login rejected: %s", resp)
	}

	return nil
}

func (p *IMAPPooler) command(cmd string) (string, error) {
	p.tag++
	tagStr := fmt.Sprintf("A%03d", p.tag)
	line := fmt.Sprintf("%s %s\r\n", tagStr, cmd)

	if _, err := io.WriteString(p.conn, line); err != nil {
		return "", fmt.Errorf("imap: write error: %w", err)
	}

	var result strings.Builder
	for {
		resp, err := p.readLine()
		if err != nil {
			return "", err
		}
		result.WriteString(resp)
		result.WriteString("\n")
		if strings.HasPrefix(resp, tagStr+" ") {
			return result.String(), nil
		}
	}
}

func (p *IMAPPooler) readLine() (string, error) {
	p.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("imap: read error: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// GetOTP retrieves an OTP code from the IMAP mailbox for the specified email address.
func (p *IMAPPooler) GetOTP(ctx context.Context, emailAddr string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("imap: OTP timeout after %v for %s", timeout, emailAddr)
		}

		code, err := p.searchOTP(emailAddr)
		if err != nil {
			// Try reconnect on connection errors
			if reconnErr := p.reconnect(); reconnErr != nil {
				return "", fmt.Errorf("imap: reconnect failed: %w (original: %v)", reconnErr, err)
			}
			// Retry after reconnect
			code, err = p.searchOTP(emailAddr)
			if err != nil {
				return "", err
			}
		}

		if code != "" {
			return code, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(p.pollInterval):
		}
	}
}

func (p *IMAPPooler) searchOTP(emailAddr string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// SELECT INBOX
	resp, err := p.command("SELECT INBOX")
	if err != nil {
		return "", err
	}
	if !strings.Contains(resp, "OK") {
		return "", fmt.Errorf("imap: SELECT INBOX failed: %s", resp)
	}

	// SEARCH for unseen messages TO the target email
	searchResp, err := p.command(fmt.Sprintf(`SEARCH UNSEEN TO "%s"`, emailAddr))
	if err != nil {
		return "", err
	}

	// Parse SEARCH response for message IDs
	msgIDs := parseSearchResponse(searchResp)
	if len(msgIDs) == 0 {
		return "", nil // No matching messages yet
	}

	// FETCH the latest matching message body
	lastID := msgIDs[len(msgIDs)-1]
	fetchResp, err := p.command(fmt.Sprintf("FETCH %s BODY[TEXT]", lastID))
	if err != nil {
		return "", err
	}

	// Extract 6-digit OTP
	matches := otpRegex.FindStringSubmatch(fetchResp)
	if len(matches) < 1 {
		return "", nil // No OTP found in body
	}

	// Mark as read to avoid reprocessing
	_, _ = p.command(fmt.Sprintf(`STORE %s +FLAGS (\Seen)`, lastID))

	return matches[0], nil
}

func (p *IMAPPooler) reconnect() error {
	if p.conn != nil {
		p.conn.Close()
	}
	return p.connect()
}

// Close logs out and closes the IMAP connection.
func (p *IMAPPooler) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		_, _ = p.command("LOGOUT")
		return p.conn.Close()
	}
	return nil
}

// parseSearchResponse extracts message IDs from an IMAP SEARCH response.
func parseSearchResponse(resp string) []string {
	var ids []string
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "* SEARCH") {
			parts := strings.Fields(line)
			if len(parts) > 2 {
				ids = append(ids, parts[2:]...)
			}
		}
	}
	return ids
}
