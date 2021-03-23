package memsql

import (
	"net/http"
	"time"

	C "factors/config"
	"factors/model/model"
	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetReplicationMetadataByTable(tableName string) (*model.ReplicationMetadata, int) {
	if tableName == "" {
		log.Error("Empty table name in GetReplicationMetadataByTable.")
		return nil, http.StatusInternalServerError
	}

	var metadata model.ReplicationMetadata

	db := C.GetServices().Db
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

	db := C.GetServices().Db
	err := db.Model(&model.ReplicationMetadata{}).Where("table_name = ?", tableName).Update(fields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update the fields on replication_metadata.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func createReplicationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	logCtx := log.WithField("table", tableName).WithField("last_run_at", lastRunAt)

	db := C.GetServices().Db
	metadata := model.ReplicationMetadata{TableName: tableName, LastRunAt: *lastRunAt, Count: count}
	err := db.Create(&metadata).Error
	if err != nil {
		if U.IsPostgresIntegrityViolationError(err) {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create replication metadata.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) CreateOrUpdateReplicationMetadataByTable(tableName string, lastRunAt *time.Time, count uint64) int {
	status := createReplicationMetadataByTable(tableName, lastRunAt, count)
	if status == http.StatusConflict {
		return updateReplicationMetadataByTable(tableName, lastRunAt, count)
	}

	return status
}
