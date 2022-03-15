-- Add settings column in dashboards table

ALTER TABLE public.dashboards ADD COLUMN settings jsonb;
-- ALTER TABLE public.dashboards DROP COLUMN settings;