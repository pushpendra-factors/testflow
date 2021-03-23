package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	emoji "github.com/tmdvs/Go-Emoji-Utils"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm/dialects/postgres"

	C "factors/config"
	U "factors/util"

	"factors/model/model"
	"factors/model/store"
)

var memSQLDB *gorm.DB
var migrateAllProjects bool
var migrationPageSize int
var migrationRoutines int
var dedupeByQuery bool

const (
	tableUserProperties                = "user_properties"
	tableUsers                         = "users"
	tableEventNames                    = "event_names"
	tableEvents                        = "events"
	tableProjects                      = "projects"
	tableAgents                        = "agents"
	tableProjectAgentMappings          = "project_agent_mappings"
	tableProjectBillingAccountMappings = "project_billing_account_mappings"
	tableProjectSettings               = "project_settings"
	tableAdwordsDocuments              = "adwords_documents"
	tableBigquerySettings              = "bigquery_settings"
	tableBillingAccounts               = "billing_accounts"
	tableDashboardUnits                = "dashboard_units"
	tableDashboards                    = "dashboards"
	tableFacebookDocuments             = "facebook_documents"
	tableFactorsGoals                  = "factors_goals"
	tableFactorsTrackedEvents          = "factors_tracked_events"
	tableFactorsTrackedUserProperties  = "factors_tracked_user_properties"
	tableHubspotDocuments              = "hubspot_documents"
	tableSalesforceDocuments           = "salesforce_documents"
	tableQueries                       = "queries"
	tableReports                       = "reports"
	tableScheduledTasks                = "scheduled_tasks"
)

type TableRecord struct {
	ProjectID uint64      `json:"project_id"`
	ID        interface{} `json:"id"`      // Using interface, as type of ID can be uint64 or uuid.
	UUID      interface{} `json:"uuid"`    // Agents table is using uuid instead of id.
	UserID    string      `json:"user_id"` // For analytics tables.
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
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
	memSQLDSN := flag.String(
		"memsql_dsn",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			*memSQLUser, *memSQLPass, *memSQLHost, *memSQLPort, *memSQLName),
		"",
	)

	projectIDStringList := flag.String("project_ids", "", "")
	pageSize := flag.Int("page_size", 0, "No.of records per page.")
	routines := flag.Int("routines", 10, "No.of parallel routines to use for each page.")
	dedupeByQueryForNonUniqueTables := flag.Bool("dedupe_by_query", false,
		"Dedupe's by firing a get query everytime for events and user_properties table.")

	// Migrates events table with dependencies for a timerange of events.
	migrateEventsWithDependencies := flag.Bool("events_with_dep", false, "")
	eventsStartTimestamp := flag.Int64("events_start_timestamp", 0, "")
	eventsEndTimestamp := flag.Int64("events_end_timestamp", 0, "")
	eventsNumRouties := flag.Int("events_num_routines", 1, "")
	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	log.SetFormatter(&log.JSONFormatter{})
	if *projectIDStringList == "" {
		log.Fatal("Invalid project_id.")
	}

	if *pageSize == 0 {
		log.Fatal("Migration page size cannot be zero.")
	}
	migrationPageSize = *pageSize
	migrationRoutines = *routines
	dedupeByQuery = *dedupeByQueryForNonUniqueTables

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

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}

	maxOpenConns := (30 * migrationRoutines) + 10
	initMemSQLDB(*env, *memSQLDSN, maxOpenConns)

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDStringList, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if allProjects {
		migrateAllProjects = true
	}

	if len(projectIDs) == 0 && !migrateAllProjects {
		log.Fatal("Invalid value on project_ids flag. Should be list of project_ids or *.")
	}

	if *migrateEventsWithDependencies {
		if len(projectIDs) == 0 {
			log.Fatal("No projects given for migrating events table with dependencies.")
		}

		for i := range projectIDs {
			log.Infof("Running in migrate events with dependencies mode for project %d.", projectIDs[i])
			migrateEventsTableAndDepedencies(projectIDs[i], *eventsStartTimestamp, *eventsEndTimestamp, *eventsNumRouties)
		}
		return
	}

	log.WithField("project_ids", projectIDs).Info("Running in replication mode.")
	migrateAllTables(projectIDs)
}

