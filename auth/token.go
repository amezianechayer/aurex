package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const (
	PrefixKey     = "crn_"
	PrefixSession = "crs_"
)

func GenerateToken(prefix string) string {
	b := make([]byte, 32)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
