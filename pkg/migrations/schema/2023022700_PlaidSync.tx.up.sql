CREATE TABLE "plaid_syncs" (
  plaid_sync_id BIGSERIAL   NOT NULL,
  plaid_link_id BIGINT      NOT NULL,
  timestamp     TIMESTAMPTZ NOT NULL,
  trigger       TEXT        NOT NULL,
  cursor        TEXT        NOT NULL,
  added         INT         NOT NULL,
  modified      INT         NOT NULL,
  removed       INT         NOT NULL,
  CONSTRAINT pk_plaid_syncs PRIMARY KEY ("plaid_sync_id"),
  CONSTRAINT fk_plaid_syncs_link FOREIGN KEY ("plaid_link_id") REFERENCES "plaid_links" ("plaid_link_id")
);

CREATE INDEX "ix_plaid_syncs_timestamp"
ON "plaid_syncs" ("plaid_link_id", "timestamp" DESC);

ALTER TABLE "plaid_links" ADD COLUMN "use_plaid_sync" BOOLEAN DEFAULT false;
UPDATE "plaid_links" SET "use_plaid_sync" = false;
ALTER TABLE "plaid_links" ALTER COLUMN "use_plaid_sync" SET NOT NULL;
