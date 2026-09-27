-- Fix orphaned lunch_flow_transactions that became disconnected from their
-- transactions row when the PUT handler failed to preserve the
-- lunch_flow_transaction_id field.
--
-- Reconnect by matching through bank_accounts on original_name/description,
-- amount, and date. ROW_NUMBER on both sides ensures 1:1 pairing when multiple
-- rows share the same matching key, which is required by the unique constraint
-- on transactions.lunch_flow_transaction_id.
--
-- Any orphaned lunch_flow_transactions that cannot be reconnected are deleted
-- to prevent duplicate key errors on future syncs.

WITH orphaned AS (
    SELECT lft."lunch_flow_transaction_id",
           lft."account_id",
           lft."lunch_flow_bank_account_id",
           lft."description",
           lft."amount",
           lft."date",
           ROW_NUMBER() OVER (
               PARTITION BY lft."lunch_flow_bank_account_id",
                            lft."description", lft."amount", lft."date"
               ORDER BY lft."lunch_flow_transaction_id"
           ) AS rn
    FROM "lunch_flow_transactions" lft
    WHERE NOT EXISTS (
        SELECT 1 FROM "transactions" t
        WHERE t."lunch_flow_transaction_id" = lft."lunch_flow_transaction_id"
          AND t."account_id" = lft."account_id"
    )
),
candidates AS (
    SELECT t."transaction_id",
           t."account_id",
           ba."lunch_flow_bank_account_id",
           t."original_name",
           t."amount",
           t."date",
           ROW_NUMBER() OVER (
               PARTITION BY ba."lunch_flow_bank_account_id",
                            t."original_name", t."amount", t."date"
               ORDER BY t."transaction_id"
           ) AS rn
    FROM "transactions" t
    JOIN "bank_accounts" ba
      ON ba."bank_account_id" = t."bank_account_id"
      AND ba."account_id" = t."account_id"
    WHERE t."source" = 'lunch_flow'
      AND t."lunch_flow_transaction_id" IS NULL
      AND ba."lunch_flow_bank_account_id" IS NOT NULL
)
UPDATE "transactions" t
SET "lunch_flow_transaction_id" = o."lunch_flow_transaction_id"
FROM orphaned o
JOIN candidates c
  ON c."lunch_flow_bank_account_id" = o."lunch_flow_bank_account_id"
  AND c."original_name" = o."description"
  AND c."amount" = o."amount"
  AND c."date" = o."date"
  AND c."account_id" = o."account_id"
  AND c."rn" = o."rn"
WHERE t."transaction_id" = c."transaction_id"
  AND t."account_id" = c."account_id";

-- Delete any remaining orphaned lunch_flow_transactions that could not be
-- reconnected to prevent duplicate key errors on future syncs.
DELETE FROM "lunch_flow_transactions" lft
WHERE NOT EXISTS (
    SELECT 1 FROM "transactions" t
    WHERE t."lunch_flow_transaction_id" = lft."lunch_flow_transaction_id"
      AND t."account_id" = lft."account_id"
);
