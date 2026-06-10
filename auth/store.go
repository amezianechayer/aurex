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

// Store persists credentials in a dedicated database, separate from the
// per-ledger stores: authentication is global.
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

func nullable(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

// --- users ---

func (s *Store) CreateUser(u User) error {
	id, err := s.nextID("auth_users")
	if err != nil {
		return err
	}
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("auth_users"))
	ib.Cols("id", "username", "password_hash", "role", "created_at", "disabled_at")
	ib.Values(id, u.Username, u.PasswordHash, u.Role, u.CreatedAt, nullable(u.DisabledAt))
	q, args := ib.BuildWithFlavor(s.flavor)
	_, err = s.db.Exec(q, args...)
	return err
}

func (s *Store) scanUser(row *sql.Row) (User, error) {
	var u User
	var disabled sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &disabled)
	if err == sql.ErrNoRows {
		return u, ErrNoRow
	}
	u.DisabledAt = disabled.String
	return u, err
}

func (s *Store) GetUserByUsername(username string) (User, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "username", "password_hash", "role", "created_at", "disabled_at")
	sb.From(s.table("auth_users"))
	sb.Where(sb.Equal("username", username))
	q, args := sb.BuildWithFlavor(s.flavor)
	return s.scanUser(s.db.QueryRow(q, args...))
}

func (s *Store) GetUserByID(id int64) (User, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "username", "password_hash", "role", "created_at", "disabled_at")
	sb.From(s.table("auth_users"))
	sb.Where(sb.Equal("id", id))
	q, args := sb.BuildWithFlavor(s.flavor)
	return s.scanUser(s.db.QueryRow(q, args...))
}

func (s *Store) ListUsers() ([]User, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "username", "password_hash", "role", "created_at", "disabled_at")
	sb.From(s.table("auth_users"))
	sb.OrderBy("id").Asc()
	q, args := sb.BuildWithFlavor(s.flavor)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		var disabled sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &disabled); err != nil {
			return nil, err
		}
		u.DisabledAt = disabled.String
		users = append(users, u)
	}
	return users, nil
}

// --- api keys ---

func (s *Store) CreateKey(k APIKey) error {
	id, err := s.nextID("auth_keys")
	if err != nil {
		return err
	}
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("auth_keys"))
	ib.Cols("id", "key_hash", "label", "role", "ledgers", "created_at", "revoked_at")
	ib.Values(id, k.KeyHash, k.Label, k.Role, k.Ledgers, k.CreatedAt, nullable(k.RevokedAt))
	q, args := ib.BuildWithFlavor(s.flavor)
	_, err = s.db.Exec(q, args...)
	return err
}

func (s *Store) GetKeyByHash(hash string) (APIKey, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "key_hash", "label", "role", "ledgers", "created_at", "revoked_at")
	sb.From(s.table("auth_keys"))
	sb.Where(sb.Equal("key_hash", hash))
	q, args := sb.BuildWithFlavor(s.flavor)

	var k APIKey
	var revoked sql.NullString
	err := s.db.QueryRow(q, args...).Scan(&k.ID, &k.KeyHash, &k.Label, &k.Role, &k.Ledgers, &k.CreatedAt, &revoked)
	if err == sql.ErrNoRows {
		return k, ErrNoRow
	}
	k.RevokedAt = revoked.String
	return k, err
}

func (s *Store) ListKeys() ([]APIKey, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "key_hash", "label", "role", "ledgers", "created_at", "revoked_at")
	sb.From(s.table("auth_keys"))
	sb.OrderBy("id").Asc()
	q, args := sb.BuildWithFlavor(s.flavor)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []APIKey{}
	for rows.Next() {
		var k APIKey
		var revoked sql.NullString
		if err := rows.Scan(&k.ID, &k.KeyHash, &k.Label, &k.Role, &k.Ledgers, &k.CreatedAt, &revoked); err != nil {
			return nil, err
		}
		k.RevokedAt = revoked.String
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *Store) RevokeKey(id int64, ts string) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("auth_keys"))
	ub.Set(ub.Assign("revoked_at", ts))
	ub.Where(ub.Equal("id", id))
	q, args := ub.BuildWithFlavor(s.flavor)
	_, err := s.db.Exec(q, args...)
	return err
}

// --- sessions ---

func (s *Store) CreateSession(sess Session) error {
	id, err := s.nextID("auth_sessions")
	if err != nil {
		return err
	}
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("auth_sessions"))
	ib.Cols("id", "user_id", "token_hash", "expires_at", "created_at", "revoked_at")
	ib.Values(id, sess.UserID, sess.TokenHash, sess.ExpiresAt, sess.CreatedAt, nullable(sess.RevokedAt))
	q, args := ib.BuildWithFlavor(s.flavor)
	_, err = s.db.Exec(q, args...)
	return err
}

func (s *Store) GetSessionByHash(hash string) (Session, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "user_id", "token_hash", "expires_at", "created_at", "revoked_at")
	sb.From(s.table("auth_sessions"))
	sb.Where(sb.Equal("token_hash", hash))
	q, args := sb.BuildWithFlavor(s.flavor)

	var sess Session
	var revoked sql.NullString
	err := s.db.QueryRow(q, args...).Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.ExpiresAt, &sess.CreatedAt, &revoked)
	if err == sql.ErrNoRows {
		return sess, ErrNoRow
	}
	sess.RevokedAt = revoked.String
	return sess, err
}

func (s *Store) RevokeSession(tokenHash, ts string) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("auth_sessions"))
	ub.Set(ub.Assign("revoked_at", ts))
	ub.Where(ub.Equal("token_hash", tokenHash))
	q, args := ub.BuildWithFlavor(s.flavor)
	_, err := s.db.Exec(q, args...)
	return err
}
