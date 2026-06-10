package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/amezianechayer/corren/auth"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

type fixtures struct {
	adminKey    string
	operatorKey string // scoped to ledger "demo"
	readonlyKey string
	router      *gin.Engine
}

func setup(t *testing.T) fixtures {
	t.Helper()
	viper.Set("storage.driver", "sqlite")
	viper.Set("storage.dir", t.TempDir())
	viper.Set("auth.enabled", true)
	viper.Set("server.http.basic_auth", nil)
	t.Cleanup(func() { viper.Set("auth.enabled", false) })

	svc, err := auth.NewServiceFromConfig()
	if err != nil {
		t.Fatal(err)
	}

	adminKey, _, _ := svc.CreateKey("admin", auth.RoleAdmin, nil)
	operatorKey, _, _ := svc.CreateKey("op", auth.RoleOperator, []string{"demo"})
	readonlyKey, _, _ := svc.CreateKey("ro", auth.RoleReadonly, nil)

	m := NewAuthMiddleware(svc)
	r := gin.New()
	r.Use(m.AuthMiddleware(r))
	ok := func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) }
	r.GET("/:ledger/contracts", ok)
	r.POST("/:ledger/contracts", ok)
	r.GET("/auth/admin/keys", ok)
	r.POST("/auth/login", ok)

	return fixtures{adminKey, operatorKey, readonlyKey, r}
}

func request(t *testing.T, r *gin.Engine, method, url, token string) int {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestAuthMatrix(t *testing.T) {
	f := setup(t)

	cases := []struct {
		name        string
		method, url string
		token       string
		expected    int
	}{
		{"no header", "GET", "/demo/contracts", "", http.StatusUnauthorized},
		{"garbage token", "GET", "/demo/contracts", "crn_bogus", http.StatusUnauthorized},
		{"readonly GET ok", "GET", "/demo/contracts", "RO", http.StatusOK},
		{"readonly POST forbidden", "POST", "/demo/contracts", "RO", http.StatusForbidden},
		{"operator POST ok on scoped ledger", "POST", "/demo/contracts", "OP", http.StatusOK},
		{"operator POST forbidden on other ledger", "POST", "/prod/contracts", "OP", http.StatusForbidden},
		{"operator GET forbidden on other ledger", "GET", "/prod/contracts", "OP", http.StatusForbidden},
		{"operator forbidden on admin route", "GET", "/auth/admin/keys", "OP", http.StatusForbidden},
		{"admin ok on admin route", "GET", "/auth/admin/keys", "ADMIN", http.StatusOK},
		{"login exempt from auth", "POST", "/auth/login", "", http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token := c.token
			switch token {
			case "RO":
				token = f.readonlyKey
			case "OP":
				token = f.operatorKey
			case "ADMIN":
				token = f.adminKey
			}
			if code := request(t, f.router, c.method, c.url, token); code != c.expected {
				t.Fatalf("expected %d, got %d", c.expected, code)
			}
		})
	}
}

func TestAuthDisabledPassesThrough(t *testing.T) {
	f := setup(t)
	viper.Set("auth.enabled", false)

	// non-regression: with auth disabled, no header is required anywhere
	if code := request(t, f.router, "POST", "/demo/contracts", ""); code != http.StatusOK {
		t.Fatalf("expected 200 with auth disabled, got %d", code)
	}
	if code := request(t, f.router, "GET", "/auth/admin/keys", ""); code != http.StatusOK {
		t.Fatalf("expected 200 with auth disabled, got %d", code)
	}
}
