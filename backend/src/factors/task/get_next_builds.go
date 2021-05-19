package task

import (
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	serviceEtcd "factors/services/etcd"
	"net/http"
	"time"

	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

const (
	ModelTypeMonth   = "m"
	ModelTypeWeek    = "w"
	ModelTypeQuarter = "q"
)

const (
	ModelTypeAll       = "all"
	ModelTypeWeekly    = "weekly"
	ModelTypeMonthly   = "monthly"
	ModelTypeQuarterly = "quarterly"
)
const OneSec = 1

type Build struct {
	ProjectId      uint64 `json:"pid"`
	ModelType      string `json:"mt"`
	StartTimestamp int64  `json:"st"`
	EndTimestamp   int64  `json:"et"`
	Exists         bool   `json:"exists"`
}

var gnbLog = taskLog.WithField("prefix", "Task#GetNextBuilds")

// Returns last build timestamp lookup map for each project by type.
func makeLastBuildTimestampMap(projectData []model.ProjectModelMetadata,
	buildStartTimestamp int64) *map[uint64]map[string]int64 {

	minEndTimestamp := buildStartTimestamp
	projectLatestModel := make(map[uint64]map[string]int64, 0)

	for _, p := range projectData {
		if p.EndTime > minEndTimestamp {
			if _, exist := projectLatestModel[p.ProjectId]; !exist {
				projectLatestModel[p.ProjectId] = make(map[string]int64, 0)
			}
			if _, exist := projectLatestModel[p.ProjectId][p.ModelType]; !exist {
				projectLatestModel[p.ProjectId][p.ModelType] = 0
			}

			if p.EndTime > projectLatestModel[p.ProjectId][p.ModelType] {
				projectLatestModel[p.ProjectId][p.ModelType] = p.EndTime
			}
		}
	}

	// map[projectId]map[modelType] = lastEndTimestamp
	return &projectLatestModel
}

func isFutureTimestamp(timestamp int64) bool {
	return timestamp > time.Now().Unix()
}

func floorTimestampByType(modelType string, timestamp int64) int64 {
	if modelType == ModelTypeWeek {
		return now.New(time.Unix(timestamp, 0).UTC()).BeginningOfWeek().Unix()
	}

	if modelType == ModelTypeMonth {
		return now.New(time.Unix(timestamp, 0).UTC()).BeginningOfMonth().Unix()
	}

	if modelType == ModelTypeQuarter {
		return now.New(time.Unix(timestamp, 0).UTC()).BeginningOfQuarter().Unix()
	}

	gnbLog.Error("Unknown floor timestamp type.")
	return 0
}

func ceilTimestampByType(modelType string, timestamp int64) int64 {
	if modelType == ModelTypeWeek {
		return now.New(time.Unix(timestamp, 0).UTC()).EndOfWeek().Unix()
	}

	if modelType == ModelTypeMonth {
		return now.New(time.Unix(timestamp, 0).UTC()).EndOfMonth().Unix()
	}

	if modelType == ModelTypeQuarter {
		return now.New(time.Unix(timestamp, 0).UTC()).EndOfQuarter().Unix()
	}

	gnbLog.Error("Unknown ceil timestamp type.")
	return 0
}

func addPendingIntervalsForProjectByType(builds *[]Build, projectId uint64, modelType string,
	initTimestamp int64, limitTimestamp int64) {

	logCtx := gnbLog.WithFields(log.Fields{"ProjectId": projectId, "ModelType": modelType,
		"InitTimestamp": initTimestamp, "LimitTimestamp": limitTimestamp})

	if isFutureTimestamp(initTimestamp) {
		logCtx.Error("Future time is not allowed as interval init timestamp")
		return
	}

	actInitTimestamp := floorTimestampByType(modelType, initTimestamp)
	actLimitTimestamp := ceilTimestampByType(modelType, limitTimestamp)
	logCtx = logCtx.WithFields(log.Fields{"ActInitTimestamp": actInitTimestamp,
		"ActLimitTimestamp": actLimitTimestamp})

	startTimestamp := actInitTimestamp
	for startTimestamp <= actLimitTimestamp {
		endTimestamp := ceilTimestampByType(modelType, startTimestamp)

		if isFutureTimestamp(endTimestamp) {
			logCtx.WithFields(log.Fields{"startTimestamp": startTimestamp,
				"endTimestamp": endTimestamp}).Info("Skipping interval with future endTimestamp.")
			return
		}

		*builds = append(*builds, Build{
			ProjectId:      projectId,
			StartTimestamp: startTimestamp,
			EndTimestamp:   endTimestamp,
			ModelType:      modelType,
		})
		startTimestamp = endTimestamp + OneSec
	}
}

func addNextIntervalsForProjectByType(builds *[]Build, projectId uint64, modelType string,
	prevBuildEndTime int64, startTimestamp, endTimestamp int64) {

	gnbLog.WithFields(log.Fields{"ProjectId": projectId, "ModelType": modelType,
		"PrevBuildEndTime": prevBuildEndTime}).Debug("Adding next intervals to build.")

	if prevBuildEndTime > 0 {
		addPendingIntervalsForProjectByType(builds, projectId, modelType,
			prevBuildEndTime+OneSec, endTimestamp)
	} else {
		addPendingIntervalsForProjectByType(builds, projectId, modelType,
			startTimestamp, endTimestamp)
	}
}

func getUniqueModelBuildID(projectID uint64, modelType string, startTimestamp, endTimestamp int64) string {
	return fmt.Sprintf("pid:%d:typ:%s:st:%d:et:%d", projectID, modelType, startTimestamp, endTimestamp)
}

// GetNextBuilds - Gets next batch of intervals by project, for building models.
func GetNextBuilds(db *gorm.DB, cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient, projectIDs map[uint64]bool, modelType string,
	startTimestamp, endTimestamp int64) ([]Build, map[string]bool, error) {

	existingBuilds := make(map[string]bool, 0)

	if db == nil {
		return nil, existingBuilds, fmt.Errorf("db cannot be nil, get build info failed")
	}

	builds := make([]Build, 0, 0)
	projectsMeta := make([]model.ProjectModelMetadata, 0)
	for projId, _ := range projectIDs {
		projMetadata, err, msg := store.GetStore().GetProjectModelMetadata(projId)
		if err != http.StatusFound {
			gnbLog.Error(msg)
			return nil, existingBuilds, nil
		}
		projectsMeta = append(projectsMeta, projMetadata...)
	}

	for _, meta := range projectsMeta {
		buildID := getUniqueModelBuildID(meta.ProjectId, meta.ModelType, meta.StartTime, meta.EndTime)
		existingBuilds[buildID] = true
	}

	// Adds intervals for existing projects on meta with respect
	// to last build of project and model_type.
	lastBuildOfProjects := makeLastBuildTimestampMap(projectsMeta, startTimestamp)
	for projectID, lastBuildEndTimestampByType := range *lastBuildOfProjects {
		addNextIntervalsForProjectByModelType(&builds, projectID,
			lastBuildEndTimestampByType, startTimestamp, endTimestamp, modelType)
	}

	// Adds intervals for non-existing projects on metadata with
	// respect to given start and end timestamp.
	noMetaProjects := make([]uint64, 0, 0)
	for projectID := range projectIDs {
		if _, exist := (*lastBuildOfProjects)[projectID]; !exist {
			noMetaProjects = append(noMetaProjects, projectID)
		}
	}

	for _, projectID := range noMetaProjects {
		addPendingIntervalsForProjectByModelType(&builds, projectID,
			startTimestamp, endTimestamp, modelType)
	}

	return builds, existingBuilds, nil
}

// A wrapper over addNextIntervalsForProjectByType() to allow model type parameter.
func addNextIntervalsForProjectByModelType(builds *[]Build, projectID uint64, lastBuildEndTimestampByType map[string]int64,
	startTimestamp, endTimestamp int64, modelType string) {

	switch modelType {
	case ModelTypeAll:
		addNextIntervalsForProjectByType(builds, projectID, ModelTypeWeek,
			lastBuildEndTimestampByType[ModelTypeWeek], startTimestamp, endTimestamp)
		addNextIntervalsForProjectByType(builds, projectID, ModelTypeMonth,
			lastBuildEndTimestampByType[ModelTypeMonth], startTimestamp, endTimestamp)
		break
	case ModelTypeWeekly:
		addNextIntervalsForProjectByType(builds, projectID, ModelTypeWeek,
			lastBuildEndTimestampByType[ModelTypeWeek], startTimestamp, endTimestamp)
		break
	case ModelTypeMonthly:
		addNextIntervalsForProjectByType(builds, projectID, ModelTypeMonth,
			lastBuildEndTimestampByType[ModelTypeMonth], startTimestamp, endTimestamp)
		break
	case ModelTypeQuarterly:
		addNextIntervalsForProjectByType(builds, projectID, ModelTypeQuarter,
			lastBuildEndTimestampByType[ModelTypeQuarter], startTimestamp, endTimestamp)
		break
	default:
		log.WithField("project_id", projectID).WithField("type", modelType).
			Error("Invalid model type on addNextIntervalsForProjectByModelType.")
		break
	}
}

/** A wrapper over addPendingIntervalsForProjectByType() to allow model type parameter **/
func addPendingIntervalsForProjectByModelType(builds *[]Build, projectID uint64,
	startTimestamp, endTimestamp int64, modelType string) {

	switch modelType {
	case ModelTypeAll:
		addPendingIntervalsForProjectByType(builds, projectID, ModelTypeWeek, startTimestamp, endTimestamp)
		addPendingIntervalsForProjectByType(builds, projectID, ModelTypeMonth, startTimestamp, endTimestamp)
		break
	case ModelTypeWeekly:
		addPendingIntervalsForProjectByType(builds, projectID, ModelTypeWeek, startTimestamp, endTimestamp)
		break
	case ModelTypeMonthly:
		addPendingIntervalsForProjectByType(builds, projectID, ModelTypeMonth, startTimestamp, endTimestamp)
		break
	case ModelTypeQuarterly:
		addPendingIntervalsForProjectByType(builds, projectID, ModelTypeQuarter, startTimestamp, endTimestamp)
		break
	default:
		log.WithField("project_id", projectID).WithField("type", modelType).
			Error("Invalid model type on addNextIntervalsForProjectByModelType.")
		break
	}
}
