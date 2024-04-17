ALTER TABLE alert_templates ADD COLUMN is_workflow BOOLEAN;
UPDATE alert_templates SET is_workflow=FALSE;
ALTER TABLE alert_templates ADD COLUMN workflow_config JSON DEFAULT NULL;