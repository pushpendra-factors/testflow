package main

import (
	C "factors/config"
	"factors/model/model"
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
	FullAccessExpiry                          = math.MaxInt64 // expiry of features with full access
	FullAccessLimit                           = math.MaxInt64 // limit of features with full access
	NonFeatureExpiry                          = 0             // expiry of configurations or integrations
	NonFeatureLimit                           = 0             // limit of configurations or integrations
	FreeLimitOnMonthlyAccountsIdentified      = 100           // free plan monthly limit on accounts identified
	BasicPlanMonthlyAccountsIdentified        = 500
	GrowthPlanMonthlyAccountsIdentified       = 5000
	ProfessionalPlanMonthlyAccountsIdentified = 10000
	FreeLimitOnMTUs                           = 5000          // free plan limit on MTUs
	FreePlanExpiry                            = math.MaxInt64 // Free plan never expires
	Create                                    = "create"
	Update                                    = "update"
	Disable                                   = "disable"
	Enable                                    = "enable"
	UpdateCustom                              = "update_custom"
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
	M.FEATURE_FACTORS_DEANONYMISATION,
	M.FEATURE_WEBHOOK,
	M.FEATURE_ACCOUNT_PROFILES,
	M.FEATURE_PEOPLE_PROFILES,
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
		if feature.Name == M.FEATURE_FACTORS_DEANONYMISATION {
			featureList[idx].Limit = FreeLimitOnMonthlyAccountsIdentified
			if planType == model.PLAN_BASIC {
				featureList[idx].Limit = BasicPlanMonthlyAccountsIdentified
			} else if planType == model.PLAN_GROWTH {
				featureList[idx].Limit = GrowthPlanMonthlyAccountsIdentified
			} else if planType == model.PLAN_PROFESSIONAL {
				featureList[idx].Limit = ProfessionalPlanMonthlyAccountsIdentified
			}
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

	case M.PLAN_GROWTH:
		features = GetGrowthPlanFeatures()

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
		M.FEATURE_REPORT_SHARING,

		// Slack
		M.FEATURE_SLACK,

		// Teams
		M.FEATURE_TEAMS,

		// Factors deanonymisation
		M.FEATURE_FACTORS_DEANONYMISATION,

		// Website Visitor Identification
		M.FEATURE_SIX_SIGNAL,
		M.FEATURE_SIX_SIGNAL_REPORT,

		// Adwords

		// Data service for ad words
	}
}

func GetGrowthPlanFeatures() []string {
	return []string{
		M.FEATURE_EVENTS,
		M.FEATURE_FUNNELS,
		M.FEATURE_KPIS,
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
		M.FEATURE_FILTERS,
		M.FEATURE_SHAREABLE_URL,
		M.FEATURE_CUSTOM_METRICS,
		M.FEATURE_SMART_PROPERTIES,
		M.FEATURE_CONTENT_GROUPS,
		M.FEATURE_DISPLAY_NAMES,
		M.FEATURE_WEEKLY_INSIGHTS,
		M.FEATURE_KPI_ALERTS,
		M.FEATURE_EVENT_BASED_ALERTS,
		M.FEATURE_REPORT_SHARING,
		M.FEATURE_SLACK,
		M.FEATURE_SEGMENT,
		M.FEATURE_ARCHIVE_EVENTS,
		M.FEATURE_BIG_QUERY_UPLOAD,
		M.FEATURE_IMPORT_ADS,
		M.FEATURE_LEADGEN,
		M.FEATURE_TEAMS,
		M.FEATURE_SIX_SIGNAL_REPORT,
		M.FEATURE_ACCOUNT_SCORING,
		M.FEATURE_FACTORS_DEANONYMISATION,
		M.FEATURE_WEBHOOK,
		M.FEATURE_ACCOUNT_PROFILES,
		M.FEATURE_PEOPLE_PROFILES,
		M.FEATURE_CLICKABLE_ELEMENTS,
		M.FEATURE_WEB_ANALYTICS_DASHBOARD,
		M.FEATURE_G2,
		M.FEATURE_RUDDERSTACK,
		M.CONF_CUSTOM_PROPERTIES,
		M.CONF_CUSTOM_EVENTS,
	}
}

