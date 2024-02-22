UPDATE "transactions" AS "transactions"
SET "amount" = "transactions"."amount" * -1
FROM "teller_transactions" 
INNER JOIN "teller_bank_accounts" ON "teller_transactions"."teller_bank_account_id" = "teller_bank_accounts"."teller_bank_account_id" AND "teller_transactions"."account_id" = "teller_bank_accounts"."account_id"
WHERE "transactions"."teller_transaction_id" = "teller_transactions"."teller_transaction_id" AND 
      "transactions"."account_id" = "teller_transactions"."account_id" AND
      "teller_bank_accounts"."type" = 'credit' AND "teller_bank_accounts"."sub_type" = 'credit_card';

UPDATE "teller_transactions" AS "teller_transactions"
SET "amount" = "teller_transactions"."amount" * -1
FROM "teller_bank_accounts" 
WHERE "teller_transactions"."teller_bank_account_id" = "teller_bank_accounts"."teller_bank_account_id" AND "teller_transactions"."account_id" = "teller_bank_accounts"."account_id" AND
      "teller_bank_accounts"."type" = 'credit' AND "teller_bank_accounts"."sub_type" = 'credit_card';
