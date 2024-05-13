package main

import (
	"bufio"
	C "factors/config"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"factors/model/model"

	log "github.com/sirupsen/logrus"
)

// TODO later with from and to.
func main() {
	var env string
	defer U.NotifyOnPanic("Script#events_cube_aggregation_deploy", env)

	overrideHealthcheckPingID, jobForLargeTimerange, env, statusCode := parseFlagsAndInitConfig()
	if statusCode != http.StatusOK {
		log.Error("Failed in parsing flags")
		os.Exit(1)
	}

	defaultHealthcheckPingID := C.HealthcheckEventCubeAggregationPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, overrideHealthcheckPingID)
	jobReport := getDefaultJobReport()
	isError, errorString, key, tableCreated := false, "", "", true

	allowedProjectIdsMap, allProjects, errString, statusCode2 := getAllowedProjectIDsAndAllProjectsTimezone()
	if statusCode2 != http.StatusOK {
		C.PingHealthcheckForFailure(healthcheckPingID, errString)
		os.Exit(1)
	}

	storeSelected := store.GetStore()

	tableCreated, statusCode = storeSelected.TableOfWebsiteAggregationExists()
	if statusCode != http.StatusOK {
		C.PingHealthcheckForFailure(healthcheckPingID, "Failed to get response of TableOfWebsiteAggregationExists - events_cube_aggregation_deploy")
		os.Exit(1)
	}

	// TODO Compare unix of maxDataPresentTimestamp and beginning of day(maxDataPresentTimestamp) and show error.
	// Since DB doesnt store timezone, we always get result of max timestamp of present data in UTC.
	// We compute beginning of date in respective timezone from it.
	for index, project := range allProjects {
		key = fmt.Sprintf("%v:%v", index, project.ID)
		if _, exists := allowedProjectIdsMap[project.ID]; exists {
			minTimestampToFillData, maxTimestampToFillData, location, errString := GetMinMaxTimestampToFillAndLocation(project, tableCreated)
			if errString != "" {
				addToJobReport(jobReport, true, errString, key)
				jobReport["failures"][key] = errString
				continue
			}

			if !U.IsAllowedToRunInThisJob(minTimestampToFillData, maxTimestampToFillData, jobForLargeTimerange) {
				continue
			}

			dashboardCreationStatusCode := storeSelected.CreatePredefWebAggDashboardIfNotExists(project.ID)
			if dashboardCreationStatusCode != http.StatusCreated && dashboardCreationStatusCode != http.StatusFound {
				log.WithField("project", project.ID).Error("Failed during dashboard creation")
				addToJobReport(jobReport, true, "Failed during dashboard creation", key)
			}

			splitTimeRanges := U.GetSplitTimerangesIn7DayRanges(minTimestampToFillData, maxTimestampToFillData, project.TimeZone, location)

			numberOfDays := U.GetNumberOfDays(minTimestampToFillData, maxTimestampToFillData)
			jobStartTimestamp := time.Now().In(location)

			// Logic: compute minTimestamp to fill Data. set currentTo to minTimestamp + 7DaysEod.
			// If currentTo < todays beginning, fill for 7 days.
			// else currentTo will be now with buffer.

			for _, fromAndTo := range splitTimeRanges {
				if isError, errorString, tableCreated = executeDbtCommand(project, fromAndTo[0].Unix(), fromAndTo[1].Unix(), tableCreated); isError {
					break
				}
			}

			// Metrics calculation
			jobEndTimestamp := time.Now().In(location)
			addToJobReport(jobReport, isError, errorString, key)
			addMetricsToReportIfCrossesThreshold(jobReport, jobEndTimestamp.Sub(jobStartTimestamp).Seconds(), numberOfDays, project.ID)
		}
	}

	// printDebugLogFile()
	log.WithField("jobReport", jobReport).Warn("Complete job report")
	dupJobReport := getDuplicateJobReportWithoutSuccessKeys(jobReport)
	hasFailure := len(jobReport["failures"]) > 0 || len(jobReport["long_run_projects"]) > 0
	if hasFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, dupJobReport)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, dupJobReport)
	}
}

func parseFlagsAndInitConfig() (string, bool, string, int) {
	env := flag.String("env", "development", "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	jobForLargeTimerange := flag.Bool("job_for_large_timerange", false, "Job for large timerange.")

	appName := "events_cube_aggregation_deploy"

	flag.Parse()
	if *env != "development" && *env != "staging" && *env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
		return "", false, *env, http.StatusBadRequest
	}

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
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
		log.Fatal("Init Config failed.")
		return "", false, "", http.StatusBadRequest
	}

	return *overrideHealthcheckPingID, *jobForLargeTimerange, *env, http.StatusOK
}

func getAllowedProjectIDsAndAllProjectsTimezone() (map[int64]bool, []model.Project, string, int) {
	storeSelected := store.GetStore()
	allowedProjectIDs, err := storeSelected.GetAllProjectsWithFeatureEnabled(model.FEATURE_WEB_ANALYTICS_DASHBOARD, false)
	if err != nil {
		errString := "Failed in fetching projects with this feature flag enabled - events_cube_aggregation_deploy"
		log.WithField("err", err).Warn(errString)
		return make(map[int64]bool), make([]model.Project, 0), errString, http.StatusInternalServerError
	}
	log.WithField("allowedProjectIDs len", len(allowedProjectIDs)).WithField("allowedProjectIDs", allowedProjectIDs).Warn("testing feature flag")
	allowedProjectIdsMap := make(map[int64]bool)
	for _, projectID := range allowedProjectIDs {
		allowedProjectIdsMap[projectID] = true
	}

	allProjects, statusCode := storeSelected.GetIDAndTimezoneForAllProjects()
	if statusCode != http.StatusFound {
		errString := "Failed to get projects"
		log.Warn("Failed to get projects - events_cube_aggregation_deploy")
		return make(map[int64]bool), make([]model.Project, 0), errString, http.StatusInternalServerError
	}
	log.WithField("allProjects", allProjects).WithField("allowedProjectIdsMap", allowedProjectIdsMap).Warn("testing feature flag 1")

	return allowedProjectIdsMap, allProjects, "", http.StatusOK
}

