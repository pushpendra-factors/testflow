package main

import (
	C "factors/config"
	M "factors/model/model"
	store "factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	FullAccessExpiry                     = math.MaxInt64 // expiry of features with full access
	FullAccessLimit                      = math.MaxInt64 // limit of features with full access
	NonFeatureExpiry                     = 0             // expiry of configurations or integrations
	NonFeatureLimit                      = 0             // limit of configurations or integrations
	FreeLimitOnMonthlyAccountsIdentified = 100           // free plan monthly limit on accounts identified
	FreeLimitOnMTUs                      = 5000          // free plan limit on MTUs
	FreePlanExpiry                       = math.MaxInt64 // Free plan never expires
	Create                               = "create"
	Update                               = "update"
)

// Add new features for the custom plan other than the features already in the free plan
var CustomFeatureList = []string{
	M.FEATURE_EVENTS,
	M.FEATURE_FUNNELS,
	M.FEATURE_KPIS,
	M.FEATURE_ATTRIBUTION,
	M.FEATURE_PROFILES,
	M.FEATURE_TEMPLATES,

	M.FEATURE_HUBSPOT,
	M.FEATURE_SALESFORCE,
	M.FEATURE_LEADSQUARED,
	M.FEATURE_GOOGLE_ADS,
	M.FEATURE_FACEBOOK,
	M.FEATURE_LINKEDIN,
	M.FEATURE_GOOGLE_ORGANIC,
	M.FEATURE_BING_ADS,
	M.FEATURE_MARKETO,
	M.FEATURE_DRIFT,
	M.FEATURE_CLEARBIT,
	M.FEATURE_SIX_SIGNAL,
	M.FEATURE_DASHBOARD,
	M.FEATURE_OFFLINE_TOUCHPOINTS,
	M.FEATURE_SAVED_QUERIES,
	M.FEATURE_EXPLAIN, // explain is a keyword in memsql.
	M.FEATURE_FILTERS,
	M.FEATURE_SHAREABLE_URL,
	M.FEATURE_CUSTOM_METRICS,
	M.FEATURE_SMART_EVENTS,
	M.FEATURE_SMART_PROPERTIES,
	M.FEATURE_CONTENT_GROUPS,
	M.FEATURE_DISPLAY_NAMES,
	M.FEATURE_WEEKLY_INSIGHTS,
	M.FEATURE_KPI_ALERTS,
	M.FEATURE_EVENT_BASED_ALERTS,
	M.FEATURE_REPORT_SHARING,
	M.FEATURE_REPORT_SHARING,
	M.FEATURE_SLACK,
	M.FEATURE_SEGMENT,
	M.FEATURE_PATH_ANALYSIS,
	M.FEATURE_ARCHIVE_EVENTS,
	M.FEATURE_BIG_QUERY_UPLOAD,
	M.FEATURE_IMPORT_ADS,
	M.FEATURE_LEADGEN,
	M.FEATURE_TEAMS,
	M.FEATURE_SIX_SIGNAL_REPORT,
	M.FEATURE_ACCOUNT_SCORING,
	M.FEATURE_ENGAGEMENT,
	M.FEATURE_FACTORS_DEANONYMISATION,
	M.FEATURE_WEBHOOK,
	M.FEATURE_ACCOUNT_PROFILES,
	M.FEATURE_PEOPLE_PROFILES,

	// INTEGRATION
	M.INT_SHOPFIY,
	M.INT_ADWORDS,
	M.INT_GOOGLE_ORGANIC,
	M.INT_FACEBOOK,
	M.INT_LINKEDIN,
	M.INT_SALESFORCE,
	M.INT_HUBSPOT,
	M.INT_DELETE,
	M.INT_SLACK,
	M.INT_TEAMS,
	M.INT_SEGMENT,
	M.INT_RUDDERSTACK,
	M.INT_MARKETO,
	M.INT_DRIFT,
	M.INT_BING_ADS,
	M.INT_CLEARBIT,
	M.INT_LEADSQUARED,
	M.INT_SIX_SIGNAL,
	M.INT_FACTORS_DEANONYMISATION,
	M.INT_G2,

	// DATA SERVICE
	M.DS_ADWORDS,
	M.DS_GOOGLE_ORGANIC,
	M.DS_HUBSPOT,
	M.DS_FACEBOOK,
	M.DS_LINKEDIN,
	M.DS_METRICS,

	// CONFIGURATIONS
	M.CONF_ATTRUBUTION_SETTINGS,
	M.CONF_CUSTOM_EVENTS,
	M.CONF_CUSTOM_PROPERTIES,
	M.CONF_CONTENT_GROUPS,
	M.CONF_TOUCHPOINTS,
	M.CONF_CUSTOM_KPIS,
	M.CONF_ALERTS,
}

