-- UP
SET GLOBAL default_table_type = rowstore;

CREATE DATABASE IF NOT EXISTS factors;

USE factors;

CREATE TABLE IF NOT EXISTS events (
    id text NOT NULL,
    project_id bigint NOT NULL,
    customer_event_id text,
    user_id text,
    user_properties_id text,
    event_name_id text,
    count bigint,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    user_properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    session_id text,
    timestamp bigint,
    properties_updated_timestamp bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (project_id, event_name_id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (id) USING HASH,
    KEY (user_id) USING HASH,
    KEY (customer_event_id) USING HASH,
    SHARD KEY (user_id)

    -- Required constraints.
    -- Unique (project_id, customer_event_id)
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, event_name_id) -> event_names (project_id, id)
    -- Ref (project_id, user_id) -> users (project_id, id)
    -- Ref (project_id, user_id, user_properties_id) -> user_properties (project_id, user_id, id)

    -- Additional constraint.
    -- Ref (project_id, session_id) -> events (project_id, id) WHERE event is session.
);
ALTER TABLE events AUTOSTATS_SAMPLING = OFF;

CREATE TABLE IF NOT EXISTS users (
    id text NOT NULL,
    project_id bigint NOT NULL,
    customer_user_id text,
    segment_anonymous_id text,
    amp_user_id text,
    properties_id text,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    properties_updated_timestamp bigint,
    join_timestamp bigint,
    is_group_user boolean,
    group_1_id text,
    group_1_user_id text,
    group_2_id text,
    group_2_user_id text,
    group_3_id text,
    group_3_user_id text,
    group_4_id text,
    group_4_user_id text,
    group_5_id text,
    group_5_user_id text,
    group_6_id text,
    group_6_user_id text,
    group_7_id text,
    group_7_user_id text,
    group_8_id text,
    group_8_user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    source int,
    customer_user_id_source int,
    -- COLUMNSTORE key is sort key, can we add an incremental numerical column to the end?
    -- Initial parts of the indices are still useful when don't use the last column which is an incremental value.
    KEY (project_id, source, join_timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (id) USING HASH,
    KEY (project_id) USING HASH,
    KEY (customer_user_id) USING HASH,
    KEY (segment_anonymous_id) USING HASH,
    KEY (amp_user_id) USING HASH,
    KEY (join_timestamp) USING HASH,
    KEY (is_group_user) USING HASH,
    KEY (group_1_id) USING HASH,
    KEY (group_2_id) USING HASH,
    KEY (group_3_id) USING HASH,
    KEY (group_4_id) USING HASH,
    KEY (source) USING HASH,
    UNIQUE KEY (project_id, id) USING HASH,
    SHARD KEY (id)

    -- Required constraints.
    -- Unique (project_id, segment_anonymous_id)
    -- Unique (project_id, amp_user_id)
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, properties_id) -> user_properties(project_id, id)
);
ALTER TABLE users AUTOSTATS_SAMPLING = OFF;

CREATE TABLE IF NOT EXISTS event_names (
    id text, -- UUID
    project_id bigint,
    name text,
    type varchar(2),
    filter_expr varchar(500),
    deleted bool NOT NULL DEFAULT false,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (id),
    KEY (project_id, name) USING CLUSTERED COLUMNSTORE,
    KEY (project_id) USING HASH,
    KEY (id) USING HASH,
    KEY (name) USING HASH

    -- Required constraints.
    -- Unique (project_id, name, type) WHERE type != 'FE'
    -- Unique (project_id, type, filter_expr)
    -- Unique (project_id, id)
    -- Ref (project_id) -> projects(id)
);


CREATE TABLE IF NOT EXISTS adwords_documents (
    id text,
    project_id bigint,
    customer_account_id text,
    type int,
    timestamp bigint,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    ad_group_id bigint,
    ad_id bigint,
    keyword_id bigint,
    campaign_id bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id, id),
    KEY (project_id, customer_account_id, type, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (project_id) USING HASH,
    KEY (customer_account_id) USING HASH,
    KEY (type) USING HASH,
    KEY (updated_at) USING HASH
    -- Required constraints.
    -- Unique (project_id, customer_account_id, timestamp, id)
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, customer_account_id) -> project_settings(project_id, int_adwords_customer_account_id)
);

