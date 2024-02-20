ALTER TABLE "jobs" 
ADD COLUMN "attempt" INT,
ADD COLUMN "timestamp" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "previous_job_id" BIGINT;

UPDATE "jobs" SET
  "attempt" = 1,
  "timestamp" = "created_at";

ALTER TABLE "jobs"
ALTER COLUMN "attempt" SET NOT NULL,
ALTER COLUMN "timestamp" SET NOT NULL;

CREATE INDEX "ix_jobs_timestamped" 
ON "jobs" ("timestamp", "status", "queue") 
WHERE "status" = 'pending';
