#!/bin/sh

# Function to recreate redis-db
function recreate_redis_database() {
    redis-cli -h "$VM_HOST" -p "$VM_REDIS_PORT" FLUSHDB
    if [ $? -eq 0 ]; then
        echo "VM's Redis database dropped."
    else
        echo "Failed to drop the VM's Redis database."
    fi
}

# Function to recreate sql db
function recreate_sql_db() {
    mysql -h $VM_HOST  --port 3306  -u root -pdbfactors123 -e "DROP DATABASE IF EXISTS factors;"
    drop_status=$?
    
    mysql -h $VM_HOST --port 3306  -u root -pdbfactors123 -e "CREATE DATABASE factors;"
    create_status=$?
    
    if [ $drop_status -eq 0 ] && [ $create_status -eq 0 ]; then
        echo "Database dropped and recreated successfully."
    else
        echo "Failed to drop and recreate the database."
    fi
}


# Check if the flag is set to recreate vm's redis-db and sql
if [ "$RECREATE" = true ]; then
    recreate_redis_database
    recreate_sql_db
fi

# Fetch keys matching the pattern from the source Redis server
keys=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" --raw keys "*:pid:$PROJECT_ID:*")

# Transfer keys to the destination Redis server using MIGRATE with COPY option
for key in $keys; do
    # Migrate each key to the destination Redis server using COPY option
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" MIGRATE "$VM_HOST" "$VM_REDIS_PORT" "$key" 0 5000 COPY
done

if [ $? -eq 0 ]; then
    echo "Keys transferred successfully to VM"
else
    echo "Failed to transfer Keys to VM"
fi

# Create schema 
mysql -h $VM_HOST --port 3306  -u root -pdbfactors123 < "1_create_schema.sql"

sleep 15

mysqldump  -u $DB_USER_NAME -h $DB_HOST -p$DB_PASSWORD factors projects --where="id=$PROJECT_ID" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump  -u $DB_USER_NAME -h $DB_HOST -p$DB_PASSWORD factors account_scoring_ranges ads_import adwords_documents alerts bigquery_settings clickable_elements content_groups crm_activities crm_groups crm_properties crm_relationships crm_settings crm_users custom_metrics dash_query_results dashboard_units dashboards data_availabilities display_name_labels display_names event_names event_properties_json event_trigger_alerts explain_v2 facebook_documents feature_gates feedbacks fivetran_mappings form_fills g2_documents  group_relationships groups  integration_documents leadgen_settings leadsquared_markers linkedin_documents otp_rules pathanalysis project_agent_mappings project_billing_account_mappings project_model_metadata project_plan_mappings project_settings property_details property_mappings property_overrides queries salesforce_documents scheduled_tasks segments shareable_url_audits shareable_urls smart_properties smart_property_rules task_execution_details templates upload_filter_files user_properties_json users website_aggregation weekly_insights_metadata --where="project_id=$PROJECT_ID" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump  -u $DB_USER_NAME -h $DB_HOST -p$DB_PASSWORD factors currency dashboard_templates plan_details | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump  -u $DB_USER_NAME -h $DB_HOST -p$DB_PASSWORD factors agents --where="uuid IN (select agent_uuid from project_agent_mappings where project_id=$PROJECT_ID)" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump  -u $DB_USER_NAME -h $DB_HOST -p$DB_PASSWORD factors  billing_accounts --where="id IN (select billing_account_id from project_billing_account_mappings where project_id=$PROJECT_ID)" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump -u $DB_USER_NAME  -h $DB_HOST -p$DB_PASSWORD factors events  --where="project_id=$PROJECT_ID" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump -u $DB_USER_NAME  -h $DB_HOST -p$DB_PASSWORD factors google_organic_documents  --where="project_id=$PROJECT_ID" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

mysqldump -u $DB_USER_NAME  -h $DB_HOST -p$DB_PASSWORD factors hubspot_documents --where="project_id=$PROJECT_ID" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors

wait -f

# mysqldump -u $DB_USER_NAME  -h $DB_HOST -p$DB_PASSWORD factors factors_goals factors_tracked_events factors_tracked_user_properties --where="project_id=$PROJECT_ID" | mysql -h $VM_HOST --port 3306 -u root -pdbfactors123 -D factors