CREATE TABLE IF NOT EXISTS agents (
    uuid text,
    first_name varchar(100),
    last_name varchar(100),
    email varchar(100),
    is_email_verified bool,
    phone varchar(100),
    company_url text,
    salt varchar(100),
    password varchar(100),
    password_created_at timestamp(6),
    invited_by text,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    last_logged_in_at timestamp(6), -- Milliseconds precision required.
    login_count bigint,
    subscribe_newsletter bool,
    int_adwords_refresh_token text,
    int_salesforce_refresh_token text,
    int_salesforce_instance_url text,
    int_google_organic_refresh_token text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_onboarding_flow_seen boolean,
    is_auth0_user boolean NOT NULL DEFAULT false,
    value json,
    slack_access_tokens JSON,
    teams_access_tokens JSON,
    last_logged_out bigint DEFAULT 0,
    SHARD KEY (uuid),
    PRIMARY KEY (uuid),
    KEY (updated_at),
    KEY (email)

    -- Required constraints.
    -- Unique (email)
    -- Ref (invited_by) -> agents(uuid) WHERE uuid != invited_by
);

CREATE TABLE IF NOT EXISTS bigquery_settings (
    id text,
    project_id bigint,
    bq_project_id text,
    bq_dataset_name text,
    bq_credentials_json text,
    last_run_at bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    KEY (updated_at),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY (project_id, bq_project_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);

CREATE TABLE IF NOT EXISTS billing_accounts (
    id text,
    plan_id bigint,
    agent_uuid text,
    organization_name text,
    billing_address text,
    pincode text,
    phone_no text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    PRIMARY KEY (agent_uuid, id)

    -- Required constraints.
    -- Ref (agent_uuid) -> agents(id)
);

CREATE TABLE IF NOT EXISTS dashboard_units (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    dashboard_id bigint,
    description text,
    presentation varchar(5),
    query_id bigint,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, dashboard_id, id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, dashboard_id) -> dashboards(project_id, id)
    -- Ref (project_id, query_id) -> queries(project_id, id)
);

CREATE TABLE IF NOT EXISTS dashboards (
    id bigint AUTO_INCREMENT,
    project_id bigint NOT NULL,
    agent_uuid text,
    name text,
    units_position json,
    description text,
    type varchar(5),
    class text,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    settings json,
    created_at timestamp(6),
    updated_at timestamp(6),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
	KEY (project_id, agent_uuid),
    KEY (updated_at),
    UNIQUE KEY unique_project_id_agent_uuid_id_idx(project_id, agent_uuid, id)

    -- Required constraits.
    -- Ref (project_id) -> projects(id)
    -- Ref (agent_uuid) -> agents(uuid) - This cannot be bound to project_id using project_agent_mappings. As removing agent from project is allowed.
);

CREATE TABLE IF NOT EXISTS facebook_documents (
    id text,
    project_id bigint,
    customer_ad_account_id text,
    platform text,
    type int,
    timestamp bigint,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    campaign_id text,
    ad_set_id text,
    ad_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id),
    KEY (project_id, customer_ad_account_id, platform, timestamp) USING CLUSTERED COLUMNSTORE

    -- Required constraints.
    -- Unique (project_id, customer_ad_account_id, platform, type, timestamp, id)
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, customer_ad_account_id) -> project_settings(project_id, int_facebook_ad_account)
);

CREATE TABLE IF NOT EXISTS factors_goals (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    name text,
    rule json,
    type varchar(2),
    created_by text,
    last_tracked_at timestamp(6),
    is_active boolean,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
	UNIQUE KEY unique_project_id_name_idx(project_id, name)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (created_by) -> agents (uuid)
);

CREATE TABLE IF NOT EXISTS factors_tracked_events (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    event_name_id text,
    type varchar(2),
    created_by text,
    last_tracked_at timestamp(6),
    is_active boolean,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
	UNIQUE KEY (project_id, event_name_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, event_name_id) -> event_names(project_id, id)
    -- Ref (created_by) -> agents(uuid)
);

