ALTER TABLE logins ALTER COLUMN crypt SET NOT NULL;
ALTER TABLE logins DROP CONSTRAINT ck_valid_password;
ALTER TABLE logins DROP COLUMN password_hash;
