ALTER TABLE "jobs" DROP CONSTRAINT "uq_jobs_signature";
ALTER TABLE "jobs" ADD CONSTRAINT "uq_jobs_queue_signature" UNIQUE ("queue", "signature");
