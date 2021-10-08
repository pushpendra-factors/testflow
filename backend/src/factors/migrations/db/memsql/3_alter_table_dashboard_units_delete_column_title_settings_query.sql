UPDATE queries, dashboard_units
SET queries.Settings=dashboard_units.Settings
WHERE queries.id=dashboard_units.query_id and queries.project_id=dashboard_units.project_id
    and dashboard_units.is_deleted=false and dashboard_units.Settings is not null;


UPDATE queries, dashboard_units
SET queries.title=dashboard_units.title
WHERE queries.id=dashboard_units.query_id and queries.project_id=dashboard_units.project_id
    and dashboard_units.is_deleted=false;


ALTER TABLE dashboard_units DROP title;
--ALTER TABLE dashboard_units ADD title text;

ALTER TABLE dashboard_units DROP settings;
--ALTER TABLE dashboard_units ADD settings jsonb;

ALTER TABLE dashboard_units DROP query;
--ALTER TABLE dashboard_units ADD query jsonb;