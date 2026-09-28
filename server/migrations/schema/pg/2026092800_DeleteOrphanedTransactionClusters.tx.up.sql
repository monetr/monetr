-- The cluster ID backfill only gave each transaction one cluster, so some
-- clusters ended up with nothing pointing at them. The clustering job never sees
-- those, and they block the same signature + centroid from being inserted again.
DELETE FROM "transaction_clusters" AS "txc"
WHERE NOT EXISTS (
  SELECT 1
  FROM "transactions" AS "t"
  WHERE "t"."transaction_cluster_id" = "txc"."transaction_cluster_id"
    AND "t"."account_id" = "txc"."account_id"
    AND "t"."bank_account_id" = "txc"."bank_account_id"
);
