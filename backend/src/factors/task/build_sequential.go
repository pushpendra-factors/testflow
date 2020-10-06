package task

import (
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	U "factors/util"

	"factors/util"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const taskID = "Task#BuildSequential"
const oneDayInSecs = 24 * 60 * 60

var bsLog = taskLog.WithField("prefix", taskID)

type BuildFailure struct {
	Build   Build  `json:"build"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type BuildSuccess struct {
	Build               Build `json:"build"`
	PullEventsTimeInMS  int64 `json:"pulled_events_in_ms"`
	PatternMineTimeInMS int64 `json:"mined_patterns_in_ms"`
	NumberOfChunks      int   `json:"num_chunks"`
	NumberOfEvents      int   `json:"num_events"`
}

// BuildSequential - runs model building sequenitally for all project intervals.
func BuildSequential(env string, db *gorm.DB, cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient, diskManger *serviceDisk.DiskDriver,
	bucketName string, noOfPatternWorkers int, projectIdToRun map[uint64]bool,
	projectIdsToSkip map[uint64]bool, maxModelSize int64, modelType string,
	lookBackPeriodInDays, noOfDays int64, countOccurence bool) error {

	defer util.NotifyOnPanic(taskID, env)

	if lookBackPeriodInDays == 0 {
		lookBackPeriodInDays = 30
	}
	lookBackPeriodInSecs := lookBackPeriodInDays * oneDayInSecs

	currentTimestamp := U.TimeNowUnix()
	startTimestamp := currentTimestamp - lookBackPeriodInSecs
	endTimestamp := startTimestamp + (noOfDays * oneDayInSecs)

	// end_timestamp cannot be same as start_timestamp (when no.of days is 0)
	// or higher than current_timestamp.
	if startTimestamp == endTimestamp || endTimestamp > currentTimestamp {
		endTimestamp = currentTimestamp
	}

	// Idea: []Builds from this can be queued and workers can process.
	builds, existingModelBuilds, err := GetNextBuilds(db, cloudManager, etcdClient, projectIdToRun,
		modelType, startTimestamp, endTimestamp)
	if err != nil {
		bsLog.WithError(err).Error("Failed to get next build info.")
		return err
	}

	bsLog.Infof("Queueing %d builds required.", len(builds))

	success := make([]BuildSuccess, 0, 0)
	failures := make([]BuildFailure, 0, 0)

	for _, build := range builds {
		// Skip builds for projects not in projectToRun.
		if _, ok := projectIdToRun[build.ProjectId]; !ok {
			bsLog.WithField("ProjectId", build.ProjectId).Info("Skipping build for the non-given project.")
			continue
		}
		if _, ok := projectIdsToSkip[build.ProjectId]; ok {
			bsLog.WithField("ProjectId", build.ProjectId).Info("Skipping build for the project.")
			continue
		}

		logCtx := bsLog.WithFields(log.Fields{"ProjectId": build.ProjectId,
			"StartTime": build.StartTimestamp, "EndTime": build.EndTimestamp,
			"Type": build.ModelType})
		// Readable time for debug.
		logCtx = logCtx.WithFields(log.Fields{
			"ReadableStartTime": util.UnixToHumanTime(build.StartTimestamp),
			"ReadableEndTime":   util.UnixToHumanTime(build.EndTimestamp),
		})

		buildID := getUniqueModelBuildID(build.ProjectId, build.ModelType, build.StartTimestamp, build.EndTimestamp)
		if _, exists := existingModelBuilds[buildID]; exists {
			build.Exists = true
			success = append(success, BuildSuccess{Build: build})
			continue
		}

		// Pull events
		startAt := time.Now().UnixNano()
		modelId, eventsCount, err := PullEvents(db, cloudManager, diskManger,
			build.ProjectId, build.StartTimestamp, build.EndTimestamp)
		if err != nil {
			logCtx.WithField("error", err).Error("Failed to pull events.")
			failures = append(failures, BuildFailure{Build: build, Error: fmt.Sprintf("%s", err), Message: "Pattern mining failure"})
			continue
		}
		if eventsCount == 0 {
			logCtx.Info("Zero events. Skipping pattern mine.")
			continue
		}
		logCtx = logCtx.WithFields(log.Fields{"ModelId": modelId, "EventsCount": eventsCount})
		timeTakenToPullEvents := (time.Now().UnixNano() - startAt) / 1000000
		logCtx = logCtx.WithField("TimeTakenToPullEventsInMS", timeTakenToPullEvents)

		// Patten mine
		startAt = time.Now().UnixNano()
		newProjectMetaVersion, numChunks, err := PatternMine(db, etcdClient, cloudManager, diskManger,
			bucketName, noOfPatternWorkers, build.ProjectId, modelId, build.ModelType,
			build.StartTimestamp, build.EndTimestamp, maxModelSize, countOccurence)
		if err != nil {
			logCtx.Error("Failed to mine patterns.")
			failures = append(failures, BuildFailure{Build: build, Error: fmt.Sprintf("%s", err), Message: "Pattern mining failure"})
			continue
		}
		logCtx = logCtx.WithFields(log.Fields{"NewProjectMetaVersion": newProjectMetaVersion})
		timeTakenToMinePatterns := (time.Now().UnixNano() - startAt) / 1000000
		logCtx = logCtx.WithField("TimeTakenToMinePatternsInMS", timeTakenToMinePatterns)
		success = append(success, BuildSuccess{
			Build:               build,
			NumberOfChunks:      numChunks,
			NumberOfEvents:      eventsCount,
			PullEventsTimeInMS:  timeTakenToPullEvents,
			PatternMineTimeInMS: timeTakenToMinePatterns})
	}

	buildStatus := map[string]interface{}{
		"success":  success,
		"failures": failures,
	}
	if err := util.NotifyThroughSNS(taskID, env, buildStatus); err != nil {
		log.WithError(err).Error("Failed to notify build status.")
	} else {
		log.Info("Notified build status.")
	}

	return nil
}
