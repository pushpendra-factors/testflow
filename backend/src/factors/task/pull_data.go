package task

import (
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	"factors/pull"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var peLog = taskLog.WithField("prefix", "Task#PullEvents")

const DATA_FAILURE_ALERT_LIMIT = 2 //number of days after which data unavailability is considered to be an error

// Pull daily data (created_at in a day) into respective folders(based on timestamp) in archive bucket
func PullAllDataV2(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	hardPull := configs["hardPull"].(*bool)
	fileTypes := configs["fileTypes"].(map[int64]bool)
	splitRangeProjectIds := configs["splitRangeProjectIds"].([]int64)
	noOfSplits := configs["noOfSplits"].(int)

	status := make(map[string]interface{})
	if projectId == 0 {
		status["error"] = "invalid project_id"
		return status, false
	}
	if startTimestamp == 0 {
		status["error"] = "invalid start timestamp"
		return status, false
	}
	if endTimestamp == 0 || endTimestamp > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}

	projectDetails, _ := store.GetStore().GetProject(projectId)
	startTimestampInProjectTimezone := startTimestamp
	endTimestampInProjectTimezone := endTimestamp
	if projectDetails.TimeZone != "" {
		// Input time is in UTC. We need the same time in the other timezone
		// if 2021-08-30 00:00:00 is UTC then we need the epoch equivalent in 2021-08-30 00:00:00 IST(project time zone)
		offset := U.FindOffsetInUTC(U.TimeZoneString(projectDetails.TimeZone))
		startTimestampInProjectTimezone = startTimestamp - int64(offset)
		endTimestampInProjectTimezone = endTimestamp - int64(offset)
	}

	if endTimestampInProjectTimezone > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp (project timezone)"
		return status, false
	}

	logCtx := peLog.WithFields(log.Fields{"ProjectId": projectId,
		"StartTime": startTimestampInProjectTimezone, "EndTime": endTimestampInProjectTimezone})

	success := true

	allIntegrationsSupported := []string{M.HUBSPOT, M.SALESFORCE, M.ADWORDS, M.BINGADS, M.FACEBOOK, M.LINKEDIN, M.GOOGLE_ORGANIC}
	var pullFileTypes map[string]bool
	var err error
	pullFileTypes, success, err = checkIntegrationsDataAvailabilityAndHardPull(allIntegrationsSupported, projectId, startTimestamp, endTimestamp, endTimestampInProjectTimezone, cloudManager, fileTypes, *hardPull, status, logCtx)
	if err != nil {
		return status, false
	}

	// EVENTS
	if pullFileTypes["events"] {
		if _, ok := pull.PullDataForEvents(projectId, cloudManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, splitRangeProjectIds, noOfSplits, status, logCtx); !ok {
			return status, false
		}
	}

	// AD REPORTS
	for _, channel := range []string{M.ADWORDS, M.BINGADS, M.FACEBOOK, M.GOOGLE_ORGANIC, M.LINKEDIN} {
		if pullFileTypes[channel] {
			if _, ok := pull.PullDataForChannel(channel, projectId, cloudManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, status, logCtx); !ok {
				return status, false
			}
		}
	}

	//USERS
	if pullFileTypes[M.USERS] {
		if _, ok := pull.PullUsersDataForCustomMetrics(projectId, cloudManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, hardPull, status, logCtx); !ok {
			return status, false
		}
	}

	return status, success
}

