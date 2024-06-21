package model

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type PlanDetails struct {
	ID                 int64           `gorm:"column:id; primary_key:true; autoIncrement:true" json:"id"`
	BillingPlanID      string          `json:"billing_plan_id"`
	BillingPlanPriceID string          `json:"billing_plan_price_id"`
	Name               string          `gorm:"column:name; primary_key:true" json:"name"`
	MtuLimit           int64           `json:"mtu_limit"`
	FeatureList        *postgres.Jsonb `gorm:"column:feature_list" json:"feature_list"` //[]FeatureDetails
}

type FeatureDetails struct {
	Limit            int64  `json:"limit"`
	Granularity      string `json:"granularity"`
	Expiry           int64  `json:"expiry"`
	IsEnabledFeature bool   `json:"is_enabled_feature"` //is feature an addition to the plan or deletion from it
}

type FeatureDetailsTemp struct { // only used for overwrites until migration to transform existing data is run
	Name             string `json:"name"`
	Limit            int64  `json:"limit"`
	Granularity      string `json:"granularity"`
	Expiry           int64  `json:"expiry"`
	IsEnabledFeature bool   `json:"is_enabled_feature"` //is feature an addition to the plan or deletion from it
}

type FeatureList map[string]FeatureDetails

type OverWrite map[string]FeatureDetails

type OverWriteOld []FeatureDetailsTemp

type DisplayPlanDetails struct {
	ProjectID     string        `json:"project_id"`
	Plan          PlanDetails   `json:"plan"`
	DisplayName   string        `json:"display_name"`
	AddOns        OverWriteOld  `json:"add_ons"`
	LastRenewedOn time.Time     `json:"last_renewed_on"`
	SixSignalInfo SixSignalInfo `json:"six_signal_info"`
}
type SixSignalInfo struct {
	IsEnabled bool  `json:"is_enabled"`
	Usage     int64 `json:"usage"`
	Limit     int64 `json:"limit"`
}

const (
	PLAN_FREE         = "FREE"
	PLAN_STARTUP      = "STARTUP"
	PLAN_BASIC        = "BASIC"
	PLAN_PROFESSIONAL = "PROFESSIONAL"
	PLAN_CUSTOM       = "CUSTOM"
	PLAN_GROWTH       = "GROWTH"
)
const (
	PLAN_ID_FREE         = 1
	PLAN_ID_BASIC        = 2
	PLAN_ID_GROWTH       = 3
	PLAN_ID_PROFESSIONAL = 4
	PLAN_ID_CUSTOM       = 5
)

var planToFeatureMap map[string][]string

func GetPlanIDFromPlanName(planName string) (int, error) {
	switch planName {
	case PLAN_FREE:
		return PLAN_ID_FREE, nil
	case PLAN_BASIC:
		return PLAN_ID_BASIC, nil
	case PLAN_GROWTH:
		return PLAN_ID_GROWTH, nil
	case PLAN_PROFESSIONAL:
		return PLAN_ID_PROFESSIONAL, nil
	case PLAN_CUSTOM:
		return PLAN_ID_CUSTOM, nil
	default:
		return 0, errors.New("Invalid Plan Name")
	}
}

func TransformFeatureListMaptoFeatureListArray(featureListMap FeatureList) []FeatureDetailsTemp {
	var res []FeatureDetailsTemp

	for featureName := range featureListMap {
		feature := featureListMap[featureName]
		res = append(res, FeatureDetailsTemp{
			Name:             featureName,
			Limit:            feature.Limit,
			Granularity:      feature.Granularity,
			Expiry:           feature.Expiry,
			IsEnabledFeature: feature.IsEnabledFeature,
		})
	}

	return res
}

// func initPlanToFeatureMapping() {
// 	planToFeatureMap = make(map[string][]string)
// 	planToFeatureMap[PLAN_FREE] = GetFeaturesForPlanType(PLAN_FREE)
// 	planToFeatureMap[PLAN_STARTUP] = GetFeaturesForPlanType(PLAN_STARTUP)
// 	planToFeatureMap[PLAN_BASIC] = GetFeaturesForPlanType(PLAN_BASIC)
// 	planToFeatureMap[PLAN_PROFESSIONAL] = GetFeaturesForPlanType(PLAN_PROFESSIONAL)
// 	planToFeatureMap[PLAN_CUSTOM] = GetFeaturesForPlanType(PLAN_CUSTOM)
// }
// func GetFeaturesForPlanType(PlanType string) []string {
// 	features := []string{}
// 	switch PlanType {
// 	case PLAN_FREE:
// 		features = []string{mid.FEATURE_DASHBOARD,
// 			mid.FEATURE_PROFILES,
// 			mid.FEATURE_ACCOUNT_PROFILES,
// 			mid.FEATURE_PEOPLE_PROFILES,
// 			mid.FEATURE_SIX_SIGNAL,
// 			mid.FEATURE_SIX_SIGNAL_REPORT,
// 			mid.FEATURE_KPI_ALERTS,
// 			mid.FEATURE_EVENT_BASED_ALERTS,
// 			mid.FEATURE_SlACK,
// 			mid.FEATURE_TEAMS,
// 			mid.FEATURE_FACTORS_DEANONYMISATION,
// 		}
// 	case PLAN_STARTUP:
// 		features = []string{mid.FEATURE_DASHBOARD,
// 			mid.FEATURE_PROFILES,
// 			mid.FEATURE_ACCOUNT_PROFILES,
// 			mid.FEATURE_PEOPLE_PROFILES,
// 			mid.FEATURE_SIX_SIGNAL,
// 			mid.FEATURE_SIX_SIGNAL_REPORT,
// 			mid.FEATURE_KPI_ALERTS,
// 			mid.FEATURE_EVENT_BASED_ALERTS,
// 		}
// 	case PLAN_BASIC:
// 		features = []string{}
// 	case PLAN_PROFESSIONAL:
// 		features = []string{}
// 	case PLAN_CUSTOM:
// 		features = []string{}
// 	}
// 	return features
// }
