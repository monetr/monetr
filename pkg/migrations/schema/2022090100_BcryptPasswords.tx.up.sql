ALTER TABLE logins ADD COLUMN crypt BYTEA;
ALTER TABLE logins ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE logins ADD CONSTRAINT ck_valid_password CHECK (crypt IS NOT NULL OR password_hash IS NOT NULL);
