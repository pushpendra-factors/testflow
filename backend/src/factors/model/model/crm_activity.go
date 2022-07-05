package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CRMActivity struct {
	ID                 string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID          int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ExternalActivityID string          `gorm:"primary_key:true;auto_increment:false" json:"external_activity_id"`
	Source             U.CRMSource     `gorm:"primary_key:true;auto_increment:false" json:"source"`
	Name               string          `gorm:"primary_key:true;auto_increment:false" json:"name"`
	Type               int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	ActorType          int             `gorm:"primary_key:true;auto_increment:false" json:"actor_type"`
	ActorID            string          `gorm:"primary_key:true;auto_increment:false" json:"actor_id"`
	Timestamp          int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	Properties         *postgres.Jsonb `json:"properties"`
	Synced             bool            `json:"synced"`
	SyncID             string          `json:"sync_id"`
	UserID             string          `json:"user_id"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// GetCRMActivityBatchedOrderedRecordsByID return list of document in batches. Order is maintained on records id.
func GetCRMActivityBatchedOrderedRecordsByID(records []CRMActivity, batchSize int) []map[string]interface{} {

	if len(records) < 0 {
		return nil
	}

	recordsMap := make(map[string][]CRMActivity)
	for i := range records {
		if _, exist := recordsMap[records[i].ID]; !exist {
			recordsMap[records[i].ID] = make([]CRMActivity, 0)
		}
		recordsMap[records[i].ID] = append(recordsMap[records[i].ID], records[i])
	}

	batchedRecordsByID := make([]map[string]interface{}, 1)
	isBatched := make(map[string]bool)
	batchLen := 0
	batchedRecordsByID[batchLen] = make(map[string]interface{})
	for i := range records {
		if isBatched[records[i].ID] {
			continue
		}

		if len(batchedRecordsByID[batchLen]) >= batchSize {
			batchedRecordsByID = append(batchedRecordsByID, make(map[string]interface{}))
			batchLen++
		}

		batchedRecordsByID[batchLen][records[i].ID] = recordsMap[records[i].ID]
		isBatched[records[i].ID] = true
	}

	return batchedRecordsByID
}
