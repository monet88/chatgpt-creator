package phone

import (
	"context"
	"time"
)

// PhoneProvider abstracts SMS-based phone verification for registration flows.
type PhoneProvider interface {
	// CheckBalance returns the current account balance in VND.
	CheckBalance(ctx context.Context) (int64, error)
	// RentNumber rents a phone number for the given service ID.
	RentNumber(ctx context.Context, serviceID int) (*RentResult, error)
	// WaitForOTP polls until an OTP code is received or timeout is reached.
	WaitForOTP(ctx context.Context, requestID string, timeout time.Duration) (string, error)
}

// RentResult holds the result of renting a phone number.
type RentResult struct {
	PhoneNumber   string `json:"phone_number"`
	RePhoneNumber string `json:"re_phone_number"`
	RequestID     string `json:"request_id"`
	CountryISO    string `json:"country_iso"`
	CountryCode   string `json:"country_code"`
	Balance       int64  `json:"balance"`
}
