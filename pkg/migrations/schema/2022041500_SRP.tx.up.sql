-- At the moment this migration cannot be reversed. There is no going back from this.

-- Add the two columns needed for SRP.
ALTER TABLE "logins"
    ADD COLUMN "verifier" BYTEA;
ALTER TABLE "logins"
    ADD COLUMN "salt" BYTEA;

-- Add a constraint to make sure that either the password_hash must be defined, OR the verifier and salt must be provided.
ALTER TABLE "logins"
    ADD CONSTRAINT "ck_hash_or_verifier" CHECK ("password_hash" IS NOT NULL OR
                                                ("verifier" IS NOT NULL AND "salt" IS NOT NULL));

-- Then remove the NOT NULL constraint from password_hash so we can slowly migrate to SRP.
ALTER TABLE "logins"
    ALTER COLUMN "password_hash" DROP NOT NULL;
