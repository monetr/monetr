CREATE TABLE "files" (
  "file_id"            BIGSERIAL                NOT NULL,
  "account_id"         BIGINT                   NOT NULL,
  "bank_account_id"    BIGINT                   NOT NULL,
  "name"               TEXT                     NOT NULL,
  "content_type"       TEXT                     NOT NULL,
  "size"               BIGINT                   NOT NULL,
  "object_uri"         TEXT                     NOT NULL,
  "created_at"         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  "created_by_user_id" BIGINT                   NOT NULL,

  CONSTRAINT "pk_files"                          PRIMARY KEY ("file_id", "account_id", "bank_account_id"),
  CONSTRAINT "fk_files_account"                  FOREIGN KEY ("account_id")                    REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_files_bank_account"             FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
  CONSTRAINT "fk_files_users_created_by_user_id" FOREIGN KEY ("created_by_user_id")            REFERENCES "users" ("user_id")
);
