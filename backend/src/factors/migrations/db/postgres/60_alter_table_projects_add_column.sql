-- Add settings column in projects table

ALTER TABLE public.projects ADD COLUMN channel_group_rules jsonb;
-- ALTER TABLE public.projects DROP COLUMN channel_group_rules;