var PlanFreeLimitMap = map[string]int64{
	M.FEATURE_DASHBOARD:               FullAccessLimit,
	M.FEATURE_PROFILES:                FullAccessLimit,
	M.FEATURE_ACCOUNT_PROFILES:        FreeLimitOnMonthlyAccountsIdentified,
	M.FEATURE_PEOPLE_PROFILES:         FreeLimitOnMonthlyAccountsIdentified,
	M.FEATURE_SIX_SIGNAL:              NonFeatureLimit,
	M.FEATURE_SIX_SIGNAL_REPORT:       NonFeatureLimit,
	M.FEATURE_KPI_ALERTS:              NonFeatureLimit,
	M.FEATURE_EVENT_BASED_ALERTS:      NonFeatureLimit,
	M.FEATURE_SLACK:                   NonFeatureLimit,
	M.FEATURE_TEAMS:                   NonFeatureLimit,
	M.FEATURE_FACTORS_DEANONYMISATION: NonFeatureLimit,
}

func GetPlanDetails(planType string) M.PlanDetails {
	var featureList []M.FeatureDetails
	if planType == M.PLAN_CUSTOM {
		return M.PlanDetails{}
	}
	featureList = GetFeaturesDetailsList(planType)

	for idx, feature := range featureList {
		if feature.Name == M.FEATURE_FACTORS_DEANONYMISATION && planType == M.PLAN_FREE {
			featureList[idx].Limit = FreeLimitOnMonthlyAccountsIdentified
		}
	}

	features, err := U.EncodeStructTypeToPostgresJsonb(featureList)
	if err != nil {
		log.Error("Json encoding failure")
		return M.PlanDetails{}
	}

	planDetails := M.PlanDetails{
		Name:        planType,
		FeatureList: features,
	}
	return planDetails
}

func GetFeaturesDetailsList(planType string) []M.FeatureDetails {
	featureNameList := GetFeaturesForPlanType(planType)
	featureDetailsList := make([]M.FeatureDetails, 0)

	for _, feature := range featureNameList {
		featureDetail := M.FeatureDetails{
			Name:             feature,
			Limit:            PlanFreeLimitMap[feature],
			IsEnabledFeature: true,
			IsConnected:      true,
			Expiry:           FreePlanExpiry,
		}
		featureDetailsList = append(featureDetailsList, featureDetail)
	}
	return featureDetailsList
}

func GetFeaturesForPlanType(PlanType string) []string {
	var features []string
	switch PlanType {

	case M.PLAN_FREE:
		features = GetFreePlanFeatures()

	case M.PLAN_STARTUP:
		features = GetStartupPlanFeatures()

	case M.PLAN_BASIC:
		features = GetBasicPlanFeatures()

	case M.PLAN_PROFESSIONAL:
		features = GetProfessionalPlanFeatures()

	case M.PLAN_CUSTOM:
		features = GetCustomPlanFeatures()

	}
	return features
}

