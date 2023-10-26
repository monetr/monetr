CREATE TABLE "institutions"
(
    institution_id       BIGSERIAL NOT NULL,
    name                 TEXT      NOT NULL,
    plaid_institution_id TEXT,
    plaid_products       TEXT[],
    url                  TEXT,
    primary_color        TEXT,
    logo                 TEXT,
    CONSTRAINT "pk_institutions" PRIMARY KEY ("institution_id"),
    CONSTRAINT "uq_institutions_plaid_institution_id" UNIQUE ("plaid_institution_id")
);

ALTER TABLE "links" ADD COLUMN "institution_id" BIGINT NULL;
ALTER TABLE "links" ADD CONSTRAINT "fk_links_institution" FOREIGN KEY ("institution_id") REFERENCES "institutions" ("institution_id") ON DELETE SET NULL;

-- -- Will use this later to add the proper institutions to the link records.
-- UPDATE "links" AS "link"
-- SET "institution_id" = "instituion"."institution_id"
-- FROM "plaid_links" AS "plaid_link"
-- JOIN "institutions" AS "institution" ON "institution"."plaid_institution_id" = "plaid_link"."institution_id"
-- WHERE "plaid_link"."plaid_link_id" = "link"."plaid_link_id"

