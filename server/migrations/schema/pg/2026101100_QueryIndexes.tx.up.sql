-- Completed jobs stick around for 15 days, without this every dequeue has to
-- scan all of them just to find the few that are still pending.
CREATE INDEX "ix_jobs_pending"
ON "jobs" ("job_id")
WHERE "status" = 'pending';

CREATE INDEX "ix_jobs_processing"
ON "jobs" ("started_at")
WHERE "status" = 'processing';

-- The primary keys on these tables lead with their own id, so lookups by
-- account were scanning every tenant's rows.
CREATE INDEX "ix_bank_accounts_link" ON "bank_accounts" ("account_id", "link_id");

CREATE INDEX "ix_links_account" ON "links" ("account_id");

CREATE INDEX "ix_links_plaid_link"
ON "links" ("plaid_link_id")
WHERE "plaid_link_id" IS NOT NULL;

CREATE INDEX "ix_plaid_bank_accounts_plaid_link" ON "plaid_bank_accounts" ("account_id", "plaid_link_id");
