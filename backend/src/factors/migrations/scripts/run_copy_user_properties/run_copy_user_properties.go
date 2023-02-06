package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var migrateAllProjects bool
var migrationPageSize int
var migrationRoutines int

// UserPropertiesMigrationMetadata - Table struct for check pointing.
type UserPropertiesMigrationMetadata struct {
	TableName string    `json:"table_name"`
	LastRunAt time.Time `json:"last_run_at"`
	Count     uint64    `json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MigrationRecord - Generic struct for fields from events and users.
type MigrationRecord struct {
	ProjectID        uint64
	ID               string
	UserPropertiesID string
	UserID           string
	UpdatedAt        time.Time
}

type MigrationState struct {
	Status    int
	UpdatedAt time.Time
}

type MigrationStates struct {
	mu      sync.Mutex
	mstates []MigrationState
}

func (s *MigrationStates) AddToState(mstate MigrationState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mstates = append(s.mstates, mstate)
}

const tableEvents = "events"
const tableUsers = "users"

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

func getOnTableUserPropertiesTableRecord(tableName string) interface{} {
	var record interface{}

	switch tableName {
	case tableEvents:
		record = &model.Event{}
	case tableUsers:
		record = &model.User{}
	default:
		// Critical error. Should not proceed without adding support for table.
		log.Fatalf("Unsupported table name %s on get_record_interface_by_table_name", tableName)
	}

	return record
}

func migrateByTableName(projectIDs []uint64, tableName string, lastPageUpdatedAt *time.Time) (*time.Time, int, int) {

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

	query := fmt.Sprintf("SELECT * FROM %s", tableName)

	var where bool
	if !migrateAllProjects {
		var whereProjectCondition string
		query = query + " " + "WHERE"
		anyProjectsCondition := fmt.Sprintf("ANY(ARRAY[%s])", buildStringList(projectIDs))
		whereProjectCondition = fmt.Sprintf("project_id = %s", anyProjectsCondition)

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

	var rowsCount int
	recordIntfs := make([]interface{}, 0, 0)
	for rows.Next() {
		record := getOnTableUserPropertiesTableRecord(tableName)
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
	var states MigrationStates
	recordChunks := util.GetInterfaceListAsBatch(recordIntfs, migrationRoutines)
	for i := range recordChunks {
		wg.Add(1)
		go copyUserPropertiesToTableRecordWorker(tableName, recordChunks[i], &wg, &states)
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

func copyUserPropertiesToTableRecordWorker(tableName string, records []interface{},
	wg *sync.WaitGroup, states *MigrationStates) {

	defer wg.Done()

	logCtx := log.WithField("table_name", tableName)
	for i := range records {
		migrationRecord, errCode := copyUserPropertiesToTableRecord(tableName, records[i])
		if errCode != http.StatusOK {
			// Do not proceed on failure. This stops the routine of the table.
			logCtx.WithField("err_code", errCode).
				WithField("record", records[i]).Error("Failed to copy user properties for record.")
			states.AddToState(MigrationState{Status: http.StatusInternalServerError})
			return
		}

		states.AddToState(MigrationState{Status: http.StatusCreated, UpdatedAt: migrationRecord.UpdatedAt})
	}
}

func getMigrationRecord(tableName string, record interface{}) *MigrationRecord {
	var migrationRecord MigrationRecord

	switch tableName {
	case tableEvents:
		event := record.(*model.Event)
		migrationRecord.UpdatedAt = event.UpdatedAt
		migrationRecord.ProjectID = event.ProjectId
		migrationRecord.UserID = event.UserId
		migrationRecord.UserPropertiesID = event.UserPropertiesId
		migrationRecord.ID = event.ID

	case tableUsers:
		user := record.(*model.User)
		migrationRecord.UpdatedAt = user.UpdatedAt
		migrationRecord.ProjectID = user.ProjectId
		migrationRecord.UserID = user.ID
		migrationRecord.UserPropertiesID = user.PropertiesId
		migrationRecord.ID = user.ID
	}

	return &migrationRecord
}

func updateUserPropertiesStateOnTableByID(tableName string, projectID uint64, id string,
	properties postgres.Jsonb, propertiesTimestamp int64) int {

	logCtx := log.WithField("project_id", projectID).WithField("id", id).
		WithField("properties", properties).WithField("properties_timestamp", propertiesTimestamp)

	var update map[string]interface{}
	if tableName == tableEvents {
		update = map[string]interface{}{
			"user_properties": properties,
		}
	} else if tableName == tableUsers {
		update = map[string]interface{}{
			"properties":                   properties,
			"properties_updated_timestamp": propertiesTimestamp,
		}
	} else {
		logCtx.Error("Invalid table name.")
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db
	err := db.Table(tableName).Where("project_id = ? AND id = ?", projectID, id).Update(update).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to copy properties to table")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func copyUserPropertiesToTableRecord(tableName string, record interface{}) (*MigrationRecord, int) {
	logCtx := log.WithField("record", record)

	migrationRecord := getMigrationRecord(tableName, record)
	if migrationRecord.UserPropertiesID == "" {
		return migrationRecord, http.StatusOK
	}

	userPropertiesRecord, errCode := store.GetStore().GetUserPropertiesRecord(
		migrationRecord.ProjectID, migrationRecord.UserID, migrationRecord.UserPropertiesID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to find the user_properties record.")
		return migrationRecord, http.StatusInternalServerError
	}

	errCode = updateUserPropertiesStateOnTableByID(tableName, migrationRecord.ProjectID,
		migrationRecord.ID, userPropertiesRecord.Properties, userPropertiesRecord.UpdatedTimestamp)
	if errCode != http.StatusAccepted {
		logCtx.WithField("err_code", errCode).Error("Failed to update the user_properties on record.")
		return migrationRecord, http.StatusInternalServerError
	}

	return migrationRecord, http.StatusOK
}

func getUserPropertiesMigrationMetadata(tableName string) (*UserPropertiesMigrationMetadata, int) {
	if tableName == "" {
		log.Error("Empty table name in getUserPropertiesMigrationMetadata.")
		return nil, http.StatusInternalServerError
	}

	var metadata UserPropertiesMigrationMetadata

	db := C.GetServices().Db
	err := db.Model(&UserPropertiesMigrationMetadata{}).Limit(1).
		Where("table_name = ?", tableName).Find(&metadata).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithField("table", tableName).WithError(err).
			Error("Failed to get migration metadata for table on getUserPropertiesMigrationMetadata.")
		return nil, http.StatusInternalServerError
	}

	return &metadata, http.StatusFound
}

func migrateTableByNameWithPagination(table string, projectIDs []uint64) int {
	logCtx := log.WithField("table", table)

	for {
		var lastRunAt *time.Time
		metadata, status := getUserPropertiesMigrationMetadata(table)
		if status == http.StatusFound {
			lastRunAt = &metadata.LastRunAt
		}
		if status == http.StatusInternalServerError {
			logCtx.Error("Failed to get last run metadata.")
			return http.StatusInternalServerError
		}

		newLastRunAt, rowCount, status := migrateByTableName(projectIDs, table, lastRunAt)
		if status != http.StatusOK {
			return http.StatusInternalServerError
		}

		// Breaks. if the row count on current page is 0.
		// But the poller will keep migrating.
		if rowCount == 0 {
			break
		}

		// Total no.of rows migrated so far.
		totalCount := uint64(rowCount)
		if metadata != nil {
			totalCount = metadata.Count + totalCount
		}

		// Check point is updated after processing each pull of data (page).
		if status := createOrUpdateMigrationMetadataByTable(table, newLastRunAt, totalCount); status != http.StatusAccepted {
			log.WithField("table", table).Error("Failed to update last run metadata.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func updateMigrationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	logCtx := log.WithField("table", tableName).WithField("last_run_at", lastRunAt)

	fields := make(map[string]interface{}, 0)
	if lastRunAt != nil {
		fields["last_run_at"] = *lastRunAt
	}
	if count > 0 {
		fields["count"] = count
	}

	db := C.GetServices().Db
	err := db.Model(&UserPropertiesMigrationMetadata{}).Where("table_name = ?", tableName).Update(fields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update the fields on migration metadata.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func createMigrationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	logCtx := log.WithField("table", tableName).WithField("last_run_at", lastRunAt)

	db := C.GetServices().Db
	metadata := UserPropertiesMigrationMetadata{TableName: tableName, LastRunAt: *lastRunAt, Count: count}
	err := db.Create(&metadata).Error
	if err != nil {
		if U.IsPostgresIntegrityViolationError(err) {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create migration metadata.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func createOrUpdateMigrationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	status := createMigrationMetadataByTable(tableName, lastRunAt, count)
	if status == http.StatusConflict {
		return updateMigrationMetadataByTable(tableName, lastRunAt, count)
	}

	return status
}

func migrateTableContiniously(table string, projectIDs []uint64) {
	logCtx := log.WithField("table_name", table)

	for {
		logCtx.Infof("Migrating %s table.", table)
		status := migrateTableByNameWithPagination(table, projectIDs)
		if status == http.StatusInternalServerError {
			logCtx.Error("Migration failure. Stopping the migration.")
			break
		}

		// Sleep before next poll (cool-off).
		time.Sleep(30 * time.Second)
	}
}

func migrateAllTables(projectIDs []uint64) {
	tables := []string{
		tableUsers,
		tableEvents,
	}

	// Runs migration continiously for each table
	// on a separate go routine.
	for i := range tables {
		go migrateTableContiniously(tables[i], projectIDs)
	}

	select {} // Keeps main routine alive forever.
}

const appName = "copy_user_properties_migration"

func main() {
	env := flag.String("env", "development", "")

	projectIDStringList := flag.String("project_ids", "", "")
	pageSize := flag.Int("page_size", 0, "No.of records per page.")
	routines := flag.Int("routines", 10, "No.of parallel routines to use for each page.")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()
	defer util.NotifyOnPanic("copy_user_properties_migration", *env)

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

	config := &C.Configuration{
		Env:       *env,
		AppName:   appName,
		SentryDSN: *sentryDSN,
	}
	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 900, 250)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db.")
	}

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDStringList, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if allProjects {
		migrateAllProjects = true
	}

	if len(projectIDs) == 0 && !migrateAllProjects {
		log.Fatal("Invalid value on project_ids flag. Should be list of project_ids or *.")
	}

	log.WithField("project_ids", projectIDs).Info("Running continious migration.")
	migrateAllTables(projectIDs)
}
