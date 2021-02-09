package model

import "time"

type ProjectBillingAccountMapping struct {
	ProjectID        uint64 `gorm:"primary_key:true"`
	BillingAccountID uint64 `gorm:"primary_key:true"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
