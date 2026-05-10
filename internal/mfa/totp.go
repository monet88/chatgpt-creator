package mfa

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func GenerateTOTP(secret string) (string, error) {
	return GenerateTOTPAt(secret, time.Now())
}

func GenerateTOTPAt(secret string, t time.Time) (string, error) {
	clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	clean = strings.TrimRight(clean, "=")
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(clean)
	if err != nil {
		if rem := len(clean) % 8; rem != 0 {
			clean += strings.Repeat("=", 8-rem)
		}
		key, err = base32.StdEncoding.DecodeString(clean)
		if err != nil {
			return "", fmt.Errorf("mfa: decode TOTP secret: %w", err)
		}
	}

	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(t.Unix())/30)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", code%1_000_000), nil
}
