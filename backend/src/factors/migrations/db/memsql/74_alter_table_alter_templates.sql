ALTER TABLE alert_templates ADD COLUMN is_workflow BOOLEAN;
UPDATE alert_templates SET COLUMN is_workflow=FALSE;