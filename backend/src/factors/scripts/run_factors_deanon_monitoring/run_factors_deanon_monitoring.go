package main

import (
	"factors/company_enrichment/factors_deanon"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	thresholdValue := flag.Float64("threshold_value", 0.3, "Percent threshold value for alerts")

	flag.Parse()

	defaultAppName := "factors_deanon_monitoring_job"
	defaultHealthcheckPingID := C.HealthcheckFactorsDeanonAlertPingID

	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)

	defer C.PingHealthcheckForPanic(defaultAppName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: defaultAppName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     defaultAppName,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize db.")
	}
	db := C.GetServices().Db
	defer db.Close()

	projectIds, errCode, errMsg, err := store.GetStore().GetAllProjectIdsUsingPaidPlan()
	if errCode != http.StatusFound {
		log.WithError(err).Error(errMsg)
		C.PingHealthcheckForFailure(healthcheckPingID, "Failed to fetch project ids for monitoring factors deanonymisation.")
		return
	}

	alertMap := make(map[string]interface{})
	for _, projectId := range projectIds {

		logCtx := log.WithField("project_id", projectId)
		msgMap := make(map[string]interface{})
		projectSettings, errCode := store.GetStore().GetProjectSetting(projectId)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to fetch project details.")
			continue
		}

		isEligible, err := IsProjectEligibleForFactorsDeanonAlerts(projectSettings)
		if err != nil {
			logCtx.Error("Failed to check eligibilty.")
			continue
		}
		if !isEligible {
			continue
		}

		projectCreatedAt := projectSettings.CreatedAt.Unix()
		currentTime := time.Now().Unix()
		diffTime := currentTime - projectCreatedAt

		if diffTime > 8*24*60*60 && diffTime <= 15*24*60*60 {
			totalCountDiff, successfulCountDiff, err := CheckFactorsDeanonymisationAlertForRecentProjects(projectId)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get API count difference.")
				continue
			}
			if math.Abs(totalCountDiff) >= *thresholdValue && math.Abs(successfulCountDiff) >= *thresholdValue {
				msgMap["total_count_diff"] = totalCountDiff
				msgMap["success_count_diff"] = successfulCountDiff
			}

		} else if diffTime > 15*24*60*60 {
			totalCountDiff, successfulCountDiff, err := CheckFactorsDeanonymisationAlertForOlderProjects(projectId)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get API count difference.")
				continue
			}

			if math.Abs(totalCountDiff) >= *thresholdValue && math.Abs(successfulCountDiff) >= *thresholdValue {
				msgMap["total_count_diff"] = totalCountDiff
				msgMap["success_count_diff"] = successfulCountDiff
			}
		}

		projectIdString := fmt.Sprintf("%v", projectId)
		alertMap[projectIdString] = msgMap

	}

	C.PingHealthcheckForSuccess(healthcheckPingID, alertMap)

}

// IsProjectEligibleForFactorsDeanonAlerts checks if the project passes the creteria for factors deanon monitoring and alerts.
func IsProjectEligibleForFactorsDeanonAlerts(projectSettings *model.ProjectSetting) (bool, error) {

	projectId := projectSettings.ProjectId

	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_FACTORS_DEANONYMISATION, false)
	if err != nil {
		return false, err
	}
	if !featureFlag {
		return false, nil
	}

	isDeanonQuotaAvailable, err := factors_deanon.CheckingFactorsDeanonQuotaLimit(projectId)
	if err != nil {
		return false, err
	}
	if !isDeanonQuotaAvailable {
		return false, nil
	}

	intFactorsDeanon := projectSettings.IntFactorsSixSignalKey

	// to check if project is churned.
	intRudderstack := projectSettings.IntRudderstack
	intSegment := projectSettings.IntSegment
	intSdk := projectSettings.AutoTrack

	isProjectRetain := *intRudderstack || *intSdk || *intSegment

	eligible := featureFlag && isDeanonQuotaAvailable && *intFactorsDeanon && isProjectRetain

	return eligible, nil
}

