-- Resetting teller again after bugfix.
DELETE FROM "transactions" WHERE "teller_transaction_id" IS NOT NULL;
DELETE FROM "bank_accounts" WHERE "teller_bank_account_id" IS NOT NULL;
DELETE FROM "teller_syncs";
DELETE FROM "teller_transactions";
DELETE FROM "teller_bank_accounts";
