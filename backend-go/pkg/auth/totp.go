package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// GenerateTOTPSecret generates a random 20-byte base32 secret.
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateTOTPCode generates a 6-digit TOTP code for a given timestamp.
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", err
	}

	interval := uint64(t.Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, interval)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	code := (binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff) % 1000000

	return fmt.Sprintf("%06d", code), nil
}

// VerifyTOTPCode verifies a 6-digit TOTP code allowing for 1-step clock skew (+-30 seconds).
func VerifyTOTPCode(secret string, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}

	now := time.Now()
	for _, offset := range []int64{-30, 0, 30} {
		testTime := now.Add(time.Duration(offset) * time.Second)
		gen, err := GenerateTOTPCode(secret, testTime)
		if err == nil && gen == code {
			return true
		}
	}
	return false
}

// GetOTPAuthURL returns the otpauth:// URI for QR code generation in Google Authenticator.
func GetOTPAuthURL(issuer, accountName, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.PathEscape(issuer),
		url.PathEscape(accountName),
		secret,
		url.QueryEscape(issuer),
	)
}
