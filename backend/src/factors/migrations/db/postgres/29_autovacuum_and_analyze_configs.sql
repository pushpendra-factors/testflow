-- UP

-- Update auto vacuum config to 2% from default value of 20% and analyze config to 1% from default value of 10% for heavy tables.
ALTER TABLE events SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);
ALTER TABLE user_properties SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);
ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);
ALTER TABLE hubspot_documents SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);
ALTER TABLE adwords_documents SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);
ALTER TABLE event_names SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);

-- DOWN
-- ALTER TABLE events RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
-- ALTER TABLE user_properties RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
-- ALTER TABLE users RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
-- ALTER TABLE hubspot_documents RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
-- ALTER TABLE adwords_documents RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
-- ALTER TABLE event_names RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
