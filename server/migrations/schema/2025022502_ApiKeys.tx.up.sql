DO $$ BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'users'
    ) THEN
        CREATE TABLE IF NOT EXISTS "api_keys"
        (
            "api_key_id"   BIGSERIAL   NOT NULL,
            "user_id"      VARCHAR(32) NOT NULL,
            "name"         TEXT        NOT NULL,
            "key_hash"     TEXT        NOT NULL,
            "created_at"   TIMESTAMPTZ NOT NULL,
            "last_used_at" TIMESTAMPTZ,
            "expires_at"   TIMESTAMPTZ,
            "is_active"    BOOLEAN     NOT NULL DEFAULT TRUE,
            CONSTRAINT "pk_api_keys" PRIMARY KEY ("api_key_id"),
            CONSTRAINT "fk_api_keys_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("user_id")
        );
    ELSE
        RAISE EXCEPTION 'Table "users" does not exist. Please run the users migration first.';
    END IF;
END $$;