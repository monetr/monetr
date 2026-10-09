ALTER TABLE "transaction_recurring" ADD COLUMN "auto_assign" BOOLEAN NOT NULL DEFAULT false;

-- Auto assign spends new charges from the linked spending, so it can't be on
-- without one. For now that means it is only for expenses, this will need to
-- change once funding schedules can be auto assigned too. Deleting spending
-- has to turn this off before the FK nulls out the spending ID.
ALTER TABLE "transaction_recurring" ADD CONSTRAINT "ck_transaction_recurring_auto_assign_spending"
CHECK ("auto_assign" = false OR "spending_id" IS NOT NULL);
