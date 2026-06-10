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
