CREATE TABLE IF NOT EXISTS "users"
(
    "user_id"    BIGSERIAL NOT NULL,
    "login_id"   BIGINT    NOT NULL,
    "account_id" BIGINT    NOT NULL,
    "first_name" TEXT      NOT NULL,
    "last_name"  TEXT,
    CONSTRAINT "pk_users" PRIMARY KEY ("user_id"),
    CONSTRAINT "uq_users_login_id_account_id" UNIQUE ("login_id", "account_id")
);
