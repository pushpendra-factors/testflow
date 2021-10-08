
UPDATE queries
SET Settings=dashboard_units.Settings
    FROM dashboard_units
WHERE queries.id=dashboard_units.query_id and queries.project_id=dashboard_units.project_id
  and dashboard_units.is_deleted=false and dashboard_units.Settings is not null;
--UPDATE dashboard_units
--SET Settings=queries.Settings
--FROM queries
--WHERE queries.id=dashboard_units.query_id;

UPDATE queries
SET title=dashboard_units.title
    FROM dashboard_units
WHERE queries.id=dashboard_units.query_id and queries.project_id=dashboard_units.project_id
  and dashboard_units.is_deleted=false;
--UPDATE dashboard_units
--SET title=queries.title
--FROM queries
--WHERE queries.id=dashboard_units.query_id;

ALTER TABLE public.dashboard_units DROP title;
--ALTER TABLE public.dashboard_units ADD title text;

ALTER TABLE public.dashboard_units DROP settings;
--ALTER TABLE public.dashboard_units ADD settings jsonb;

ALTER TABLE public.dashboard_units DROP query;
--ALTER TABLE public.dashboard_units ADD query jsonb;