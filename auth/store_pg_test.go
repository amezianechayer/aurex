package auth

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestPGAuthStoreRoundTrip(t *testing.T) {
	conn := os.Getenv("CORREN_TEST_PG_CONN")
	if conn == "" {
		t.Skip("CORREN_TEST_PG_CONN not set")
	}
	viper.Set("storage.driver", "postgres")
	viper.Set("storage.postgres.conn_string", conn)
	defer viper.Set("storage.driver", "sqlite")

	s, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	// fresh schema for idempotent runs
	if _, err := s.db.Exec(`DROP SCHEMA IF EXISTS "corren_auth" CASCADE`); err != nil {
		t.Fatal(err)
	}
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}

	u := User{Username: "pgalice", PasswordHash: "h", Role: RoleAdmin, CreatedAt: "2026-06-10T00:00:00Z"}
	if err := s.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetUserByUsername("pgalice")
	if err != nil || got.Role != RoleAdmin {
		t.Fatalf("user: %+v err %v", got, err)
	}
	if _, err := s.GetUserByUsername("missing"); err != ErrNoRow {
		t.Fatalf("expected ErrNoRow, got %v", err)
	}

	k := APIKey{KeyHash: "pgkh", Label: "pg", Role: RoleOperator, Ledgers: "*", CreatedAt: "2026-06-10T00:00:00Z"}
	if err := s.CreateKey(k); err != nil {
		t.Fatal(err)
	}
	gotKey, err := s.GetKeyByHash("pgkh")
	if err != nil || gotKey.Label != "pg" {
		t.Fatalf("key: %+v err %v", gotKey, err)
	}
	if err := s.RevokeKey(gotKey.ID, "2026-06-11T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	gotKey, _ = s.GetKeyByHash("pgkh")
	if gotKey.RevokedAt == "" {
		t.Fatal("expected revoked")
	}

	sess := Session{UserID: got.ID, TokenHash: "pgth", ExpiresAt: "2026-06-10T12:00:00Z", CreatedAt: "2026-06-10T00:00:00Z"}
	if err := s.CreateSession(sess); err != nil {
		t.Fatal(err)
	}
	gotSess, err := s.GetSessionByHash("pgth")
	if err != nil || gotSess.UserID != got.ID {
		t.Fatalf("session: %+v err %v", gotSess, err)
	}
	if err := s.RevokeSession("pgth", "2026-06-10T01:00:00Z"); err != nil {
		t.Fatal(err)
	}
}