func getDefaultJobReport() map[string]map[string]interface{} {
	jobReport := make(map[string]map[string]interface{})
	jobReport["success"] = make(map[string]interface{})
	jobReport["success"]["count"] = 0
	jobReport["failures"] = make(map[string]interface{})
	jobReport["long_run_projects"] = make(map[string]interface{})
	return jobReport
}

// minTimestamp to fill data should consider max data present timestamp because it might have partial data.
// Initial Run - minTimestamp is 32 days earlier from today.
func GetMinMaxTimestampToFillAndLocation(project model.Project, tableCreated bool) (time.Time, time.Time, *time.Location, string) {
	location, _ := time.LoadLocation(project.TimeZone)
	maxDataPresentTimestamp, statusCode := getMaxTimestampOfDataPresent(project, tableCreated)
	if statusCode != http.StatusOK {
		errString := "Failed in	getting maxDataPresentTimestamp"
		return maxDataPresentTimestamp, maxDataPresentTimestamp, location, errString
	}

	todayBeginningTimestamp, nowTimestampWithBuffer := U.GetTodayBeginningTimestampAndNowInBuffer(project.TimeZone, location)
	yesterdayBeginningTimestamp := todayBeginningTimestamp.AddDate(0, 0, -1)
	var minTimestampToFillData time.Time

	if maxDataPresentTimestamp.Unix() >= yesterdayBeginningTimestamp.Unix() {
		minTimestampToFillData = yesterdayBeginningTimestamp
	} else {
		minTimestampToFillData = maxDataPresentTimestamp
	}

	return minTimestampToFillData, nowTimestampWithBuffer, location, ""
}

func getMaxTimestampOfDataPresent(project model.Project, tableCreated bool) (time.Time, int) {
	storeSelected := store.GetStore()
	if tableCreated {
		maxTimestamp, statusCode := storeSelected.GetMaxTimestampOfDataPresenceFromWebsiteAggregation(project.ID, project.TimeZone)
		return maxTimestamp, statusCode
	}
	maxTimeBeginningOfDay := U.GetBeginningOfDayTimeZ(time.Now().Unix(), U.TimeZoneString(project.TimeZone))
	return maxTimeBeginningOfDay.AddDate(0, 0, -32), http.StatusOK
}

// add fmt.Println later to check how it work.
func executeDbtCommand(project model.Project, from, to int64, tableCreated bool) (bool, string, bool) {
	log.WithField("project", project.ID).WithField("from", from).WithField("to", to).Warn("DBT ran for this params")
	command := fmt.Sprintf("dbt run --vars '{project_id: %v, time_zone: %v, from: %v, to: %v}'", project.ID, project.TimeZone, from, to)
	cmd := exec.Command("/bin/sh", "-c", command)
	_, stdErr := cmd.Output()
	// log.WithField("stdOut", stdOut).Warn("Printing dbt output")
	if stdErr != nil {
		log.WithField("project_id", project.ID).WithField("err", stdErr.Error()).Warn("Failed in Dbt run")
		return true, string(stdErr.Error()), tableCreated
	}
	return false, "dbt completed", true
}

// This is printing out of order in gcp. TODO handle.
func printDebugLogFile() {
	f, err := os.Open("./logs/dbt.log")
	if err != nil {
		log.Error("Failed in opening dbt log - events_cube_aggregation_deploy" + err.Error())
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Error(err)
		}
	}()
	log.Warn("Start of the debug logfile.")
	s := bufio.NewScanner(f)
	for s.Scan() {
		log.Warn(s.Text())
	}
	err = s.Err()
	if err != nil {
		log.Error("Failed during scan of dbt log - events_cube_aggregation_deploy" + err.Error())
	}
	log.Warn("End of the debug logfile.")
}

func addToJobReport(jobReport map[string]map[string]interface{}, isError bool, errorString string, key string) {

	if isError {
		jobReport["failures"][key] = errorString
	}
	if _, exists := jobReport["success"][key]; !exists {
		jobReport["success"]["count"] = jobReport["success"]["count"].(int) + 1
		jobReport["success"][key] = ""
	}
}

func addMetricsToReportIfCrossesThreshold(jobReport map[string]map[string]interface{}, jobRunTimeInSeconds float64, numberOfDays float64, projectID int64) {
	if jobRunTimeInSeconds > numberOfDays*60 {
		key := fmt.Sprintf("%v:%v", projectID, numberOfDays)
		jobReport["long_run_projects"][key] = jobRunTimeInSeconds
	}
}

func getDuplicateJobReportWithoutSuccessKeys(jobReport map[string]map[string]interface{}) map[string]map[string]interface{} {
	dupJobReport := make(map[string]map[string]interface{})
	dupJobReport["success"] = make(map[string]interface{})
	dupJobReport["success"]["count"] = 0
	dupJobReport["failures"] = make(map[string]interface{})
	dupJobReport["long_run_projects"] = make(map[string]interface{})

	dupJobReport["success"]["count"] = jobReport["success"]["count"]
	dupJobReport["failures"] = jobReport["failures"]
	dupJobReport["long_run_projects"] = jobReport["long_run_projects"]
	return dupJobReport
}
