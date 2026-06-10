# Phase 1 — Auth & Roles Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add production-grade authentication to corren: API keys with roles for machine clients, revocable login sessions for Horizon, role+ledger-scope enforcement on every route.

**Architecture:** New `auth/` package with its own database (`corren_auth.db` SQLite file or `corren_auth` Postgres schema) accessed through `database/sql` (sqlite3 driver + pgx stdlib adapter — same modules, no new dependency). Opaque random tokens stored hashed (sha256), bcrypt for passwords. The existing `AuthMiddleware` is rewritten to enforce a role matrix (admin/operator/readonly) and per-ledger scope; `auth.enabled=false` (default) keeps today's behavior byte-for-byte.

**Tech Stack:** Go 1.16, gin, go-sqlbuilder, viper, fx, `golang.org/x/crypto/bcrypt` (already in go.sum), `crypto/rand` + `crypto/sha256` stdlib.

**Spec:** `docs/superpowers/specs/2026-06-10-pilot-hardening-design.md` (Phase 1 section). Phases 2–4 get their own plans.

**Branch:** `feature/pilot-hardening` (already created, stacked on `feature/sharia-murabaha-v1`). NEVER commit to main.

---

## File map

| File | Responsibility |
|---|---|
| `auth/token.go` (new) | token generation (`crn_`/`crs_` prefixes) and sha256 hashing |
| `auth/types.go` (new) | `User`, `APIKey`, `Session`, `Identity`, role constants + rank |
| `auth/migration/v001.sql` (new) | the three tables, `--statement` format |
| `auth/store.go` (new) | `Store` over `database/sql`, both flavors, CRUD |
| `auth/service.go` (new) | `Service`: CreateUser/Login/Logout/Authenticate/CreateKey/RevokeKey |
| `api/middlewares/auth_middleware.go` (modify) | bearer extraction, role matrix, ledger scope |
| `api/actions/auth_controller.go` (new) | `/auth/login|logout|me`, `/auth/admin/*` |
| `api/routes/routes.go` (modify) | auth routes + controller wiring |
| `api/actions/controllers.go` (modify) | fx provide |
| `api/api.go` (modify) | provide `*auth.Service` via fx |
| `cmd/root.go` (modify) | `corren auth init` command |

Roles: `readonly` < `operator` < `admin`. GET/HEAD need readonly+, mutating verbs need operator+, `/auth/admin/*` needs admin. Key scope: list of ledger names or `*`.

---

### Task 1: Token generation and hashing

**Files:** Create `auth/token.go`, Test `auth/token_test.go`

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run** `go test ./auth/` — expect FAIL (undefined: GenerateToken)

- [ ] **Step 3: Implement**

```go
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
```

- [ ] **Step 4: Run** `go test ./auth/` — expect PASS
- [ ] **Step 5: Commit** `git add auth/ && git commit -m "feat(auth): token generation and hashing"`

---

### Task 2: Types and roles

**Files:** Create `auth/types.go`, Test `auth/types_test.go`

- [ ] **Step 1: Failing test**

```go
package auth

import "testing"

func TestRoleRank(t *testing.T) {
	if !RoleAtLeast(RoleAdmin, RoleOperator) || !RoleAtLeast(RoleOperator, RoleReadonly) {
		t.Fatal("admin>=operator>=readonly expected")
	}
	if RoleAtLeast(RoleReadonly, RoleOperator) {
		t.Fatal("readonly must not satisfy operator")
	}
	if RoleAtLeast("bogus", RoleReadonly) {
		t.Fatal("unknown role must rank below everything")
	}
}

func TestIdentityAllowsLedger(t *testing.T) {
	all := Identity{Role: RoleOperator, Ledgers: []string{"*"}}
	one := Identity{Role: RoleOperator, Ledgers: []string{"demo"}}
	if !all.AllowsLedger("anything") || !one.AllowsLedger("demo") || one.AllowsLedger("prod") {
		t.Fatal("ledger scope broken")
	}
}
```

- [ ] **Step 2: Run** `go test ./auth/` — FAIL
- [ ] **Step 3: Implement**

