package phone

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(handler http.HandlerFunc) (*ViOTPClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := NewViOTPClient("test-token")
	client.baseURL = server.URL
	client.pollInterval = 10 * time.Millisecond
	return client, server
}

func jsonResponse(statusCode int, message string, data any) []byte {
	dataBytes, _ := json.Marshal(data)
	resp := viOTPResponse{
		StatusCode: statusCode,
		Success:    statusCode == 200,
		Message:    message,
		Data:       dataBytes,
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestCheckBalance_Success(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/balance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("token") != "test-token" {
			t.Fatalf("unexpected token: %s", r.URL.Query().Get("token"))
		}
		w.Write(jsonResponse(200, "successful", balanceData{Balance: 50000}))
	})
	defer server.Close()

	balance, err := client.CheckBalance(context.Background())
	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}
	if balance != 50000 {
		t.Fatalf("balance = %d, want 50000", balance)
	}
}

func TestCheckBalance_AuthError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(401, "unauthorized", nil))
	})
	defer server.Close()

	_, err := client.CheckBalance(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestRentNumber_Success(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/request/getv2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("serviceId") != "42" {
			t.Fatalf("unexpected serviceId: %s", r.URL.Query().Get("serviceId"))
		}
		w.Write(jsonResponse(200, "success", rentData{
			PhoneNumber:   "987654321",
			RePhoneNumber: "84987654321",
			CountryISO:    "VN",
			CountryCode:   "84",
			Balance:       49200,
			RequestID:     122314,
		}))
	})
	defer server.Close()

	result, err := client.RentNumber(context.Background(), 42)
	if err != nil {
		t.Fatalf("RentNumber() error = %v", err)
	}
	if result.PhoneNumber != "987654321" {
		t.Fatalf("PhoneNumber = %q", result.PhoneNumber)
	}
	if result.RequestID != "122314" {
		t.Fatalf("RequestID = %q", result.RequestID)
	}
	if result.Balance != 49200 {
		t.Fatalf("Balance = %d", result.Balance)
	}
}

func TestRentNumber_InsufficientBalance(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(-2, "insufficient balance", nil))
	})
	defer server.Close()

	_, err := client.RentNumber(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}

func TestRentNumber_OutOfStock(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(-3, "out of stock", nil))
	})
	defer server.Close()

	_, err := client.RentNumber(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for out of stock")
	}
}

func TestWaitForOTP_Success(t *testing.T) {
	callCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			// First 2 calls: waiting
			w.Write(jsonResponse(200, "successful", sessionData{Status: 0, Code: ""}))
		} else {
			// 3rd call: OTP received
			w.Write(jsonResponse(200, "successful", sessionData{Status: 1, Code: "486460"}))
		}
	})
	defer server.Close()

	code, err := client.WaitForOTP(context.Background(), "122314", 5*time.Second)
	if err != nil {
		t.Fatalf("WaitForOTP() error = %v", err)
	}
	if code != "486460" {
		t.Fatalf("code = %q, want 486460", code)
	}
	if callCount != 3 {
		t.Fatalf("callCount = %d, want 3", callCount)
	}
}

func TestWaitForOTP_Expired(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(200, "successful", sessionData{Status: 2, Code: ""}))
	})
	defer server.Close()

	_, err := client.WaitForOTP(context.Background(), "122314", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for expired session")
	}
}

func TestWaitForOTP_Timeout(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(200, "successful", sessionData{Status: 0, Code: ""}))
	})
	defer server.Close()

	_, err := client.WaitForOTP(context.Background(), "122314", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWaitForOTP_ContextCancellation(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(200, "successful", sessionData{Status: 0, Code: ""}))
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := client.WaitForOTP(ctx, "122314", 10*time.Second)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
