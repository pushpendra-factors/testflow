-- UP
CREATE INDEX projects_updated_at ON projects(updated_at ASC);
CREATE INDEX agents_updated_at ON agents(updated_at ASC);
CREATE INDEX project_agent_mappings_updated_at ON project_agent_mappings(updated_at ASC);
CREATE INDEX project_billing_account_mappings_updated_at ON project_billing_account_mappings(updated_at ASC);
CREATE INDEX project_settings_updated_at ON project_settings(updated_at ASC);
CREATE INDEX bigquery_settings_updated_at ON bigquery_settings(updated_at ASC);
CREATE INDEX billing_accounts_updated_at ON billing_accounts(updated_at ASC);
CREATE INDEX dashboard_units_updated_at ON dashboard_units(updated_at ASC);
CREATE INDEX dashboards_updated_at ON dashboards(updated_at ASC);
CREATE INDEX facebook_documents_updated_at ON facebook_documents(updated_at ASC);
CREATE INDEX factors_goals_updated_at ON factors_goals(updated_at ASC);
CREATE INDEX factors_tracked_events_updated_at ON factors_tracked_events(updated_at ASC);
CREATE INDEX factors_tracked_user_properties_updated_at ON factors_tracked_user_properties(updated_at ASC);
CREATE INDEX queries_updated_at ON queries(updated_at ASC);
CREATE INDEX reports_updated_at ON reports(updated_at ASC);
CREATE INDEX scheduled_tasks_updated_at ON scheduled_tasks(updated_at ASC);

CREATE INDEX user_properties_updated_at ON user_properties(updated_at ASC);
CREATE INDEX users_updated_at ON users(updated_at ASC);
CREATE INDEX event_names_updated_at ON event_names(updated_at ASC);
CREATE INDEX salesforce_documents_updated_at ON salesforce_documents(updated_at ASC);

CREATE INDEX events_updated_at ON events(updated_at ASC);
CREATE INDEX adwords_documents_updated_at ON adwords_documents(updated_at ASC);
CREATE INDEX hubspot_documents_updated_at ON hubspot_documents(updated_at ASC);

-- DOWN
DROP INDEX projects_updated_at;
DROP INDEX agents_updated_at;
DROP INDEX project_agent_mappings_updated_at;
DROP INDEX project_billing_account_mappings_updated_at;
DROP INDEX project_settings_updated_at;
DROP INDEX bigquery_settings_updated_at;
DROP INDEX billing_accounts_updated_at;
DROP INDEX dashboard_units_updated_at;
DROP INDEX dashboards_updated_at;
DROP INDEX facebook_documents_updated_at;
DROP INDEX factors_goals_updated_at;
DROP INDEX factors_tracked_events_updated_at;
DROP INDEX factors_tracked_user_properties_updated_at;
DROP INDEX queries_updated_at;
DROP INDEX reports_updated_at;
DROP INDEX scheduled_tasks_updated_at;

DROP INDEX user_properties_updated_at;
DROP INDEX users_updated_at;
DROP INDEX event_names_updated_at;
DROP INDEX events_updated_at;
DROP INDEX adwords_documents_updated_at;
DROP INDEX hubspot_documents_updated_at;
DROP INDEX salesforce_documents_updated_at;