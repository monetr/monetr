BEGIN;
SELECT plan(4);

SELECT has_table('logins');
SELECT has_table('registrations');
SELECT has_table('email_verifications');
SELECT has_table('phone_verifications');

SELECT *
FROM finish();
ROLLBACK;
