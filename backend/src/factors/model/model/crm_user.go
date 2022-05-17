package model

import (
	"errors"
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CRMUser struct {
	ID         string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID  uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source     U.CRMSource     `gorm:"primary_key:true;auto_increment:false" json:"source"`
	Type       int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Timestamp  int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	Email      string          `gorm:"default:null" json:"email"`
	Phone      string          `gorm:"default:null" json:"phone"`
	Action     CRMAction       `gorm:"auto_increment:false;not null" json:"action"`
	Metadata   *postgres.Jsonb `json:"metadata"`
	Properties *postgres.Jsonb `json:"properties"`
	Synced     bool            `gorm:"default:false" json:"synced"`
	SyncID     string          `gorm:"default:null" json:"sync_id"`
	UserID     string          `gorm:"default:null" json:"user_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type CRMAction int

const (
	CRMActionCreated CRMAction = 1
	CRMActionUpdated CRMAction = 2
	CRMActionDeleted CRMAction = 2
)

var ALLOWED_CRM_SOURCES = map[U.CRMSource]bool{
	U.CRM_SOURCE_HUBSPOT:    true,
	U.CRM_SOURCE_SALESFORCE: true,
	U.CRM_SOURCE_MARKETO:    true,
}

var CRM_SOURCE = map[U.CRMSource]string{
	U.CRM_SOURCE_HUBSPOT:    U.CRM_SOURCE_NAME_HUBSPOT,
	U.CRM_SOURCE_SALESFORCE: U.CRM_SOURCE_NAME_SALESFORCE,
	U.CRM_SOURCE_MARKETO:    U.CRM_SOURCE_NAME_MARKETO,
}

func AllowedCRMBySource(crmSource U.CRMSource) bool {
	return ALLOWED_CRM_SOURCES[crmSource]
}

func IsCRMSource(source string) bool {
	for _, crmSource := range CRM_SOURCE {
		if source == crmSource {
			return true
		}
	}
	return false
}

func GetCRMSourceByAliasName(sourceAlias string) (U.CRMSource, error) {
	for sourceType, alias := range CRM_SOURCE {
		if sourceAlias == alias {
			return sourceType, nil
		}
	}

	return 0, errors.New("invalid source alias")
}

// GetCRMUserBatchedOrderedRecordsByID return list of document in batches. Order is maintained on records id.
func GetCRMUserBatchedOrderedRecordsByID(records []CRMUser, batchSize int) []map[string]interface{} {

	if len(records) < 0 {
		return nil
	}

	recordsMap := make(map[string][]CRMUser)
	for i := range records {
		if _, exist := recordsMap[records[i].ID]; !exist {
			recordsMap[records[i].ID] = make([]CRMUser, 0)
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
