package auth

import (
	"testing"

	"github.com/spf13/viper"
)

func withTestStore(t *testing.T, f func(s *Store)) {
	t.Helper()
	viper.Set("storage.driver", "sqlite")
	viper.Set("storage.dir", t.TempDir())
	s, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}
	f(s)
}

func TestUserRoundTrip(t *testing.T) {
	withTestStore(t, func(s *Store) {
		u := User{Username: "alice", PasswordHash: "h", Role: RoleAdmin, CreatedAt: "2026-06-10T00:00:00Z"}
		if err := s.CreateUser(u); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetUserByUsername("alice")
		if err != nil || got.Role != RoleAdmin || got.PasswordHash != "h" {
			t.Fatalf("got %+v err %v", got, err)
		}
		if got.ID == 0 {
			t.Fatal("expected assigned id")
		}
		byID, err := s.GetUserByID(got.ID)
		if err != nil || byID.Username != "alice" {
			t.Fatalf("byID %+v err %v", byID, err)
		}
		if err := s.CreateUser(u); err == nil {
			t.Fatal("duplicate username must fail")
		}
		if _, err := s.GetUserByUsername("nobody"); err != ErrNoRow {
			t.Fatalf("expected ErrNoRow, got %v", err)
		}
		users, err := s.ListUsers()
		if err != nil || len(users) != 1 {
			t.Fatalf("list users: %v %v", users, err)
		}
	})
}

func TestKeyRoundTrip(t *testing.T) {
	withTestStore(t, func(s *Store) {
		k := APIKey{KeyHash: "kh", Label: "fintech", Role: RoleOperator, Ledgers: "demo", CreatedAt: "2026-06-10T00:00:00Z"}
		if err := s.CreateKey(k); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetKeyByHash("kh")
		if err != nil || got.Label != "fintech" || got.ID == 0 {
			t.Fatalf("got %+v err %v", got, err)
		}
		if _, err := s.GetKeyByHash("missing"); err != ErrNoRow {
			t.Fatalf("expected ErrNoRow, got %v", err)
		}
		keys, err := s.ListKeys()
		if err != nil || len(keys) != 1 {
			t.Fatalf("list: %v %v", keys, err)
		}
		if err := s.RevokeKey(got.ID, "2026-06-11T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetKeyByHash("kh")
		if got.RevokedAt == "" {
			t.Fatal("expected revoked")
		}
	})
}

func TestSessionRoundTrip(t *testing.T) {
	withTestStore(t, func(s *Store) {
		sess := Session{UserID: 1, TokenHash: "th", ExpiresAt: "2026-06-10T12:00:00Z", CreatedAt: "2026-06-10T00:00:00Z"}
		if err := s.CreateSession(sess); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetSessionByHash("th")
		if err != nil || got.UserID != 1 {
			t.Fatalf("got %+v err %v", got, err)
		}
		if _, err := s.GetSessionByHash("missing"); err != ErrNoRow {
			t.Fatalf("expected ErrNoRow, got %v", err)
		}
		if err := s.RevokeSession("th", "2026-06-10T01:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetSessionByHash("th")
		if got.RevokedAt == "" {
			t.Fatal("expected revoked")
		}
	})
}
