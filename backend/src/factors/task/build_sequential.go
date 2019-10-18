package task

import (
	"factors/filestore"
	M "factors/model"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"

	"factors/util"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const taskID = "Task#BuildSequential"

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

func notifyOnPanic(env string) {
	if pe := recover(); pe != nil {
		if ne := util.NotifyThroughSNS(taskID, env, map[string]interface{}{"panic_error": pe, "stacktrace": string(debug.Stack())}); ne != nil {
			log.Fatal(ne, pe)
		}
		log.Fatal(pe)
	}
}

// BuildSequential - runs model building sequenitally for all project intervals.
func BuildSequential(env string, db *gorm.DB, cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient, diskManger *serviceDisk.DiskDriver,
	bucketName string, noOfPatternWorkers int, projectId uint64,
	projectIdsToSkip map[uint64]bool, maxModelSize int64) error {

	defer notifyOnPanic(env)

	// Todo(Dinesh): Add success and failure notification.
	// Idea: []Builds from this can be queued and workers can process.
	builds, activeProjects, err := GetNextBuilds(db, cloudManager, etcdClient)
	if err != nil {
		bsLog.WithField("error", err).Error("Failed to get next build info.")
	}

	bsLog.Infof("Queueing %d builds required.", len(builds))

	success := make([]BuildSuccess, 0, 0)
	failures := make([]BuildFailure, 0, 0)

	for _, build := range builds {
		// Build model, for projectId if given, else for all.
		if projectId > 0 && build.ProjectId != projectId {
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
			"ReadableStartTime": unixToHumanTime(build.StartTimestamp),
			"ReadableEndTime":   unixToHumanTime(build.EndTimestamp),
		})

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
			build.StartTimestamp, build.EndTimestamp, maxModelSize)
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

	newActiveProjects := make([]M.ProjectEventsInfo, 0, 0)
	for _, pei := range activeProjects {
		// Adds new projects with first event timestamp within last 24 hours.
		if pei.FirstEventTimestamp >= util.UnixTimeBeforeDuration(time.Hour*24) {
			newActiveProjects = append(newActiveProjects, pei)
		}
	}

	buildStatus := map[string]interface{}{
		"success":             success,
		"failures":            failures,
		"all_active_projects": activeProjects,
		"new_active_projects": newActiveProjects,
	}
	if err := util.NotifyThroughSNS(taskID, env, buildStatus); err != nil {
		log.WithError(err).Error("Failed to notify build status.")
	} else {
		log.Info("Notified build status.")
	}

	return nil
}
