-- Resetting teller again after bugfix.
DELETE FROM "transactions" WHERE "teller_transaction_id" IS NOT NULL;
DELETE FROM "transaction_clusters"
       USING "bank_accounts"
       WHERE "bank_accounts"."account_id" = "transaction_clusters"."account_id" AND
             "bank_accounts"."bank_account_id" = "transaction_clusters"."bank_account_id" AND
             "bank_accounts"."teller_bank_account_id" IS NOT NULL;
DELETE FROM "bank_accounts" WHERE "teller_bank_account_id" IS NOT NULL;
DELETE FROM "teller_syncs";
DELETE FROM "teller_transactions";
DELETE FROM "teller_bank_accounts";
