package model

import "time"

type PlansAndAddOnsPrices []SubscriptionProductPrice

type SubscriptionProductPrice struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	ExternalName string `json:"external_name"`
	ID           string `json:"id"`
	Price        int64  `json:"price"`
	Period       int    `json:"period"`
}

type Subscription struct {
	Status              string               `json:"status"`
	RenewsOn            time.Time            `json:"renews_on"`
	BillingPeriod       int                  `json:"period"`
	BillingPeriodUnit   string               `json:"billing_period_unit"`
	SubscriptionDetails []SubscriptionDetail `json:"subscription_details"`
}

type SubscriptionDetail struct {
	Type         string `json:"type"`
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	ExternalName string `json:"external_name"`
}

type UpdateSubscriptionParams struct {
	UpdatedPlanID string         `json:"updated_plan_id"`
	Addons        []AddOnsUpdate `json:"add_ons"`
}

type AddOnsUpdate struct {
	AddOnID  string `json:"addon_id"`
	Quantity int32  `json:"quantity"`
}