// CheckFactorsDeanonymisationAlertForRecentProjects checks alert data for projects that are 8 to 15 days older.
func CheckFactorsDeanonymisationAlertForRecentProjects(projectId int64) (float64, float64, error) {

	yesterdayDate := time.Now().AddDate(0, 0, -1).Format(util.DATETIME_FORMAT_YYYYMMDD)
	eightDaysAgoDate := time.Now().AddDate(0, 0, -8).Format(util.DATETIME_FORMAT_YYYYMMDD)

	yesterdayUint64, _ := strconv.ParseUint(yesterdayDate, 10, 64)
	eightDaysAgoUint64, _ := strconv.ParseUint(eightDaysAgoDate, 10, 64)

	// Fetching total api count for n-1 and n-8
	yesterdayTotalApiCount, err := model.GetFactorsDeanonAPITotalHitCountResult(projectId, yesterdayUint64)
	if err != nil {
		return 0, 0, err
	}
	eightDaysAgoTotalApiCount, err := model.GetFactorsDeanonAPITotalHitCountResult(projectId, eightDaysAgoUint64)
	if err != nil {
		return 0, 0, err
	}

	// Fetching successful(domain is present) api count for n-1 and n-8
	yesterdaySuccessfulApiCount, err := model.GetFactorsDeanonAPICountResult(projectId, yesterdayUint64)
	if err != nil {
		return 0, 0, err
	}
	eightDaysAgoSuccessfulApiCount, err := model.GetFactorsDeanonAPICountResult(projectId, eightDaysAgoUint64)
	if err != nil {
		return 0, 0, err
	}

	totalCountDiff := float64(yesterdayTotalApiCount-eightDaysAgoTotalApiCount) / float64(eightDaysAgoTotalApiCount)
	successfulCountDiff := float64(yesterdaySuccessfulApiCount-eightDaysAgoSuccessfulApiCount) / float64(eightDaysAgoSuccessfulApiCount)

	if eightDaysAgoTotalApiCount <= 0 {
		return 0, 0, nil
	}

	return totalCountDiff, successfulCountDiff, nil

}

// CheckFactorsDeanonymisationAlertForOlderProjects checks alert data for projects that are older than 15 days.
func CheckFactorsDeanonymisationAlertForOlderProjects(projectId int64) (float64, float64, error) {

	currentDate := time.Now()
	var last14DaysTotalApiCount []int
	var last14DaysSuccessfulApiCount []int
	for i := 1; i <= 14; i++ {
		// Subtract i days from the current date
		date, _ := strconv.ParseUint(currentDate.AddDate(0, 0, -i).Format(util.DATETIME_FORMAT_YYYYMMDD), 10, 64)

		totalApiCount, err := model.GetFactorsDeanonAPITotalHitCountResult(projectId, date)
		if err != nil {
			return 0, 0, err
		}
		successfulApiCount, err := model.GetFactorsDeanonAPICountResult(projectId, date)
		if err != nil {
			return 0, 0, err
		}

		last14DaysTotalApiCount = append(last14DaysTotalApiCount, totalApiCount)
		last14DaysSuccessfulApiCount = append(last14DaysSuccessfulApiCount, successfulApiCount)
	}

	totalCountThisWeekSum := util.Sum(last14DaysTotalApiCount[:7])
	totalCountPastWeekSum := util.Sum(last14DaysTotalApiCount[7:])
	totalCountDiff := float64(totalCountThisWeekSum-totalCountPastWeekSum) / float64(totalCountPastWeekSum)

	successfulCountThisWeekSum := util.Sum(last14DaysSuccessfulApiCount[:7])
	successfulCountPastWeekSum := util.Sum(last14DaysSuccessfulApiCount[7:])
	successfulCountDiff := float64(successfulCountThisWeekSum-successfulCountPastWeekSum) / float64(successfulCountPastWeekSum)

	if totalCountPastWeekSum <= 0 {
		return 0, 0, nil
	}

	return totalCountDiff, successfulCountDiff, nil
}
