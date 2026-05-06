package email

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeIMAPMessage struct {
	ID   string
	To   string
	Body string
	Seen bool
}

type fakeIMAPServer struct {
	listener          net.Listener
	acceptWG          sync.WaitGroup
	connWG            sync.WaitGroup
	mu                sync.Mutex
	messages          []fakeIMAPMessage
	dropFirstSearch   bool
	droppedSearchConn bool
}

func newFakeIMAPServer(t *testing.T, messages []fakeIMAPMessage, dropFirstSearch bool) *fakeIMAPServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	copied := make([]fakeIMAPMessage, len(messages))
	copy(copied, messages)

	s := &fakeIMAPServer{
		listener:        listener,
		messages:        copied,
		dropFirstSearch: dropFirstSearch,
	}

	s.acceptWG.Add(1)
	go s.acceptLoop()

	return s
}

func (s *fakeIMAPServer) addr() string {
	return s.listener.Addr().String()
}

func (s *fakeIMAPServer) close() error {
	err := s.listener.Close()
	s.acceptWG.Wait()
	s.connWG.Wait()
	return err
}

func (s *fakeIMAPServer) acceptLoop() {
	defer s.acceptWG.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			return
		}
		s.connWG.Add(1)
		go s.handleConn(conn)
	}
}

func (s *fakeIMAPServer) handleConn(conn net.Conn) {
	defer s.connWG.Done()
	defer conn.Close()

	_, _ = fmt.Fprint(conn, "* OK fake-imap ready\r\n")
	r := bufio.NewReader(conn)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		tag := parts[0]
		cmd := strings.ToUpper(parts[1])

		switch cmd {
		case "LOGIN":
			_, _ = fmt.Fprintf(conn, "%s OK LOGIN completed\r\n", tag)
		case "SELECT":
			_, _ = fmt.Fprint(conn, "* 3 EXISTS\r\n")
			_, _ = fmt.Fprintf(conn, "%s OK [READ-WRITE] SELECT completed\r\n", tag)
		case "SEARCH":
			if s.shouldDropSearchConn() {
				return
			}
			target := s.extractSearchTo(line)
			ids := s.findUnseenByRecipient(target)
			if len(ids) == 0 {
				_, _ = fmt.Fprint(conn, "* SEARCH\r\n")
			} else {
				_, _ = fmt.Fprintf(conn, "* SEARCH %s\r\n", strings.Join(ids, " "))
			}
			_, _ = fmt.Fprintf(conn, "%s OK SEARCH completed\r\n", tag)
		case "FETCH":
			if len(parts) < 3 {
				_, _ = fmt.Fprintf(conn, "%s BAD FETCH missing id\r\n", tag)
				continue
			}
			id := parts[2]
			body := s.bodyForID(id)
			_, _ = fmt.Fprintf(conn, "* %s FETCH (BODY[TEXT] {%d}\r\n%s\r\n)\r\n", id, len(body), body)
			_, _ = fmt.Fprintf(conn, "%s OK FETCH completed\r\n", tag)
		case "STORE":
			if len(parts) >= 3 {
				s.markSeen(parts[2])
			}
			_, _ = fmt.Fprintf(conn, "%s OK STORE completed\r\n", tag)
		case "LOGOUT":
			_, _ = fmt.Fprint(conn, "* BYE logging out\r\n")
			_, _ = fmt.Fprintf(conn, "%s OK LOGOUT completed\r\n", tag)
			return
		default:
			_, _ = fmt.Fprintf(conn, "%s BAD unknown command\r\n", tag)
		}
	}
}

func (s *fakeIMAPServer) shouldDropSearchConn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.dropFirstSearch || s.droppedSearchConn {
		return false
	}
	s.droppedSearchConn = true
	return true
}

func (s *fakeIMAPServer) extractSearchTo(line string) string {
	idx := strings.Index(line, " TO ")
	if idx == -1 {
		return ""
	}
	toPart := strings.TrimSpace(line[idx+4:])
	return strings.Trim(toPart, `"`)
}

func (s *fakeIMAPServer) findUnseenByRecipient(emailAddr string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0)
	for _, msg := range s.messages {
		if msg.Seen {
			continue
		}
		if strings.EqualFold(msg.To, emailAddr) {
			ids = append(ids, msg.ID)
		}
	}
	slices.Sort(ids)
	return ids
}

func (s *fakeIMAPServer) bodyForID(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, msg := range s.messages {
		if msg.ID == id {
			return msg.Body
		}
	}
	return ""
}

func (s *fakeIMAPServer) markSeen(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.messages {
		if s.messages[i].ID == id {
			s.messages[i].Seen = true
			return
		}
	}
}

func newIMAPPoolerForTest(t *testing.T, serverAddr string) *IMAPPooler {
	t.Helper()
	host, portStr, err := net.SplitHostPort(serverAddr)
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}
	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		t.Fatalf("LookupPort() error = %v", err)
	}

	pooler, err := NewIMAPPooler(IMAPConfig{
		Host:     host,
		Port:     port,
		User:     "user",
		Password: "pass",
		UseTLS:   false,
	})
	if err != nil {
		t.Fatalf("NewIMAPPooler() error = %v", err)
	}
	pooler.pollInterval = 10 * time.Millisecond
	return pooler
}