// Check which integrations are present, whether data is available and if the required file already exists or not
// Return a bool map (pullFileTypes) telling which files to actually pull from db and a bool (success) telling whether job is successful till now
func checkIntegrationsDataAvailabilityAndHardPull(allIntegrationsSupported []string, projectId, startTimestamp, endTimestamp, endTimestampInProjectTimezone int64, cloudManager *filestore.FileManager, fileTypes map[int64]bool, hardPull bool, status map[string]interface{}, logCtx *log.Entry) (map[string]bool, bool, error) {
	var pullFileTypes = make(map[string]bool)
	success := true
	if integrationsStatus, err := store.GetStore().GetLatestDataStatus(allIntegrationsSupported, projectId, false); err != nil {
		logCtx.WithError(err).Error("error getting integrations status")
		status["error"] = err.Error()
		return pullFileTypes, false, err
	} else {
		for fileType, fileTypeNum := range pull.FileType {
			if !fileTypes[fileTypeNum] {
				continue
			}
			if fileType == "events" {
				if !hardPull {
					if ok, _ := pull.CheckEventFileExists(cloudManager, projectId, startTimestamp, startTimestamp, endTimestamp); ok {
						status["events-info"] = "File already exists"
						continue
					}
				}
				pullFileTypes[fileType] = true
				var eventsPull bool = true
				var errStr string
				for _, integration := range []string{M.HUBSPOT, M.SALESFORCE} {
					intStatus := integrationsStatus[integration]
					if !intStatus.IntegrationStatus {
						status[integration+"-info"] = "Not Integrated"
					} else {
						if int64(intStatus.LatestData) < endTimestampInProjectTimezone {
							eventsPull = false
							success = false
							errStr += integration + " "
							noOfDaysFromNow := (U.TimeNowUnix() - int64(intStatus.LatestData)) / U.Per_day_epoch
							var key string
							if noOfDaysFromNow > DATA_FAILURE_ALERT_LIMIT {
								key = integration + "-error"
							} else {
								key = integration + "-info"
							}
							status[key] = fmt.Sprintf("Data not available after %s", time.Unix(int64(intStatus.LatestData), 0).Format("01-02-2006 15:04:05"))
						}
					}
				}
				if !eventsPull {
					status[fileType+"-info"] = errStr + "data availability error"
					pullFileTypes[fileType] = false
				}
			} else if fileType == "users" {
				pullFileTypes[fileType] = true
			} else {
				if !hardPull {
					if ok, _ := pull.CheckChannelFileExists(fileType, cloudManager, projectId, startTimestamp, startTimestamp, endTimestamp); ok {
						status[fileType+"-info"] = "File already exists"
						continue
					}
				}
				intStatus := integrationsStatus[fileType]
				pullFileTypes[fileType] = true
				if !intStatus.IntegrationStatus {
					pullFileTypes[fileType] = false
					status[fileType+"-info"] = "Not Integrated"
				} else {
					if int64(intStatus.LatestData) < endTimestampInProjectTimezone {
						pullFileTypes[fileType] = false
						success = false
						noOfDays := (U.TimeNowUnix() - int64(intStatus.LatestData)) / U.Per_day_epoch
						var key string
						if noOfDays > DATA_FAILURE_ALERT_LIMIT {
							key = fileType + "-error"
						} else {
							key = fileType + "-info"
						}
						status[key] = fmt.Sprintf("Data not available after %s", time.Unix(int64(intStatus.LatestData), 0).Format("01-02-2006 15:04:05"))
					}
				}
			}
		}
	}
	return pullFileTypes, success, nil
}

func MergeAndWriteSortedFileTask(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	hardPull := configs["hardPull"].(*bool)
	fileTypes := configs["fileTypes"].(map[int64]bool)
	beamConfig := configs["beamConfig"].(*merge.RunBeamConfig)

	status := make(map[string]interface{})
	if projectId == 0 {
		status["error"] = "invalid project_id"
		return status, false
	}
	if startTimestamp == 0 {
		status["error"] = "invalid start timestamp"
		return status, false
	}
	if endTimestamp == 0 || endTimestamp > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}

	success := true
	for ftype, _ := range fileTypes {
		if ftype == pull.FileType["events"] {
			_, _, err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeEvent, "", startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, cloudManager, diskManager, beamConfig, *hardPull, 0, true, false, false)
			if err != nil {
				status["events-error"] = err
				success = false
			}
		} else if ftype == pull.FileType["users"] {
			uniqueDateFileds := make(map[string]bool)
			{
				customMetrics, errStr, getStatus := store.GetStore().GetCustomMetricByProjectIdAndQueryType(projectId, M.ProfileQueryType)
				if getStatus != http.StatusFound {
					peLog.WithField("error", errStr).Error("Pull users failed. get custom metrics failed.")
					status["users-error"] = errStr
					return status, false
				}
				for _, customMetric := range customMetrics {
					var customMetricTransformation M.CustomMetricTransformation
					err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
					if err != nil {
						status["users-error"] = "Error during decode of custom metrics transformations"
						return status, false
					}
					if _, ok := uniqueDateFileds[customMetricTransformation.DateField]; !ok {
						uniqueDateFileds[customMetricTransformation.DateField] = true
					}
				}
			}
			for dateField, _ := range uniqueDateFileds {
				_, _, err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeUser, dateField, startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, cloudManager, diskManager, beamConfig, *hardPull, 0, true, false, false)
				if err != nil {
					status["users-"+dateField+"-error"] = err
					success = false
				}
			}
		} else {
			for channel, ft := range pull.FileType {
				if ft == ftype {
					_, _, err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeAdReport, channel, startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, cloudManager, diskManager, beamConfig, *hardPull, 0, true, false, false)
					if err != nil {
						status[channel+"-error"] = err
						success = false
					}
				}
			}
		}
	}

	return status, success
}
