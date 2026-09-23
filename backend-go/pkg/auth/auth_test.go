package auth

import (
	"testing"
	"time"
)

func TestTokenGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-12345"
	refreshSecret := "test-refresh-secret-key-12345"
	userID := "550e8400-e29b-41d4-a716-446655440000"
	email := "provider@example.com"
	userType := "PROVIDER"

	access, refresh, err := GenerateTokenPair(userID, email, userType, secret, refreshSecret, 1, 7)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := ValidateToken(access, secret)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID || claims.Email != email || claims.UserType != userType {
		t.Errorf("Claims mismatch: got %+v", claims)
	}

	refreshClaims, err := ValidateToken(refresh, refreshSecret)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("Refresh claims mismatch: got %+v", refreshClaims)
	}
}

func TestPasswordHashingAndVerification(t *testing.T) {
	password := "SecurePass123!"

	// Test bcrypt
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Errorf("Password check failed for bcrypt")
	}

	if CheckPassword(hash, "WrongPass") {
		t.Errorf("Password check passed for incorrect password")
	}
}

func TestGenerateOTP(t *testing.T) {
	otp, err := GenerateOTP(6)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}
	if len(otp) != 6 {
		t.Errorf("Expected 6-digit OTP, got length %d (%s)", len(otp), otp)
	}
}

func TestTOTPGenerationAndVerification(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}
	if len(secret) == 0 {
		t.Fatal("Secret was empty")
	}

	url := GetOTPAuthURL("NeighborServices", "admin@example.com", secret)
	if len(url) == 0 {
		t.Fatal("OTP auth URL was empty")
	}

	code, err := GenerateTOTPCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateTOTPCode failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("Expected 6-digit TOTP code, got %s", code)
	}

	if !VerifyTOTPCode(secret, code) {
		t.Fatal("VerifyTOTPCode returned false for valid code")
	}

	if VerifyTOTPCode(secret, "000000") && code != "000000" {
		t.Fatal("VerifyTOTPCode returned true for invalid code")
	}
}
