package main

// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_db_create.go

import (
	C "factors/config"
	"factors/model/model"
	"factors/util"
	"flag"
	"fmt"
	"os"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	isDevSetup := flag.Bool("dev_setup", false, "")

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

	C.InitConf(config)
	// Initialize configs and connections.
	err := C.InitDB(*config)
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
	if err := db.CreateTable(&model.Project{}).Error; err != nil {
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
	if err := db.CreateTable(&model.ProjectSetting{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table creation failed.")
	} else {
		log.Info("Created project_settings table.")
	}
	// Add foreign key constraint by project.
	if err := db.Model(&model.ProjectSetting{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table association with projects table failed.")
	} else {
		log.Info("project_settings table is associated with projects table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.ProjectSetting{}).AddForeignKey("int_adwords_enabled_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings_table table association with agents table failed.")
	} else {
		log.Info("project_settings table is associated with agents table.")
	}

	if err := db.Model(&model.ProjectSetting{}).AddForeignKey("int_salesforce_enabled_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create project_settings int_salesforce_enabled_agent_uuid to agents(uuid) foriegn key.")
	} else {
		log.Info("Created project_settings int_salesforce_enabled_agent_uuid to agents(uuid) foriegn key.")
	}

	// Create users table.
	if err := db.CreateTable(&model.User{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table creation failed.")
	} else {
		log.Info("Created users table")
	}
	// Add foreign key constraints.
	if err := db.Model(&model.User{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
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

	if err := db.Exec("CREATE INDEX users_project_id_customer_user_id_idx ON users(project_id, customer_user_id)").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("users table users_project_id_customer_user_id_idx index creation failed.")
	} else {
		log.Info("Created users table users_project_id_customer_user_id_idx index.")
	}

	// Create event_names table.
	if err := db.CreateTable(&model.EventName{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("event_names table creation failed.")
	} else {
		log.Info("Created event_names table")
		if err := db.Exec("ALTER TABLE event_names ALTER COLUMN id TYPE bigint USING id::bigint;").Error; err != nil {
			log.WithError(err).Error("Failed to alter type of id on event_names.")
		} else {
			log.Info("Altered type of id of event_names table.")
		}
	}
	// Add foreign key constraints.
	if err := db.Model(&model.EventName{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
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
	if err := db.CreateTable(&model.Event{}).Error; err != nil {
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
	if err := db.Model(&model.Event{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with projects table failed.")
	} else {
		log.Info("events table is associated with projects table.")
	}
	// Adding composite foreign key with users table.
	if err := db.Model(&model.Event{}).AddForeignKey("project_id, user_id", "users(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("events table association with users table failed.")
	} else {
		log.Info("events table is associated with users table.")
	}
	// Adding composite foreign key with event_names table.
	if err := db.Model(&model.Event{}).AddForeignKey("project_id, event_name_id", "event_names(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
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
	if err := db.CreateTable(&model.Agent{}).Error; err != nil {
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
	if err := db.CreateTable(&model.ProjectAgentMapping{}).Error; err != nil {
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
	if err := db.Model(&model.ProjectAgentMapping{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table association with projects table failed.")
	} else {
		log.Info("project_agent_mappings table is associated with projects table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.ProjectAgentMapping{}).AddForeignKey("agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table association with agents table failed.")
	} else {
		log.Info("project_agent_mappings table is associated with agents table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.ProjectAgentMapping{}).AddForeignKey("invited_by", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_agent_mappings table association with agents table failed.")
	} else {
		log.Info("project_agent_mappings table is associated with agents table.")
	}

	// Create dashboard table.
	if err := db.CreateTable(&model.Dashboard{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard table creation failed.")
	} else {
		log.Info("Created dashboard table")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.Dashboard{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
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
	if err := db.CreateTable(&model.DashboardUnit{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard_unit table creation failed.")
	} else {
		log.Info("Created dashboard_unit table")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.DashboardUnit{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard_unit table association with projects table failed.")
	} else {
		log.Info("dashboard_unit table is associated with projects table.")
	}

	// Add foreign key constraints dashboard_id, dashboards(project_id, id).
	if err := db.Model(&model.DashboardUnit{}).AddForeignKey("project_id, dashboard_id", "dashboards(project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("dashboard_unit table association with dashboards table failed.")
	} else {
		log.Info("dashboard_unit table is associated with dashboards table.")
	}

	// Create billing_accounts table
	if err := db.CreateTable(&model.BillingAccount{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("billing_accounts table creation failed.")
	} else {
		log.Info("Created billing_accounts table")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.BillingAccount{}).AddForeignKey("agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
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
	if err := db.CreateTable(&model.ProjectBillingAccountMapping{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_billing_account_mappings table creation failed.")
	} else {
		log.Info("project_billing_account_mappings table creation failed.")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.ProjectBillingAccountMapping{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_billing_account_mappings table association with projects table failed.")
	} else {
		log.Info("project_billing_account_mappings table is associated with projects table.")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.ProjectBillingAccountMapping{}).AddForeignKey("billing_account_id", "billing_accounts(id)", "RESTRICT", "RESTRICT").Error; err != nil {
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

	// Create adwords document table
	if err := db.CreateTable(&model.AdwordsDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("adwords document table creation failed.")
	} else {
		log.Info("created adwords document table.")
	}

	// Create hubspot documents table.
	if err := db.CreateTable(&model.HubspotDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("hubspot_documents table creation failed.")
	} else {
		log.Info("created hubspot documents table.")
	}

	// Adding hubspot_documents_project_id_type_user_id_timestamp_idx index.
	if err := db.Exec("CREATE INDEX hubspot_documents_project_id_type_timestamp_user_id_idx ON hubspot_documents(project_id, type, timestamp DESC, user_id)").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create index hubspot_documents_project_id_type_timestamp_user_id_idx.")
	} else {
		log.Info("Created index hubspot_documents_project_id_type_timestamp_user_id_idx.")
	}

	// Adding hubspot_documents_project_id_type_user_id_timestamp_idx index.
	if err := db.Exec("CREATE INDEX hubspot_documents_project_id_type_user_id_timestamp_idx ON hubspot_documents(project_id, type, user_id, timestamp DESC)").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create index hubspot_documents_project_id_type_user_id_timestamp_idx.")
	} else {
		log.Info("Created index hubspot_documents_project_id_type_user_id_timestamp_idx.")
	}

	// Create facebook documents table
	if err := db.CreateTable(&model.FacebookDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("facebook_documents table creation failed.")
	} else {
		log.Info("created facebook documents table.")
	}

	// Add foreign key constraints agents(uuid).
	if err := db.Model(&model.ProjectSetting{}).AddForeignKey("int_facebook_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table association with agents table failed.")
	} else {
		log.Info("project_settings table is associated with agents table.")
	}

	// Create linkedin documents table
	if err := db.CreateTable(&model.LinkedinDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("linkedin_documents table creation failed.")
	} else {
		log.Info("created linkedin documents table.")
	}

	// Add foreign key constraints agents(uuid).
	if err := db.Model(&model.ProjectSetting{}).AddForeignKey("int_linkedin_agent_uuid", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_settings table association with agents table failed.")
	} else {
		log.Info("project_settings table is associated with agents table.")
	}

	// Create salesforce_documents documents table
	if err := db.CreateTable(&model.SalesforceDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("salesforce_documents table creation failed.")
	} else {
		log.Info("created salesforce_documents documents table.")
	}

	// Add foreign key constraints projects(id).
	if err := db.Model(&model.SalesforceDocument{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("salesforce_documents table association with projects table failed.")
	} else {
		log.Info("salesforce_documents table is associated with projects table.")
	}

	// Adding salesforce_documents_project_id_type_timestamp_user_id_idx index on salesforce_documents table.
	if err := db.Exec("CREATE INDEX salesforce_documents_project_id_type_timestamp_user_id_idx ON salesforce_documents(project_id, type, timestamp DESC, user_id)").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create index salesforce_documents_project_id_type_timestamp_user_id_idx.")
	} else {
		log.Info("Created index salesforce_documents_project_id_type_timestamp_user_id_idx.")
	}

	// Adding salesforce_documents_project_id_type_user_id_timestamp_idx index on salesforce_documents table.
	if err := db.Exec("CREATE INDEX salesforce_documents_project_id_type_user_id_timestamp_idx ON salesforce_documents(project_id, type, user_id, timestamp DESC)").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create index salesforce_documents_project_id_type_user_id_timestamp_idx.")
	} else {
		log.Info("Created index salesforce_documents_project_id_type_user_id_timestamp_idx.")
	}

	// Create bigquery settings table.
	if err := db.CreateTable(&model.BigquerySetting{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("bigquery_settings table creation failed.")
	} else {
		log.Info("created bigquery_settings table.")
	}
	// Create queries table.
	if err := db.CreateTable(&model.Queries{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("queries table creation failed.")
	} else {
		log.Info("Created queries table")
	}

	// Add foreign key constraints.
	if err := db.Model(&model.Queries{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("queries table association with projects table failed.")
	} else {
		log.Info("queries table is associated with projects table.")
	}
	// Add index on queries table
	if err := db.Exec("CREATE INDEX queries_title_trgm_idx ON queries USING gin (title gin_trgm_ops);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create index on queries(title).")
	} else {
		log.Info("Created index on queries(title).")
	}

	// Create scheduled_tasks table.
	if err := db.CreateTable(&model.ScheduledTask{}).Error; err != nil {
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

	// tracked events
	if err := db.CreateTable(&model.FactorsTrackedEvent{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_events table creation failed.")
	} else {
		log.Info("created factors_tracked_events table.")
	}

	// Adding unique index.
	if err := db.Exec("CREATE UNIQUE INDEX projectid_eventnameid_unique_idx ON factors_tracked_events (project_id, event_name_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_Events unique index creation failed ")
	} else {
		log.Info("Created unique index on tracked events table")
	}

	// Add foreign key constraint
	if err := db.Model(&model.FactorsTrackedEvent{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_events table association with projects table failed.")
	} else {
		log.Info("factors_tracked_events table is associated with projects table.")
	}

	// Add foreign key constraint
	if err := db.Model(&model.FactorsTrackedEvent{}).AddForeignKey("created_by", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_events table association with agents table failed.")
	} else {
		log.Info("factors_tracked_events table is associated with agents table.")
	}

	// tracked user properties
	if err := db.CreateTable(&model.FactorsTrackedUserProperty{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_user_properties table creation failed.")
	} else {
		log.Info("created factors_tracked_user_properties table.")
	}

	// Adding unique index.
	if err := db.Exec("CREATE UNIQUE INDEX projectid_userpropertyname_unique_idx ON factors_tracked_user_properties (project_id, user_property_name);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_user_properties unique index creation failed ")
	} else {
		log.Info("Created unique index on tracked user properties table")
	}

	// Add foreign key constraint
	if err := db.Model(&model.FactorsTrackedUserProperty{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_user_properties table association with projects table failed.")
	} else {
		log.Info("factors_tracked_user_properties table is associated with projects table.")
	}

	// Add foreign key constraint
	if err := db.Model(&model.FactorsTrackedUserProperty{}).AddForeignKey("created_by", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_tracked_user_properties table association with agents table failed.")
	} else {
		log.Info("factors_tracked_user_properties table is associated with agents table.")
	}

	// saved goals
	if err := db.CreateTable(&model.FactorsGoal{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_goals table creation failed.")
	} else {
		log.Info("created factors_goals table.")
	}

	// Adding unique index.
	if err := db.Exec("CREATE UNIQUE INDEX name_projectId_unique_idx ON factors_goals (project_id, \"name\");").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_goals unique index creation failed ")
	} else {
		log.Info("Created unique index on factors_goals table")
	}

	// Add foreign key constraint
	if err := db.Model(&model.FactorsGoal{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_goals table association with projects table failed.")
	} else {
		log.Info("factors_goals table is associated with projects table.")
	}

	// Add foreign key constraint
	if err := db.Model(&model.FactorsGoal{}).AddForeignKey("created_by", "agents(uuid)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("factors_goals table association with agents table failed.")
	} else {
		log.Info("factors_goals table is associated with agents table.")
	}

	// Create property_details table.
	if err := db.CreateTable(&model.PropertyDetail{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("property_details table creation failed.")
	} else {
		log.Info("Created configured properties table.")
	}

	// Add foreign key constraint
	if err := db.Model(&model.PropertyDetail{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("property_details table association with projects table failed.")
	} else {
		log.Info("property_details table is associated with projects table.")
	}

	// Add foreign key constraint
	if err := db.Model(&model.PropertyDetail{}).AddForeignKey("project_id, event_name_id", "event_names (project_id, id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("property_details table association with events table failed.")
	} else {
		log.Info("property_details table is associated with events table.")
	}

	// Create smart_property_rules table
	if err := db.CreateTable(&model.SmartPropertyRules{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("smart_property_rules table creation failed.")
	} else {
		log.Info("Created smart properties rules table.")
	}

	// Create smart_properties table
	if err := db.CreateTable(&model.SmartProperties{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("smart_properties table creation failed.")
	} else {
		log.Info("Created smart properties table.")
	}

	// Create google_organic_documents table
	if err := db.CreateTable(&model.GoogleOrganicDocument{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("google_organic_documents table creation failed.")
	} else {
		log.Info("Created google_organic_documents table.")
	}
	// task management
	if err := db.CreateTable(&model.TaskDetails{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_details table creation failed.")
	} else {
		log.Info("Created task_details table.")
	}

	if err := db.CreateTable(&model.TaskExecutionDependencyDetails{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_dependency_Details table creation failed.")
	} else {
		log.Info("Created task_dependency_Details table.")
	}

	// Add foreign key constraint by TaskId.
	if err := db.Model(&model.TaskExecutionDependencyDetails{}).AddForeignKey("task_id", "task_details(task_id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_dependency_Details table association with projects table failed.")
	} else {
		log.Info("task_dependency_Details table is associated with projects table.")
	}

	// Add foreign key constraint by DependenctTaskId.
	if err := db.Model(&model.TaskExecutionDependencyDetails{}).AddForeignKey("dependent_task_id", "task_details(task_id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_dependency_Details table association with projects table failed.")
	} else {
		log.Info("task_dependency_Details table is associated with projects table.")
	}

	if err := db.CreateTable(&model.TaskExecutionDetails{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_execution_Details table creation failed.")
	} else {
		log.Info("Created task_execution_Details table.")
	}

	// Add foreign key constraint by task_id.
	if err := db.Model(&model.TaskExecutionDetails{}).AddForeignKey("task_id", "task_details(task_id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_execution_Details table association with projects table failed.")
	} else {
		log.Info("task_execution_Details table is associated with projects table.")
	}

	// Add unique index on task_details tokens.
	if err := db.Exec("CREATE UNIQUE INDEX task_name_unique_idx on task_details(task_name);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_details table token unique indexing failed.")
	} else {
		log.Info("task_details table token unique index created.")
	}

	// Add unique index on task_execution_details tokens.
	if err := db.Exec("CREATE UNIQUE INDEX task_id_delta_unique_idx on task_execution_details(task_id, delta, project_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_execution_details table token unique indexing failed.")
	} else {
		log.Info("task_execution_details table token unique index created.")
	}

	// Add unique index on task_execution_dependency_details tokens.
	if err := db.Exec("CREATE UNIQUE INDEX task_id_dependent_task_id_unique_idx on task_execution_dependency_details(task_id, dependent_task_id);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("task_execution_dependency_details table token unique indexing failed.")
	} else {
		log.Info("task_execution_dependency_details table token unique index created.")
	}

	// project model metadata
	if err := db.CreateTable(&model.ProjectModelMetadata{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_model_metadata table creation failed.")
	} else {
		log.Info("Created project_model_metadata table.")
	}

	// Add unique index on project_model_metadata tokens.
	if err := db.Exec("CREATE UNIQUE INDEX project_model_metadata_project_id_stdate_enddate_unique_idx on project_model_metadata(project_id, start_time, end_time);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_model_metadata table token unique indexing failed.")
	} else {
		log.Info("project_model_metadata table token unique index created.")
	}

	// Add foreign key constraint by project.
	if err := db.Model(&model.ProjectModelMetadata{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("project_model_metadata table association with projects table failed.")
	} else {
		log.Info("project_model_metadata table is associated with projects table.")
	}
	// display name
	if err := db.CreateTable(&model.DisplayName{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("displayname table creation failed.")
	} else {
		log.Info("Created display name table.")
	}

	// Add foreign key constraint by project.
	if err := db.Model(&model.DisplayName{}).AddForeignKey("project_id", "projects(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("displayname table association with projects table failed.")
	} else {
		log.Info("display_name table is associated with projects table.")
	}

	// Add unique index on display_names tokens.
	if err := db.Exec("CREATE UNIQUE INDEX display_names_project_id_object_group_entity_tag_unique_idx on display_names(project_id, entity_type, group_name, group_object_name, display_name);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("display_names table token unique indexing failed.")
	} else {
		log.Info("display_names table token unique index created.")
	}

	// Add unique index on display_names tokens.
	if err := db.Exec("CREATE UNIQUE INDEX display_names_project_id_event_name_property_name_tag_unique_idx on display_names(project_id, event_name, property_name, tag);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("display_names table token unique indexing failed.")
	} else {
		log.Info("display_names table token unique index created.")
	}

	// templates
	if err := db.CreateTable(&model.Template{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("template table creation failed.")
	} else {
		log.Info("Created template table.")
	}

	// Create Feedback Table
	if err := db.CreateTable(&model.Feedback{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("feedback table creation failed.")
	} else {
		log.Info("Created feedback table.")
	}
	if err := db.Exec("ALTER TABLE events DROP CONSTRAINT events_project_id_event_name_id_event_names_project_id_id_foreign_key;").Error; err != nil {
		log.WithError(err).Error("Failed to drop constraint on events table")
	} else {
		log.Info("Altered type of id of event_names table.")
	}
	if err := db.Exec("ALTER TABLE events ALTER COLUMN event_name_id TYPE bigint USING event_name_id::bigint;").Error; err != nil {
		log.WithError(err).Error("Failed to alter type of event_name_id on events.")
	} else {
		log.Info("Dropped constraint on events for fixing event_name_id type.")
	}

	if err := db.Exec("ALTER TABLE property_details DROP CONSTRAINT property_details_project_id_event_name_id_event_names_project_id_id_foreign_key;").Error; err != nil {
		log.WithError(err).Error("Failed to drop constraint on property_details")
	} else {
		log.Info("Dropped constraint on property_details for fixing event_name_id type.")
	}
	if err := db.Exec("ALTER TABLE property_details ALTER COLUMN event_name_id TYPE bigint USING event_name_id::bigint;").Error; err != nil {
		log.WithError(err).Error("Failed to alter type of event_name_id on property_details.")
	} else {
		log.Info("Altered type of event_name_id of property_details table.")
	}

	// Adding auto increment to event_names
	if err := db.Exec("CREATE SEQUENCE public.event_names_id_seq INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;").Error; err != nil {
		log.WithError(err).Error("Failed to create event_names sequence")
	} else {
		log.Info("Create event_names sequence success")
		if err := db.Exec("alter table event_names drop column id cascade").Error; err != nil {
			log.WithError(err).Error("dropping id column failed")
		} else {
			log.Info("dropping id column")
		}
		if err := db.Exec("alter table event_names add id bigint NOT NULL DEFAULT nextval('event_names_id_seq'::regclass)").Error; err != nil {
			log.WithError(err).Error("event_names id - attach sequence failed")
		} else {
			log.Info("event_names id - attach sequence")
		}
	}

	// Create Groups Table
	if err := db.CreateTable(&model.Group{}).Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Group table creation failed.")
	} else {
		log.Info("Created Groups table.")
	}

	if *isDevSetup {
		InitialiseDevSetup()
	}
}
func InitialiseDevSetup() {
	db := C.GetServices().Db
	defer db.Close()
	// inserting data to agents table for (dev setup purpose only)
	if err := db.Exec("insert into agents ( uuid, first_name, last_name, email, is_email_verified, phone, salt, password, password_created_at, invited_by, created_at, updated_at, is_deleted, last_logged_in_at, login_count, int_adwords_refresh_token, int_salesforce_instance_url, int_salesforce_refresh_token, company_url, subscribe_newsletter, int_google_organic_refresh_token ) values ( '058aa637-5846-4309-b961-267970e6a939', 'Test', 'Test', 'test@factors.ai', TRUE, '1234567890', 'mwTMiEQNeIoIOO8qVhPTHpg6ZA3JrrRO', '$2a$14$5ALo.F9YdJYPGB36D6iFEufTs04fW4rqY5UGXrcB1QTTMEYNOKC46', '2021-10-21 06:21:16.694155+00 ', NULL, ' 2021-10-21 06:20:45.149699+00 ', ' 2021-10-21 06:21:25.074638+00 ', FALSE, ' 2021-10-21 06:21:25.072506+00 ', 1, '', '', '', '', FALSE, '' );").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("creating agent account failed")
	} else {
		log.Info("created account email: test@factors.ai pwd: Factors123")
	}

	// insert agent data to billing accounts (dev setup)
	if err := db.Exec("insert into billing_accounts (id,plan_id,agent_uuid,organization_name,billing_address,pincode,phone_no,created_at,updated_at) values ('c9cb0bde-1f87-477b-a224-1704111c0f97',1,'058aa637-5846-4309-b961-267970e6a939','','','','','2021-10-21 06:20:45.211408+00','2021-10-21 06:20:45.211408+00');").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create billing account record")
	} else {
		log.Info("created billing account ")
	}

	// create project
	if err := db.Exec("INSERT INTO projects (name,token,private_token,created_at,updated_at,project_uri,time_format,date_format,time_zone, jobs_metadata) VALUES ('My Project','78ycpg9dsgok7o4dbk58jtl2rgt0tg0o','kstyafhqje219c59utjwzivhwrfy3lgn','2021-10-29 08:23:29','2021-10-29 08:23:29','','','','Asia/Kolkata','{\"next_session_start_timestamp\":1548844122}');").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create project")
	} else {
		log.Info("created project test with project id = 1 ")
	}

	// map project to agent
	if err := db.Exec("insert into project_agent_mappings (agent_uuid,project_id,role) values('058aa637-5846-4309-b961-267970e6a939',1,2);").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to map agent to project")
	} else {
		log.Info("mapped agent to project_id =1  ")
	}

	// adding project settings
	if err := db.Exec("insert into project_settings (project_id, auto_track, exclude_bot,int_adwords_enabled_agent_uuid,int_adwords_customer_account_id,int_facebook_agent_uuid,int_facebook_ad_account,int_linkedin_ad_account) values(1, true, true,'058aa637-5846-4309-b961-267970e6a939','2368493227','058aa637-5846-4309-b961-267970e6a939','act_2737160626567083','508557389');").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to add project settings ")
	} else {
		log.Info("added project settings for project_id =1  ")
	}
	// mapping project to billing account
	if err := db.Exec("insert into project_billing_account_mappings (project_id,billing_account_id) values (1,'c9cb0bde-1f87-477b-a224-1704111c0f97');").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to map project to billing accounts ")
	} else {
		log.Info("mapped project to billing accounts  ")
	}

	if err := db.Exec("INSERT INTO dashboards (project_id,agent_uuid,name,description,type,class,units_position,created_at,updated_at) VALUES ('1','058aa637-5846-4309-b961-267970e6a939','My Dashboard','No Description','pr','user_created',NULL,'2021-10-29 08:23:29','2021-10-29 08:23:29');").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to map project to billing accounts ")
	} else {
		log.Info("dashboard inserted")
	}

	if err := db.Exec("insert into task_details (id, task_id, task_name, frequency, frequency_interval, skip_start_index, skip_end_index, offset_start_minutes, recurrence, is_project_enabled, created_at, updated_at) values ('c54ad9f9-dc7b-4a59-afd4-0112fcb4a73e', 1, 'RollUpSortedSet', 2, 1, 0, -1, 0, true, false, now(), now() );").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to insert into task_details ")
	} else {
		log.Info("inserted task-details")
	}

	if err := db.Exec("insert into project_model_metadata (id,project_id,model_id,model_type,start_time,end_time,chunks,created_at,updated_at) values ('c54ad9f9-dc7b-4a59-afd4-0112fcb4a73e', 1, 1596524994039, 'w', 1612656000, 1613260799, '1', now(), now());").Error; err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to insert into project_model_metadata ")
	} else {
		log.Info("inserted project_model_metadata")
	}

}

