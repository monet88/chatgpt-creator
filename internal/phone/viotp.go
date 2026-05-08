package phone

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.viotp.com"
const defaultPollInterval = 5 * time.Second

// ViOTPClient implements PhoneProvider using the ViOTP API.
type ViOTPClient struct {
	token        string
	httpClient   *http.Client
	baseURL      string
	pollInterval time.Duration
}

// NewViOTPClient creates a new ViOTP API client.
func NewViOTPClient(token string) *ViOTPClient {
	return &ViOTPClient{
		token:        token,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		baseURL:      defaultBaseURL,
		pollInterval: defaultPollInterval,
	}
}

type viOTPResponse struct {
	StatusCode int             `json:"status_code"`
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
}

type balanceData struct {
	Balance int64 `json:"balance"`
}

type rentData struct {
	PhoneNumber   string `json:"phone_number"`
	RePhoneNumber string `json:"re_phone_number"`
	CountryISO    string `json:"countryISO"`
	CountryCode   string `json:"countryCode"`
	Balance       int64  `json:"balance"`
	RequestID     any    `json:"request_id"` // API returns int or string
}

type sessionData struct {
	Status int    `json:"Status"`
	Code   string `json:"Code"`
}

func (c *ViOTPClient) doGet(ctx context.Context, path string, params map[string]string) (*viOTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("viotp: failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Set("token", c.token)
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("viotp: request failed: %w", err)
	}
	defer resp.Body.Close()

	var result viOTPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("viotp: failed to decode response: %w", err)
	}

	return &result, nil
}

// CheckBalance returns the current ViOTP account balance.
func (c *ViOTPClient) CheckBalance(ctx context.Context) (int64, error) {
	result, err := c.doGet(ctx, "/users/balance", nil)
	if err != nil {
		return 0, err
	}
	if result.StatusCode != 200 {
		return 0, fmt.Errorf("viotp: balance check failed (code=%d): %s", result.StatusCode, result.Message)
	}

	var data balanceData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return 0, fmt.Errorf("viotp: failed to parse balance data: %w", err)
	}
	return data.Balance, nil
}

// RentNumber rents a phone number for the specified service.
func (c *ViOTPClient) RentNumber(ctx context.Context, serviceID int) (*RentResult, error) {
	result, err := c.doGet(ctx, "/request/getv2", map[string]string{
		"serviceId": strconv.Itoa(serviceID),
	})
	if err != nil {
		return nil, err
	}

	switch result.StatusCode {
	case 200:
		// success
	case -2:
		return nil, fmt.Errorf("viotp: insufficient balance")
	case -3:
		return nil, fmt.Errorf("viotp: phone numbers out of stock")
	case -4:
		return nil, fmt.Errorf("viotp: service not found or suspended (serviceId=%d)", serviceID)
	case 429:
		return nil, fmt.Errorf("viotp: rate limit exceeded")
	default:
		return nil, fmt.Errorf("viotp: rent failed (code=%d): %s", result.StatusCode, result.Message)
	}

	var data rentData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, fmt.Errorf("viotp: failed to parse rent data: %w", err)
	}

	// request_id can be int or string from the API
	requestID := fmt.Sprintf("%v", data.RequestID)

	return &RentResult{
		PhoneNumber:   data.PhoneNumber,
		RePhoneNumber: data.RePhoneNumber,
		RequestID:     requestID,
		CountryISO:    data.CountryISO,
		CountryCode:   data.CountryCode,
		Balance:       data.Balance,
	}, nil
}

// WaitForOTP polls the ViOTP session API until an OTP is received or timeout.
func (c *ViOTPClient) WaitForOTP(ctx context.Context, requestID string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("viotp: OTP timeout after %v", timeout)
		}

		result, err := c.doGet(ctx, "/session/getv2", map[string]string{
			"requestId": requestID,
		})
		if err != nil {
			return "", err
		}

		if result.StatusCode != 200 {
			return "", fmt.Errorf("viotp: session check failed (code=%d): %s", result.StatusCode, result.Message)
		}

		var data sessionData
		if err := json.Unmarshal(result.Data, &data); err != nil {
			return "", fmt.Errorf("viotp: failed to parse session data: %w", err)
		}

		switch data.Status {
		case 1: // OTP received
			if data.Code == "" {
				return "", fmt.Errorf("viotp: OTP status=1 but code is empty")
			}
			return data.Code, nil
		case 2: // Expired
			return "", fmt.Errorf("viotp: phone session expired (requestId=%s)", requestID)
		case 0: // Waiting
			// continue polling
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(c.pollInterval):
		}
	}
}
