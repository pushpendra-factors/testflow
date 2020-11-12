-- Add settings column in dashboard_units table

ALTER TABLE public.dashboard_units ADD COLUMN settings jsonb;
-- ALTER TABLE public.dashboard_units DROP COLUMN settings;