CREATE TABLE IF NOT EXISTS factors_tracked_user_properties (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    user_property_name text,
    type varchar(2),
    created_by text,
    last_tracked_at timestamp(6),
    is_active boolean,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
	UNIQUE KEY (project_id, user_property_name)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (created_by) -> agents(uuid)
);

CREATE TABLE IF NOT EXISTS hubspot_documents (
    id text,
	project_id bigint,
    type int,
    action int,
    timestamp bigint,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    group_user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, type, id),
    KEY (project_id, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (user_id) USING HASH,
    KEY (type) USING HASH,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_id_type_action_timestamp_unique_idx(project_id, id, type,action,timestamp) USING HASH

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id, id, type, action, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);

CREATE TABLE IF NOT EXISTS project_agent_mappings (
    project_id bigint,
    agent_uuid text,
    role bigint,
    invited_by text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, agent_uuid)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (agent_uuid) -> agents(uuid)
    -- Ref (invited_by) -> agents(uuid)
);

CREATE TABLE IF NOT EXISTS project_billing_account_mappings (
    project_id bigint,
    billing_account_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, billing_account_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (billing_account_id) -> billing_accounts(id)
);

CREATE TABLE IF NOT EXISTS project_settings (
    project_id bigint,
    attribution_config json,
    auto_track boolean NOT NULL DEFAULT FALSE,
    auto_track_spa_page_view boolean NOT NULL DEFAULT FALSE,
    auto_form_capture boolean NOT NULL DEFAULT FALSE,
    auto_click_capture boolean NOT NULL DEFAULT FALSE,
    auto_capture_form_fills boolean NOT NULL DEFAULT FALSE,
    exclude_bot boolean NOT NULL DEFAULT FALSE,
    int_segment boolean NOT NULL DEFAULT FALSE,
    int_rudderstack boolean NOT NULL DEFAULT FALSE,
    int_adwords_enabled_agent_uuid text,
    int_adwords_customer_account_id text,
    int_adwords_client_manager_map json,
    int_hubspot boolean NOT NULL DEFAULT FALSE,
    int_hubspot_api_key text,
    int_hubspot_refresh_token text,
    int_hubspot_sync_info json,
    int_hubspot_portal_id int,
    int_hubspot_first_time_synced boolean NOT NULL DEFAULT FALSE,
    int_facebook_email text,
    int_facebook_access_token text,
    int_facebook_agent_uuid text,
    int_facebook_user_id text,
    int_facebook_ad_account text,
    int_facebook_token_expiry bigint,
    int_linkedin_ad_account text,
    int_linkedin_access_token text,
    cache_settings json,
    int_linkedin_access_token_expiry bigint,
    int_linkedin_refresh_token text,
    int_linkedin_refresh_token_expiry bigint,
    int_linkedin_agent_uuid text,
    archive_enabled boolean NOT NULL DEFAULT FALSE,
    bigquery_enabled boolean NOT NULL DEFAULT FALSE,
    int_salesforce_enabled_agent_uuid text,
    int_drift boolean NOT NULL DEFAULT FALSE,
    int_google_organic_enabled_agent_uuid text,
    int_google_organic_url_prefixes text,
    int_google_ingestion_timezone text,
    int_facebook_ingestion_timezone text,
    int_clear_bit boolean NOT NULL DEFAULT FALSE,
    clearbit_key text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    lead_squared_config json,
    is_weekly_insights_enabled boolean,
    is_explain_enabled boolean,
    timelines_config json,
    client6_signal_key text,
    factors6_signal_key text,
    int_client_six_signal_key boolean NOT NULL DEFAULT FALSE,
    int_factors_six_signal_key boolean NOT NULL DEFAULT FALSE,
    integration_bits varchar(32) DEFAULT '00000000000000000000000000000000',
    project_currency varchar(10),
    is_path_analysis_enabled boolean,
    filter_ips JSON,
    is_deanonymization_requested boolean,
    is_onboarding_completed boolean,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id)


    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (int_adwords_enabled_agent_uuid) -> agents(uuid)
    -- Ref (int_facebook_agent_uuid) -> agents(uuid)
    -- Ref (int_linkedin_agent_uuid) -> agents(uuid)
    -- Ref (int_salesforce_enabled_agent_uuid) -> agents(uuid)
    -- Ref (int_google_organic_enabled_agent_uuid) -> agents(uuid)
);

