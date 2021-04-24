BEGIN;
SELECT plan(4);

SELECT has_table('accounts');
SELECT has_table('bank_accounts');
SELECT has_table('funding_schedules');
SELECT has_table('jobs');
SELECT has_table('links');
SELECT has_table('logins');
SELECT has_table('plaid_links');
SELECT has_table('spending');
SELECT has_table('transactions');
SELECT has_table('users');

SELECT *
FROM finish();
ROLLBACK;
