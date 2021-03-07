-- UP

ALTER TABLE events SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);
ALTER TABLE user_properties SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);
ALTER TABLE users SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);
ALTER TABLE event_names SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);
ALTER TABLE hubspot_documents SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);
ALTER TABLE salesforce_documents SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);
ALTER TABLE adwords_documents SET (autovacuum_vacuum_scale_factor = 0, autovacuum_analyze_scale_factor = 0.01, autovacuum_vacuum_threshold = 1000000);

-- DOWN
-- ALTER TABLE events RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
-- ALTER TABLE user_properties RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
-- ALTER TABLE users RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
-- ALTER TABLE event_names RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
-- ALTER TABLE hubspot_documents RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
-- ALTER TABLE salesforce_documents RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
-- ALTER TABLE adwords_documents RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor, autovacuum_vacuum_threshold);
