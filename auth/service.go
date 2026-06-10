package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

var ErrUnauthorized = errors.New("auth: invalid credentials")

type Service struct {
	store      *Store
	SessionTTL time.Duration
}

func NewService(s *Store) *Service {
	return &Service{store: s, SessionTTL: 12 * time.Hour}
}

// NewServiceFromConfig opens and initializes the auth store using viper
// configuration. Used by fx and the CLI.
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

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

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

func (svc *Service) ListUsers() ([]User, error) { return svc.store.ListUsers() }

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
		UserID:    u.ID,
		TokenHash: HashToken(token),
		ExpiresAt: expiresAt,
		CreatedAt: now(),
	})
	if err != nil {
		return "", "", err
	}
	return token, expiresAt, nil
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
	k := APIKey{
		KeyHash:   HashToken(plain),
		Label:     label,
		Role:      role,
		Ledgers:   strings.Join(ledgers, ","),
		CreatedAt: now(),
	}
	if err := svc.store.CreateKey(k); err != nil {
		return "", APIKey{}, err
	}
	saved, err := svc.store.GetKeyByHash(k.KeyHash)
	return plain, saved, err
}

func (svc *Service) RevokeKey(id int64) error { return svc.store.RevokeKey(id, now()) }

func (svc *Service) ListKeys() ([]APIKey, error) { return svc.store.ListKeys() }

// Authenticate resolves a bearer token (api key or session) into an
// Identity. RFC3339 strings compare lexicographically, hence the string
// comparison for expiry — same convention as the rest of the codebase.
func (svc *Service) Authenticate(token string) (Identity, error) {
	switch {
	case strings.HasPrefix(token, PrefixKey):
		k, err := svc.store.GetKeyByHash(HashToken(token))
		if err != nil || k.RevokedAt != "" {
			return Identity{}, ErrUnauthorized
		}
		return Identity{
			Subject: k.Label,
			Role:    k.Role,
			Ledgers: strings.Split(k.Ledgers, ","),
			Kind:    "key",
		}, nil

	case strings.HasPrefix(token, PrefixSession):
		s, err := svc.store.GetSessionByHash(HashToken(token))
		if err != nil || s.RevokedAt != "" || s.ExpiresAt <= now() {
			return Identity{}, ErrUnauthorized
		}
		u, err := svc.store.GetUserByID(s.UserID)
		if err != nil || u.DisabledAt != "" {
			return Identity{}, ErrUnauthorized
		}
		return Identity{
			Subject: u.Username,
			Role:    u.Role,
			Ledgers: []string{"*"},
			Kind:    "session",
		}, nil
	}
	return Identity{}, ErrUnauthorized
}
