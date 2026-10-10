-- User defined ordering of the bank accounts within a link in the UI, stored as
-- a list of bank account IDs.
ALTER TABLE "links" ADD COLUMN "bank_account_order" VARCHAR(32)[];