```go
package auth

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleReadonly = "readonly"
)

var roleRank = map[string]int{RoleReadonly: 1, RoleOperator: 2, RoleAdmin: 3}

func RoleIsValid(r string) bool { return roleRank[r] > 0 }

func RoleAtLeast(have, need string) bool { return roleRank[have] >= roleRank[need] }

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	CreatedAt    string `json:"created_at"`
	DisabledAt   string `json:"disabled_at,omitempty"`
}

type APIKey struct {
	ID        int64  `json:"id"`
	KeyHash   string `json:"-"`
	Label     string `json:"label"`
	Role      string `json:"role"`
	Ledgers   string `json:"ledgers"` // CSV or "*"
	CreatedAt string `json:"created_at"`
	RevokedAt string `json:"revoked_at,omitempty"`
}

type Session struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt string
	CreatedAt string
	RevokedAt string
}

// Identity is what the middleware puts in the gin context.
type Identity struct {
	Subject string   `json:"subject"` // username or key label
	Role    string   `json:"role"`
	Ledgers []string `json:"ledgers"`
	Kind    string   `json:"kind"` // "key" | "session"
}

func (i Identity) AllowsLedger(name string) bool {
	for _, l := range i.Ledgers {
		if l == "*" || l == name {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run** `go test ./auth/` — PASS
- [ ] **Step 5: Commit** `git commit -am "feat(auth): identity types and role ranking"`

---

### Task 3: Auth store (migrations + CRUD, both flavors)

**Files:** Create `auth/migration/v001.sql`, `auth/store.go`, Test `auth/store_test.go`

- [ ] **Step 1: Migration file** `auth/migration/v001.sql`

```sql
--statement
CREATE TABLE IF NOT EXISTS auth_users (
  "id"            integer,
  "username"      varchar,
  "password_hash" varchar,
  "role"          varchar,
  "created_at"    varchar,
  "disabled_at"   varchar,
  UNIQUE("id"),
  UNIQUE("username")
);
--statement
CREATE TABLE IF NOT EXISTS auth_keys (
  "id"         integer,
  "key_hash"   varchar,
  "label"      varchar,
  "role"       varchar,
  "ledgers"    varchar,
  "created_at" varchar,
  "revoked_at" varchar,
  UNIQUE("id"),
  UNIQUE("key_hash")
);
--statement
CREATE TABLE IF NOT EXISTS auth_sessions (
  "id"         integer,
  "user_id"    integer,
  "token_hash" varchar,
  "expires_at" varchar,
  "created_at" varchar,
  "revoked_at" varchar,
  UNIQUE("id"),
  UNIQUE("token_hash")
);
```

(Portable SQL: identical on Postgres; the store prefixes the schema there.)

- [ ] **Step 2: Failing store test** (sqlite, temp dir)

```go
package auth

import (
	"os"
	"path"
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
		if err := s.CreateUser(u); err == nil {
			t.Fatal("duplicate username must fail")
		}
		if _, err := s.GetUserByUsername("nobody"); err != ErrNoRow {
			t.Fatalf("expected ErrNoRow, got %v", err)
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
		if err != nil || got.Label != "fintech" {
			t.Fatalf("got %+v err %v", got, err)
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
		if err := s.RevokeSession("th", "2026-06-10T01:00:00Z"); err != nil {
			t.Fatal(err)
		}
		got, _ = s.GetSessionByHash("th")
		if got.RevokedAt == "" {
			t.Fatal("expected revoked")
		}
	})
}
```

- [ ] **Step 3: Run** `go test ./auth/` — FAIL (OpenStore undefined)
- [ ] **Step 4: Implement `auth/store.go`** — `database/sql` with both drivers; IDs assigned by `count+1` like the rest of the codebase:

```go
package auth

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

//go:embed migration
var migrations embed.FS

var ErrNoRow = errors.New("auth: not found")

type Store struct {
	db     *sql.DB
	flavor sqlbuilder.Flavor
	schema string // "" for sqlite, "corren_auth" for postgres
}

