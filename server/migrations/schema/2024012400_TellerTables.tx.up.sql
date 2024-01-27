CREATE TABLE "teller_links" (
  "teller_link_id"         BIGSERIAL NOT NULL,
  "account_id"             BIGINT    NOT NULL,
  "enrollment_id"          TEXT      NOT NULL,
  "teller_user_id"         TEXT      NOT NULL,
  "status"                 INT       NOT NULL,
  "error_code"             TEXT,
  "institution_name"       TEXT      NOT NULL,
  "last_manual_sync"       TIMESTAMP WITH TIME ZONE,
  "last_successful_update" TIMESTAMP WITH TIME ZONE,
  "last_attempted_update"  TIMESTAMP WITH TIME ZONE,
  "updated_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_by_user_id"     BIGINT    NOT NULL,
  CONSTRAINT "pk_teller_links"                          PRIMARY KEY ("teller_link_id", "account_id"),
  CONSTRAINT "fk_teller_links_account"                  FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_teller_links_users_created_by_user_id" FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("user_id"),
  CONSTRAINT "uq_teller_links_enrollment"               UNIQUE ("account_id", "enrollment_id")
);

ALTER TABLE "links" 
ADD COLUMN "teller_link_id" BIGINT,
ADD CONSTRAINT "fk_links_teller_link" FOREIGN KEY ("teller_link_id", "account_id") REFERENCES "teller_links" ("teller_link_id", "account_id");

CREATE TABLE "teller_bank_accounts" (
  "teller_bank_account_id" BIGSERIAL NOT NULL,
  "account_id"             BIGINT    NOT NULL,
  "teller_link_id"         BIGINT    NOT NULL,
  "teller_id"              TEXT      NOT NULL,
  "institution_id"         TEXT      NOT NULL,
  "institution_name"       TEXT      NOT NULL,
  "mask"                   TEXT      NOT NULL,
  "name"                   TEXT      NOT NULL,
  "type"                   TEXT      NOT NULL,
  "sub_type"               TEXT      NOT NULL,
  "status"                 INT       NOT NULL,
  "updated_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_by_user_id"     BIGINT    NOT NULL,
  CONSTRAINT "pk_teller_bank_accounts"                          PRIMARY KEY ("teller_bank_account_id", "account_id"),
  CONSTRAINT "fk_teller_bank_accounts_account"                  FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_teller_bank_accounts_teller_link"              FOREIGN KEY ("teller_link_id", "account_id") REFERENCES "teller_links" ("teller_link_id", "account_id"),
  CONSTRAINT "fk_teller_bank_accounts_users_created_by_user_id" FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("user_id"),
  CONSTRAINT "uq_teller_bank_accounts_teller_id"                UNIQUE ("account_id", "teller_id")
);

ALTER TABLE "bank_accounts"
ADD COLUMN "teller_bank_account_id" BIGINT,
ADD CONSTRAINT "fk_bank_accounts_teller_bank_account" FOREIGN KEY ("teller_bank_account_id", "account_id") REFERENCES "teller_bank_accounts" ("teller_bank_account_id", "account_id");
