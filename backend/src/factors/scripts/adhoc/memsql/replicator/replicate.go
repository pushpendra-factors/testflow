package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm/dialects/postgres"

	C "factors/config"
	U "factors/util"

	"factors/model/model"
)

var memSQLDB *gorm.DB
var migrateAllProjects bool
var migrationPageSize int
var migrationRoutinesHeavy int
var migrationRoutinesOther int
var disableSyncColumns bool
var sourceDatastore *string

var migrageAllTables bool
var includeTablesMap map[string]bool
var excludeTablesMap map[string]bool

var modeMigrate bool
var isReverseReplication bool
var skipPrimaryKeyDedupeTables []string

const (
	tableUsers                          = "users"
	tableEventNames                     = "event_names"
	tableEvents                         = "events"
	tableProjects                       = "projects"
	tableAgents                         = "agents"
	tableProjectAgentMappings           = "project_agent_mappings"
	tableProjectBillingAccountMappings  = "project_billing_account_mappings"
	tableProjectSettings                = "project_settings"
	tableAdwordsDocuments               = "adwords_documents"
	tableBigquerySettings               = "bigquery_settings"
	tableBillingAccounts                = "billing_accounts"
	tableDashboardUnits                 = "dashboard_units"
	tableDashboards                     = "dashboards"
	tableFacebookDocuments              = "facebook_documents"
	tableFactorsGoals                   = "factors_goals"
	tableFactorsTrackedEvents           = "factors_tracked_events"
	tableFactorsTrackedUserProperties   = "factors_tracked_user_properties"
	tableHubspotDocuments               = "hubspot_documents"
	tableSalesforceDocuments            = "salesforce_documents"
	tableQueries                        = "queries"
	tableScheduledTasks                 = "scheduled_tasks"
	tableLinkedInDocuments              = "linkedin_documents"
	tablePropertyDetails                = "property_details"
	tableSmartProperties                = "smart_properties"
	tableSmartPropertyRules             = "smart_property_rules"
	tableDisplayNames                   = "display_names"
	tableProjectModelMetadata           = "project_model_metadata"
	tableGoogleOrganicDocuments         = "google_organic_documents"
	tableTaskDetails                    = "task_details"
	tableTaskExecutionDetails           = "task_execution_details"
	tableTaskExecutionDependencyDetails = "task_execution_dependency_details"
	tableWeekyInsightsMetadata          = "weekly_insights_metadata"
	tableTemplates                      = "templates"
	tableFeedback                       = "feedbacks"
	tableGroups                         = "groups"
	healthcheckPingID                   = "e6e3735b-82a3-4534-82be-b621470c4c69"
)

var supportedTables = []string{
	tableUsers,
	tableEventNames,
	tableEvents,
	tableProjects,
	tableAgents,
	tableProjectAgentMappings,
	tableProjectBillingAccountMappings,
	tableProjectSettings,
	tableAdwordsDocuments,
	tableBigquerySettings,
	tableBillingAccounts,
	tableDashboardUnits,
	tableDashboards,
	tableFacebookDocuments,
	tableFactorsGoals,
	tableFactorsTrackedEvents,
	tableFactorsTrackedUserProperties,
	tableHubspotDocuments,
	tableSalesforceDocuments,
	tableQueries,
	tableScheduledTasks,
	tableLinkedInDocuments,
	tablePropertyDetails,
	tableSmartProperties,
	tableSmartPropertyRules,
	tableDisplayNames,
	tableProjectModelMetadata,
	tableGoogleOrganicDocuments,
	tableTaskDetails,
	tableTaskExecutionDetails,
	tableTaskExecutionDependencyDetails,
	tableWeekyInsightsMetadata,
	tableTemplates,
	tableFeedback,
	tableGroups,
}

// Disabled tables to avoid accidental replication.
var permanantlyDisabledTables = []string{
	tableUsers,
	tableEvents,
	tableGroups,
}

var heavyTables = []string{tableEvents, tableUsers, tableAdwordsDocuments, tableHubspotDocuments}

