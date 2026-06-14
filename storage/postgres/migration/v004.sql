--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".guard_rules (
  "id"           varchar,
  "kind"         varchar,
  "params"       varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "enabled"      integer,
  "created_at"   varchar,
  "updated_at"   varchar,
  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".guard_events (
  "id"           bigint,
  "rule_id"      varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "tx_reference" varchar,
  "payload"      varchar,
  "created_at"   varchar
);
--statement
CREATE INDEX IF NOT EXISTS guard_events_created ON "VAR_LEDGER_NAME".guard_events ("created_at");
