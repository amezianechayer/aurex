--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".sharia_contracts (
  "id"               varchar,
  "type"             varchar,
  "state"            varchar,
  "params"           varchar,
  "template_version" varchar,
  "created_at"       varchar,
  "updated_at"       varchar,

  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".sharia_schedule (
  "contract_id"    varchar,
  "seq"            integer,
  "due_date"       varchar,
  "amount"         bigint,
  "principal_part" bigint,
  "profit_part"    bigint,
  "status"         varchar,
  "paid_tx_id"     bigint,
  "paid_at"        varchar,

  UNIQUE("contract_id","seq")
);
--statement
CREATE TABLE IF NOT EXISTS "VAR_LEDGER_NAME".sharia_audit (
  "contract_id"  varchar,
  "seq"          integer,
  "event"        varchar,
  "transition"   varchar,
  "decision"     varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "tx_id"        bigint,
  "payload"      varchar,
  "prev_hash"    varchar,
  "hash"         varchar,
  "created_at"   varchar,

  UNIQUE("contract_id","seq")
);
--statement
CREATE INDEX IF NOT EXISTS sh_sched_due ON "VAR_LEDGER_NAME".sharia_schedule (
  "status",
  "due_date"
);
