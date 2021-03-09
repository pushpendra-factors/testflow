-- Add settings column in queries table

ALTER TABLE public.queries ADD COLUMN settings jsonb;
-- ALTER TABLE public.queries DROP COLUMN settings;

