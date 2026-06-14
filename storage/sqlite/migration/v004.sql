--statement
CREATE TABLE IF NOT EXISTS guard_rules (
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
CREATE TABLE IF NOT EXISTS guard_events (
  "id"           integer,
  "rule_id"      varchar,
  "action"       varchar,
  "reason"       varchar,
  "standard_ref" varchar,
  "tx_reference" varchar,
  "payload"      varchar,
  "created_at"   varchar
);
--statement
CREATE INDEX IF NOT EXISTS 'guard_events_created' ON "guard_events" ("created_at");
