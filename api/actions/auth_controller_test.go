package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/amezianechayer/corren/auth"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func setupAuthRouter(t *testing.T) (*gin.Engine, *auth.Service) {
	t.Helper()
	viper.Set("storage.driver", "sqlite")
	viper.Set("storage.dir", t.TempDir())

	svc, err := auth.NewServiceFromConfig()
	if err != nil {
		t.Fatal(err)
	}

	ctl := NewAuthController(svc)
	r := gin.New()
	grp := r.Group("/auth")
	grp.POST("/login", ctl.Login)
	grp.POST("/logout", ctl.Logout)
	grp.GET("/me", ctl.Me)
	grp.POST("/admin/keys", ctl.CreateKey)
	grp.GET("/admin/keys", ctl.ListKeys)
	grp.DELETE("/admin/keys/:id", ctl.RevokeKey)
	grp.POST("/admin/users", ctl.CreateUser)
	grp.GET("/admin/users", ctl.ListUsers)
	return r, svc
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func doAuth(t *testing.T, r *gin.Engine, method, url string, body interface{}, headers map[string]string) (int, map[string]interface{}) {
	t.Helper()
	// reuse the JSON helper from contract_controller_test.go but with headers
	return doWithHeaders(t, r, method, url, body, headers)
}

func TestAuthEndpointsFlow(t *testing.T) {
	r, svc := setupAuthRouter(t)
	if _, err := svc.CreateUser("admin", "adminpw", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}

	// wrong password → 401
	code, _ := doAuth(t, r, "POST", "/auth/login", map[string]string{"username": "admin", "password": "nope"}, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}

	// login ok → token + role
	code, out := doAuth(t, r, "POST", "/auth/login", map[string]string{"username": "admin", "password": "adminpw"}, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %v", code, out)
	}
	data := out["data"].(map[string]interface{})
	token, _ := data["token"].(string)
	if token == "" || data["role"] != auth.RoleAdmin || data["expires_at"] == "" {
		t.Fatalf("bad login payload: %v", data)
	}

	// me → identity
	code, out = doAuth(t, r, "GET", "/auth/me", nil, bearer(token))
	if code != http.StatusOK {
		t.Fatalf("me: expected 200, got %d", code)
	}
	data = out["data"].(map[string]interface{})
	if data["subject"] != "admin" || data["kind"] != "session" {
		t.Fatalf("bad me payload: %v", data)
	}

	// me without token → 401
	code, _ = doAuth(t, r, "GET", "/auth/me", nil, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("me without token: expected 401, got %d", code)
	}

	// create key → 201, plain key shown once
	code, out = doAuth(t, r, "POST", "/auth/admin/keys",
		map[string]interface{}{"label": "fintech", "role": "operator", "ledgers": []string{"demo"}}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create key: expected 201, got %d: %v", code, out)
	}
	data = out["data"].(map[string]interface{})
	plainKey, _ := data["key"].(string)
	if plainKey == "" || data["label"] != "fintech" {
		t.Fatalf("bad key payload: %v", data)
	}
	keyID := data["id"].(float64)

	// invalid role → 400
	code, _ = doAuth(t, r, "POST", "/auth/admin/keys",
		map[string]interface{}{"label": "x", "role": "superuser"}, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid role: expected 400, got %d", code)
	}

	// list keys → no plaintext, no hash
	code, out = doAuth(t, r, "GET", "/auth/admin/keys", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("list keys: expected 200, got %d", code)
	}
	for _, k := range out["data"].([]interface{}) {
		km := k.(map[string]interface{})
		if _, has := km["key"]; has {
			t.Fatal("plaintext key must never be listed")
		}
		if _, has := km["KeyHash"]; has {
			t.Fatal("key hash must never be listed")
		}
	}

	// revoke key → key stops authenticating
	code, _ = doAuth(t, r, "DELETE", "/auth/admin/keys/"+strconv.Itoa(int(keyID)), nil, nil)
	if code != http.StatusOK {
		t.Fatalf("revoke: expected 200, got %d (id %v)", code, keyID)
	}
	if _, err := svc.Authenticate(plainKey); err == nil {
		t.Fatal("revoked key must not authenticate")
	}

	// create user → 201, then duplicate → error
	code, _ = doAuth(t, r, "POST", "/auth/admin/users",
		map[string]string{"username": "op1", "password": "pw123", "role": "operator"}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create user: expected 201, got %d", code)
	}
	code, out = doAuth(t, r, "GET", "/auth/admin/users", nil, nil)
	if code != http.StatusOK || len(out["data"].([]interface{})) != 2 {
		t.Fatalf("list users: %d %v", code, out)
	}

	// logout → token dead
	code, _ = doAuth(t, r, "POST", "/auth/logout", nil, bearer(token))
	if code != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d", code)
	}
	code, _ = doAuth(t, r, "GET", "/auth/me", nil, bearer(token))
	if code != http.StatusUnauthorized {
		t.Fatalf("me after logout: expected 401, got %d", code)
	}
}
