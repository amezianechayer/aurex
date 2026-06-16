--statement
CREATE TABLE IF NOT EXISTS payments (
  "id"          varchar,
  "psp"         varchar,
  "direction"   varchar,
  "wallet_id"   varchar,
  "asset"       varchar,
  "amount"      integer,
  "status"      varchar,
  "reference"   varchar,
  "external_id" varchar,
  "created_at"  varchar,
  "updated_at"  varchar,
  UNIQUE("id")
);
--statement
CREATE INDEX IF NOT EXISTS 'payments_wallet' ON "payments" ("wallet_id");
