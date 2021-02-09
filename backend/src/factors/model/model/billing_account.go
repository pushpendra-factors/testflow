package model

import "time"

type BillingAccount struct {
	ID        uint64 `gorm:"primary_key:true;" json:"id"`
	PlanID    uint64 `gorm:"not null;" json:"plan_id"`
	AgentUUID string `gorm:"not null;" json:"agent_uuid"`

	OrganizationName string `json:"organization_name"`
	BillingAddress   string `json:"billing_address"`
	Pincode          string `json:"pincode"`
	PhoneNo          string `json:"phone_no"` // Optional

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
