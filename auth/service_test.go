package auth

import (
	"testing"
	"time"
)

func withService(t *testing.T, f func(svc *Service)) {
	withTestStore(t, func(s *Store) { f(NewService(s)) })
}

func TestLoginFlow(t *testing.T) {
	withService(t, func(svc *Service) {
		if _, err := svc.CreateUser("alice", "s3cret", RoleOperator); err != nil {
			t.Fatal(err)
		}
		token, expiresAt, err := svc.Login("alice", "s3cret")
		if err != nil {
			t.Fatal(err)
		}
		if expiresAt == "" {
			t.Fatal("expected expiry")
		}
		id, err := svc.Authenticate(token)
		if err != nil || id.Subject != "alice" || id.Role != RoleOperator || id.Kind != "session" {
			t.Fatalf("id %+v err %v", id, err)
		}
		if _, _, err := svc.Login("alice", "wrong"); err == nil {
			t.Fatal("bad password must fail")
		}
		if _, _, err := svc.Login("bob", "x"); err == nil {
			t.Fatal("unknown user must fail")
		}
		if err := svc.Logout(token); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Authenticate(token); err == nil {
			t.Fatal("revoked session must fail")
		}
	})
}

func TestSessionExpiry(t *testing.T) {
	withService(t, func(svc *Service) {
		svc.SessionTTL = -time.Minute // sessions are born expired
		svc.CreateUser("carol", "pw", RoleReadonly)
		token, _, _ := svc.Login("carol", "pw")
		if _, err := svc.Authenticate(token); err == nil {
			t.Fatal("expired session must fail")
		}
	})
}

func TestCreateUserInvalidRole(t *testing.T) {
	withService(t, func(svc *Service) {
		if _, err := svc.CreateUser("dave", "pw", "superuser"); err == nil {
			t.Fatal("invalid role must fail")
		}
	})
}

func TestAPIKeyFlow(t *testing.T) {
	withService(t, func(svc *Service) {
		plain, key, err := svc.CreateKey("fintech", RoleOperator, []string{"demo"})
		if err != nil {
			t.Fatal(err)
		}
		if key.Label != "fintech" || key.ID == 0 {
			t.Fatalf("key %+v", key)
		}
		id, err := svc.Authenticate(plain)
		if err != nil || id.Kind != "key" || id.Role != RoleOperator || !id.AllowsLedger("demo") || id.AllowsLedger("prod") {
			t.Fatalf("id %+v err %v", id, err)
		}

		// default scope is all ledgers
		plainAll, _, err := svc.CreateKey("backoffice", RoleAdmin, nil)
		if err != nil {
			t.Fatal(err)
		}
		idAll, _ := svc.Authenticate(plainAll)
		if !idAll.AllowsLedger("anything") {
			t.Fatal("default key scope must be *")
		}

		if err := svc.RevokeKey(key.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Authenticate(plain); err == nil {
			t.Fatal("revoked key must fail")
		}
		if _, err := svc.Authenticate("crn_doesnotexist"); err == nil {
			t.Fatal("unknown token must fail")
		}
		if _, err := svc.Authenticate("garbage"); err == nil {
			t.Fatal("unprefixed token must fail")
		}
		if _, _, err := svc.CreateKey("x", "bogusrole", nil); err == nil {
			t.Fatal("invalid role must fail")
		}
	})
}