func GetFreePlanFeatures() []string {
	return []string{
		// Dashboards
		M.FEATURE_DASHBOARD,

		// Analyse
		M.FEATURE_EVENTS,
		M.FEATURE_FUNNELS,
		M.FEATURE_KPIS,
		M.FEATURE_TEMPLATES,
		M.FEATURE_PROFILES,

		// Account Profile
		M.FEATURE_ACCOUNT_PROFILES,

		// People Profiles
		M.FEATURE_PEOPLE_PROFILES,

		// Alerts
		M.FEATURE_KPI_ALERTS,
		M.FEATURE_EVENT_BASED_ALERTS,
		M.CONF_ALERTS,
		M.FEATURE_REPORT_SHARING,

		// Slack
		M.FEATURE_SLACK,
		M.INT_SLACK,

		// Teams
		M.FEATURE_TEAMS,
		M.INT_TEAMS,

		// Factors deanonymisation
		M.INT_FACTORS_DEANONYMISATION,
		M.FEATURE_FACTORS_DEANONYMISATION,

		// Website Visitor Identification
		M.FEATURE_SIX_SIGNAL,
		M.FEATURE_SIX_SIGNAL_REPORT,

		// Adwords
		M.INT_ADWORDS,
		M.INT_GOOGLE_ORGANIC,
		M.INT_FACEBOOK,
		M.INT_LINKEDIN,

		// Data service for ad words
		M.DS_ADWORDS,
		M.DS_GOOGLE_ORGANIC,
		M.DS_FACEBOOK,
		M.DS_LINKEDIN,
	}
}

func GetStartupPlanFeatures() []string {
	return []string{
		// Dashboards
		M.FEATURE_DASHBOARD,

		// Analyse
		M.FEATURE_EVENTS,
		M.FEATURE_FUNNELS,
		M.FEATURE_KPIS,
		M.FEATURE_TEMPLATES,
		M.FEATURE_PROFILES,

		// Account Profile
		M.FEATURE_ACCOUNT_PROFILES,

		// People Profiles
		M.FEATURE_PEOPLE_PROFILES,

		// Custom Events
		M.CONF_CUSTOM_EVENTS,

		// Custom Properties
		M.CONF_CUSTOM_PROPERTIES,

		// Content Groups
		M.CONF_CONTENT_GROUPS,

		// Touch points
		M.CONF_TOUCHPOINTS,

		// Custom KPIs
		M.CONF_CUSTOM_KPIS,

		// Top Event and properties

		// Alerts
		M.FEATURE_KPI_ALERTS,
		M.FEATURE_EVENT_BASED_ALERTS,
		M.CONF_ALERTS,

		// Slack
		M.FEATURE_SLACK,
		M.INT_SLACK,

		// Teams
		M.FEATURE_TEAMS,
		M.INT_TEAMS,

		// Google Ads
		M.FEATURE_GOOGLE_ADS,
		M.DS_ADWORDS,

		// Facebook
		M.FEATURE_FACEBOOK,
		M.INT_FACEBOOK,
		M.DS_FACEBOOK,

		// LinkedIn
		M.FEATURE_LINKEDIN,
		M.INT_LINKEDIN,
		M.DS_LINKEDIN,

		// Drift
		M.FEATURE_DRIFT,
		M.INT_DRIFT,

		// Google search Console
		M.FEATURE_GOOGLE_ORGANIC,
		M.INT_GOOGLE_ORGANIC,
		M.DS_GOOGLE_ORGANIC,

		// Bing Ads
		M.FEATURE_BING_ADS,
		M.INT_BING_ADS,

		// Clearbit Reveal
		M.FEATURE_CLEARBIT,
		M.INT_CLEARBIT,

		// Leadsquared
		M.FEATURE_LEADSQUARED,
		M.INT_LEADSQUARED,

		// Factors deanonymisation
		M.FEATURE_FACTORS_DEANONYMISATION,

		// Website Visitor Identification
		M.FEATURE_SIX_SIGNAL,
		M.FEATURE_SIX_SIGNAL_REPORT,

		// Six signal
		M.INT_SIX_SIGNAL,
	}
}

