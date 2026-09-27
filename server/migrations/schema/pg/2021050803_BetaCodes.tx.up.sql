CREATE TABLE betas
(
    beta_id         BIGSERIAL NOT NULL,
    code_hash       TEXT   NOT NULL,
    used_by_user_id BIGINT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT "pk_betas" PRIMARY KEY ("beta_id"),
    CONSTRAINT "uq_betas_code_hash" UNIQUE ("code_hash"),
    CONSTRAINT "fk_betas_used_by" FOREIGN KEY ("used_by_user_id") REFERENCES "users" ("user_id")
);