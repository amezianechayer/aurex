package auth

import (
	"strings"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	k1 := GenerateToken(PrefixKey)
	k2 := GenerateToken(PrefixKey)
	s1 := GenerateToken(PrefixSession)

	if !strings.HasPrefix(k1, "crn_") || !strings.HasPrefix(s1, "crs_") {
		t.Fatalf("bad prefixes: %q %q", k1, s1)
	}
	if k1 == k2 {
		t.Fatal("tokens must be unique")
	}
	if len(k1) != len("crn_")+64 { // 32 random bytes hex-encoded
		t.Fatalf("unexpected token length %d", len(k1))
	}
}

func TestHashToken(t *testing.T) {
	tok := GenerateToken(PrefixKey)
	h1, h2 := HashToken(tok), HashToken(tok)
	if h1 != h2 {
		t.Fatal("hash must be deterministic")
	}
	if len(h1) != 64 {
		t.Fatalf("expected sha256 hex (64), got %d", len(h1))
	}
	if h1 == HashToken(GenerateToken(PrefixKey)) {
		t.Fatal("different tokens must hash differently")
	}
}
