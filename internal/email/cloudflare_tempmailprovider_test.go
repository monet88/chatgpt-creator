package email

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCloudflareTempEmail_ReturnsEscapedMailboxURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want %s", r.Method, http.MethodPost)
			return
		}
		if r.URL.Path != "/api/v1/email/generate" {
			t.Errorf("path = %s, want /api/v1/email/generate", r.URL.Path)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"mary.baker.19e3@monet.uno","user":"mary.baker.19e3","domain":"monet.uno"}}`))
	}))
	defer server.Close()

	createEmail := CreateCloudflareTempEmail(server.URL)
	emailAddr, mailboxURL, err := createEmail("monet.uno")
	if err != nil {
		t.Fatalf("CreateCloudflareTempEmail() error = %v", err)
	}
	if emailAddr != "mary.baker.19e3@monet.uno" {
		t.Fatalf("emailAddr = %q, want mary.baker.19e3@monet.uno", emailAddr)
	}

	wantMailboxURL := server.URL + "/#mary.baker.19e3%40monet.uno"
	if mailboxURL != wantMailboxURL {
		t.Fatalf("mailboxURL = %q, want %q", mailboxURL, wantMailboxURL)
	}
}

func TestCreateCloudflareTempEmail_ReturnsEmptyValuesOnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusInternalServerError)
	}))
	defer server.Close()

	createEmail := CreateCloudflareTempEmail(server.URL)
	emailAddr, mailboxURL, err := createEmail("monet.uno")
	if err == nil {
		t.Fatal("CreateCloudflareTempEmail() error = nil, want error")
	}
	if emailAddr != "" {
		t.Fatalf("emailAddr = %q, want empty", emailAddr)
	}
	if mailboxURL != "" {
		t.Fatalf("mailboxURL = %q, want empty", mailboxURL)
	}
}
