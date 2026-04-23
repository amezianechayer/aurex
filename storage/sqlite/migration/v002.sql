--statement
CREATE TABLE IF NOT EXISTS assets (
  "id"           varchar PRIMARY KEY,
  "name"         varchar NOT NULL,
  "precision"    integer NOT NULL DEFAULT 2,
  "category"     varchar NOT NULL,
  "aaoifi_class" varchar,
  "metadata"     text,
  "created_at"   varchar NOT NULL
);
--statement
CREATE TABLE IF NOT EXISTS sharia_contracts (
  "id"         varchar PRIMARY KEY,
  "type"       varchar NOT NULL,
  "status"     varchar NOT NULL DEFAULT 'PENDING',
  "ledger"     varchar NOT NULL,
  "parties"    text NOT NULL,
  "terms"      text NOT NULL,
  "aaoifi_fas" varchar,
  "created_at" varchar NOT NULL,
  "updated_at" varchar NOT NULL
);
--statement
CREATE INDEX IF NOT EXISTS 'scon_i0' ON "sharia_contracts" ("status");
--statement
CREATE INDEX IF NOT EXISTS 'scon_i1' ON "sharia_contracts" ("ledger");
--statement
CREATE TABLE IF NOT EXISTS sharia_certificates (
  "id"          varchar PRIMARY KEY,
  "contract_id" varchar,
  "txid"        integer,
  "constraint"  varchar NOT NULL,
  "result"      varchar NOT NULL,
  "issued_at"   varchar NOT NULL,
  "authority"   varchar NOT NULL DEFAULT 'CORREN'
);
--statement
CREATE INDEX IF NOT EXISTS 'cert_i0' ON "sharia_certificates" ("contract_id");
--statement
CREATE INDEX IF NOT EXISTS 'cert_i1' ON "sharia_certificates" ("txid");
--statement
CREATE TABLE IF NOT EXISTS api_keys (
  "key_hash"   varchar PRIMARY KEY,
  "name"       varchar NOT NULL,
  "role"       varchar NOT NULL DEFAULT 'reader',
  "created_at" varchar NOT NULL,
  "expires_at" varchar
);