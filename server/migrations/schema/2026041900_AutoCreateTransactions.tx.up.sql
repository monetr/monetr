ALTER TABLE "spending" ADD COLUMN "auto_create_transaction" BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE "funding_schedules" ADD COLUMN "auto_create_transaction" BOOLEAN NOT NULL DEFAULT false;