func initMemSQLDB(env, dsn string, maxOpenConns int) {
	var err error
	// dsn sample admin:LpAHQyAMyI@tcp(svc-2b9e36ee-d5d0-4082-9779-2027e39fcbab-ddl.gcp-virginia-1.db.memsql.com:3306)/factors?charset=utf8mb4&parseTime=True&loc=Local
	memSQLDB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.WithError(err).Fatal("Failed connecting to memsql.")
	}
	// Disables before callback for create method.
	memSQLDB.Callback().Create().Remove("gorm:before_create")
	// Removes unneccesary select after insert triggered by gorm.
	memSQLDB.Callback().Create().Remove("gorm:force_reload_after_create")
	// Removes emoji and cleans up string and postgres.Jsonb columns.
	memSQLDB.Callback().Create().Before("gorm:create").Register("cleanup", gormCleanupCallback)

	if C.IsDevelopment() {
		memSQLDB.LogMode(true)
	} else {
		memSQLDB.LogMode(false)
		memSQLDB.DB().SetMaxOpenConns(maxOpenConns)
		memSQLDB.DB().SetMaxIdleConns(100)
	}
}

func sanitizeStringValue(s string) string {
	return emoji.RemoveAll(s)
}

// Custom GORM Plugin for cleaning up field values.
func gormCleanupCallback(scope *gorm.Scope) {
	for _, field := range scope.Fields() {
		switch field.Field.Type().String() {
		case "string":
			fieldValue := field.Field.Interface().(string)
			err := field.Set(sanitizeStringValue(fieldValue))
			if err != nil {
				log.WithError(err).Error("Failed to cleanup string field value.")
				return
			}
		case "postgres.Jsonb":
			fieldValue := field.Field.Interface().(postgres.Jsonb)
			jsonAsString := string(fieldValue.RawMessage)
			fieldValue.RawMessage = []byte(sanitizeStringValue(jsonAsString))
			err := field.Set(fieldValue)
			if err != nil {
				log.WithError(err).Error("Failed to cleanup postgres jsonb field value.")
				return
			}
		}
	}
}

func getDedupeMap(tableName string) *sync.Map {
	switch tableName {
	case tableEvents:
		return &eventsMap
	case tableEventNames:
		return &eventNamesMap
	case tableUsers:
		return &usersMap
	case tableUserProperties:
		return &userPropertiesMap
	default:
		log.Fatalf("Unsupported table %s on getDedupeMap.", tableName)
		return nil
	}
}

// Move this method to model, only if it is being used on production.
func getEventNameByID(projectId uint64, id uint64) (*model.EventName, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "event_name_id": id})

	var eventName model.EventName
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&eventName).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get event_name by id.")
		return nil, http.StatusInternalServerError
	}

	return &eventName, http.StatusFound
}

// Move this method to model, only if it is being used on production.
func getUserByID(projectId uint64, id string) (*model.User, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": id})

	var user model.User
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get user by id.")
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

func doesExistByUserAndIDOnMemSQLDB(tableName string, projectID uint64,
	userID string, id interface{}) (bool, *TableRecord, error) {

	logCtx := log.WithField("project_id", projectID).
		WithField("id", id).WithField("user_id", userID).
		WithField("table", tableName)

	var record TableRecord
	// Select only id to use index only read and avoid disk read.
	err := memSQLDB.Table(tableName).Select("id, updated_at").Limit(1).
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

func createOnMemSQLWithInMemoryDedupe(projectID uint64, tableName string, record interface{}) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName)

	pgTableRecord, err := convertToTableRecord(tableName, record)
	if err != nil {
		logCtx.WithError(err).Error("Failed to convert pg record to table record.")
		return http.StatusInternalServerError
	}

	return createOnMemSQL(projectID, tableName, record, pgTableRecord.ID, true)
}

