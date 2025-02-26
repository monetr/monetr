CREATE TABLE IF NOT EXISTS "cron_jobs"
(
    "job_id"      BIGSERIAL   NOT NULL,
    "schedule"     TEXT        NOT NULL,
    "command"      TEXT        NOT NULL,
    "last_run_at"  TIMESTAMPTZ,
    "next_run_at"  TIMESTAMPTZ,
    "is_active"    BOOLEAN     NOT NULL DEFAULT TRUE,
    CONSTRAINT "pk_cron_jobs" PRIMARY KEY ("job_id")
);
