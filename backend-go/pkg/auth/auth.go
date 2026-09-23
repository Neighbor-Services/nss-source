package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

type Claims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	UserType    string `json:"user_type"`
	IsStaff     bool   `json:"is_staff,omitempty"`
	IsSuperuser bool   `json:"is_superuser,omitempty"`
	jwt.RegisteredClaims
}

// GenerateTokenPair generates an access token and refresh token matching Django SimpleJWT claims.
func GenerateTokenPair(userID, email, userType, accessSecret, refreshSecret string, accessExpiryHours, refreshExpiryDays int) (accessToken string, refreshToken string, err error) {
	now := time.Now()

	// Access Token
	accessClaims := &Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(accessExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	accessObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessObj.SignedString([]byte(accessSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(refreshExpiryDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	refreshObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshObj.SignedString([]byte(refreshSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken parses and validates a JWT token.
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// HashPassword hashes a password with bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a password against either bcrypt or Django's pbkdf2_sha256 format.
func CheckPassword(hashedPassword, plainPassword string) bool {
	// 1. Try standard bcrypt
	if strings.HasPrefix(hashedPassword, "$2a$") || strings.HasPrefix(hashedPassword, "$2b$") || strings.HasPrefix(hashedPassword, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword)) == nil
	}

	// 2. Try Django pbkdf2_sha256: pbkdf2_sha256$iterations$salt$hash
	if strings.HasPrefix(hashedPassword, "pbkdf2_sha256$") {
		parts := strings.Split(hashedPassword, "$")
		if len(parts) != 4 {
			return false
		}
		iterations, err := strconv.Atoi(parts[1])
		if err != nil {
			return false
		}
		salt := parts[2]
		expectedHash := parts[3]

		derivedKey := pbkdf2.Key([]byte(plainPassword), []byte(salt), iterations, 32, sha256.New)
		computedHash := base64.StdEncoding.EncodeToString(derivedKey)

		return subtle.ConstantTimeCompare([]byte(computedHash), []byte(expectedHash)) == 1
	}

	return false
}

// GenerateOTP generates a random numeric OTP string.
func GenerateOTP(length int) (string, error) {
	const digits = "0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = digits[int(bytes[i])%len(digits)]
	}
	return string(bytes), nil
}
