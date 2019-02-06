package task

import (
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var bsLog = taskLog.WithField("prefix", "Task#BuildSequential")

// BuildSequential - runs model building sequenitally for all project
// intervals.
func BuildSequential(db *gorm.DB, cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient, diskManger *serviceDisk.DiskDriver,
	bucketName string, noOfPatternWorkers int, projectId uint64) error {

	// Todo(Dinesh): Add success and failure notification.
	// Idea: []Builds from this can be queued and workers can process.
	builds, err := GetNextBuilds(db, cloudManager, etcdClient)
	if err != nil {
		bsLog.WithField("error", err).Error("Failed to get next build info.")
	}

	bsLog.Infof("Queueing %d builds required.", len(builds))

	for _, build := range builds {
		// Build model, for projectId if given, else for all.
		if projectId > 0 && build.ProjectId != projectId {
			bsLog.WithField("ProjectId", build.ProjectId).Info("Skipping build for the non-given project.")
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
		startAt := time.Now().Unix()
		modelId, eventsCount, err := PullEvents(db, cloudManager, diskManger,
			build.ProjectId, build.StartTimestamp, build.EndTimestamp)
		if err != nil {
			logCtx.WithField("error", err).Error("Failed to pull events.")
			continue
		}
		if eventsCount == 0 {
			logCtx.Info("Zero events. Skipping pattern mine.")
			continue
		}
		logCtx = logCtx.WithFields(log.Fields{"ModelId": modelId, "EventsCount": eventsCount})
		timeTakenToPullEvents := (time.Now().Unix() - startAt)
		logCtx = logCtx.WithField("TimeTakenToPullEventsInSecs", timeTakenToPullEvents)

		// Patten mine
		startAt = time.Now().Unix()
		newProjectMetaVersion, err := PatternMine(db, etcdClient, cloudManager, diskManger,
			bucketName, noOfPatternWorkers, build.ProjectId, modelId, build.ModelType,
			build.StartTimestamp, build.EndTimestamp)
		if err != nil {
			logCtx.Error("Failed to mine patterns.")
			continue
		}
		logCtx = logCtx.WithFields(log.Fields{"NewProjectMetaVersion": newProjectMetaVersion})
		timeTakenToMinePatterns := (time.Now().Unix() - startAt)
		logCtx = logCtx.WithField("TimeTakenToMinePatternsInSecs", timeTakenToMinePatterns)
	}

	return nil
}