func OpenStore() (*Store, error) {
	switch viper.GetString("storage.driver") {
	case "postgres":
		db, err := sql.Open("pgx", viper.GetString("storage.postgres.conn_string"))
		if err != nil {
			return nil, err
		}
		return &Store{db: db, flavor: sqlbuilder.PostgreSQL, schema: "corren_auth"}, nil
	default:
		dbpath := fmt.Sprintf("file:%s?_journal=WAL",
			path.Join(viper.GetString("storage.dir"), "corren_auth.db"))
		db, err := sql.Open("sqlite3", dbpath)
		if err != nil {
			return nil, err
		}
		return &Store{db: db, flavor: sqlbuilder.SQLite}, nil
	}
}

func (s *Store) table(name string) string {
	if s.schema != "" {
		return fmt.Sprintf("%q.%q", s.schema, name)
	}
	return name
}

func (s *Store) Initialize() error {
	if s.schema != "" {
		if _, err := s.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q", s.schema)); err != nil {
			return err
		}
	}
	b, err := migrations.ReadFile("migration/v001.sql")
	if err != nil {
		return err
	}
	plain := string(b)
	if s.schema != "" {
		for _, tbl := range []string{"auth_users", "auth_keys", "auth_sessions"} {
			plain = strings.ReplaceAll(plain, tbl, fmt.Sprintf("%q.%q", s.schema, tbl))
		}
	}
	for i, stmt := range strings.Split(plain, "--statement") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("auth migration statement %d: %w", i, err)
		}
	}
	return nil
}

func (s *Store) Close() { s.db.Close() }

func (s *Store) nextID(tbl string) (int64, error) {
	var n int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM " + s.table(tbl)).Scan(&n)
	return n + 1, err
}
```

then CRUD with go-sqlbuilder using `s.flavor` and `s.table(...)`:
`CreateUser`, `GetUserByUsername` (maps `sql.ErrNoRows`→`ErrNoRow`),
`CreateKey`, `GetKeyByHash`, `ListKeys`, `RevokeKey(id, ts)`,
`CreateSession`, `GetSessionByHash`, `RevokeSession(tokenHash, ts)` —
nullable columns (`disabled_at`, `revoked_at`) scanned via `sql.NullString`,
inserted as `""`.

- [ ] **Step 5: Run** `go test ./auth/` — PASS. **Commit** `git add auth/ && git commit -m "feat(auth): credential store with embedded migrations (sqlite+postgres)"`

---

### Task 4: Service (bcrypt login, sessions, key issuance)

**Files:** Create `auth/service.go`, Test `auth/service_test.go`

- [ ] **Step 1: Failing test**

```go
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
		token, _, err := svc.Login("alice", "s3cret")
		if err != nil {
			t.Fatal(err)
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
		svc.SessionTTL = -time.Minute // already expired
		svc.CreateUser("carol", "pw", RoleReadonly)
		token, _, _ := svc.Login("carol", "pw")
		if _, err := svc.Authenticate(token); err == nil {
			t.Fatal("expired session must fail")
		}
	})
}

