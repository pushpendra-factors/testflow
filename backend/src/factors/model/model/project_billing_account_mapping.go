package model

import "time"

type ProjectBillingAccountMapping struct {
	ProjectID        int64  `gorm:"primary_key:true" json:"project_id"`
	BillingAccountID string `gorm:"primary_key:true" json:"billing_account_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
