package model

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CRMUser struct {
	ID         string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID  uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source     CRMSource       `gorm:"primary_key:true;auto_increment:false" json:"source"`
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

type CRMSource int

const (
	CRM_SOURCE_HUBSPOT         CRMSource = 1
	CRM_SOURCE_SALESFORCE      CRMSource = 2
	CRM_SOURCE_MARKETO         CRMSource = 3
	CRM_SOURCE_NAME_HUBSPOT              = "hubspot"
	CRM_SOURCE_NAME_SALESFORCE           = "salesforce"
	CRM_SOURCE_NAME_MARKETO              = "marketo"
)

var ALLOWED_CRM_SOURCES = map[CRMSource]bool{
	CRM_SOURCE_HUBSPOT:    true,
	CRM_SOURCE_SALESFORCE: true,
	CRM_SOURCE_MARKETO:    true,
}

var CRM_SOURCE = map[CRMSource]string{
	CRM_SOURCE_HUBSPOT:    CRM_SOURCE_NAME_HUBSPOT,
	CRM_SOURCE_SALESFORCE: CRM_SOURCE_NAME_SALESFORCE,
	CRM_SOURCE_MARKETO:    CRM_SOURCE_NAME_MARKETO,
}

func AllowedCRMBySource(crmSource CRMSource) bool {
	return ALLOWED_CRM_SOURCES[crmSource]
}

func GetCRMSourceByAliasName(sourceAlias string) (CRMSource, error) {
	for sourceType, alias := range CRM_SOURCE {
		if sourceAlias == alias {
			return sourceType, nil
		}
	}

	return 0, errors.New("invalid source alias")
}
