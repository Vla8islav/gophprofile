package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidAuthToken = errors.New("invalid auth token")
	ErrExpiredAuthToken = errors.New("expired auth token")
)

// пока что будет константой
const authTokenTTL = 24 * time.Hour

func CreateAuthToken(userID int64, secret []byte) (string, error) {
	expiresAt := time.Now().Add(authTokenTTL).Unix()

	payload := fmt.Sprintf("%d:%d", userID, expiresAt)
	signature := signAuthTokenPayload(payload, secret)

	return base64.RawURLEncoding.EncodeToString([]byte(payload + ":" + signature)), nil
}

func ParseAuthToken(token string, secret []byte) (int64, error) {
	rawToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, ErrInvalidAuthToken
	}

	parts := strings.Split(string(rawToken), ":")
	if len(parts) != 3 {
		return 0, ErrInvalidAuthToken
	}

	payload := parts[0] + ":" + parts[1]
	expectedSignature := signAuthTokenPayload(payload, secret)
	actualSignature := parts[2]

	if !hmac.Equal([]byte(actualSignature), []byte(expectedSignature)) {
		return 0, ErrInvalidAuthToken
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ErrInvalidAuthToken
	}

	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, ErrInvalidAuthToken
	}

	if time.Now().Unix() > expiresAt {
		return 0, ErrExpiredAuthToken
	}

	return userID, nil
}

func signAuthTokenPayload(payload string, authTokenSecret []byte) string {
	mac := hmac.New(sha256.New, authTokenSecret)
	mac.Write([]byte(payload))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
