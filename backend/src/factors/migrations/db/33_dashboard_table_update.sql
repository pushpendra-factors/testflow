-- Add description column in dashboard_units table

ALTER TABLE public.dashboard_units ADD COLUMN description text;
-- ALTER TABLE public.dashboard_units DROP COLUMN description;


-- Add description column in dashboards table
ALTER TABLE public.dashboards ADD COLUMN description text;

-- ALTER TABLE public.dashboards DROP COLUMN description;