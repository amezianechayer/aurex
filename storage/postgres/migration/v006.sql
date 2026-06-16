--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".payments (
  "id"          varchar,
  "psp"         varchar,
  "direction"   varchar,
  "wallet_id"   varchar,
  "asset"       varchar,
  "amount"      bigint,
  "status"      varchar,
  "reference"   varchar,
  "external_id" varchar,
  "created_at"  varchar,
  "updated_at"  varchar,
  UNIQUE("id")
);
--statement
CREATE INDEX IF NOT EXISTS payments_wallet ON "VAR_LEDGER_NAME".payments ("wallet_id");
