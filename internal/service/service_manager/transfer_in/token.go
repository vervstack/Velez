package transfer_in

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	tokenByteLen = 32
	tokenTtl     = time.Hour
)

func newToken() (string, error) {
	raw := make([]byte, tokenByteLen)

	_, err := rand.Read(raw)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading random bytes")
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

func tokenExpiry(now time.Time) time.Time {
	return now.Add(tokenTtl)
}
