ALTER TABLE project_settings DROP COLUMN int_six_signal;
ALTER TABLE project_settings ADD COLUMN int_client_six_signal_key bool DEFAULT FALSE;
ALTER TABLE project_settings ADD COLUMN int_factors_six_signal_key bool DEFAULT FALSE;