func TestIMAPPooler_GetOTP_FiltersTargetRecipient(t *testing.T) {
	server := newFakeIMAPServer(t, []fakeIMAPMessage{
		{ID: "1", To: "wrong@example.com", Body: "Your verification code is 111111"},
		{ID: "2", To: "target@example.com", Body: "Your verification code is 654321"},
	}, false)
	defer server.close()

	pooler := newIMAPPoolerForTest(t, server.addr())
	defer pooler.Close()

	otp, err := pooler.GetOTP(context.Background(), "target@example.com", 300*time.Millisecond)
	if err != nil {
		t.Fatalf("GetOTP() error = %v", err)
	}
	if otp != "654321" {
		t.Fatalf("otp = %q, want %q", otp, "654321")
	}
}

func TestIMAPPooler_GetOTP_ReconnectAfterDroppedConnection(t *testing.T) {
	server := newFakeIMAPServer(t, []fakeIMAPMessage{
		{ID: "3", To: "target@example.com", Body: "Your verification code is 222333"},
	}, true)
	defer server.close()

	pooler := newIMAPPoolerForTest(t, server.addr())
	defer pooler.Close()

	otp, err := pooler.GetOTP(context.Background(), "target@example.com", 600*time.Millisecond)
	if err != nil {
		t.Fatalf("GetOTP() error = %v", err)
	}
	if otp != "222333" {
		t.Fatalf("otp = %q, want %q", otp, "222333")
	}
}

func TestIMAPPooler_GetOTP_ConcurrentCalls(t *testing.T) {
	server := newFakeIMAPServer(t, []fakeIMAPMessage{
		{ID: "10", To: "a@example.com", Body: "Your verification code is 100100"},
		{ID: "11", To: "b@example.com", Body: "Your verification code is 200200"},
		{ID: "12", To: "c@example.com", Body: "Your verification code is 300300"},
	}, false)
	defer server.close()

	pooler := newIMAPPoolerForTest(t, server.addr())
	defer pooler.Close()

	type result struct {
		email string
		otp   string
		err   error
	}
	results := make(chan result, 3)
	emails := []string{"a@example.com", "b@example.com", "c@example.com"}

	var wg sync.WaitGroup
	for _, emailAddr := range emails {
		emailAddr := emailAddr
		wg.Add(1)
		go func() {
			defer wg.Done()
			otp, err := pooler.GetOTP(context.Background(), emailAddr, 500*time.Millisecond)
			results <- result{email: emailAddr, otp: otp, err: err}
		}()
	}
	wg.Wait()
	close(results)

	expected := map[string]string{
		"a@example.com": "100100",
		"b@example.com": "200200",
		"c@example.com": "300300",
	}
	for res := range results {
		if res.err != nil {
			t.Fatalf("GetOTP(%s) error = %v", res.email, res.err)
		}
		if res.otp != expected[res.email] {
			t.Fatalf("otp for %s = %q, want %q", res.email, res.otp, expected[res.email])
		}
	}
}

func TestIMAPPooler_GetOTP_ContextCancellation(t *testing.T) {
	server := newFakeIMAPServer(t, nil, false)
	defer server.close()

	pooler := newIMAPPoolerForTest(t, server.addr())
	defer pooler.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := pooler.GetOTP(ctx, "target@example.com", 3*time.Second)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context deadline exceeded", err)
	}
	if time.Since(start) > 400*time.Millisecond {
		t.Fatalf("GetOTP returned too late: %v", time.Since(start))
	}
}

func TestParseSearchResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single message",
			input:    "* SEARCH 42\nA001 OK SEARCH completed",
			expected: []string{"42"},
		},
		{
			name:     "multiple messages",
			input:    "* SEARCH 1 5 10 23\nA001 OK SEARCH completed",
			expected: []string{"1", "5", "10", "23"},
		},
		{
			name:     "no matches",
			input:    "* SEARCH\nA001 OK SEARCH completed",
			expected: nil,
		},
		{
			name:     "empty response",
			input:    "A001 OK SEARCH completed",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseSearchResponse(tc.input)
			if len(result) != len(tc.expected) {
				t.Fatalf("len = %d, want %d", len(result), len(tc.expected))
			}
			for i, id := range result {
				if id != tc.expected[i] {
					t.Errorf("id[%d] = %q, want %q", i, id, tc.expected[i])
				}
			}
		})
	}
}

func TestOTPRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard OTP",
			input:    "Your verification code is 486460. Do not share this code.",
			expected: "486460",
		},
		{
			name:     "OTP at start",
			input:    "123456 is your verification code",
			expected: "123456",
		},
		{
			name:     "no OTP",
			input:    "Hello, your account is ready",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := otpRegex.FindStringSubmatch(tc.input)
			if tc.expected == "" {
				if len(matches) >= 1 {
					t.Fatalf("expected no match, got %q", matches[0])
				}
				return
			}
			if len(matches) < 1 {
				t.Fatal("expected OTP match, got none")
			}
			if matches[0] != tc.expected {
				t.Fatalf("OTP = %q, want %q", matches[0], tc.expected)
			}
		})
	}
}

func TestGeneratorEmailProvider_Close(t *testing.T) {
	p := &GeneratorEmailProvider{}
	if err := p.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
