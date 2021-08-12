-- Add settings column in projects table

ALTER TABLE public.projects ADD COLUMN salesforce_touch_points jsonb;
-- ALTER TABLE public.projects DROP COLUMN salesforce_touch_points;

ALTER TABLE public.projects ADD COLUMN hubspot_touch_points jsonb;
-- ALTER TABLE public.projects DROP COLUMN hubspot_touch_points;

