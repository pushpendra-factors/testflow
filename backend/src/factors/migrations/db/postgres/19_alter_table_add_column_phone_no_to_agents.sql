--UP
ALTER TABLE public.agents ADD COLUMN phone text;

-- DOWN
-- ALTER TABLE public.agents DROP COLUMN phone RESTRICT;