func TestAPIKeyFlow(t *testing.T) {
	withService(t, func(svc *Service) {
		plain, key, err := svc.CreateKey("fintech", RoleOperator, []string{"demo"})
		if err != nil {
			t.Fatal(err)
		}
		id, err := svc.Authenticate(plain)
		if err != nil || id.Kind != "key" || id.Role != RoleOperator || !id.AllowsLedger("demo") || id.AllowsLedger("prod") {
			t.Fatalf("id %+v err %v", id, err)
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
		if _, _, err := svc.CreateKey("x", "bogusrole", nil); err == nil {
			t.Fatal("invalid role must fail")
		}
	})
}
```

- [ ] **Step 2: Run** — FAIL
- [ ] **Step 3: Implement `auth/service.go`**

```go
package auth

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrUnauthorized = errors.New("auth: invalid credentials")

type Service struct {
	store      *Store
	SessionTTL time.Duration
}

func NewService(s *Store) *Service { return &Service{store: s, SessionTTL: 12 * time.Hour} }

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (svc *Service) CreateUser(username, password, role string) (User, error) {
	if !RoleIsValid(role) {
		return User{}, errors.New("auth: invalid role")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	u := User{Username: username, PasswordHash: string(hash), Role: role, CreatedAt: now()}
	if err := svc.store.CreateUser(u); err != nil {
		return User{}, err
	}
	return svc.store.GetUserByUsername(username)
}

func (svc *Service) Login(username, password string) (token string, expiresAt string, err error) {
	u, err := svc.store.GetUserByUsername(username)
	if err != nil || u.DisabledAt != "" {
		return "", "", ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", "", ErrUnauthorized
	}
	token = GenerateToken(PrefixSession)
	expiresAt = time.Now().UTC().Add(svc.SessionTTL).Format(time.RFC3339)
	err = svc.store.CreateSession(Session{
		UserID: u.ID, TokenHash: HashToken(token), ExpiresAt: expiresAt, CreatedAt: now(),
	})
	return token, expiresAt, err
}

func (svc *Service) Logout(token string) error {
	return svc.store.RevokeSession(HashToken(token), now())
}

func (svc *Service) CreateKey(label, role string, ledgers []string) (string, APIKey, error) {
	if !RoleIsValid(role) {
		return "", APIKey{}, errors.New("auth: invalid role")
	}
	if len(ledgers) == 0 {
		ledgers = []string{"*"}
	}
	plain := GenerateToken(PrefixKey)
	k := APIKey{KeyHash: HashToken(plain), Label: label, Role: role,
		Ledgers: strings.Join(ledgers, ","), CreatedAt: now()}
	if err := svc.store.CreateKey(k); err != nil {
		return "", APIKey{}, err
	}
	saved, err := svc.store.GetKeyByHash(k.KeyHash)
	return plain, saved, err
}

func (svc *Service) RevokeKey(id int64) error { return svc.store.RevokeKey(id, now()) }

func (svc *Service) Authenticate(token string) (Identity, error) {
	switch {
	case strings.HasPrefix(token, PrefixKey):
		k, err := svc.store.GetKeyByHash(HashToken(token))
		if err != nil || k.RevokedAt != "" {
			return Identity{}, ErrUnauthorized
		}
		return Identity{Subject: k.Label, Role: k.Role,
			Ledgers: strings.Split(k.Ledgers, ","), Kind: "key"}, nil
	case strings.HasPrefix(token, PrefixSession):
		s, err := svc.store.GetSessionByHash(HashToken(token))
		if err != nil || s.RevokedAt != "" || s.ExpiresAt <= now() {
			return Identity{}, ErrUnauthorized
		}
		u, err := svc.store.GetUserByID(s.UserID)
		if err != nil || u.DisabledAt != "" {
			return Identity{}, ErrUnauthorized
		}
		return Identity{Subject: u.Username, Role: u.Role,
			Ledgers: []string{"*"}, Kind: "session"}, nil
	}
	return Identity{}, ErrUnauthorized
}
```

(Requires adding `GetUserByID` to the store + test. RFC3339 strings compare lexicographically — same trick as the scheduler.)

- [ ] **Step 4: Run** `go test ./auth/` — PASS. `go build ./...` (x/crypto moves from indirect to direct in go.mod — already in go.sum, no download).
- [ ] **Step 5: Commit** `git commit -am "feat(auth): service with bcrypt login, revocable sessions and api keys"`

---

### Task 5: Middleware enforcement

**Files:** Modify `api/middlewares/auth_middleware.go`, Test `api/middlewares/auth_middleware_test.go`

- [ ] **Step 1: Failing test** — httptest matrix:

```go
package middlewares

// setup: viper auth.enabled=true, temp sqlite store, service with:
//   admin key, operator key scoped to "demo", readonly key, expired session
// router: gin.New() + AuthMiddleware + routes:
//   GET  /:ledger/contracts        -> 200 stub
//   POST /:ledger/contracts        -> 200 stub
//   GET  /auth/admin/keys          -> 200 stub
// matrix assertions:
//   no header                       -> 401
//   garbage token                   -> 401
//   readonly GET  /demo/contracts   -> 200
//   readonly POST /demo/contracts   -> 403
//   operator POST /demo/contracts   -> 200
//   operator POST /prod/contracts   -> 403 (ledger scope)
//   operator GET  /auth/admin/keys  -> 403
//   admin    GET  /auth/admin/keys  -> 200
//   POST /auth/login                -> reaches handler without auth (exempt)
// plus: auth.enabled=false -> everything passes with no header (non-regression)
```

(Write it as real Go table-driven code following `contract_controller_test.go` patterns: build requests with `httptest.NewRecorder`, assert codes.)

- [ ] **Step 2: Run** `go test ./api/middlewares/` — FAIL
- [ ] **Step 3: Implement** — replace the body of `AuthMiddleware`:

```go
package middlewares

import (
	"net/http"
	"strings"

	"github.com/amezianechayer/corren/auth"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type AuthMiddleware struct {
	service *auth.Service
}

func NewAuthMiddleware(service *auth.Service) AuthMiddleware {
	return AuthMiddleware{service: service}
}

func (m AuthMiddleware) AuthMiddleware(engine *gin.Engine) gin.HandlerFunc {
	// kept for backward compat: legacy basic_auth still honored when set
	if basic := viper.Get("server.http.basic_auth"); basic != nil {
		segment := strings.Split(basic.(string), ":")
		engine.Use(gin.BasicAuth(gin.Accounts{segment[0]: segment[1]}))
	}

	return func(c *gin.Context) {
		if !viper.GetBool("auth.enabled") {
			return // pilot off: today's behavior, untouched
		}
		p := c.FullPath()
		if p == "/auth/login" || p == "/healthz" {
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abort(c, http.StatusUnauthorized, "missing bearer token")
			return
		}
		id, err := m.service.Authenticate(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			abort(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		need := auth.RoleReadonly
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			need = auth.RoleOperator
		}
		if strings.HasPrefix(p, "/auth/admin") {
			need = auth.RoleAdmin
		}
		if !auth.RoleAtLeast(id.Role, need) {
			abort(c, http.StatusForbidden, "role "+id.Role+" cannot access this resource")
			return
		}
		if ledger := c.Param("ledger"); ledger != "" && !id.AllowsLedger(ledger) {
			abort(c, http.StatusForbidden, "key not scoped to ledger "+ledger)
			return
		}
		c.Set("identity", id)
	}
}

func abort(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"ok": false, "error": true,
		"error_code": status, "error_message": msg})
}
```

- [ ] **Step 4: Run** `go test ./api/...` — PASS (existing controller tests must stay green: they don't set `auth.enabled`)
- [ ] **Step 5: Commit** `git commit -am "feat(auth): role and ledger-scope enforcement middleware"`

---

### Task 6: Auth HTTP endpoints

**Files:** Create `api/actions/auth_controller.go`, Test `api/actions/auth_controller_test.go`; Modify `api/routes/routes.go`, `api/actions/controllers.go`, `api/api.go`

- [ ] **Step 1: Failing test** — httptest flow (pattern of `contract_controller_test.go`): login wrong password → 401; login ok → token; `GET /auth/me` with token → username+role; admin creates key (`POST /auth/admin/keys {"label","role","ledgers"}`) → 201 with `key` shown once; `GET /auth/admin/keys` → list without hashes; `DELETE /auth/admin/keys/:id` → revoked; logout → 200 then `me` → 401.
- [ ] **Step 2: Run** — FAIL
- [ ] **Step 3: Implement controller**

```go
package actions

// AuthController holds *auth.Service.
// POST /auth/login   {username,password} -> 200 {token, expires_at, role} | 401
// POST /auth/logout  (Bearer)            -> 200
// GET  /auth/me      (Bearer)            -> 200 {subject, role, kind, ledgers}
//   (me/logout re-authenticate the bearer directly via the service so they
//    work even though the middleware only sets context when auth.enabled)
// POST   /auth/admin/keys      {label, role, ledgers[]} -> 201 {key, id, ...}
// GET    /auth/admin/keys      -> 200 list
// DELETE /auth/admin/keys/:id  -> 200
// POST   /auth/admin/users     {username, password, role} -> 201
// GET    /auth/admin/users     -> 200 list (no hashes)
```

Routes in `routes.go` (outside the `/:ledger` group):

```go
authGroup := engine.Group("/auth")
{
	authGroup.POST("/login", r.authController.Login)
	authGroup.POST("/logout", r.authController.Logout)
	authGroup.GET("/me", r.authController.Me)
	authGroup.POST("/admin/keys", r.authController.CreateKey)
	authGroup.GET("/admin/keys", r.authController.ListKeys)
	authGroup.DELETE("/admin/keys/:id", r.authController.RevokeKey)
	authGroup.POST("/admin/users", r.authController.CreateUser)
	authGroup.GET("/admin/users", r.authController.ListUsers)
}
```

fx wiring: `api/api.go` Module gains `fx.Provide(auth.NewServiceFromConfig)` where

```go
// auth/service.go
func NewServiceFromConfig() (*Service, error) {
	store, err := OpenStore()
	if err != nil {
		return nil, err
	}
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	svc := NewService(store)
	if ttl := viper.GetDuration("auth.session_ttl"); ttl > 0 {
		svc.SessionTTL = ttl
	}
	return svc, nil
}
```

plus `ListUsers` in the store (no password hashes in JSON — `json:"-"` already handles it).

- [ ] **Step 4: Run** `go test ./api/...` — PASS
- [ ] **Step 5: Commit** `git commit -am "feat(auth): login/logout/me and admin key/user endpoints"`

---

### Task 7: Bootstrap CLI `corren auth init`

**Files:** Modify `cmd/root.go`

- [ ] **Step 1: Implement** (CLI = manual verification, no unit test; logic already covered by service tests)

```go
authCmd := &cobra.Command{Use: "auth"}
authCmd.AddCommand(&cobra.Command{
	Use:   "init",
	Short: "Create the first admin user and api key (printed once)",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		svc, err := auth.NewServiceFromConfig()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		password := auth.GenerateToken("")[:16]
		if _, err := svc.CreateUser("admin", password, auth.RoleAdmin); err != nil {
			fmt.Println("error (already initialized?):", err)
			os.Exit(1)
		}
		key, _, err := svc.CreateKey("bootstrap-admin", auth.RoleAdmin, nil)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("admin user created — shown ONCE, store these now:")
		fmt.Printf("  username: admin\n  password: %s\n  api key:  %s\n", password, key)
		fmt.Println("enable auth with: corren config set auth.enabled true (or in corren.yaml)")
	},
})
root.AddCommand(authCmd)
```

- [ ] **Step 2: Manual check**

```bash
go build -o /tmp/corren-auth . && CORREN_STORAGE_DIR=$(mktemp -d) /tmp/corren-auth auth init
```

Expected: prints username/password/key; second run errors ("already initialized").

- [ ] **Step 3: Commit** `git commit -am "feat(auth): corren auth init bootstrap command"`

---

### Task 8: End-to-end + non-regression

- [ ] **Step 1:** Add `auth.enabled=true` e2e test in `api/actions/auth_controller_test.go`: full server flow — readonly key gets 403 on `POST .../transitions/sell`, operator key succeeds on contract creation. (Real assertion of the role matrix against the *actual* contract routes, not stubs.)
- [ ] **Step 2:** `go test ./... -count=1` — everything green.
- [ ] **Step 3:** Live smoke: start server with `CORREN_AUTH_ENABLED=true`, run `auth init`, verify `curl` without token → 401, with operator key → contract created, with readonly key on sell → 403. Verify `demo/murabaha_demo.sh` still works with auth disabled.
- [ ] **Step 4:** Commit + update `sharia/README.md` API section with one line pointing to auth (Bearer).

---

## Self-review notes

- Spec coverage: storage dedicated DB ✓ (T3), 3 roles ✓ (T2/T5), two access modes ✓ (T4/T6), bootstrap ✓ (T7), `auth.enabled` compat ✓ (T5/T8), endpoints ✓ (T6), tests incl. readonly-cannot-sell ✓ (T8).
- `GetUserByID` added in T4 step 3 note + store test required in T3 extension — included in T4 step 4 run.
- Phases 2–4 of the spec are intentionally separate plans.
