package model

import "time"

type BillingAccount struct {
	ID        string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	PlanID    uint64 `gorm:"not null;" json:"plan_id"`
	AgentUUID string `gorm:"not null;" json:"agent_uuid"`

	OrganizationName string `json:"organization_name"`
	BillingAddress   string `json:"billing_address"`
	Pincode          string `json:"pincode"`
	PhoneNo          string `json:"phone_no"` // Optional

	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	BillingLastSyncedAt time.Time `json:"billing_last_synced_at`
}
