package main

// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_db_create.go

import (
	C "factors/config"
	M "factors/model"
	"factors/util"
	"flag"
	"fmt"
	"os"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()

	defer util.NotifyOnPanic("Task#DbCreate", *env)

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config.Env)
	// Initialize configs and connections.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.Error("Failed to initialize.")
		os.Exit(1)
	}

	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Error("Not Development Environment. Aborting")
		return
	}

	db := C.GetServices().Db
	defer db.Close()

	// Create projects table.
	if err := db.CreateTable(&M.Project{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("projects table creation failed.")
	} else {
		log.Info("Created projects table")
	}
	// Add unique index on project tokens.
	if err := db.Exec("CREATE UNIQUE INDEX token_unique_idx ON projects(token);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("projects table token unique indexing failed.")
	} else {
		log.Info("projects table token unique index created.")
	}

	// Add unique index on project private_tokens.
	if err := db.Exec("CREATE UNIQUE INDEX private_token_unique_idx ON projects(private_token);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("projects table private_token unique indexing failed.")
	} else {
		log.Info("projects table private_token unique index created.")
	}

	// Create project settings table.
	if err := db.CreateTable(&M.ProjectSetting{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table creation failed.")
	} else {
		log.Info("Created project_settings table.")
	}
	// Add foreign key constraint by project.
	if err := db.Model(&M.ProjectSetting{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table association with projects table failed.")
	} else {
		log.Info("project_settings table is associated with projects table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.ProjectSetting{}).AddForeignKey("int_adwords_enabled_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings_table table association with agents table failed.")
	} else {
		log.Info("project_settings table is associated with agents table.")
	}

	if err := db.Model(&M.ProjectSetting{}).AddForeignKey("int_salesforce_enabled_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create project_settings int_salesforce_enabled_agent_uuid to agents(uuid) foriegn key.")
	} else {
		log.Info("Created project_settings int_salesforce_enabled_agent_uuid to agents(uuid) foriegn key.")
	}

	// Create users table.
	if err := db.CreateTable(&M.User{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table creation failed.")
	} else {
		log.Info("Created users table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.User{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table association with projects table failed.")
	} else {
		log.Info("users table is associated with projects table.")
	}
	// Add unique index users_project_id_segment_anonymous_uidx.
	if err := db.Exec("CREATE UNIQUE INDEX users_project_id_segment_anonymous_uidx ON users(project_id, segment_anonymous_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table add unique index users_project_id_segment_anonymous_uidx failed")
	} else {
		log.Info("users table unique index users_project_id_segment_anonymous_uidx crated")
	}
	// Add unique index users_project_id_amp_user_idx.
	if err := db.Exec("CREATE UNIQUE INDEX users_project_id_amp_user_idx ON users(project_id, amp_user_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table add unique index users_project_id_amp_user_idx failed")
	} else {
		log.Info("users table unique index users_project_id_amp_user_idx crated")
	}

	// Create user_properties table.
	if err := db.CreateTable(&M.UserProperties{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("user_properties table creation failed.")
	} else {
		log.Info("Created user_propeties table")
	}
	// Add foreign key with projects.
	if err := db.Model(&M.UserProperties{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("user_properties table association with projects table failed.")
	} else {
		log.Info("user_properties table is associated with projects table.")
	}
	// Adding composite foreign key with users table.
	if err := db.Model(&M.UserProperties{}).AddForeignKey("project_id, user_id", "users(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("user_properties table association with users table failed.")
	} else {
		log.Info("user_properties table is associated with users table.")
	}

	if err := db.Exec("CREATE INDEX users_project_id_customer_user_id_idx ON users(project_id, customer_user_id)").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table users_project_id_customer_user_id_idx index creation failed.")
	} else {
		log.Info("Created users table users_project_id_customer_user_id_idx index.")
	}

	// Index for user_property $hubspot_contact_lead_guid.
	if err := db.Exec("CREATE INDEX user_property_hubspot_contact_lead_guid_indx ON user_properties USING gin ((properties->'$hubspot_contact_lead_guid'))").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("user_properties table user_property_hubspot_contact_lead_guid_indx index creation failed.")
	} else {
		log.Info("Created user_properties table user_property_hubspot_contact_lead_guid_indx index.")
	}

	// Create event_names table.
	if err := db.CreateTable(&M.EventName{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table creation failed.")
	} else {
		log.Info("Created event_names table")
	}
	// Add foreign key constraints.
	if err := db.Model(&M.EventName{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table association with projects table failed.")
	} else {
		log.Info("event_names table is associated with projects table.")
	}
	// Adding unique index on project_id, type and filter_expr.
	if err := db.Exec("CREATE UNIQUE INDEX project_type_fexpr_unique_idx ON event_names (project_id, type, filter_expr);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names project type filter_expr unique index creation failed.")
	} else {
		log.Info("Created project type filter_expr unique index created.")
	}
	// Adding unique index on project_id, name, type partial index for non FE type.
	if err := db.Exec("	CREATE UNIQUE INDEX event_names_project_id_name_type_not_fe_uindx ON event_names(project_id, name, type) WHERE type != 'FE';").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names event_names_project_id_name_type_not_fe_uindx unique index creation failed.")
	} else {
		log.Info("Created event_names_project_id_name_type_not_fe_uindx index created.")
	}

	// Create events table.
	if err := db.CreateTable(&M.Event{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table creation failed.")
	} else {
		log.Info("Created events table")
	}

	// Adding unique index on project_id, customer_event_id
	if err := db.Exec("CREATE UNIQUE INDEX project_id_customer_event_id_unique_idx ON events (project_id, customer_event_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events project_id customer_event_id unique index creation failed.")
	} else {
		log.Info("Created project_id customer_event_id unique index created.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.Event{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with projects table failed.")
	} else {
		log.Info("events table is associated with projects table.")
	}
	// Adding composite foreign key with users table.
	if err := db.Model(&M.Event{}).AddForeignKey("project_id, user_id", "users(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with users table failed.")
	} else {
		log.Info("events table is associated with users table.")
	}
	// Adding composite foreign key with event_names table.
	if err := db.Model(&M.Event{}).AddForeignKey("project_id, event_name_id", "event_names(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with event_names table failed.")
	} else {
		log.Info("events table is associated with event_names table.")
	}
	// Add index on project_id, event_name_id, timestamp.
	if err := db.Exec("CREATE INDEX project_id_event_name_id_timestamp_idx ON events(project_id, event_name_id, timestamp DESC);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table project_id:event_name_id:timestamp index failed.")
	} else {
		log.Info("events table project_id:event_name_id:timestamp index created.")
	}
	// Add index on project_id, event_name_id, user_id.
	if err := db.Exec("CREATE INDEX project_id_event_name_id_user_id_idx ON events(project_id, event_name_id, user_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table project_id:event_name_id:user_id index failed.")
	} else {
		log.Info("events table project_id:event_name_id:user_id index created.")
	}

	// Create agents table.
	if err := db.CreateTable(&M.Agent{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("agents table creation failed.")
	} else {
		log.Info("Created agents table")
	}

	// Adding unique index on email.
	if err := db.Exec("CREATE UNIQUE INDEX agent_email_unique_idx ON agents (email);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("agents email unique index creation failed.")
	} else {
		log.Info("Created email unique index on agents table.")
	}

	// Create project_agent_mappings table
	if err := db.CreateTable(&M.ProjectAgentMapping{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table creation failed.")
	} else {
		log.Info("Created project_agent_mappings table")
	}

	// Add sort index on agent_uuid, project_id
	if err := db.Exec("CREATE INDEX project_id_agent_uuid_idx ON project_agent_mappings (project_id, agent_uuid NULLS FIRST) ;").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table agent_uuid_project_id_idx sort index failed.")
	} else {
		log.Info("events table agent_uuid_project_id_idx sort index created.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.ProjectAgentMapping{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table association with projects table failed.")
	} else {
		log.Info("project_agent_mappings table is associated with projects table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.ProjectAgentMapping{}).AddForeignKey("agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table association with agents table failed.")
	} else {
		log.Info("project_agent_mappings table is associated with agents table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.ProjectAgentMapping{}).AddForeignKey("invited_by", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table association with agents table failed.")
	} else {
		log.Info("project_agent_mappings table is associated with agents table.")
	}

	// Create dashboard table.
	if err := db.CreateTable(&M.Dashboard{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard table creation failed.")
	} else {
		log.Info("Created dashboard table")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.Dashboard{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard table association with projects table failed.")
	} else {
		log.Info("dashboard table is associated with projects table.")
	}

	// Adding unique index on dashboards id, project_id for dashboard units foreign referrence.
	if err := db.Exec("CREATE UNIQUE INDEX project_id_id_unique_idx ON dashboards (project_id, id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create unique index dashboards(project_id, id).")
	} else {
		log.Info("Created unique index on dashboards(project_id, id).")
	}

	// Create dashboard_unit table.
	if err := db.CreateTable(&M.DashboardUnit{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard_unit table creation failed.")
	} else {
		log.Info("Created dashboard_unit table")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.DashboardUnit{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard_unit table association with projects table failed.")
	} else {
		log.Info("dashboard_unit table is associated with projects table.")
	}

	// Add foreign key constraints dashboard_id, dashboards(project_id, id).
	if err := db.Model(&M.DashboardUnit{}).AddForeignKey("project_id, dashboard_id", "dashboards(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard_unit table association with dashboards table failed.")
	} else {
		log.Info("dashboard_unit table is associated with dashboards table.")
	}

	// Create billing_accounts table
	if err := db.CreateTable(&M.BillingAccount{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("billing_accounts table creation failed.")
	} else {
		log.Info("Created billing_accounts table")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.BillingAccount{}).AddForeignKey("agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("billing_accounts table association with agents table failed.")
	} else {
		log.Info("billing_accounts table is associated with projects table.")
	}

	// Adding unique index on billing_accounts agent_uuid.
	if err := db.Exec("CREATE UNIQUE INDEX agent_uuid_unique_idx ON billing_accounts(agent_uuid);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create unique index billing_accounts(agent_uuid).")
	} else {
		log.Info("Created unique index on billing_accounts(agent_uuid).")
	}

	// Create project_billing_account_mappings table
	if err := db.CreateTable(&M.ProjectBillingAccountMapping{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_billing_account_mappings table creation failed.")
	} else {
		log.Info("project_billing_account_mappings table creation failed.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.ProjectBillingAccountMapping{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_billing_account_mappings table association with projects table failed.")
	} else {
		log.Info("project_billing_account_mappings table is associated with projects table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.ProjectBillingAccountMapping{}).AddForeignKey("billing_account_id", "billing_accounts(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_billing_account_mappings table association with billing_accounts table failed.")
	} else {
		log.Info("project_billing_account_mappings table is associated with billing_accounts table.")
	}

	// Add sort index on billing_account_id, project_id
	if err := db.Exec("CREATE INDEX billing_account_id_project_id_idx ON project_billing_account_mappings (billing_account_id ,project_id) ;").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_billing_account_mappings table billing_account_id_project_id_idx sort index failed.")
	} else {
		log.Info("project_billing_account_mappings table billing_account_id_project_id_idx sort index created.")
	}

	// Create reports table
	if err := db.CreateTable(&M.DBReport{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("reports table creation failed.")
	} else {
		log.Info("reports table creation failed.")
	}

	// Adding index on reports table.
	if err := db.Exec("CREATE INDEX report_project_id_dashboard_id_type_st_et_unique_idx ON reports(project_id, dashboard_id, type, start_time, end_time);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create unique index report_project_id_dashboard_id_type_st_et_unique_idx(project_id, dashboard_id, type, start_time, end_time).")
	} else {
		log.Info("Created index on reports(project_id, dashboard_id, type, start_time, end_time).")
	}

	// Add foreign key constraints.
	if err := db.Model(&M.DBReport{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("reports table association with projects table failed.")
	} else {
		log.Info("reports table is associated with projects table.")
	}

	// Add foreign key constraints dashboards(project_id, id).
	if err := db.Model(&M.DBReport{}).AddForeignKey("project_id, dashboard_id", "dashboards(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("reports table association with dashboards table failed.")
	} else {
		log.Info("reports table is associated with dashboards table.")
	}

	// Create adwords document table
	if err := db.CreateTable(&M.AdwordsDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("adwords document table creation failed.")
	} else {
		log.Info("created adwords document table.")
	}

	// Create hubspot documents table.
	if err := db.CreateTable(&M.HubspotDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("hubspot_documents table creation failed.")
	} else {
		log.Info("created hubspot documents table.")
	}

	// Create facebook documents table
	if err := db.CreateTable(&M.FacebookDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("facebook_documents table creation failed.")
	} else {
		log.Info("created facebook documents table.")
	}

	// Add foreign key constraints agents(uuid).
	if err := db.Model(&M.ProjectSetting{}).AddForeignKey("int_facebook_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table association with agents table failed.")
	} else {
		log.Info("project_settings table is associated with agents table.")
	}

	// Create salesforce_documents documents table
	if err := db.CreateTable(&M.SalesforceDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("salesforce_documents table creation failed.")
	} else {
		log.Info("created salesforce_documents documents table.")
	}

	// Add foreign key constraints projects(id).
	if err := db.Model(&M.SalesforceDocument{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("salesforce_documents table association with projects table failed.")
	} else {
		log.Info("salesforce_documents table is associated with projects table.")
	}

	// Create bigquery settings table.
	if err := db.CreateTable(&M.BigquerySetting{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("bigquery_settings table creation failed.")
	} else {
		log.Info("created bigquery_settings table.")
	}

	// Create scheduled_tasks table.
	if err := db.CreateTable(&M.ScheduledTask{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("scheduled_tasks table creation failed.")
	} else {
		log.Info("created scheduled_tasks table.")
	}

	// Update auto vacuum and analyze settings for heavy tables.
	for _, tableName := range [...]string{"events", "users", "user_properties", "hubspot_documents", "adwords_documents", "event_names"} {
		query := fmt.Sprintf("ALTER TABLE %s SET (autovacuum_vacuum_scale_factor = 0.02, autovacuum_analyze_scale_factor = 0.01);", tableName)
		if err := db.Exec(query).Error; err != nil {
			log.WithError(err).Errorf("Failed to update vacuum config for %s", tableName)
		} else {
			log.Infof("Updated config for %s table", tableName)
		}
	}
}
