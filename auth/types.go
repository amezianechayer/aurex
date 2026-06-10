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
	Ledgers   string `json:"ledgers"` // CSV of ledger names, or "*"
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

// Identity is what the middleware puts in the gin context after a
// successful authentication.
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
