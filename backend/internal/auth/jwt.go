package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// HashPassword hashes plaintext password with SHA-256.
func HashPassword(password string) string {
	digest := sha256.Sum256([]byte(password))
	return hex.EncodeToString(digest[:])
}

// ComparePassword compares plaintext password with stored hash.
func ComparePassword(hash, password string) bool {
	return hash == HashPassword(password)
}

type tokenPayload struct {
	Subject   string   `json:"sub"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	Issuer    string   `json:"iss"`
	Audience  []string `json:"aud"`
	UserID    int      `json:"userId"`
}

// GenerateToken creates a JWT signed with the provided secret using HS256.
func GenerateToken(userID int, email, secret string, ttl time.Duration) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	payload := tokenPayload{
		Subject:   email,
		ExpiresAt: time.Now().Add(ttl).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "fullstack-go-api",
		Audience:  []string{"fullstack-go-api"},
		UserID:    userID,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerPart + "." + payloadPart

	signature := sign(signingInput, secret)
	return signingInput + "." + signature, nil
}

// ParseToken validates the provided JWT and returns the user id and email.
func ParseToken(tokenString, secret string) (int, string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return 0, "", errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := sign(signingInput, secret)
	if !secureCompare(expectedSig, parts[2]) {
		return 0, "", errors.New("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, "", errors.New("invalid token payload")
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return 0, "", errors.New("invalid token payload")
	}

	if payload.ExpiresAt < time.Now().Unix() {
		return 0, "", errors.New("token expired")
	}

	if payload.Subject == "" || payload.UserID == 0 {
		return 0, "", errors.New("invalid token claims")
	}

	return payload.UserID, payload.Subject, nil
}

func sign(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	signature := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}

func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