CREATE TABLE IF NOT EXISTS projects (
    id bigint AUTO_INCREMENT,
    name text,
    token varchar(32),
    private_token varchar(32),
    project_uri text,
    time_format text,
    time_zone text,
    date_format text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    interaction_settings json,
    salesforce_touch_points json,
    hubspot_touch_points json,
    jobs_metadata json,
    channel_group_rules json,
    profile_picture text,
    KEY (updated_at),
    PRIMARY KEY (id),
    KEY (token),
    KEY (private_token)

    -- Required constraints.
    -- Unique (token)
    -- Unique (private_token)
);

CREATE TABLE IF NOT EXISTS queries (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    title text, -- Add trigram index for like queries.
    query json,
    settings json,
    type int,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    created_by text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    id_text text,
    converted boolean,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY (project_id, type, id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (created_by) -> agents(uuid)
);

CREATE TABLE IF NOT EXISTS salesforce_documents (
    id text,
    project_id bigint,
    type int,
    action int,
    timestamp bigint,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    group_user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, type, id),
    KEY (project_id, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (user_id) USING HASH,
    KEY (type) USING HASH,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_id_type_timestamp_unique_idx(project_id, id, type,timestamp) USING HASH

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id, id, type, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);

CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id text,
    project_id bigint,
    job_id text,
    task_type text,
    task_status text,
    task_start_time bigint,
    task_end_time bigint,
    task_details json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);

CREATE TABLE IF NOT EXISTS linkedin_documents (
    id text NOT NULL,
    project_id bigint NOT NULL,
    customer_ad_account_id text NOT NULL,
    type int NOT NULL,
    timestamp bigint NOT NULL,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    creative_id text,
    campaign_group_id text,
    campaign_id text,
    is_backfilled boolean default FALSE NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id),
    KEY (project_id, customer_ad_account_id, timestamp) USING CLUSTERED COLUMNSTORE

    -- Required constraints.
    -- Unique (project_id, customer_ad_account_id, type, timestamp, id)
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, customer_ad_account_id) -> project_settings(project_id, int_facebook_ad_account)
);

CREATE TABLE IF NOT EXISTS smart_property_rules (
    id text,
    project_id bigint NOT NULL,
    type bigint NOT NULL,
    description text,
    name text NOT NULL,
    rules json NOT NULL,
    evaluation_status int NOT NULL,
    is_deleted bool DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id)
);

CREATE TABLE IF NOT EXISTS smart_properties (
    project_id bigint NOT NULL,
    source text NOT NULL,
    object_id text NOT NULL,
    object_type bigint NOT NULL,
    object_property json NOT NULL,
    properties json NOT NULL,
    rules_ref json NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, object_id, object_type, source)
);

CREATE TABLE IF NOT EXISTS property_details (
    project_id bigint NOT NULL,
    event_name_id text,
    `key` text NOT NULL,
    `type` text NOT NULL,
    entity integer NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    UNIQUE KEY property_details_project_id_event_name_id_key_unique_idx(project_id, event_name_id,`key`)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref.(project_id,event_name_id) -> event_names(project_id,id)
);

CREATE TABLE IF NOT EXISTS display_names (
    id text,
    project_id bigint NOT NULL,
    event_name text NULL,
    property_name text NULL,
    entity_type integer NOT NULL,
    display_name text NOT NULL,
    tag text NOT NULL,
    group_name text NOT NULL,
    group_object_name text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY  display_names_project_id_event_name_property_name_tag_unique_idx(project_id, event_name, property_name, tag),
    UNIQUE KEY  display_names_project_id_object_group_entity_tag_unique_idx(project_id, group_name, entity_type, group_object_name, display_name)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);

CREATE TABLE IF NOT EXISTS google_organic_documents (
    id text NOT NULL,
    project_id bigint NOT NULL,
    url_prefix text NOT NULL,
    type int,
    timestamp bigint NOT NULL,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id),
    KEY (project_id, url_prefix, timestamp, type, id) USING CLUSTERED COLUMNSTORE

    -- Required constraints.
    -- Unique (project_id, customer_ad_account_id, type, timestamp, id)
    -- Ref (project_id) -> projects(id)
);

