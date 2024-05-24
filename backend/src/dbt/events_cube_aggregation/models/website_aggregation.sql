{{ config(
    materialized='incremental',
    sort_key=['project_id','event_type','timestamp_at_day'],
    pre_hook = '
        {% if is_incremental() %}
            Delete from {{this}} where project_id = {{var(\'project_id\')}} 
            AND timestamp_at_day BETWEEN {{ var(\'from\') }} AND {{ var(\'to\') }}
        {% endif %}
        ',
    )
}}

/*
    TODO Consideration to keep the system performant - 
    TODO How to get ready for debugging.

    Issues faced -
    Tmp table being created in between is a rowstore and it doesnt allow sort key.

    Conversion logic time is being considered in that timezone and respective date is assigned.
    Eg - 1stOct 4:30 IST will be assigned 1stOct if Asia/Kolkata is considered and 30Th Sept if UTC is considered.
    Epoch will be same between the different timezones as date trunc will not have any timezone and epoch will be presented as if its UTC.
*/

with session_event_name as (
    select id, project_id from event_names where name = '$session' AND project_id = {{ var('project_id') }}
),

session_data as (
SELECT
    events.project_id as project_id,
    CONVERT(UNIX_TIMESTAMP(date_trunc('day', CONVERT_TZ(FROM_UNIXTIME(timestamp), 'UTC', '{{ var('time_zone') }}')) ), UNSIGNED) as timestamp_at_day,

    'session' as event_name,
    'session' as event_type,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$source\')') }} as source,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$medium\')') }} as medium,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$campaign\')') }} as campaign,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$initial_referrer_url\')') }} as referrer_url,    
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$initial_page_url\')') }} as landing_page_url,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$session_latest_page_url\')') }} as latest_page_url,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$country\')') }} as country,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$region\')') }} as region,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$city\')') }} as city,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$browser\')') }} as browser,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$browser_version\')') }} as browser_version,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$os\')') }} as os,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$os_version\')') }} as os_version,

    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$device_name\')') }} as device,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_industry\')') }} as 6signal_industry,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_employee_range\')') }} as 6signal_employee_range,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_revenue_range\')') }} as 6signal_revenue_range,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_naics_description\')') }} as 6signal_naics_description,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_sic_description\')') }} as 6signal_sic_description,

    count(*) as count_of_records,
    sum({{ null_or_empty_to_zero('JSON_EXTRACT_STRING(events.properties, \'$session_spent_time\')' ) }}) as spent_time,
    max(events.updated_at) as max_updated_at

FROM 
    events INNER JOIN session_event_name ON events.project_id = session_event_name.project_id AND events.event_name_id = session_event_name.id
    AND events.timestamp BETWEEN {{ var('from') }} AND {{ var('to') }}

GROUP BY events.project_id, timestamp_at_day, source, medium, campaign, landing_page_url, referrer_url, country, region,
    city, browser, browser_version, os, os_version, 6signal_industry, 6signal_employee_range, 6signal_revenue_range,
    6signal_naics_description, 6signal_sic_description
),

page_view_event_name as (
    select id, name, project_id from event_names where type = 'AT' AND project_id = {{ var('project_id') }}
), 

page_view_data as (
SELECT
    events.project_id as project_id,
    CONVERT(UNIX_TIMESTAMP(date_trunc('day', CONVERT_TZ(FROM_UNIXTIME(timestamp), 'UTC', '{{ var('time_zone') }}')) ), UNSIGNED) as timestamp_at_day,

    page_view_event_name.name as event_name,
    'page_view' as event_type,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$source\')') }} as source,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$medium\')') }} as medium,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$campaign\')') }} as campaign,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.properties, \'$referrer_url\')') }} as referrer_url,    
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$initial_page_url\')') }} as landing_page_url,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$latest_page_url\')') }} as latest_page_url,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$country\')') }} as country,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$region\')') }} as region,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$city\')') }} as city,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$browser\')') }} as browser,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$browser_version\')') }} as browser_version,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$os\')') }} as os,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$os_version\')') }} as os_version,
    
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$device_name\')') }} as device,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_industry\')') }} as 6signal_industry,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_employee_range\')') }} as 6signal_employee_range,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_revenue_range\')') }} as 6signal_revenue_range,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_naics_description\')') }} as 6signal_naics_description,
    {{ null_or_empty_to_none('JSON_EXTRACT_STRING(events.user_properties, \'$6Signal_sic_description\')') }} as 6signal_sic_description,

    count(*) as count_of_records,
    sum({{ null_or_empty_to_zero('JSON_EXTRACT_STRING(events.properties, \'$page_spent_time\') ') }}) as spent_time,
    max(events.updated_at) as max_updated_at

FROM 
    events INNER JOIN page_view_event_name ON events.project_id = page_view_event_name.project_id AND events.event_name_id = page_view_event_name.id
    AND events.timestamp BETWEEN {{ var('from') }} AND {{ var('to') }}

GROUP BY events.project_id, timestamp_at_day, source, medium, campaign, landing_page_url, referrer_url, country, region,
    city, browser, browser_version, os, os_version, 6signal_industry, 6signal_employee_range, 6signal_revenue_range,
    6signal_naics_description, 6signal_sic_description
),

union_data as (
    SELECT * FROM session_data UNION ALL SELECT * FROM page_view_data
)

select * from union_data