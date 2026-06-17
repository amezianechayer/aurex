--statement
-- Webhook idempotency: a (psp, external_id) pair settles at most once. Partial
-- so the many payout rows that start with an empty external_id don't collide.
CREATE UNIQUE INDEX IF NOT EXISTS payments_psp_extid
  ON payments ("psp", "external_id")
  WHERE "external_id" != '';
