-- Any spending object that happens after the daylight savings time transition,
-- reset it to the previous recurrence. The process spending job will then
-- correct this by recalculating the next recurrence within 30 minutes.
UPDATE "spending"
SET "next_recurrence" = now()
WHERE "spending_type" = 0 AND 
      "next_recurrence" > '2024-11-02 00:00:00.000000+00'::TIMESTAMP AND 
      "next_recurrence" < '2025-03-10 00:00:00.000000+00'::TIMESTAMP;
