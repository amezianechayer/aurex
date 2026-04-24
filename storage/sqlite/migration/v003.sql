--statement
ALTER TABLE api_keys ADD COLUMN "tier" varchar NOT NULL DEFAULT 'sandbox';
