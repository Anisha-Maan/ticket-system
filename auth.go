package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var jwtSecret = []byte("super-secret-key-change-me")

type Claims struct {
	UserID int   `json:"user_id"`
	Exp    int64 `json:"exp"`
}

func generateSalt() string {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func verifyPassword(password, salt, storedHash string) bool {
	return hashPassword(password, salt) == storedHash
}

func GenerateJWT(userID int) (string, error) {

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	payload := Claims{
		UserID: userID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsignedToken := headerEncoded + "." + payloadEncoded

	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(unsignedToken))

	signature := h.Sum(nil)

	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	token := unsignedToken + "." + signatureEncoded

	return token, nil
}

func ValidateJWT(token string) (*Claims, error) {

	parts := splitToken(token)

	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	unsignedToken := parts[0] + "." + parts[1]

	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(unsignedToken))

	expectedSignature := h.Sum(nil)

	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}

	if !hmac.Equal(expectedSignature, providedSignature) {
		return nil, errors.New("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid payload")
	}

	var claims Claims

	err = json.Unmarshal(payloadBytes, &claims)
	if err != nil {
		return nil, errors.New("invalid claims")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func splitToken(token string) []string {

	var parts []string
	current := ""

	for _, ch := range token {

		if ch == '.' {
			parts = append(parts, current)
			current = ""
			continue
		}

		current += string(ch)
	}

	parts = append(parts, current)

	return parts
}