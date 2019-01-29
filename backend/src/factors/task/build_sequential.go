package task

import (
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// BuildSequential - runs model building sequenitally for all project
// intervals.
func BuildSequential(db *gorm.DB, cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient, diskManger *serviceDisk.DiskDriver,
	localDiskTmpDir string, bucketName string, noOfPatternWorkers int, projectId uint64) error {

	// Todo(Dinesh): Add success and failure notification.
	// Idea: []Builds from this can be queued and workers can process.
	builds, err := GetNextBuilds(db, cloudManager, etcdClient)
	if err != nil {
		log.WithField("error", err).Error("Failed to get next build info.")
	}

	for _, build := range builds {
		// Build model, for projectId if given, else for all.
		if projectId > 0 && build.ProjectId != projectId {
			log.WithField("ProjectId", build.ProjectId).Info("Skipping build for the non-given project.")
			continue
		}

		logCtx := log.WithFields(log.Fields{"ProjectId": build.ProjectId,
			"StartTime": build.StartTimestamp, "EndTime": build.EndTimestamp,
			"Type": build.ModelType})
		// Readable time for debug.
		logCtx = logCtx.WithFields(log.Fields{
			"ReadableStartTime": unixToHumanTime(build.StartTimestamp),
			"ReadableEndTime":   unixToHumanTime(build.EndTimestamp),
		})

		// Pull events
		startAt := time.Now().Unix()
		logCtx.Info("**** Starting to pull events *****")
		modelId, eventsCount, err := PullEvents(db, cloudManager, diskManger, localDiskTmpDir,
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
		logCtx.Info("**** Pulled events successfully ****")

		// Patten mine
		startAt = time.Now().Unix()
		logCtx.Info("***** Starting to mine patterns *****")
		newProjectMetaVersion, err := PatternMine(db, etcdClient, cloudManager, diskManger,
			localDiskTmpDir, bucketName, noOfPatternWorkers, build.ProjectId, modelId,
			build.ModelType, build.StartTimestamp, build.EndTimestamp)
		if err != nil {
			logCtx.Error("Failed to mine patterns.")
			continue
		}
		logCtx = logCtx.WithFields(log.Fields{"NewProjectMetaVersion": newProjectMetaVersion})
		timeTakenToMinePatterns := (time.Now().Unix() - startAt)
		logCtx = logCtx.WithField("TimeTakenToMinePatternsInSecs", timeTakenToMinePatterns)
		logCtx.Info("**** Mined patterns successfully and update version metadata ****")
	}

	return nil
}
