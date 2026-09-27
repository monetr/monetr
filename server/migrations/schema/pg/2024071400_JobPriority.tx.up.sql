ALTER TABLE "jobs" ADD COLUMN "priority" BIGINT DEFAULT extract(epoch from now() at time zone 'utc')::integer;

UPDATE "jobs" 
SET "priority" = extract(epoch from "created_at" at time zone 'utc')::integer
WHERE "priority" IS NULL;

ALTER TABLE "jobs" ALTER COLUMN "priority" SET NOT NULL;