CREATE TABLE IF NOT EXISTS project_model_metadata
(
    id text NOT NULL,
    project_id bigint NOT NULL,
    model_id bigint NOT NULL,
    model_type text NOT NULL,
    start_time  bigint NOT NULL,
    end_time bigint NOT NULL,
    chunks text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    UNIQUE KEY  project_model_metadata_project_id_stdate_enddate_unique_idx(project_id, start_time, end_time),
    KEY (project_id) USING HASH

    -- Add Foreign Key for project_id
);

CREATE TABLE IF NOT EXISTS task_details
(
    id text NOT NULL,
    task_id bigint AUTO_INCREMENT,
    task_name text NOT NULL,
    source text NULL,
    frequency integer NOT NULL,
    frequency_interval integer, -- There are 4 types hourly/daily/weekly/stateless
    skip_start_index integer,
    skip_end_index integer,
    offset_start_minutes integer,
    recurrence boolean,
    metadata json,
    is_project_enabled boolean,
    delay_alert_threshold_hours integer,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (task_id),
    PRIMARY KEY (task_id)
);

CREATE TABLE IF NOT EXISTS task_execution_details
(
    id text NOT NULL,
    task_id bigint NOT NULL,
    project_id bigint NOT NULL,
    delta bigint NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    metadata json,
    is_completed boolean,
    KEY (updated_at),
    SHARD KEY (task_id),
    KEY (task_id) USING HASH,
    PRIMARY KEY (task_id, id)
);

CREATE TABLE IF NOT EXISTS task_execution_dependency_details
(
    id text NOT NULL,
    task_id bigint NOT NULL,
    dependent_task_id bigint NOT NULL,
    dependency_offset integer NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (task_id),
    PRIMARY KEY (task_id, id),
    KEY (task_id) USING HASH
);

CREATE TABLE IF NOT EXISTS weekly_insights_metadata
(
    id text NOT NULL,
    project_id bigint NOT NULL,
    query_id bigint NOT NULL,
    insight_type text NOT NULL,
    base_start_time  bigint NOT NULL,
    base_end_time bigint NOT NULL,
    comparison_start_time  bigint NOT NULL,
    comparison_end_time bigint NOT NULL,
    insight_id bigint NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    UNIQUE KEY  weekly_insights_metadata_project_id_stdate_enddate_unique_idx(project_id, query_id, base_start_time, base_end_time, comparison_start_time, comparison_end_time),
    KEY (project_id) USING HASH,
    PRIMARY KEY (project_id, id)

);

CREATE TABLE IF NOT EXISTS templates (
    project_id bigint NOT NULL,
    type int NOT NULL,
    thresholds JSON,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    PRIMARY KEY (project_id, type),
    SHARD KEY (project_id)
);

CREATE TABLE IF NOT EXISTS feedbacks(
    id text NOT NULL,
    project_id bigint NOT NULL,
    feature text NOT NULL,
    property json,
    vote_type integer NOT NULL,
    created_by text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    PRIMARY KEY (id,project_id),
    SHARD KEY (project_id)

);

CREATE ROWSTORE TABLE IF NOT EXISTS groups(
    project_id bigint NOT NULL,
    id int NOT NULL,
    name text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, name),
    UNIQUE KEY (project_id,id)
);

CREATE TABLE IF NOT EXISTS group_relationships(
    project_id bigint NOT NULL,
    left_group_name_id int NOT NULL,
    left_group_user_id text NOT NULL,
    right_group_name_id int NOT NULL,
    right_group_user_id text NOT NULL,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (left_group_user_id),
    KEY (project_id, left_group_user_id) USING CLUSTERED COLUMNSTORE,
    UNIQUE KEY(project_id, left_group_user_id,right_group_user_id) USING HASH
);

CREATE TABLE IF NOT EXISTS content_groups(
    id text NOT NULL,
    project_id bigint NOT NULL,
    content_group_name text,
    content_group_description text,
    rule json,
    created_by text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean,
    SHARD KEY (project_id),
    PRIMARY KEY (id, project_id),
    UNIQUE KEY (project_id,content_group_name)
);