func GetBasicPlanFeatures() []string {
	return []string{
		M.FEATURE_EVENTS,
		M.FEATURE_FUNNELS,
		M.FEATURE_KPIS,
		M.FEATURE_PROFILES,
		M.FEATURE_TEMPLATES,
		M.FEATURE_GOOGLE_ADS,
		M.FEATURE_FACEBOOK,
		M.FEATURE_LINKEDIN,
		M.FEATURE_GOOGLE_ORGANIC,
		M.FEATURE_BING_ADS,
		M.FEATURE_DRIFT,
		M.FEATURE_CLEARBIT,
		M.FEATURE_SIX_SIGNAL,
		M.FEATURE_DASHBOARD,
		M.FEATURE_OFFLINE_TOUCHPOINTS,
		M.FEATURE_SAVED_QUERIES,
		M.FEATURE_FILTERS,
		M.FEATURE_SHAREABLE_URL,
		M.FEATURE_CUSTOM_METRICS,
		M.FEATURE_SMART_PROPERTIES,
		M.FEATURE_CONTENT_GROUPS,
		M.FEATURE_DISPLAY_NAMES,
		M.FEATURE_WEEKLY_INSIGHTS,
		M.FEATURE_KPI_ALERTS,
		M.FEATURE_EVENT_BASED_ALERTS,
		M.FEATURE_REPORT_SHARING,
		M.FEATURE_SLACK,
		M.FEATURE_SEGMENT,
		M.FEATURE_ARCHIVE_EVENTS,
		M.FEATURE_BIG_QUERY_UPLOAD,
		M.FEATURE_IMPORT_ADS,
		M.FEATURE_LEADGEN,
		M.FEATURE_TEAMS,
		M.FEATURE_SIX_SIGNAL_REPORT,
		M.FEATURE_FACTORS_DEANONYMISATION,
		M.FEATURE_WEBHOOK,
		M.FEATURE_ACCOUNT_PROFILES,
		M.FEATURE_PEOPLE_PROFILES,
		M.FEATURE_CLICKABLE_ELEMENTS,
		M.FEATURE_WEB_ANALYTICS_DASHBOARD,
		M.FEATURE_RUDDERSTACK,
		M.CONF_CUSTOM_PROPERTIES,
		M.CONF_CUSTOM_EVENTS,
	}
}

