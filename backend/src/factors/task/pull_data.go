package task

import (
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	"factors/pull"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var peLog = taskLog.WithField("prefix", "Task#PullEvents")

func PullAllDataV2(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	hardPull := configs["hardPull"].(*bool)
	fileTypes := configs["fileTypes"].(map[int64]bool)

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

	logCtx := peLog.WithFields(log.Fields{"ProjectId": projectId,
		"StartTime": startTimestampInProjectTimezone, "EndTime": endTimestampInProjectTimezone})

	integrations := store.GetStore().IsIntegrationAvailable(projectId)

	success := true

	// EVENTS
	if fileTypes[pull.FileType["events"]] {
		exists := false
		if !*hardPull {
			if ok, _ := pull.CheckEventFileExists(cloudManager, projectId, startTimestamp, startTimestamp, endTimestamp); ok {
				status["events-info"] = "File already exists"
				exists = true
			}
		}

		if !exists {
			toPull := true
			var errMsg string = "data not available for "
			for _, channel := range []string{M.HUBSPOT, M.SALESFORCE} {
				if !integrations[channel] {
					status[channel+"-info"] = "Not Integrated"
				} else {
					if !store.GetStore().IsDataAvailable(projectId, channel, uint64(endTimestampInProjectTimezone)) {
						errMsg += channel + " "
						toPull = false
					}
				}
			}

			if toPull {
				if _, ok := pull.PullEventsData(projectId, cloudManager, diskManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, status, logCtx); !ok {
					return status, false
				}
			} else {
				success = false
				status["events-error"] = errMsg
			}
		}

	}

	// AD REPORTS
	for _, channel := range []string{M.ADWORDS, M.BINGADS, M.FACEBOOK, M.GOOGLE_ORGANIC, M.LINKEDIN} {
		if fileTypes[pull.FileType[channel]] {
			if !*hardPull {
				if ok, _ := pull.CheckChannelFileExists(channel, cloudManager, projectId, startTimestamp, startTimestamp, endTimestamp); ok {
					status[channel+"-info"] = "File already exists"
					continue
				}
			}
			if !integrations[channel] {
				status[channel+"-info"] = "Not Integrated"
			} else {
				if store.GetStore().IsDataAvailable(projectId, channel, uint64(endTimestampInProjectTimezone)) {
					if _, ok := pull.PullDataForChannel(channel, projectId, cloudManager, diskManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, status, logCtx); !ok {
						return status, false
					}
				} else {
					success = false
					status[channel+"-error"] = "Data not available"
				}
			}
		}
	}

	//USERS
	if fileTypes[pull.FileType["users"]] {
		if _, ok := pull.PullUsersDataForCustomMetrics(projectId, cloudManager, diskManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, hardPull, status, logCtx); !ok {
			return status, false
		}
	}

	return status, success
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

	for ftype, _ := range fileTypes {
		if ftype == pull.FileType["events"] {
			merge.MergeAndWriteSortedFile(projectId, U.DataTypeEvent, "", startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, cloudManager, diskManager, beamConfig, *hardPull, 0)
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
				merge.MergeAndWriteSortedFile(projectId, U.DataTypeUser, dateField, startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, cloudManager, diskManager, beamConfig, *hardPull, 0)
			}
		} else {
			for channel, ft := range pull.FileType {
				if ft == ftype {
					merge.MergeAndWriteSortedFile(projectId, U.DataTypeAdReport, channel, startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, cloudManager, diskManager, beamConfig, *hardPull, 0)
				}
			}
		}
	}

	return status, true
}