CREATE ROWSTORE TABLE IF NOT EXISTS custom_metrics(
    project_id bigint NOT NULL,
    id text NOT NULL,
    name text NOT NULL,
    description text,
    type_of_query int,  -- represents if kpi-profiles
    object_type text, -- represents if hubspot_contact ...
    transformations json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY unique_custom_metrics_project_id_name_idx(project_id, name) USING HASH
);
-- DOWN
-- DROP TABLE IF EXISTS custom_metrics;

CREATE ROWSTORE TABLE IF NOT EXISTS leadgen_settings (
    project_id bigint NOT NULL,
    source int NOT NULL,
    source_property text NOT NULL,
    spreadsheet_id text,
    sheet_name text,
    row_read bigint,
    timezone text,
    created_at timestamp(6),
    updated_at timestamp(6),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, source)
);
-- DOWN
-- DROP TABLE IF EXISTS leadgen_settings;

CREATE ROWSTORE TABLE IF NOT EXISTS fivetran_mappings(
    project_id bigint NOT NULL,
    id text NOT NULL,
    integration text NOT NULL,
    connector_id text NOT NULL,
    schema_id text NOT NULL,
    accounts text NOT NULL,
    status boolean,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id)
);



CREATE TABLE IF NOT EXISTS integration_documents (
    document_id text,
    project_id bigint,
    customer_account_id text,
    document_type int,
    timestamp bigint,
    source text,
    value JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id, document_id),
    KEY (updated_at) USING HASH,
    KEY (project_id, customer_account_id, document_id, document_type, source, timestamp)  USING CLUSTERED COLUMNSTORE
);

CREATE TABLE IF NOT EXISTS shareable_urls (
    id text NOT NULL,
    query_id text NOT NULL,
    entity_type integer NOT NULL,
    share_type integer NOT NULL,
    entity_id bigint NOT NULL,
    created_at timestamp(6),
    updated_at timestamp(6),
    is_deleted boolean NOT NULL DEFAULT false,
    expires_at bigint,
    project_id bigint NOT NULL,
    created_by text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS shareable_url_audits (
    id text NOT NULL,
    project_id bigint NOT NULL,
    share_id text NOT NULL,
    query_id text NOT NULL,
    entity_type integer NOT NULL,
    share_type integer NOT NULL,
    entity_id bigint NOT NULL,
    created_at timestamp(6),
    updated_at timestamp(6),
    is_deleted boolean NOT NULL DEFAULT false,
    expires_at bigint,
    accessed_by text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS alerts(
    id text NOT NULL,
    project_id bigint NOT NULL,
    alert_name text,
    created_by text,
    alert_type int,
    alert_description json,
    alert_configuration json,
    query_id bigint,
    last_alert_sent bool,
    last_run_time timestamp(6),
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean,
    SHARD KEY (project_id),
    PRIMARY KEY (id, project_id)
);

-- create crm_users table
CREATE TABLE IF NOT EXISTS crm_users (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    type int NOT NULL ,
    timestamp bigint NOT NULL,
    email text,
    phone text,
    action int NOT NULL,
    metadata JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, type, id),
    KEY (project_id, source, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id,source, id, type, action, timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id,source, id, type, action, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);


-- create crm_groups table
CREATE TABLE IF NOT EXISTS crm_groups (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    type int NOT NULL,
    timestamp bigint NOT NULL,
    action int NOT NULL,
    metadata JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, type, id),
    KEY (project_id, source, type, action, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id,source, id, type, action, timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id,source, id, type, action, timestamp)
    -- Ref (project_id, user_id) -> users(project_id, id)
);


-- create new crm_relationships table
CREATE TABLE IF NOT EXISTS crm_relationships (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    from_type int NOT NULL,
    from_id text NOT NULL,
    to_type int NOT NULL,
    to_id text NOT NULL,
    timestamp bigint NOT NULL,
    external_relationship_name text,
    external_relationship_id text,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4',
    skip_process  boolean NOT NULL DEFAULT FALSE,
    synced boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, from_type,from_id),
    KEY (project_id, source, from_type, to_type, from_id, to_id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id, source, from_type, from_id, to_type, to_id) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id, source, from_type, from_id, to_type, to_id)
);

