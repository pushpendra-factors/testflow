-- UP
CREATE DATABASE IF NOT EXISTS factors;

USE factors;

CREATE TABLE IF NOT EXISTS events (
    id text NOT NULL,
    project_id bigint NOT NULL,
    customer_event_id text, 
    user_id text,
    user_properties_id text, 
    event_name_id int, 
    count bigint,
    properties json,
    session_id text,
    timestamp bigint,
    properties_updated_timestamp bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (project_id, event_name_id, timestamp) USING CLUSTERED COLUMNSTORE,
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

CREATE TABLE IF NOT EXISTS users (
    id text NOT NULL, 
    project_id bigint NOT NULL, 
    customer_user_id text,
    segment_anonymous_id text,
    amp_user_id text,
    properties_id text,
    join_timestamp bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    -- COLUMNSTORE key is sort key, can we add an incremental numerical column to the end?
    -- Initial parts of the indices are still useful when don't use the last column which is an incremental value.
    KEY (project_id, customer_user_id) USING CLUSTERED COLUMNSTORE,
    SHARD KEY (id),
    UNIQUE KEY unique_id_idx (id) USING HASH

    -- Required constraints.
    -- Unique (project_id, segment_anonymous_id)
    -- Unique (project_id, amp_user_id)
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, properties_id) -> user_properties(project_id, id)
);

CREATE TABLE IF NOT EXISTS user_properties (
    id text NOT NULL,
    project_id bigint NOT NULL,
    user_id text,
    properties json,
    updated_timestamp bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (project_id, user_id) USING CLUSTERED COLUMNSTORE,
    SHARD KEY (user_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, user_id) -> users(project_id, id)

    -- Missing index.
    -- Index (project_id, properties::$hubspot_contact_lead_guid)
);

CREATE TABLE IF NOT EXISTS event_names (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    name text,
    type varchar(2),
    filter_expr varchar(500),
    deleted bool NOT NULL DEFAULT false,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
    UNIQUE KEY project_id_type_filter_expr_idx(project_id, type, filter_expr)

    -- Required constraints.
    -- Unique (project_id, name, type) WHERE type != 'FE'
    -- Ref (project_id) -> projects(id)
);


CREATE TABLE IF NOT EXISTS adwords_documents (
    id text,
    project_id bigint,
    customer_account_id text,
    type int,
    timestamp bigint,
    value json,
    ad_group_id bigint,
    ad_id bigint,
    keyword_id bigint,
    campaign_id bigint,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, customer_account_id, type, timestamp, id)
    
    -- Required constraints.
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
    last_logged_in_at datetime,
    login_count bigint,
    subscribe_newsletter bool,
    int_adwords_refresh_token text,
    int_salesforce_refresh_token text,
    int_salesforce_instance_url text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (uuid),
    PRIMARY KEY (uuid),
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
    PRIMARY KEY (project_id, id),
    UNIQUE KEY (project_id, bq_project_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
);

CREATE TABLE IF NOT EXISTS billing_accounts (
    id bigint AUTO_INCREMENT,
    plan_id bigint,
    agent_uuid text,
    organization_name text,
    billing_address text,
    pincode text,
    phone_no text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (agent_uuid),
    PRIMARY KEY (agent_uuid, id),
    UNIQUE KEY (agent_uuid)

    -- Required constraints.
    -- Ref (agent_uuid) -> agents(id)
);

CREATE TABLE IF NOT EXISTS dashboard_units (
    id bigint AUTO_INCREMENT,
    project_id bigint,
    dashboard_id bigint,
    title text,
    description text,
    presentation varchar(5),
    query json,
    query_id bigint,
    settings json,
    is_deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
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
    is_deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(6),
    updated_at timestamp(6),
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id),
	KEY (project_id, agent_uuid),
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
    value json,
    campaign_id text,
    ad_set_id text,
    ad_id text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, customer_ad_account_id, platform, type, timestamp, id)

    -- Required constraints.
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
    event_name_id bigint,
    type varchar(2),
    created_by text,
    last_tracked_at timestamp(6), 
    is_active boolean,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
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
    timestamp timestamp(6),
    value json,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text,
    user_id text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id, type, action, timestamp),
    KEY project_id_type_timestamp_user_id_idx(project_id, type, timestamp DESC, user_id),
    KEY project_id_type_user_id_timestamp_idx(project_id, type, user_id, timestamp DESC)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (project_id, user_id) -> users(project_id, id)
);

CREATE TABLE IF NOT EXISTS project_agent_mappings (
    project_id bigint,
    agent_uuid text,
    role bigint,
    invited_by text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, agent_uuid)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (agent_uuid) -> agents(uuid)
    -- Ref (invited_by) -> agents(uuid)
);

CREATE TABLE IF NOT EXISTS project_billing_account_mappings (
    project_id bigint,
    billing_account_id bigint,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, billing_account_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (billing_account_id) -> billing_accounts(id)
);

CREATE TABLE IF NOT EXISTS project_settings (
    project_id bigint,
    auto_track boolean NOT NULL DEFAULT FALSE, 
    auto_form_capture boolean NOT NULL DEFAULT FALSE,
    exclude_bot boolean NOT NULL DEFAULT FALSE,
    int_segment boolean NOT NULL DEFAULT FALSE,
    int_adwords_enabled_agent_uuid text,
    int_adwords_customer_account_id text,
    int_hubspot boolean NOT NULL DEFAULT FALSE,
    int_hubspot_api_key text,
    int_facebook_email text,
    int_facebook_access_token text,
    int_facebook_agent_uuid text,
    int_facebook_user_id text,
    int_facebook_ad_account text,
    int_linkedin_ad_account text,
    int_linkedin_access_token text,
    int_linkedin_access_token_expiry bigint,
    int_linkedin_refresh_token text,
    int_linkedin_refresh_token_expiry bigint,
    int_linkedin_agent_uuid text,
    archive_enabled boolean NOT NULL DEFAULT FALSE,
    bigquery_enabled boolean NOT NULL DEFAULT FALSE,
    int_salesforce_enabled_agent_uuid text,
    int_drift boolean NOT NULL DEFAULT FALSE, 
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
    -- Ref (int_adwords_enabled_agent_uuid) -> agents(uuid)
    -- Ref (int_facebook_agent_uuid) -> agents(uuid)
    -- Ref (int_linkedin_agent_uuid) -> agents(uuid)
    -- Ref (int_salesforce_enabled_agent_uuid) -> agents(uuid)
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
    jobs_metadata json,
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
    value json,
    synced boolean NOT NULL DEFAULT FALSE,
    sync_id text, 
    user_id text,
    created_at timestamp(6) NOT NULL, 
    updated_at timestamp(6) NOT NULL,
    SHARD KEY (project_id),
    PRIMARY KEY (project_id, id, type, timestamp),
    KEY project_id_type_timestamp_user_id_idx(project_id, type, timestamp DESC, user_id),
    KEY project_id_type_user_id_timestamp_idx(project_id, type, user_id, timestamp DESC)

    -- Required constraints.
    -- Ref (project_id) -> projects(id)
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
    value json,
    creative_id text,
    campaign_group_id text,
    campaign_id text,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL
);

CREATE TABLE IF NOT EXISTS property_details(
    project_id bigint NOT NULL,
    event_name_id bigint null,
    `key` text  NOT NULL,
    `type` text  NOT NULL,
    entity integer  NOT NULL
);

-- DOWN

-- DROP DATABASE factors;