-- Marks recurring transactions that were linked to spending by the matching
-- job rather than by the user, so those links can be told apart (and undone)
-- later. Linking or unlinking by the user clears it.
ALTER TABLE "transaction_recurring" ADD COLUMN "auto_matched" BOOLEAN NOT NULL DEFAULT false;
