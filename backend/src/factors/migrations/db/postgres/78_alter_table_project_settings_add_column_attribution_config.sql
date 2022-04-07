-- Add attribution_config column in project_settings table

ALTER TABLE public.project_settings ADD COLUMN attribution_config jsonb;
-- ALTER TABLE public.project_settings DROP COLUMN attribution_config;