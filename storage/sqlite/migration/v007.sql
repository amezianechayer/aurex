--statement
CREATE TABLE IF NOT EXISTS flows (
  "id"           varchar,
  "name"         varchar,
  "trigger_type" varchar,
  "schedule"     varchar,
  "steps"        varchar,
  "last_run_at"  varchar,
  "created_at"   varchar,
  "updated_at"   varchar,
  UNIQUE("id")
);
--statement
CREATE TABLE IF NOT EXISTS flow_instances (
  "id"           varchar,
  "flow_id"      varchar,
  "status"       varchar,
  "current_step" integer,
  "input"        varchar,
  "resume_at"    varchar,
  "error"        varchar,
  "created_at"   varchar,
  "updated_at"   varchar,
  UNIQUE("id")
);
--statement
CREATE INDEX IF NOT EXISTS 'flow_instances_flow' ON "flow_instances" ("flow_id");
--statement
CREATE INDEX IF NOT EXISTS 'flow_instances_status' ON "flow_instances" ("status");
