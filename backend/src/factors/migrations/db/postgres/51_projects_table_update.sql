-- Add settings column in projects table

ALTER TABLE public.projects ADD COLUMN interaction_settings jsonb;
-- ALTER TABLE public.projects DROP COLUMN interaction_settings;

