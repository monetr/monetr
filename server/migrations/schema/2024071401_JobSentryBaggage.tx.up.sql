ALTER TABLE "jobs" RENAME COLUMN "trace_id" TO "sentry_trace_id";
ALTER TABLE "jobs" ADD COLUMN "sentry_baggage" TEXT;
