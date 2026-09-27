-- Remove stray teller transactions.
DELETE FROM "teller_transactions" WHERE "teller_transaction_id" IN (
  SELECT 
    "teller_transactions"."teller_transaction_id" 
  FROM "teller_transactions"
  LEFT JOIN "transactions" ON "transactions"."teller_transaction_id" = "teller_transactions"."teller_transaction_id" AND
                              "transactions"."account_id" = "teller_transactions"."account_id"
  WHERE "transactions"."transaction_id" IS NULL
);

