package web

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
)

// SSEBroker manages a set of SSE client channels and broadcasts messages to all of them.
type SSEBroker struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

// NewSSEBroker creates an empty broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{clients: make(map[chan string]struct{})}
}

// Subscribe returns a receive channel and an unsubscribe function.
// Callers must call unsubscribe when they are done.
func (b *SSEBroker) Subscribe() (<-chan string, func()) {
	ch := make(chan string, 128)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
		close(ch)
	}
}

// Send broadcasts a message to all subscribed clients.
// Slow clients are silently dropped rather than blocking the sender.
func (b *SSEBroker) Send(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

// LineWriter returns an io.Writer that broadcasts each complete line as an SSE log event.
func (b *SSEBroker) LineWriter() io.Writer {
	return &brokerWriter{broker: b}
}

type brokerWriter struct {
	broker *SSEBroker
	mu     sync.Mutex
	buf    strings.Builder
}

func (w *brokerWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	for {
		s := w.buf.String()
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			break
		}
		line := s[:idx]
		w.buf.Reset()
		w.buf.WriteString(s[idx+1:])
		if line != "" {
			w.broker.Send(encodeLogEvent(line))
		}
	}
	return len(p), nil
}

func encodeLogEvent(line string) string {
	data, _ := json.Marshal(map[string]string{"type": "log", "line": line})
	return string(data)
}