func GetProfessionalPlanFeatures() []string {
	return []string{
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
		M.FEATURE_EXPLAIN,
		M.FEATURE_FILTERS,
		M.FEATURE_SHAREABLE_URL,
		M.FEATURE_CUSTOM_METRICS,
		M.FEATURE_SMART_PROPERTIES,
		M.FEATURE_CONTENT_GROUPS,
		M.FEATURE_DISPLAY_NAMES,
		M.FEATURE_WEEKLY_INSIGHTS,
		M.FEATURE_KPI_ALERTS,
		M.FEATURE_EVENT_BASED_ALERTS,
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
		M.FEATURE_FACTORS_DEANONYMISATION,
		M.FEATURE_WEBHOOK,
		M.FEATURE_ACCOUNT_PROFILES,
		M.FEATURE_PEOPLE_PROFILES,
		M.FEATURE_CLICKABLE_ELEMENTS,
		M.FEATURE_WEB_ANALYTICS_DASHBOARD,
		M.FEATURE_G2,
		M.FEATURE_RUDDERSTACK,
		M.CONF_CUSTOM_PROPERTIES,
		M.CONF_CUSTOM_EVENTS,
	}
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

	action := flag.String("action", "", "Enter 'create' for creating new plan else enter 'update' to update new plan enable or disable for updating feautre status")
	planType := flag.String("plan_type", "", "Enter comma separated values: 'free' for free plan; 'basic' for basic plan; 'pro' for professional plan; 'custom' for custom plan; 'professional' for startup plan")

	//update action
	planID := flag.Int64("id", -1, "Enter plan id to be updated")
	updateAdd := flag.String("add", "", "Add feature to the feature list of an existing plan")
	updateRemove := flag.String("remove", "", "Remove feature from the feature list of an existing plan")
	updateFeaturesStatus := flag.String("update_features_status", "", "")
	updateAccountLimits := flag.Bool("update_accounts_limit", false, "")
	updatedAccountLimit := flag.Int64("updated_account_limit", 100, "")
	ProjectIDs := flag.String("project_ids", "", "comma separeted list of project ids, * for all")
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
	if *action == Update { // free
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
	if *action == Disable || *action == Enable {
		// only allowing in Custom Plan
		features := C.GetTokensFromStringListAsString(*updateFeaturesStatus)
		if *updateFeaturesStatus == "*" {
			features = model.GetAllAvailableFeatures()
		}
		featureMap := make(map[string]bool)
		for _, feature := range features {
			featureMap[feature] = true
		}
		var projectIDS []int64
		var err error
		if *ProjectIDs == "*" {
			projectIDS, _, _, err = store.GetStore().GetAllProjectIdsUsingPlanId(model.PLAN_ID_CUSTOM)
			if err != nil {
				log.Error("Failed to get project ID with custom plan")
				return
			}
		} else {
			projectIDS = C.GetTokensFromStringListAsUint64(*ProjectIDs)
		}
		for _, id := range projectIDS {
			UpdateFeatureStatusForCustomPlan(id, featureMap, *action, *updateAccountLimits, *updatedAccountLimit)

		}
	}

	if *action == UpdateCustom {
		add := C.GetTokensFromStringListAsString(*updateAdd)
		// remove := C.GetTokensFromStringListAsString(*updateRemove)
		var projectIDS []int64
		var err error
		if *ProjectIDs == "*" {
			projectIDS, _, _, err = store.GetStore().GetAllProjectIdsUsingPlanId(model.PLAN_ID_CUSTOM)
			if err != nil {
				log.Error("Failed to get project ID with custom plan")
				return
			}
		} else {
			projectIDS = C.GetTokensFromStringListAsUint64(*ProjectIDs)
		}
		for _, id := range projectIDS {
			AddFeatureToCustomPlan(id, add)
		}

	}
}

func UpdateFeatureStatusForCustomPlan(ProjectID int64, featuresMap map[string]bool, action string, updateAcccountLimits bool, accountsLimit int64) { // only to update accounts identification limit
	ppmapping, err := store.GetStore().GetProjectPlanMappingforProject(ProjectID)
	if err != nil {
		log.Error("Failed to disable features for ID ", ProjectID)
		return
	}
	var addOns model.OverWrite
	if ppmapping.OverWrite != nil {
		err := U.DecodePostgresJsonbToStructType(ppmapping.OverWrite, &addOns)
		if err != nil {
			log.Error(err, " ID ", ProjectID)
			return
		}
	}
	for idx, feature := range addOns {
		if _, ok := featuresMap[feature.Name]; ok {
			if action == Enable {
				addOns[idx].IsEnabledFeature = true
				if updateAcccountLimits && feature.Name == model.FEATURE_FACTORS_DEANONYMISATION {
					addOns[idx].Limit = accountsLimit
				}
			} else if action == Disable {
				addOns[idx].IsEnabledFeature = false
			}
		}
	}
	_, err = store.GetStore().UpdateAddonsForProject(ProjectID, addOns)
	if err != nil {
		log.Error("Failed to update addons for project ", ProjectID)
	}
}

func AddFeatureToCustomPlan(ProjectID int64, features []string) {
	ppmapping, err := store.GetStore().GetProjectPlanMappingforProject(ProjectID)
	if err != nil {
		log.Error("Failed to add features for ID ", ProjectID)
		return
	}
	var addOns model.OverWrite
	if ppmapping.OverWrite != nil {
		err := U.DecodePostgresJsonbToStructType(ppmapping.OverWrite, &addOns)
		if err != nil {
			log.Error(err, " ID ", ProjectID)
			return
		}
	}

	featureMap := make(map[string]bool)
	for _, feature := range addOns {
		featureMap[feature.Name] = true
	}

	for _, feature := range features {
		if _, exists := featureMap[feature]; exists {
			log.Info("Feature already exists for project ", feature, ProjectID)
			continue
		}

		addOns = append(addOns, model.FeatureDetails{
			Name:             feature,
			IsEnabledFeature: false,
		})
	}
	_, err = store.GetStore().UpdateAddonsForProject(ProjectID, addOns)
	if err != nil {
		log.Error("Failed to update addons for project ", ProjectID)
	}
}
