package camofox

import "testing"

func TestParseProxy(t *testing.T) {
	proxy, err := parseProxy("http://user:pass@example.com:8080")
	if err != nil {
		t.Fatalf("parseProxy() error = %v", err)
	}
	if proxy["host"] != "example.com" || proxy["port"] != 8080 {
		t.Fatalf("proxy host/port = %#v", proxy)
	}
	if proxy["username"] != "user" || proxy["password"] != "pass" {
		t.Fatalf("proxy auth = %#v", proxy)
	}
}

func TestParseProxy_Invalid(t *testing.T) {
	if _, err := parseProxy("http://example.com"); err == nil {
		t.Fatal("parseProxy() error = nil, want non-nil")
	}
}
