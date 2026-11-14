CREATE TABLE "root_keys" (
  "root_key_id" VARCHAR(32) NOT NULL,
  "wrappers"    JSONB       NOT NULL,
  "fingerprint" TEXT        NOT NULL,
  "created_at"  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  CONSTRAINT "pk_root_keys"             PRIMARY KEY ("root_key_id"),
  CONSTRAINT "uq_root_keys_fingerprint" UNIQUE ("fingerprint")
);

CREATE TABLE "key_encryption_keys" (
  "key_encryption_key_id" VARCHAR(32) NOT NULL,
  "root_key_id"           VARCHAR(32) NOT NULL,
  "wrapped"               TEXT        NOT NULL,
  "created_at"            TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  CONSTRAINT "pk_key_encryption_keys"          PRIMARY KEY ("key_encryption_key_id"),
  -- Restrict so a root key can't be removed while a KEK is still wrapped by it.
  CONSTRAINT "fk_key_encryption_keys_root_key" FOREIGN KEY ("root_key_id") REFERENCES "root_keys" ("root_key_id") ON DELETE RESTRICT
);

CREATE TABLE "secretbox_keys" (
  "secretbox_key_id"      VARCHAR(32) NOT NULL,
  "fingerprint"           TEXT        NOT NULL,
  "created_at"            TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "key_encryption_key_id" VARCHAR(32) NOT NULL,
  "encrypted_data"        TEXT        NOT NULL,
  CONSTRAINT "pk_secretbox_keys"                    PRIMARY KEY ("secretbox_key_id"),
  CONSTRAINT "uq_secretbox_keys_fingerprint"        UNIQUE ("fingerprint"),
  -- Restrict so a KEK can't be removed while a DEK is still wrapped by it.
  CONSTRAINT "fk_secretbox_keys_key_encryption_key" FOREIGN KEY ("key_encryption_key_id") REFERENCES "key_encryption_keys" ("key_encryption_key_id") ON DELETE RESTRICT
);
