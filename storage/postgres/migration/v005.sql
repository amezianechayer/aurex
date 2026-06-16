--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".wallets (
  "id"         varchar,
  "owner"      varchar,
  "asset"      varchar,
  "created_at" varchar,
  "updated_at" varchar,
  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".wallet_holds (
  "id"         varchar,
  "wallet_id"  varchar,
  "asset"      varchar,
  "amount"     bigint,
  "status"     varchar,
  "reason"     varchar,
  "expires_at" varchar,
  "created_at" varchar,
  "updated_at" varchar,
  UNIQUE("id")
);
--statement
CREATE INDEX IF NOT EXISTS wallet_holds_wallet ON "VAR_LEDGER_NAME".wallet_holds ("wallet_id");
