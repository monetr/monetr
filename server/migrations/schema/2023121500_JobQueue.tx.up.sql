DROP TABLE IF EXISTS "jobs";
DROP TABLE IF EXISTS "cron_jobs";

CREATE TABLE "jobs" (
  job_id       BIGSERIAL,
  queue        TEXT NOT NULL,
  signature    TEXT NOT NULL,
  input        BYTEA,
  output       BYTEA,
  status       TEXT NOT NULL,
  created_at   TIMESTAMP WITH TIME ZONE NOT NULL,
  updated_at   TIMESTAMP WITH TIME ZONE NOT NULL,
  started_at   TIMESTAMP WITH TIME ZONE,
  completed_at TIMESTAMP WITH TIME ZONE,

  CONSTRAINT pk_jobs PRIMARY KEY (job_id),
  CONSTRAINT uq_jobs_signature UNIQUE (signature)
);

CREATE TABLE "cron_jobs" (
  queue         TEXT NOT NULL,
  cron_schedule TEXT NOT NULL,
  last_run_at   TIMESTAMP WITH TIME ZONE,
  next_run_at   TIMESTAMP WITH TIME ZONE NOT NULL,

  CONSTRAINT pk_cron_jobs PRIMARY KEY (queue)
);

