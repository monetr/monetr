BEGIN;
SELECT plan(1);

SELECT has_view('balances');

SELECT *
FROM finish();
ROLLBACK;