-- create new crm_activities table
CREATE TABLE IF NOT EXISTS crm_activities (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source int NOT NULL,
    external_activity_id text NOT NULL,
    name text NOT NULL,
    type int NOT NULL,
    actor_type int NOT NULL,
    actor_id text NOT NULL,
    timestamp bigint NOT NULL,
    properties JSON COLLATE utf8_bin OPTION 'SeekableLZ4' NOT NULL,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at) USING HASH,
    SHARD KEY (project_id, source, type, id),
    KEY (project_id, source, type, external_activity_id, id, timestamp) USING CLUSTERED COLUMNSTORE,
    KEY (user_id) USING HASH,
    KEY (synced) USING HASH,
    UNIQUE KEY project_id_source_id_type_timestamp_unique_idx(project_id,source, id, type, external_activity_id, actor_type, actor_id, timestamp) USING HASH
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Unique (project_id,source, id, type, actor_type, actor_id, timestamp)
);

CREATE TABLE IF NOT EXISTS crm_properties (
    id text NOT NULL,
    project_id bigint NOT NULL,
    source integer NOT NULL,
    `type` integer NOT NULL,
    name text NOT NULL,
    label text,
    external_data_type text,
    mapped_data_type text,
    synced boolean DEFAULT FALSE,
    timestamp bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (updated_at),
    PRIMARY KEY (project_id, id),
    SHARD KEY (project_id),
    UNIQUE KEY crm_properties_project_id_source_type_name_unique_idx(project_id, id, source,`type`,name,timestamp)
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);

CREATE TABLE IF NOT EXISTS crm_settings (
    project_id bigint NOT NULL,
    hubspot_enrich_heavy boolean NOT NULL DEFAULT FALSE,
    hubspot_enrich_heavy_max_created_at bigint,
    PRIMARY KEY (project_id)
    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);

-- DROP DATABASE factors;

CREATE TABLE IF NOT EXISTS dashboard_templates(
    id text NOT NULL,
    title text,
    description text,
    dashboard json,
    units json,
    is_deleted boolean DEFAULT false,
    similar_template_ids json,
    tags json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (id) USING HASH,
    SHARD KEY (id)
);

CREATE TABLE IF NOT EXISTS data_availabilities (
    project_id bigint NOT NULL,
    integration text,
    latest_data_timestamp bigint,
    last_polled timestamp(6) NOT NULL,
    source text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    UNIQUE KEY project_id_integration_unique_idx(project_id,integration) USING HASH
);

CREATE TABLE IF NOT EXISTS clickable_elements (
    project_id bigint NOT NULL,
    id text NOT NULL,
    display_name text NOT NULL,
    element_type text,
    element_attributes json,
    click_count int NOT NULL,
    enabled boolean DEFAULT false,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id, display_name, element_type),
    KEY (project_id, display_name, element_type) USING CLUSTERED COLUMNSTORE,
    UNIQUE KEY(project_id, display_name, element_type) USING HASH
);


CREATE TABLE IF NOT EXISTS ads_import (
    project_id bigint NOT NULL,
    id text NOT NULL,
    status boolean,
    last_processed_index json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL
);

CREATE TABLE IF NOT EXISTS otp_rules(
    id text NOT NULL,
    project_id bigint NOT NULL,
    rule_type text,
    crm_type text,
    touch_point_time_ref text,
    filters json,
    properties_map json,
    is_deleted boolean DEFAULT false,
    created_by text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (id) USING HASH,
    SHARD KEY (id)
);

CREATE TABLE IF NOT EXISTS pathanalysis(
    id TEXT NOT NULL,
    project_id BIGINT NOT NULL,
    title TEXT,
    status TEXT,
    created_by TEXT,
    query JSON,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (project_id, id),
    SHARD KEY(id)
);

CREATE TABLE IF NOT EXISTS form_fills(
    project_id bigint NOT NULL,
    id text NOT NULL,
    user_id text NOT NULL,
    form_id text NOT NULL,
    field_id text NOT NULL,
    value text,
    created_at timestamp(6),
    updated_at timestamp(6),
    PRIMARY KEY (project_id, user_id, form_id, id),
    SHARD KEY (project_id, user_id, form_id)
);