// createEventAndDependenciesOnMemSQL - creates event with all associated dependencies.
func createEventAndDependenciesOnMemSQL(event *model.Event, inMemoryDedupe bool, wg *sync.WaitGroup) {
	defer wg.Done()

	logCtx := log.WithField("event", event)

	// Skip existence check using db for inmemory dedupe.
	if !inMemoryDedupe {
		// Return, if event already exists.
		exists, tableRecord, err := doesExistByUserAndIDOnMemSQLDB(tableEvents, event.ProjectId, event.UserId, event.ID)
		if err != nil {
			return
		}
		if exists {
			if isUpdated := event.UpdatedAt.After(tableRecord.UpdatedAt); !isUpdated {
				return
			}
		}
	}

	if event.EventNameId != 0 {
		eventName, errCode := getEventNameByID(event.ProjectId, event.EventNameId)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get user associated to event.")
			return
		}

		if inMemoryDedupe {
			errCode = createOnMemSQLWithInMemoryDedupe(event.ProjectId, tableEventNames, eventName)
		} else {
			_, errCode = createIfNotExistOrUpdateIfChangedOnMemSQL(tableEventNames, eventName)
		}
		if errCode != http.StatusCreated {
			return
		}
	}

	var latestUserPropertiesID string
	if event.UserId != "" {
		logCtx = logCtx.WithField("user_id", event.UserId)

		user, errCode := getUserByID(event.ProjectId, event.UserId)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get user associated to event.")
			return
		}
		latestUserPropertiesID = user.PropertiesId

		if inMemoryDedupe {
			errCode = createOnMemSQLWithInMemoryDedupe(event.ProjectId, tableUsers, user)
		} else {
			_, errCode = createIfNotExistOrUpdateIfChangedOnMemSQL(tableUsers, user)
		}
		if errCode != http.StatusCreated {
			return
		}
	}

	// Create latest user_properties of user.
	if latestUserPropertiesID != "" {
		errCode := createUserPropertiesOnMemSQLDB(event.ProjectId, event.UserId, latestUserPropertiesID, inMemoryDedupe)
		if errCode != http.StatusCreated {
			logCtx.WithField("err_code", errCode).Error("Failed to create latest user_properties of user.")
			return
		}
	}

	if event.UserPropertiesId != "" {
		errCode := createUserPropertiesOnMemSQLDB(event.ProjectId, event.UserId, event.UserPropertiesId, inMemoryDedupe)
		if errCode != http.StatusCreated {
			logCtx.WithField("err_code", errCode).Error("Failed to create user_properties from event.")
			return
		}
	}

	if inMemoryDedupe {
		createOnMemSQLWithInMemoryDedupe(event.ProjectId, tableEvents, event)
		return
	}

	createIfNotExistOrUpdateIfChangedOnMemSQL(tableEvents, event)
}

func createUserPropertiesOnMemSQLDB(projectID uint64, userID, userPropertiesID string, inMemoryDedupe bool) int {
	logCtx := log.WithField("project_id", projectID).
		WithField("user_id", userID).
		WithField("user_properties_id", userPropertiesID)

	userProperties, errCode := store.GetStore().GetUserPropertiesRecord(projectID, userID, userPropertiesID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get user_properties associated to event.")
		return http.StatusInternalServerError
	}

	if inMemoryDedupe {
		return createOnMemSQLWithInMemoryDedupe(projectID, tableUserProperties, userProperties)
	}

	_, status := createIfNotExistOrUpdateIfChangedOnMemSQL(tableUserProperties, userProperties)
	return status
}

