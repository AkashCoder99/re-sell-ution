package utils

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

type TokenManager struct {
	Secret []byte
}

type tokenPayload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func NewTokenManager(secret string) TokenManager {
	return TokenManager{Secret: []byte(secret)}
}

func (m TokenManager) Create(userID string, expiresAt time.Time) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(tokenPayload{Sub: userID, Exp: expiresAt.Unix()})
	if err != nil {
		return "", err
	}

	head := base64.RawURLEncoding.EncodeToString(headerJSON)
	body := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsigned := head + "." + body

	h := hmac.New(sha256.New, m.Secret)
	h.Write([]byte(unsigned))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return unsigned + "." + sig, nil
}

func (m TokenManager) Parse(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token format")
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSig := sign(unsigned, m.Secret)
	providedSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", errors.New("invalid token signature")
	}
	if !hmac.Equal(expectedSig, providedSig) {
		return "", errors.New("token signature mismatch")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("invalid token payload")
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", errors.New("invalid token payload")
	}

	if payload.Sub == "" {
		return "", errors.New("missing subject")
	}
	if time.Now().Unix() >= payload.Exp {
		return "", errors.New("token expired")
	}

	return payload.Sub, nil
}

func sign(unsigned string, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(unsigned))
	return h.Sum(nil)
}

func Fingerprint(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