CREATE TABLE IF NOT EXISTS event_trigger_alerts(
    id text NOT NULL,
    project_id bigint NOT NULL,
    title text,
    created_by text,
    event_trigger_alert json,
    last_alert_at timestamp(6),
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS segments(
    id text NOT NULL,
    project_id bigint NOT NULL,
    name text NOT NULL, 
    description text, 
    query json,
    type text,
    PRIMARY KEY (project_id, id),
    SHARD KEY (project_id, id)
);

CREATE TABLE IF NOT EXISTS explain_v2(
    id text NOT NULL,
    project_id bigint NOT NULL,
    title text,
    status text,
    created_by text,
    query json,
    model_id bigint(20),
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (project_id, id)
);

CREATE TABLE IF NOT EXISTS feature_gates (
  project_id bigint,
  hubspot INT DEFAULT 2,
  salesforce INT DEFAULT 2,
  leadsquared INT DEFAULT 2,
  google_ads INT DEFAULT 2,
  facebook INT DEFAULT 2,
  linkedin INT DEFAULT 2,
  google_organic INT DEFAULT 2,
  bing_ads INT DEFAULT 2,
  marketo INT DEFAULT 2,
  drift INT DEFAULT 2,
  clearbit INT DEFAULT 2,
  six_signal INT DEFAULT 1,
  dashboard INT DEFAULT 2,
  offline_touchpoints INT DEFAULT 2,
  saved_queries INT DEFAULT 2,
  explain_feature INT DEFAULT 1,
  filters INT DEFAULT 2,
  shareable_url INT DEFAULT 2,
  custom_metrics INT DEFAULT 2,
  smart_events INT DEFAULT 2,
  templates INT DEFAULT 2,
  smart_properties INT DEFAULT 2,
  content_groups INT DEFAULT 2,
  display_names INT DEFAULT 2,
  weekly_insights INT DEFAULT 1,
  alerts INT DEFAULT 2,
  slack INT DEFAULT 2,
  teams INT DEFAULT 2,
  profiles INT DEFAULT 2,
  segment INT DEFAULT 2,
  path_analysis INT DEFAULT 1,
  archive_events INT DEFAULT 1,
  big_query_upload INT DEFAULT 1,
  import_ads INT DEFAULT 2,
  leadgen INT DEFAULT 2,
  int_shopify INT DEFAULT 2,
  int_adwords INT DEFAULT 2,
  int_google_organic INT DEFAULT 2,
  int_facebook INT DEFAULT 2,
  int_linkedin INT DEFAULT 2,
  int_salesforce INT DEFAULT 2,
  int_hubspot INT DEFAULT 2,
  int_delete INT DEFAULT 2,
  int_slack INT DEFAULT 2,
  int_teams INT DEFAULT 2,
  ds_adwords INT DEFAULT 2,
  ds_google_oraganic INT DEFAULT 2,
  ds_hubspot INT DEFAULT 2,
  ds_facebook INT DEFAULT 2,
  ds_linkedin INT DEFAULT 2,
  ds_metrics INT DEFAULT 2,
  updated_at timestamp(6) NOT NULL,
  SHARD KEY (project_id),
  PRIMARY KEY (project_id)
);

CREATE TABLE IF NOT EXISTS  currency(
    currency varchar(10), 
    date bigint, 
    inr_value double, 
    created_at timestamp(6), 
    updated_at timestamp(6)
);

CREATE TABLE IF NOT EXISTS property_mappings (
    id text NOT NULL,
    project_id bigint NOT NULL,
    name text NOT NULL, 
    display_name text NOT NULL,
    section_bit_map int NOT NULL,
    data_type text NOT NULL,
    properties json,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    is_deleted boolean NOT NULL DEFAULT false,
    KEY (updated_at),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id)
);

CREATE TABLE IF NOT EXISTS display_name_labels (
    project_id bigint NOT NULL,
    id text NOT NULL,
    source text NOT NULL,
    property_key text NOT NULL,
    value text NOT NULL,
    label text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (project_id, source, property_key, value, id) USING CLUSTERED COLUMNSTORE,
    SHARD KEY (project_id, source, id),
    UNIQUE KEY(project_id, source, id, property_key, value) USING HASH
);