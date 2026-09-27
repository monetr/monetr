-- Bump the funding schedules again. This is because it wasn't fixed properly before and now the dates are fucked.
-- This will fix the existing ones if there are any. And then itll be correct going forward.
UPDATE "funding_schedules" SET "next_occurrence_original" = "next_occurrence";
