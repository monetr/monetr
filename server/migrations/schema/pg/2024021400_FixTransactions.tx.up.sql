-- Remove corrupted transactions.
DELETE FROM "transactions" WHERE "transaction_id" IN (
  SELECT 
    "transactions"."transaction_id"
  FROM "transactions"
  INNER JOIN "bank_accounts" ON "bank_accounts"."bank_account_id" = "transactions"."bank_account_id" AND 
                                "bank_accounts"."account_id" = "transactions"."account_id"
  INNER JOIN "links" ON "links"."link_id" = "bank_accounts"."link_id" AND 
             "links"."account_id" = "bank_accounts"."account_id" 
  WHERE "links"."teller_link_id" IS NOT NULL AND 
        "transactions"."teller_transaction_id" IS NULL
);

DELETE FROM "teller_syncs" WHERE "timestamp" > '2024-02-06' AND "trigger" != 'initial';
DELETE FROM "plaid_syncs" WHERE "timestamp" > '2024-02-06';
