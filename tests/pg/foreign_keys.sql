BEGIN;
SELECT plan(2);

SELECT has_fk('transactions'); -- Make sure the transactions table has a foreign key.
SELECT col_is_fk('transactions', 'account_id'); -- Account ID should always be a foreign key.

SELECT *
FROM finish();
ROLLBACK;
