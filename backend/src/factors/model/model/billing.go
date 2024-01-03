package model

import (
	"github.com/chargebee/chargebee-go/v3/models/invoice"
	"github.com/chargebee/chargebee-go/v3/models/subscription"
	"time"
)

type PlansAndAddOnsPrices []SubscriptionProductPrice
type DifferentialPrices []DifferentialPrice

const (
	ADD_ON_ADDITIONAL_500_ACCOUNTS_MONTHLY          = "Additional-500-Accounts-USD-Monthly"
	ADD_ON_ADDITIONAL_500_ACCOUNTS_YEARLY           = "Additional-500-Accounts-USD-Yearly"
	ADD_ON_ADDITIONAL_500_ACCOUNTS_MONTHLY_ACCOUNTS = 500
	ADD_ON_ADDITIONAL_500_ACCOUNTS_YEARLY_ACCOUNTS  = 500
)

func GetNumberOfAccountsForAddOnID(addOnID string) int {
	switch addOnID {
	case ADD_ON_ADDITIONAL_500_ACCOUNTS_MONTHLY:
		return ADD_ON_ADDITIONAL_500_ACCOUNTS_MONTHLY_ACCOUNTS
	case ADD_ON_ADDITIONAL_500_ACCOUNTS_YEARLY:
		return ADD_ON_ADDITIONAL_500_ACCOUNTS_YEARLY_ACCOUNTS
	default:
		return 0
	}
}

type SubscriptionProductPrice struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	ExternalName string `json:"external_name"`
	ID           string `json:"id"`
	Price        int64  `json:"price"`
	PeriodUnit   string `json:"period_unit"`
}

type DifferentialPrice struct {
	ID           string `json:"id"`
	ItemPriceID  string `json:"item_price_id"`
	ParentItemID string `json:"parent_item_id"`
	Price        int64  `json:"price"`
}

type Subscription struct {
	Status              string               `json:"status"`
	RenewsOn            time.Time            `json:"renews_on"`
	BillingPeriodUnit   string               `json:"period_unit"`
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

type Invoice struct {
	ID          string    `json:"id"`
	BillingDate time.Time `json:"billing_date"`
	Amount      int64     `json:"amount`
	AmountPaid  int64     `json:"amount_paid"`
	AmountDue   int64     `json:"amount_due`
	Items       []string  `json:"items"`
}

type DownloadInvoice struct {
	Url       string    `json:"url"`
	ValidTill time.Time `json:"valid_till"`
}

// webhook structs
type Content struct {
	Subscription subscription.Subscription `json:"subscription"`
	Invoice      invoice.Invoice           `json:"invoice"`
}