type TableRecord struct {
	ProjectID int64       `json:"project_id"`
	ID        interface{} `json:"id"`      // Using interface, as type of ID can be uint64 or uuid.
	UUID      interface{} `json:"uuid"`    // Agents table is using uuid instead of id.
	UserID    string      `json:"user_id"` // For analytics tables.

	// Additional Primary Key Fields per table.
	// hubspot_documents, salesforce_documents
	// adwords_documents, facebook_documents, linkedin_documents
	Type      interface{} `json:"type"`
	Timestamp uint64      `json:"timestamp"`
	// adwords_documents
	CustomerAccountID string `json:"customer_acc_id"`
	// google_organic_documents
	URLPrefix string `json:"url_prefix"`
	// linkedin_documents
	CustomerAdAccountID string `json:"customer_ad_account_id"`
	// facebook_documents
	Platform string `json:"platform"`
	// hubspot_documents, salesforce_documents
	Action interface{} `json:"action"`
	// project_agent_mappings
	AgentUUID string `json:"agent_uuid"`
	// property_details
	EventNameID string `json:"event_name_id"`
	Key         string `json:"key"`
	// smart_properties
	ObjectID   string `json:"object_id"`
	ObjectType uint64 `json:"object_type"`
	Source     string `json:"source"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Columns to exclude.
	// Projects table.
	JobsMetadata *postgres.Jsonb `json:"jobs_metadata"`
	// Hubspot and salesforce documents. UserID is already added above.
	Synced bool   `json:"synced"`
	SyncId string `json:"sync_id"`
}

type dest_EventName_pg struct {
	// Composite primary key with projectId.
	ID   string `gorm:"primary_key:true;" json:"-"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId int64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type dest_Event struct {
	// Composite primary key with project_id and uuid.
	ID              string  `gorm:"primary_key:true;type:uuid" json:"id"`
	CustomerEventId *string `json:"customer_event_id"`

	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	// (project_id, event_name_id) -> event_names(project_id, id)
	ProjectId int64  `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `json:"user_id"`

	// TODO(Dinesh): Remove user_properties_id column and field after
	// user_properties table is permanantly deprecated.
	UserPropertiesId string `json:"user_properties_id"`

	UserProperties *postgres.Jsonb `json:"user_properties"`
	SessionId      *string         `json:session_id`
	EventNameId    string          `json:"-"` // DISABLED json tag for overriding on programatically.
	Count          uint64          `json:"count"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties                 postgres.Jsonb `json:"properties,omitempty"`
	PropertiesUpdatedTimestamp int64          `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
	// unix epoch timestamp in seconds.
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	eventsMap         sync.Map
	eventNamesMap     sync.Map
	usersMap          sync.Map
	userPropertiesMap sync.Map
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	projectIDStringList := flag.String("project_ids", "", "")
	pageSize := flag.Int("page_size", 0, "No.of records per page.")
	routinesHeavy := flag.Int("routines_heavy", 10, "No.of parallel routines to use for events, users, adwords_documents and hubspot_documents.")
	routinesOther := flag.Int("routines_other", 5, "No.of parallel routines to use for all other tables.")
	migrate := flag.Bool("migrate", false, "Disable's additional queries like dedupe query for migration.")

	includeTables := flag.String("include_tables", "", "Comma separated tables to run or *")
	excludeTables := flag.String("exclude_tables", "", "Comma separated tables to exclude from the run")
	disableSyncColumnsFlag := flag.Bool("disable_sync_column", false, "To disable sync columns like jobs_metadata, synced etc")

	skipDedupeTables := flag.String("skip_dedupe_tables", "", "Skip primary key deduplication for the tables.")
	isReverseReplicator := flag.Bool("is_reverse_replication", false, "Is this the primary replicator or reverse replicator")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	if *projectIDStringList == "" {
		log.Fatal("Invalid project_id.")
	}

	if *includeTables == "" {
		log.Fatal("Invalid include_tables. It should be either * or list of tables.")
	}

	if *includeTables != "" && *includeTables != "*" &&
		!isValidTables(U.CleanSplitByDelimiter(*includeTables, ",")) {
		log.Fatal("Invalid table names on include_tables.")
	}

	if *excludeTables != "" && *excludeTables != "*" &&
		!isValidTables(U.CleanSplitByDelimiter(*excludeTables, ",")) {
		log.Fatal("Invalid table names on exclude_tables.")
	}

	if *pageSize == 0 {
		log.Fatal("Migration page size cannot be zero.")
	}
	migrationPageSize = *pageSize
	migrationRoutinesHeavy = *routinesHeavy
	migrationRoutinesOther = *routinesOther
	modeMigrate = *migrate

	if !U.StringValueIn(*sourceDatastore, []string{C.DatastoreTypeMemSQL}) {
		log.WithField("source", *sourceDatastore).Fatal("Invalid source provided.")
	}

	includeTablesMap = make(map[string]bool)
	excludeTablesMap = make(map[string]bool)
	if *includeTables == "*" {
		migrageAllTables = true
	} else {
		for _, tableName := range U.CleanSplitByDelimiter(*includeTables, ",") {
			includeTablesMap[tableName] = true
		}
	}

	for _, tableName := range U.CleanSplitByDelimiter(*excludeTables, ",") {
		excludeTablesMap[tableName] = true
	}
	for _, tableName := range permanantlyDisabledTables {
		excludeTablesMap[tableName] = true
	}

	isReverseReplication = *isReverseReplicator
	disableSyncColumns = *disableSyncColumnsFlag

	skipPrimaryKeyDedupeTables = U.CleanSplitByDelimiter(*skipDedupeTables, ",")

	config := &C.Configuration{
		AppName: "memsql_replicator",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		SentryDSN: *sentryDSN,
	}
	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}
	registerReplicationCallbacks(C.GetServices().Db)

	maxOpenConns := (24 * migrationRoutinesOther) + (4 * migrationRoutinesHeavy) + 10
	memSQLDBConf := C.DBConf{
		Host:        *memSQLHost,
		Port:        *memSQLPort,
		User:        *memSQLUser,
		Name:        *memSQLName,
		Password:    *memSQLPass,
		Certificate: *memSQLCertificate,
	}
	initMemSQLDB(*env, &memSQLDBConf, maxOpenConns)

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDStringList, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if allProjects {
		migrateAllProjects = true
	}

	if len(projectIDs) == 0 && !migrateAllProjects {
		log.Fatal("Invalid value on project_ids flag. Should be list of project_ids or *.")
	}

	log.WithField("project_ids", projectIDs).Info("Running in replication mode.")
	migrateAllTables(projectIDs)
}

func isValidTables(tables []string) bool {
	for _, t := range tables {
		if !U.StringValueIn(t, supportedTables) {
			return false
		}
	}
	return true
}

func registerReplicationCallbacks(db *gorm.DB) {
	// Disables before callback for create method.
	db.Callback().Create().Remove("gorm:before_create")
	// Removes unneccesary select after insert triggered by gorm.
	db.Callback().Create().Remove("gorm:force_reload_after_create")
	// Removes emoji and cleans up string and postgres.Jsonb columns.
	db.Callback().Create().Before("gorm:create").Register("cleanup", U.GormCleanupCallback)
	db.Callback().Create().Before("gorm:update").Register("cleanup", U.GormCleanupCallback)
}

func initMemSQLDB(env string, dbConf *C.DBConf, maxOpenConns int) {
	dbConf.UseSSL = C.IsStaging() || C.IsProduction()

	var err error
	memSQLDB, err = gorm.Open("mysql", C.GetMemSQLDSNString(dbConf))
	if err != nil {
		log.WithError(err).Fatal("Failed connecting to memsql.")
	}

	registerReplicationCallbacks(memSQLDB)

	if C.IsDevelopment() {
		memSQLDB.LogMode(true)
	} else {
		memSQLDB.LogMode(false)
		memSQLDB.DB().SetMaxOpenConns(maxOpenConns)
		memSQLDB.DB().SetMaxIdleConns(maxOpenConns)
	}
}

func getDestinationDB() *gorm.DB {
	switch *sourceDatastore {
	case C.DatastoreTypeMemSQL:
		return C.GetServices().Db
	default:
		return nil
	}
}

func getSourceDB() *gorm.DB {
	switch *sourceDatastore {
	case C.DatastoreTypeMemSQL:
		return memSQLDB
	default:
		return nil
	}
}

func isDestination(store string) bool {
	return store != *sourceDatastore
}

func getDedupeMap(tableName string) *sync.Map {
	switch tableName {
	case tableEvents:
		return &eventsMap
	case tableEventNames:
		return &eventNamesMap
	case tableUsers:
		return &usersMap
	default:
		log.Fatalf("Unsupported table %s on getDedupeMap.", tableName)
		return nil
	}
}

func doesExistByUserAndIDOnMemSQLDB(tableName string, projectID int64,
	userID string, id interface{}) (bool, *TableRecord, error) {

	logCtx := log.WithField("project_id", projectID).
		WithField("id", id).WithField("user_id", userID).
		WithField("table", tableName)

	var record TableRecord
	// Select only id to use index only read and avoid disk read.
	err := getDestinationDB().Table(tableName).Select("id, updated_at").Limit(1).
		Where("project_id = ? AND user_id = ? AND id = ?", projectID, userID, id).
		Find(&record).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil, nil
		}

		logCtx.WithError(err).Error("Failed to check existence.")
		return true, nil, err
	}

	return true, &record, nil
}

func getEventsListAsBatch(list []model.Event, batchSize int) [][]model.Event {
	batchList := make([][]model.Event, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func getAdditionalPrimaryKeyConditionsByTableName(tableName string, sourceTableRecord *TableRecord) (string, []interface{}) {
	var conditions []string
	params := make([]interface{}, 0, 0)

	switch tableName {
	case tableAdwordsDocuments:
		conditions = []string{"customer_account_id", "type", "timestamp"}
		params = append(params, sourceTableRecord.CustomerAccountID, sourceTableRecord.Type, sourceTableRecord.Timestamp)

	case tableFacebookDocuments:
		conditions = []string{"customer_ad_account_id", "platform", "type", "timestamp"}
		params = append(params, sourceTableRecord.CustomerAdAccountID, sourceTableRecord.Platform,
			sourceTableRecord.Type, sourceTableRecord.Timestamp)

	case tableLinkedInDocuments:
		conditions = []string{"customer_ad_account_id", "type", "timestamp"}
		params = append(params, sourceTableRecord.CustomerAdAccountID, sourceTableRecord.Type, sourceTableRecord.Timestamp)

	case tableHubspotDocuments:
		conditions = []string{"type", "action", "timestamp"}
		params = append(params, sourceTableRecord.Type, sourceTableRecord.Action, sourceTableRecord.Timestamp)

	case tableSalesforceDocuments:
		conditions = []string{"type", "timestamp"}
		params = append(params, sourceTableRecord.Type, sourceTableRecord.Timestamp)

	case tableProjectAgentMappings:
		conditions = []string{"agent_uuid"}
		params = append(params, sourceTableRecord.AgentUUID)

	case tableQueries:
		conditions = []string{"type"}
		params = append(params, sourceTableRecord.Type)

	case tablePropertyDetails:
		conditions = []string{"event_name_id", "key"}
		params = append(params, sourceTableRecord.EventNameID, sourceTableRecord.Key)

	case tableSmartProperties:
		conditions = []string{"object_id", "object_type", "source"}
		params = append(params, sourceTableRecord.ObjectID, sourceTableRecord.ObjectType, sourceTableRecord.Source)

	case tableGoogleOrganicDocuments:
		conditions = []string{"url_prefix", "timestamp"}
		params = append(params, sourceTableRecord.URLPrefix, sourceTableRecord.Timestamp)

	}

	var condition string
	for i, c := range conditions {
		if i > 0 {
			condition = condition + " AND "
		}
		condition = condition + fmt.Sprintf("%s = ?", c)
	}

	return condition, params
}

func getPrimaryKeyConditionByTableName(tableName string, sourceTableRecord *TableRecord) (string, []interface{}) {
	idColName := "id"
	if tableName == tableAgents {
		idColName = "uuid"
	} else if isProjectAssociatedTable(tableName) {
		idColName = "project_id"
	}

	params := make([]interface{}, 0, 0)

	condition := fmt.Sprintf("%s = ?", idColName)
	params = append(params, sourceTableRecord.ID)

	addCondition, addParams := getAdditionalPrimaryKeyConditionsByTableName(tableName, sourceTableRecord)
	if addCondition != "" {
		condition = fmt.Sprintf("%s AND %s", condition, addCondition)
		params = append(params, addParams...)
	}

	return condition, params
}

func isValidProjectID(tableName string, id int64) bool {
	return id > 0 || tableName == tableTaskExecutionDetails
}

func getTableRecordByIDFromMemSQL(projectID int64, tableName string, id interface{},
	sourceTableRecord *TableRecord) (*TableRecord, int) {

	logCtx := log.WithField("project_id", projectID).WithField("id", id).
		WithField("table_name", tableName)

	if tableName == "" || id == nil {
		logCtx.Error("Invalid table_name or id on getTableRecordByIDFromMemSQL.")
		return nil, http.StatusBadRequest
	}

	if !isTableWithoutProjectID(tableName) && !isValidProjectID(tableName, projectID) {
		logCtx.Error("Invalid project_id on getTableRecordByIDFromMemSQL")
		return nil, http.StatusBadRequest
	}

	var record TableRecord
	condition, params := getPrimaryKeyConditionByTableName(tableName, sourceTableRecord)
	db := getDestinationDB().Table(tableName).Limit(1).Where(condition, params...)
	if !isTableWithoutProjectID(tableName) {
		db = db.Where("project_id = ?", projectID)
	}

	// Add user_id also part of the filter to speed up the query.
	if tableName == tableEvents {
		if sourceTableRecord.UserID == "" {
			logCtx.Error("user_id not provided on getTableRecordByIDFromMemSQL")
		} else {
			db = db.Where("user_id = ?", sourceTableRecord.UserID)
		}
	}

	err := db.Find(&record).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get table_record by id from memsql.")
		return nil, http.StatusInternalServerError
	}

	return &record, http.StatusFound
}

func deleteByIDOnMemSQL(projectID int64, tableName string, id interface{}, sourceTableRecord *TableRecord) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName).
		WithField("id", id).WithField("user_id", sourceTableRecord.UserID)

	if (!isTableWithoutProjectID(tableName) && !isValidProjectID(tableName, projectID)) || tableName == "" || id == nil {
		return http.StatusBadRequest
	}

	condition, params := getPrimaryKeyConditionByTableName(tableName, sourceTableRecord)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, condition)
	if !isTableWithoutProjectID(tableName) {
		query = fmt.Sprintf("%s AND project_id = %d", query, projectID)
	}

	// Add user_id also part of the filter to speed up the query.
	if tableName == tableEvents {
		if sourceTableRecord.UserID == "" {
			logCtx.Error("user_id not provided on deleteByIDOnMemSQL")
		} else {
			query = fmt.Sprintf("%s AND user_id = '%s'", query, sourceTableRecord.UserID)
		}
	}

	rows, err := getDestinationDB().Raw(query, params...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed deleting record in memsql.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusOK
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate") ||
		strings.Contains(err.Error(), "violates unique constraint")
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func convertToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func isDefaultValue(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func createOnMemSQL(projectID int64, tableName string, record interface{}, id interface{}, isFromUpdate bool) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName).
		WithField("record", record)

	// Avoid resetting of fields for create calls from update.
	if !isFromUpdate {
		record = disallowCreateWithSyncColumns(tableName, record)
	}

	if err := getDestinationDB().Table(tableName).Create(record).Error; err != nil {
		if isDuplicateError(err) {
			return http.StatusConflict
		}

		if strings.Contains(err.Error(), "Invalid JSON value for column") {
			logCtx.WithError(err).Error("Skipping record with invalid json value")
			return http.StatusCreated
		}

		if strings.Contains(err.Error(), "Error 1364: Field") {
			logCtx.WithError(err).Error("Skipping record with invalid field value.")
			return http.StatusCreated
		}

		logCtx.WithError(err).Error("Failed to create record on memsql.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func convertToTableRecord(tableName string, record interface{}) (*TableRecord, error) {
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	var tableRecord TableRecord
	err = json.Unmarshal([]byte(recordJSON), &tableRecord)
	if err != nil {
		return nil, err
	}

	// Agents table is using column name as UUID for ID column.
	// This maps UUID to ID for other parts of code to work.
	if tableName == tableAgents {
		tableRecord.ID = tableRecord.UUID
	}

	// Tables using project_id as id.
	if isProjectAssociatedTable(tableName) {
		tableRecord.ID = tableRecord.ProjectID
	}

	return &tableRecord, nil
}

func isTableWithoutUniquePrimaryKey(tableName string) bool {
	return U.StringValueIn(tableName, []string{
		tableEvents,
		tableEventNames,
		tableAdwordsDocuments,
		tableFacebookDocuments,
		tableLinkedInDocuments,
		tableHubspotDocuments,
		tableSalesforceDocuments,
		tableGoogleOrganicDocuments,
	})
}

func isProjectAssociatedTable(tableName string) bool {
	return U.StringValueIn(tableName, []string{
		tableProjectBillingAccountMappings,
		tableProjectSettings,
		tableProjectAgentMappings,
		// project_id + additional keys
		tableSmartProperties,
		tablePropertyDetails,
	})
}

func updateIfExistOnMemSQL(projectID int64, tableName string, pgRecord interface{}, pgTableRecord, memsqlTableRecord *TableRecord) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName)

	var status int
	if memsqlTableRecord == nil {
		memsqlTableRecord, status = getTableRecordByIDFromMemSQL(projectID, tableName, pgTableRecord.ID, pgTableRecord)
		if status == http.StatusInternalServerError || status == http.StatusBadRequest {
			logCtx.WithField("status", status).WithField("pg_table_record", pgTableRecord).
				Error("Failed to get the existing record from memsql.")
			return http.StatusInternalServerError
		}

		if status == http.StatusNotFound {
			if tableName == tableUsers {
				// Users table has duplicate id (uuid) on postgres. And it was not realised as uniqueness is
				// maintained by project_id, id But on memsql users table will use only id as unique constraint.
				logCtx.WithField("pg_table_record", pgTableRecord).Error("Possible duplicate uuid on postgres users table.")
			}
			return http.StatusOK
		}
	}

	if pgTableRecord.UpdatedAt.After(memsqlTableRecord.UpdatedAt) {
		status := deleteByIDOnMemSQL(projectID, tableName, pgTableRecord.ID, pgTableRecord)
		if status != http.StatusOK {
			return http.StatusInternalServerError
		}

		// Copy old memsql record values for selected sync columns.
		pgRecord = disallowUpdateOnSyncColumns(tableName, memsqlTableRecord, pgRecord)
		status = createOnMemSQL(projectID, tableName, pgRecord, pgTableRecord.ID, true)
		if status != http.StatusCreated && status != http.StatusConflict {
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func disallowUpdateOnSyncColumns(tableName string, memsqlTableRecord *TableRecord, pgRecord interface{}) interface{} {
	if !disableSyncColumns || (tableName != tableProjects && tableName != tableHubspotDocuments && tableName != tableSalesforceDocuments) {
		return pgRecord
	}

	if tableName == tableProjects {
		project := pgRecord.(*model.Project)
		project.JobsMetadata = memsqlTableRecord.JobsMetadata
		return project
	} else if tableName == tableHubspotDocuments {
		hubspotDocument := pgRecord.(*model.HubspotDocument)
		hubspotDocument.SyncId = memsqlTableRecord.SyncId
		hubspotDocument.Synced = memsqlTableRecord.Synced
		hubspotDocument.UserId = memsqlTableRecord.UserID
		return hubspotDocument
	} else if tableName == tableSalesforceDocuments {
		salesforceDocument := pgRecord.(*model.SalesforceDocument)
		salesforceDocument.SyncID = memsqlTableRecord.SyncId
		salesforceDocument.Synced = memsqlTableRecord.Synced
		salesforceDocument.UserID = memsqlTableRecord.UserID
		return salesforceDocument
	}
	return pgRecord
}

func disallowCreateWithSyncColumns(tableName string, pgRecord interface{}) interface{} {
	if !disableSyncColumns || (tableName != tableProjects && tableName != tableHubspotDocuments && tableName != tableSalesforceDocuments) {
		return pgRecord
	}

	if tableName == tableProjects {
		project := pgRecord.(*model.Project)
		jobsMetadata := &map[string]interface{}{
			// Set it to 1 hour before created_at to be safe.
			model.JobsMetadataKeyNextSessionStartTimestamp: project.CreatedAt.Unix() - 3600,
		}
		jobsMetadataJsonb, err := U.EncodeToPostgresJsonb(jobsMetadata)
		if err != nil {
			return pgRecord
		}
		project.JobsMetadata = jobsMetadataJsonb
		return project
	} else if tableName == tableHubspotDocuments {
		hubspotDocument := pgRecord.(*model.HubspotDocument)
		hubspotDocument.SyncId = ""
		hubspotDocument.Synced = false
		hubspotDocument.UserId = ""
		return hubspotDocument
	} else if tableName == tableSalesforceDocuments {
		salesforceDocument := pgRecord.(*model.SalesforceDocument)
		salesforceDocument.SyncID = ""
		salesforceDocument.Synced = false
		salesforceDocument.UserID = ""
		return salesforceDocument
	}
	return pgRecord
}

func createIfNotExistOrUpdateIfChangedOnMemSQL(tableName string, pgRecord interface{}) (*TableRecord, int) {
	pgTableRecord, err := convertToTableRecord(tableName, pgRecord)
	if err != nil {
		log.WithField("pg_record", pgRecord).
			WithError(err).Error("Failed to convert pg record to table record.")
		return nil, http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", pgTableRecord.ProjectID).
		WithField("table_name", tableName).
		WithField("pg_record_id", pgTableRecord.ID)

	if isDestination(C.DatastoreTypeMemSQL) && isTableWithoutUniquePrimaryKey(tableName) &&
		!modeMigrate && !U.StringValueIn(tableName, skipPrimaryKeyDedupeTables) {
		// TODO: Try to add a procedure/trigger to do the existence check before create on memsql
		// itself for long term. Now doing it on the script itself, as it is sequential.
		currentMemsqlRecord, status := getTableRecordByIDFromMemSQL(pgTableRecord.ProjectID,
			tableName, pgTableRecord.ID, pgTableRecord)
		if status == http.StatusInternalServerError || status == http.StatusBadRequest {
			logCtx.WithField("status", status).Error("Failed to get the existing record from memsql.")
			return nil, http.StatusInternalServerError
		}

		if status == http.StatusFound {
			// Uses already fetched memsql record to compare updated_at and replace.
			if status := updateIfExistOnMemSQL(pgTableRecord.ProjectID, tableName, pgRecord,
				pgTableRecord, currentMemsqlRecord); status != http.StatusOK {

				return nil, http.StatusInternalServerError
			}

			return pgTableRecord, http.StatusCreated
		}

		// If the status is not_found. It should proceed and create the record.
		if status != http.StatusNotFound {
			return pgTableRecord, http.StatusInternalServerError
		}
	}

	// If the table has unique primary key. Do not do getOnMemSQL in the beginning.
	// Only do it, when there is a conflict for checking the updatedAt.
	status := createOnMemSQL(pgTableRecord.ProjectID, tableName, pgRecord, pgTableRecord.ID, false)
	if status == http.StatusConflict {
		if status := updateIfExistOnMemSQL(pgTableRecord.ProjectID,
			tableName, pgRecord, pgTableRecord, nil); status != http.StatusOK {

			return pgTableRecord, http.StatusInternalServerError
		}

		return pgTableRecord, http.StatusCreated
	}

	if status != http.StatusCreated {
		return pgTableRecord, http.StatusInternalServerError
	}

	return pgTableRecord, http.StatusCreated
}

func getRecordInterfaceByTableName(tableName string) interface{} {
	var record interface{}

	// TODO: Use table constants and add all tables.
	switch tableName {

	// Tables with project_id.
	case tableAdwordsDocuments:
		record = &model.AdwordsDocument{}
	case tableBigquerySettings:
		record = &model.BigquerySetting{}
	case tableBillingAccounts:
		record = &model.BillingAccount{}
	case tableDashboardUnits:
		record = &model.DashboardUnit{}
	case tableDashboards:
		record = &model.Dashboard{}
	case tableFacebookDocuments:
		record = &model.FacebookDocument{}
	case tableFactorsGoals:
		record = &model.FactorsGoal{}
	case tableFactorsTrackedEvents:
		record = &model.FactorsTrackedEvent{}
	case tableFactorsTrackedUserProperties:
		record = &model.FactorsTrackedUserProperty{}
	case tableHubspotDocuments:
		record = &model.HubspotDocument{}
	case tableProjectAgentMappings:
		record = &model.ProjectAgentMapping{}
	case tableProjectBillingAccountMappings:
		record = &model.ProjectBillingAccountMapping{}
	case tableProjectSettings:
		record = &model.ProjectSetting{}
	case tableSalesforceDocuments:
		record = &model.SalesforceDocument{}
	case tableScheduledTasks:
		record = &model.ScheduledTask{}
	case tableQueries:
		record = &model.Queries{}
	case tableLinkedInDocuments:
		record = &model.LinkedinDocument{}
	case tablePropertyDetails:
		record = &model.PropertyDetail{}
	case tableSmartProperties:
		record = &model.SmartProperties{}
	case tableSmartPropertyRules:
		record = &model.SmartPropertyRules{}
	case tableDisplayNames:
		record = &model.DisplayName{}
	case tableGoogleOrganicDocuments:
		record = &model.GoogleOrganicDocument{}
	case tableTemplates:
		record = &model.Template{}
	case tableProjectModelMetadata:
		record = &model.ProjectModelMetadata{}
	case tableTaskDetails:
		record = &model.TaskDetails{}
	case tableTaskExecutionDetails:
		record = &model.TaskExecutionDetails{}
	case tableTaskExecutionDependencyDetails:
		record = &model.TaskExecutionDependencyDetails{}
	case tableWeekyInsightsMetadata:
		record = &model.WeeklyInsightsMetadata{}
	case tableGroups:
		record = &model.Group{}
	// Tables related to analytics.
	case tableEvents:
		record = &model.Event{}
	case tableUsers:
		record = &model.User{}
	case tableEventNames:
		if isReverseReplication {
			record = &model.EventName{}
		} else {
			record = &model.EventName{}
		}

	// Tables without project_id.
	case tableAgents:
		record = &model.Agent{}
	case tableProjects:
		record = &model.Project{}

	// feedback table
	case tableFeedback:
		record = &model.Feedback{}
	default:
		// Critical error. Should not proceed without adding support for table.
		log.Fatalf("Unsupported table name %s on get_record_interface_by_table_name", tableName)
	}

	return record
}

func isTableWithoutProjectID(tableName string) bool {
	return U.StringValueIn(tableName, []string{tableAgents, tableProjects, tableBillingAccounts, tableTaskDetails, tableTaskExecutionDetails, tableTaskExecutionDependencyDetails})
}

func buildStringList(ids []int64) string {
	var list string
	for i, id := range ids {
		if i > 0 {
			list = list + ", "
		}

		list = list + fmt.Sprintf("%d", id)
	}

	return list
}

func getRecordsAsChunks(list []interface{}, batchSize int) [][]interface{} {
	batchList := make([][]interface{}, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func migrateTableByName(projectIDs []int64, tableName string, lastPageUpdatedAt *time.Time) (*time.Time, int, int) {
	logCtx := log.WithField("table_name", tableName).
		WithField("project_id", projectIDs)

	if tableName == "" {
		logCtx.Error("Invalid table name.")
		return nil, 0, http.StatusBadRequest
	}

	if len(projectIDs) == 0 && !migrateAllProjects {
		logCtx.Error("Invalid project_id.")
		return nil, 0, http.StatusBadRequest
	}

	anyProjectsCondition := fmt.Sprintf("ANY(ARRAY[%s])", buildStringList(projectIDs))

	query := fmt.Sprintf("SELECT %%s FROM %s", tableName)

	var where bool
	if !migrateAllProjects {
		var whereProjectCondition string

		query = query + " " + "WHERE"
		where = true

		switch tableName {
		case tableProjects:
			whereProjectCondition = fmt.Sprintf("id = %s", anyProjectsCondition)
		case tableAgents:
			whereProjectCondition = fmt.Sprintf("uuid IN (SELECT agent_uuid FROM project_agent_mappings WHERE project_id = %s)",
				anyProjectsCondition,
			)
		case tableBillingAccounts:
			whereProjectCondition = fmt.Sprintf("id IN (SELECT billing_account_id FROM project_billing_account_mappings WHERE project_id = %s)",
				anyProjectsCondition,
			)
		case tableEventNames:
			whereProjectCondition = fmt.Sprintf("type IN ('%s', '%s', '%s') AND project_id = %s", "CS", "CH", "FE", anyProjectsCondition)
		default:
			whereProjectCondition = fmt.Sprintf("project_id = %s", anyProjectsCondition)
		}

		query = query + " " + whereProjectCondition
	} else {
		switch tableName {
		case tableEventNames:

			query = query + " " + "WHERE"
			where = true
			whereProjectCondition := fmt.Sprintf("type IN ('%s', '%s', '%s')", "CS", "CH", "FE")
			query = query + " " + whereProjectCondition
		}
	}

	var timestampSubQuery string
	if lastPageUpdatedAt != nil {
		if where {
			query = query + " " + "AND"
		} else {
			query = query + " " + "WHERE"
		}

		timestampSubQuery = fmt.Sprintf("%s updated_at > '%+v'", query, lastPageUpdatedAt.Format(time.RFC3339Nano))
	} else {
		timestampSubQuery = query
		if migrateAllProjects {
			query = query + " " + "WHERE"
		}
	}
	timestampSubQuery = fmt.Sprintf(timestampSubQuery, "updated_at")

	timestampSubQuery = timestampSubQuery + " " + fmt.Sprintf("ORDER BY updated_at ASC LIMIT %d", migrationPageSize)
	timestampQuery := fmt.Sprintf("SELECT COUNT(*), MIN(updated_at), MAX(updated_at) FROM (%s) subq", timestampSubQuery)
	db := getSourceDB()

	rows, err := db.Raw(timestampQuery).Rows()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logCtx.Info("No new records found")
			return nil, 0, http.StatusOK
		}
		logCtx.WithError(err).Error("Error fetching last updated_at range")
		return nil, 0, http.StatusInternalServerError
	}
	defer rows.Close()

	var minTimestamp, maxTimestamp sql.NullTime
	var rowCount int64
	for rows.Next() {
		if err := rows.Scan(&rowCount, &minTimestamp, &maxTimestamp); err != nil {
			logCtx.WithError(err).Error("Failed to scan min and max timestmap")
			return nil, 0, http.StatusInternalServerError
		}
		if rowCount == 0 {
			logCtx.Info("No new records found")
			return nil, 0, http.StatusOK
		}
		if !minTimestamp.Valid || !maxTimestamp.Valid {
			logCtx.Info("Invalid values for min: %v, max: %v timestamp", minTimestamp, maxTimestamp)
			return nil, 0, http.StatusOK
		}
	}
	logCtx.Info("Fetching for range %v - %v", minTimestamp, maxTimestamp)

	if lastPageUpdatedAt == nil && !migrateAllProjects {
		query = query + " AND"
	}
	query = fmt.Sprintf(query, "*") + fmt.Sprintf(" updated_at BETWEEN '%+v' AND '%+v'",
		minTimestamp.Time.Format(time.RFC3339Nano), maxTimestamp.Time.Format(time.RFC3339Nano))
	rows, err = db.Raw(query).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get records on migrate_table_by_name.")
		return nil, 0, http.StatusInternalServerError
	}
	defer rows.Close()

	// Scans and migrates each record.
	var rowsCount int
	recordIntfs := make([]interface{}, 0, 0)
	for rows.Next() {
		record := getRecordInterfaceByTableName(tableName)
		if err := db.ScanRows(rows, record); err != nil {
			logCtx.WithError(err).Error("Failed to scan the get all records from postgres table.")
			return nil, 0, http.StatusInternalServerError
		}

		recordIntfs = append(recordIntfs, record)
		rowsCount++
	}

	if err := rows.Err(); err != nil {
		logCtx.WithError(err).Error("Failure on scanner after tried scanning on get all records from postgres table.")
		return nil, 0, http.StatusInternalServerError
	}

	var wg sync.WaitGroup
	var states MStates
	var recordChunks [][]interface{}
	if U.StringValueIn(tableName, heavyTables) {
		recordChunks = getRecordsAsChunks(recordIntfs, migrationRoutinesHeavy)
	} else {
		recordChunks = getRecordsAsChunks(recordIntfs, migrationRoutinesOther)
	}
	for i := range recordChunks {
		wg.Add(1)
		go createOrUpdateOnMemSQLInParallel(tableName, recordChunks[i], &wg, &states)
	}
	wg.Wait()

	var lastUpdatedAt time.Time
	for i := range states.mstates {
		if states.mstates[i].Status == http.StatusInternalServerError {
			return nil, 0, http.StatusInternalServerError
		}

		// Largest updated_at.
		if states.mstates[i].UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = states.mstates[i].UpdatedAt
		}
	}

	return &lastUpdatedAt, rowsCount, http.StatusOK
}

type MState struct {
	Status    int
	UpdatedAt time.Time
}

type MStates struct {
	mu      sync.Mutex
	mstates []MState
}

func (s *MStates) AddToState(mstate MState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mstates = append(s.mstates, mstate)
}

func convertIDToUUIDFromInterface(id interface{}) (string, error) {
	var uuid string
	var err error
	switch id.(type) {
	case uint64:
		uuid, err = U.ConvertIntToUUID(id.(uint64))
	case string:
		uuid, err = U.ConvertIntStringToUUID(id.(string))
	case *string:
		if id == nil {
			return "", nil
		}

		idString := id.(*string)
		uuid, err = U.ConvertIntStringToUUID(*idString)
	default:
		return "", errors.New(fmt.Sprintf("Invalid type %T with value %v", id, id))
	}
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func convertStruct(sourceRecord, destRecord interface{}) error {
	recordJSON, err := json.Marshal(sourceRecord)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(recordJSON), destRecord)
	if err != nil {
		return err
	}

	return nil
}

func convertColumnTypeByTable(tableName string, record interface{}) (interface{}, error) {
	switch tableName {
	case tableEventNames:
		if isReverseReplication {
			eventName := record.(*model.EventName)
			var pgEventName dest_EventName_pg
			err := convertStruct(eventName, &pgEventName)
			if err != nil {
				return record, err
			}
			return &pgEventName, nil
		} else {
			eventName := record.(*model.EventName)
			var memsqlEventName model.EventName
			err := convertStruct(eventName, &memsqlEventName)
			if err != nil {
				return record, err
			}

			uuid, err := convertIDToUUIDFromInterface(eventName.ID)
			if err != nil {
				log.WithError(err).Error("Failed to convert event to memsql event_name.")
				return record, err
			}
			memsqlEventName.ID = uuid
			return &memsqlEventName, nil
		}

	case tableEvents:
		event := record.(*model.Event)
		var memsqlEvent dest_Event
		err := convertStruct(event, &memsqlEvent)
		if err != nil {
			log.WithError(err).Error("Failed to convert event to memsql event.")
			return record, err
		}

		uuid, err := convertIDToUUIDFromInterface(event.EventNameId)
		if err != nil {
			return record, err
		}
		memsqlEvent.EventNameId = uuid
		return &memsqlEvent, nil

	case tablePropertyDetails:
		record := record.(*model.PropertyDetail)
		if record.EventNameID == nil {
			return &record, nil
		}

		uuid, err := convertIDToUUIDFromInterface(record.EventNameID)
		if err != nil {
			return record, err
		}
		record.EventNameID = &uuid
		return &record, nil

	case tableFactorsTrackedEvents:
		record := record.(*model.FactorsTrackedEvent)
		uuid, err := convertIDToUUIDFromInterface(record.EventNameID)
		if err != nil {
			return record, err
		}
		record.EventNameID = uuid
		return &record, nil
	}

	return record, nil
}

func createOrUpdateOnMemSQLInParallel(tableName string, records []interface{}, wg *sync.WaitGroup, states *MStates) {
	defer wg.Done()

	logCtx := log.WithField("table_name", tableName)
	for i := range records {
		record, err := convertColumnTypeByTable(tableName, records[i])
		if err != nil {
			// Do not proceed on failure. This stops the routine of the table.
			logCtx.WithError(err).WithField("record", records[i]).Error("Failed to convert column type")
			states.AddToState(MState{Status: http.StatusInternalServerError})
			return
		}

		sourceTableRecord, errCode := createIfNotExistOrUpdateIfChangedOnMemSQL(tableName, record)
		if errCode != http.StatusCreated {
			// Do not proceed on failure. This stops the routine of the table.
			logCtx.WithField("record", records[i]).Error("Failed to create or update record.")
			states.AddToState(MState{Status: http.StatusInternalServerError})
			return
		}

		states.AddToState(MState{Status: http.StatusCreated, UpdatedAt: sourceTableRecord.UpdatedAt})
	}
}

func migrateAllTables(projectIDs []int64) {
	// Runs replication continiously for each table
	// on a separate go routine.
	for i := range supportedTables {
		_, includeFound := includeTablesMap[supportedTables[i]]
		_, excludeFound := excludeTablesMap[supportedTables[i]]
		if (!migrageAllTables && !includeFound) || excludeFound {
			continue
		}
		go migrateTableContiniously(supportedTables[i], projectIDs)
	}

	select {} // Keeps main routine alive forever.
}

func migrateTableContiniously(table string, projectIDs []int64) {
	logCtx := log.WithField("table_name", table)

	for {
		logCtx.Infof("Replicating %s table.", table)
		status := migrateTableByNameWithPagination(table, projectIDs)
		if status == http.StatusInternalServerError {
			logCtx.Error("Migration failure. Stopping the replication.")
			C.PingHealthcheckForFailure(healthcheckPingID, fmt.Sprintf("Replication failed for table %s", table))
			break
		}

		// Sleep before next poll (cool-off).
		time.Sleep(30 * time.Second)
	}
}

func migrateTableByNameWithPagination(table string, projectIDs []int64) int {
	logCtx := log.WithField("table", table)

	for {
		var lastRunAt *time.Time
		metadata, status := GetReplicationMetadataByTable(table)
		if status == http.StatusFound {
			lastRunAt = &metadata.LastRunAt
		}
		if status == http.StatusInternalServerError {
			logCtx.Error("Failed to get last run metadata.")
			return http.StatusInternalServerError
		}

		newLastRunAt, rowCount, status := migrateTableByName(projectIDs, table, lastRunAt)
		if status != http.StatusOK {
			return http.StatusInternalServerError
		}

		// Breaks. if the row count on current page is 0.
		// But the poller will keep replicating.
		if rowCount == 0 {
			break
		}

		// Total no.of rows replicated so far.
		totalCount := uint64(rowCount)
		if metadata != nil {
			totalCount = metadata.Count + totalCount
		}

		// Check point is updated after processing each pull of data (page).
		if status := CreateOrUpdateReplicationMetadataByTable(table, newLastRunAt, totalCount); status != http.StatusAccepted {
			log.WithField("table", table).Error("Failed to update last run metadata.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func GetReplicationMetadataByTable(tableName string) (*model.ReplicationMetadata, int) {
	if tableName == "" {
		log.Error("Empty table name in GetReplicationMetadataByTable.")
		return nil, http.StatusInternalServerError
	}

	var metadata model.ReplicationMetadata

	db := getSourceDB()
	err := db.Model(&model.ReplicationMetadata{}).Limit(1).
		Where("table_name = ?", tableName).Find(&metadata).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithField("table", tableName).WithError(err).
			Error("Failed to get replication metadata for table on GetReplicationMetadataByTable.")
		return nil, http.StatusInternalServerError
	}

	return &metadata, http.StatusFound
}

func updateReplicationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	logCtx := log.WithField("table", tableName).WithField("last_run_at", lastRunAt)

	fields := make(map[string]interface{}, 0)
	if lastRunAt != nil {
		fields["last_run_at"] = *lastRunAt
	}
	if count > 0 {
		fields["count"] = count
	}

	db := getSourceDB()
	err := db.Model(&model.ReplicationMetadata{}).Where("table_name = ?", tableName).Update(fields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update the fields on replication_metadata.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func createReplicationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	logCtx := log.WithField("table", tableName).WithField("last_run_at", lastRunAt)

	db := getSourceDB()
	metadata := model.ReplicationMetadata{TableName: tableName, LastRunAt: *lastRunAt, Count: count}
	err := db.Create(&metadata).Error
	if err != nil {
		if isDuplicateError(err) {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create replication metadata.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func CreateOrUpdateReplicationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	status := createReplicationMetadataByTable(tableName, lastRunAt, count)
	if status == http.StatusConflict {
		return updateReplicationMetadataByTable(tableName, lastRunAt, count)
	}

	return status
}