func GetBasicPlanFeatures() []string {
	return []string{
		M.FEATURE_DASHBOARD,

		// Analyse
		M.FEATURE_EVENTS,
		M.FEATURE_FUNNELS,
		M.FEATURE_KPIS,
		M.FEATURE_TEMPLATES,
		M.FEATURE_PROFILES,

		// Account Profile
		M.FEATURE_ACCOUNT_PROFILES,

		// People Profiles
		M.FEATURE_PEOPLE_PROFILES,

		// Custom Events
		M.CONF_CUSTOM_EVENTS,

		// Custom Properties
		M.CONF_CUSTOM_PROPERTIES,

		// Content Groups
		M.CONF_CONTENT_GROUPS,

		// Touch points
		M.CONF_TOUCHPOINTS,

		// Custom KPIs
		M.CONF_CUSTOM_KPIS,

		// Top Event and properties

		// Alerts
		M.FEATURE_KPI_ALERTS,
		M.FEATURE_EVENT_BASED_ALERTS,
		M.CONF_ALERTS,

		// Segment
		M.FEATURE_SEGMENT,
		M.INT_SEGMENT,

		// Rudderstack
		M.INT_RUDDERSTACK,

		// Marketo
		M.INT_MARKETO,
		M.FEATURE_MARKETO,

		// Slack
		M.FEATURE_SLACK,
		M.INT_SLACK,

		// Teams
		M.FEATURE_TEAMS,
		M.INT_TEAMS,

		// Hubspot
		M.FEATURE_HUBSPOT,
		M.INT_HUBSPOT,
		M.DS_HUBSPOT,

		// Salesforce
		M.INT_SALESFORCE,
		M.FEATURE_SALESFORCE,

		// Google Ads
		M.FEATURE_GOOGLE_ADS,
		M.DS_ADWORDS,

		// Facebook
		M.FEATURE_FACEBOOK,
		M.INT_FACEBOOK,
		M.DS_FACEBOOK,

		// LinkedIn
		M.FEATURE_LINKEDIN,
		M.INT_LINKEDIN,
		M.DS_LINKEDIN,

		// Drift
		M.FEATURE_DRIFT,
		M.INT_DRIFT,

		// Google search Console
		M.FEATURE_GOOGLE_ORGANIC,
		M.INT_GOOGLE_ORGANIC,
		M.DS_GOOGLE_ORGANIC,

		// Bing Ads
		M.FEATURE_BING_ADS,
		M.INT_BING_ADS,

		// Clearbit Reveal
		M.FEATURE_CLEARBIT,
		M.INT_CLEARBIT,

		// Leadsquared
		M.FEATURE_LEADSQUARED,
		M.INT_LEADSQUARED,

		// Factors deanonymisation
		M.FEATURE_FACTORS_DEANONYMISATION,

		// Website Visitor Identification
		M.FEATURE_SIX_SIGNAL,
		M.FEATURE_SIX_SIGNAL_REPORT,

		// Six signal
		M.INT_SIX_SIGNAL,

		// G2
		M.INT_G2,
	}
}

func GetProfessionalPlanFeatures() []string {
	return []string{}
}

func GetCustomPlanFeatures() []string {
	// freeFeature := GetFreePlanFeatures()

	feature := make([]string, 0)
	// feature = append(feature, freeFeature...)

	customFeature := CustomFeatureList
	feature = append(feature, customFeature...)

	return feature
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	action := flag.String("action", "", "Enter 'create' for creating new plan else enter 'update' to update new plan")
	planType := flag.String("plan_type", "", "Enter comma separated values: 'free' for free plan; 'basic' for basic plan; 'pro' for professional plan; 'custom' for custom plan; 'startup' for startup plan")

	//update action
	planID := flag.Int64("id", -1, "Enter plan id to be updated")
	updateAdd := flag.String("add", "", "Add feature to the feature list of an existing plan")
	updateRemove := flag.String("remove", "", "Remove feature from the feature list of an existing plan")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "plan_population"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	PlanTypes := C.GetTokensFromStringListAsString(*planType)

	if *action == Create {
		for _, argPlanType := range PlanTypes {
			planDetails := GetPlanDetails(strings.ToUpper(argPlanType))
			_, err = store.GetStore().PopulatePlanDetailsTable(planDetails)
			if err != nil {
				log.WithError(err).Error("failed to populate plan details table")
			}
		}
	}
	if *action == Update {
		for _ = range PlanTypes {
			add := C.GetTokensFromStringListAsString(*updateAdd)
			remove := C.GetTokensFromStringListAsString(*updateRemove)
			_, err = store.GetStore().UpdatePlanDetailsTable(*planID, add, true)
			if err != nil {
				log.WithError(err).Error("failed to populate plan details table")
			}
			_, err = store.GetStore().UpdatePlanDetailsTable(*planID, remove, false)
			if err != nil {
				log.WithError(err).Error("failed to populate plan details table")
			}
		}
	}
}