// Copies records from events table and dependencies user by user_id,
// user_properties by user_properties_id, event_names by event_names_id.
// Also copies the user_properties by properties_id while copying users from events table.
func migrateEventsTableAndDepedencies(projectID uint64, startTimestamp, endTimestamp int64, numRoutines int) {
	logCtx := log.WithField("project_id", projectID).
		WithField("start_timestamp", startTimestamp).
		WithField("end_timestamp", endTimestamp).
		WithField("num_routines", numRoutines)

	if startTimestamp == 0 || endTimestamp == 0 {
		log.Fatal("For start_timestamp and end_timestamp cannot be zero.")
	}

	logCtx.Info("Migrating with events with dependencies.")

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT * FROM events WHERE project_id = ? AND timestamp BETWEEN ? AND ?",
		projectID, startTimestamp, endTimestamp).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to pull events.")
		return
	}
	defer rows.Close()

	events := make([]model.Event, 0, 0)
	for rows.Next() {
		var event model.Event
		if err := db.ScanRows(rows, &event); err != nil {
			log.WithError(err).Error("Failed to scan event.")
			continue
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		log.WithError(err).Fatal("Failure on rows scanner.")
	}

	logCtx.WithField("rows", len(events))
	logCtx.Info("Fetched events. Pushing to memsql.")

	runCreateEventAndDependenciesOnMemSQL(getEventsListAsBatch(events, numRoutines))

	logCtx.Info("Successfully pushed from postgres to memsql.")
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

func runCreateEventAndDependenciesOnMemSQL(eventsList [][]model.Event) {
	for eli := range eventsList {
		var wg sync.WaitGroup
		wg.Add(len(eventsList[eli]))

		for ei := range eventsList[eli] {
			// using inmemory dedupe.
			go createEventAndDependenciesOnMemSQL(&eventsList[eli][ei], true, &wg)
		}

		wg.Wait()
	}
}

func getTableRecordByIDFromMemSQL(projectID uint64, tableName string, id interface{}, userID string) (*TableRecord, int) {
	logCtx := log.WithField("project_id", projectID).WithField("id", id).
		WithField("table", tableName)

	if tableName == "" || id == nil {
		return nil, http.StatusBadRequest
	}

	if !isTableWithoutProjectID(tableName) && projectID == 0 {
		return nil, http.StatusBadRequest
	}

	var record TableRecord

	idColName := "id"
	if tableName == tableAgents {
		idColName = "uuid"
	}
	if isProjectAssociatedTable(tableName) {
		idColName = "project_id"
	}

	db := memSQLDB.Table(tableName).Limit(1).Where(fmt.Sprintf("%s = ?", idColName), id)
	if !isTableWithoutProjectID(tableName) {
		db = db.Where("project_id = ?", projectID)
	}

	// Add user_id also part of the filter to speed up the query.
	if tableName == tableEvents || tableName == tableUserProperties {
		if userID == "" {
			logCtx.Error("user_id not provided on getTableRecordByIDFromMemSQL")
		} else {
			db = db.Where("user_id = ?", userID)
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

func deleteByIDOnMemSQL(projectID uint64, tableName string, id interface{}, userID string) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName).
		WithField("id", id).WithField("user_id", userID)

	if (!isTableWithoutProjectID(tableName) && projectID == 0) || tableName == "" || id == nil {
		return http.StatusBadRequest
	}

	idColName := "id"
	if tableName == tableAgents {
		idColName = "uuid"
	}
	if isProjectAssociatedTable(tableName) {
		idColName = "project_id"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", tableName, idColName)
	if !isTableWithoutProjectID(tableName) {
		query = fmt.Sprintf("%s AND project_id = %d", query, projectID)
	}

	// Add user_id also part of the filter to speed up the query.
	if tableName == tableEvents || tableName == tableUserProperties {
		if userID == "" {
			logCtx.Error("user_id not provided on deleteByIDOnMemSQL")
		} else {
			query = fmt.Sprintf("%s AND user_id = '%s'", query, userID)
		}
	}

	rows, err := memSQLDB.Raw(query, id).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed deleting record in memsql.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusOK
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate")
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

func createOnMemSQL(projectID uint64, tableName string, record interface{}, id interface{}, inMemoryDedupe bool) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName).
		WithField("record", record)

	if inMemoryDedupe {
		dedupeMap := getDedupeMap(tableName)
		if _, exists := dedupeMap.Load(id); exists {
			return http.StatusCreated
		}

		// Mark it as inserted quickly to avoid duplicate
		// insertion on another go routine.
		dedupeMap.Store(id, true)
	}

	if err := memSQLDB.Create(record).Error; err != nil {
		if isDuplicateError(err) {
			return http.StatusConflict
		}

		if inMemoryDedupe {
			dedupeMap := getDedupeMap(tableName)
			dedupeMap.Delete(id) // Delete, if create failed.
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

	// Project one-one tables uses project_id as ID column.
	if isProjectAssociatedTable(tableName) {
		tableRecord.ID = tableRecord.ProjectID
	}

	return &tableRecord, nil
}

func isTableWithoutUniquePrimaryKey(tableName string) bool {
	return U.StringValueIn(tableName, []string{tableEvents, tableUserProperties})
}

func isProjectAssociatedTable(tableName string) bool {
	return U.StringValueIn(tableName, []string{tableProjectBillingAccountMappings,
		tableProjectSettings, tableProjectAgentMappings})
}

func updateIfExistOnMemSQL(projectID uint64, tableName string, pgRecord interface{}, pgTableRecord, memsqlTableRecord *TableRecord) int {
	logCtx := log.WithField("project_id", projectID).WithField("table_name", tableName)

	var status int
	if memsqlTableRecord == nil {
		memsqlTableRecord, status = getTableRecordByIDFromMemSQL(projectID, tableName, pgTableRecord.ID, pgTableRecord.UserID)
		if status == http.StatusInternalServerError || status == http.StatusBadRequest {
			logCtx.WithField("status", status).Error("Failed to get the existing record from memsql.")
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
		status := deleteByIDOnMemSQL(projectID, tableName, pgTableRecord.ID, pgTableRecord.UserID)
		if status != http.StatusOK {
			return http.StatusInternalServerError
		}

		status = createOnMemSQL(projectID, tableName, pgRecord, pgTableRecord.ID, false)
		if status != http.StatusCreated && status != http.StatusConflict {
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
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

	if isTableWithoutUniquePrimaryKey(tableName) && dedupeByQuery {
		// TODO: Try to add a procedure/trigger to do the existence check before create on memsql itself for long term.
		// Now doing it on the script itself, as it is sequential.
		currentMemsqlRecord, status := getTableRecordByIDFromMemSQL(pgTableRecord.ProjectID, tableName, pgTableRecord.ID, pgTableRecord.UserID)
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
	case tableReports:
		record = &model.Report{}

	// Tables related to analytics.
	case tableEvents:
		record = &model.Event{}
	case tableUsers:
		record = &model.User{}
	case tableUserProperties:
		record = &model.UserProperties{}
	case tableEventNames:
		record = &model.EventName{}

	// Tables without project_id.
	case tableAgents:
		record = &model.Agent{}
	case tableProjects:
		record = &model.Project{}

	default:
		// Critical error. Should not proceed without adding support for table.
		log.Fatalf("Unsupported table name %s on get_record_interface_by_table_name", tableName)
	}

	return record
}

func isTableWithoutProjectID(tableName string) bool {
	return U.StringValueIn(tableName, []string{tableAgents, tableProjects, tableBillingAccounts})
}

func buildStringList(ids []uint64) string {
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

func migrateTableByName(projectIDs []uint64, tableName string, lastPageUpdatedAt *time.Time) (*time.Time, int, int) {

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

	query := fmt.Sprintf("SELECT * FROM %s", tableName)

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
		default:
			whereProjectCondition = fmt.Sprintf("project_id = %s", anyProjectsCondition)
		}

		query = query + " " + whereProjectCondition
	}

	if lastPageUpdatedAt != nil {
		if where {
			query = query + " " + "AND"
		} else {
			query = query + " " + "WHERE"
		}

		query = fmt.Sprintf("%s updated_at > '%+v'", query, lastPageUpdatedAt.Format(time.RFC3339Nano))
	}

	query = query + " " + fmt.Sprintf("ORDER BY updated_at ASC LIMIT %d", migrationPageSize)

	db := C.GetServices().Db
	rows, err := db.Raw(query).Rows()
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
	recordChunks := getRecordsAsChunks(recordIntfs, migrationRoutines)
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

func createOrUpdateOnMemSQLInParallel(tableName string, records []interface{}, wg *sync.WaitGroup, states *MStates) {
	defer wg.Done()

	logCtx := log.WithField("table_name", tableName)
	for i := range records {
		sourceTableRecord, errCode := createIfNotExistOrUpdateIfChangedOnMemSQL(tableName, records[i])
		if errCode != http.StatusCreated {
			// Do not proceed on failure. This stops the routine of the table.
			logCtx.WithField("record", records[i]).Error("Failed to create or update record.")
			states.AddToState(MState{Status: http.StatusInternalServerError})
			return
		}

		states.AddToState(MState{Status: http.StatusCreated, UpdatedAt: sourceTableRecord.UpdatedAt})
	}
}

func migrateAllTables(projectIDs []uint64) {
	tables := []string{
		tableUserProperties,
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
		tableReports,
		tableScheduledTasks,
	}

	// Runs replication continiously for each table
	// on a separate go routine.
	for i := range tables {
		go migrateTableContiniously(tables[i], projectIDs)
	}

	select {} // Keeps main routine alive forever.
}

func migrateTableContiniously(table string, projectIDs []uint64) {
	logCtx := log.WithField("table_name", table)

	for {
		logCtx.Infof("Replicating %s table.", table)
		status := migrateTableByNameWithPagination(table, projectIDs)
		if status == http.StatusInternalServerError {
			logCtx.Error("Migration failure. Stopping the replication.")
			break
		}

		// Sleep before next poll (cool-off).
		time.Sleep(30 * time.Second)
	}
}

func migrateTableByNameWithPagination(table string, projectIDs []uint64) int {
	logCtx := log.WithField("table", table)

	for {
		var lastRunAt *time.Time
		metadata, status := store.GetStore().GetReplicationMetadataByTable(table)
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
		if status := store.GetStore().CreateOrUpdateReplicationMetadataByTable(table, newLastRunAt, totalCount); status != http.StatusAccepted {
			log.WithField("table", table).Error("Failed to update last run metadata.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}
