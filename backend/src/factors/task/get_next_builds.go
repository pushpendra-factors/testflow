package task

import (
	"errors"
	"factors/filestore"
	M "factors/model"
	PMM "factors/pattern_model_meta"
	serviceEtcd "factors/services/etcd"
	"time"

	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

const (
	ModelTypeMonth = "m"
	ModelTypeWeek  = "w"
)

const OneSec = 1

type Build struct {
	ProjectId      uint64 `json:"pid"`
	ModelType      string `json:"mt"`
	StartTimestamp int64  `json:"st"`
	EndTimestamp   int64  `json:"et"`
}

// Returns last build timestamp lookup map for each project by type.
func makeLastBuildTimestampMap(projectData []PMM.ProjectData) *map[uint64]map[string]int64 {
	projectLatestModel := make(map[uint64]map[string]int64, 0)

	for _, p := range projectData {
		if _, exist := projectLatestModel[p.ID]; !exist {
			projectLatestModel[p.ID] = make(map[string]int64, 0)
		}
		if _, exist := projectLatestModel[p.ID][p.ModelType]; !exist {
			projectLatestModel[p.ID][p.ModelType] = 0
		}

		if p.EndTimestamp > projectLatestModel[p.ID][p.ModelType] {
			projectLatestModel[p.ID][p.ModelType] = p.EndTimestamp
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

	log.Error("Unknown floor timestamp type.")
	return 0
}

func ceilTimestampByType(modelType string, timestamp int64) int64 {
	if modelType == ModelTypeWeek {
		return now.New(time.Unix(timestamp, 0).UTC()).EndOfWeek().Unix()
	}

	if modelType == ModelTypeMonth {
		return now.New(time.Unix(timestamp, 0).UTC()).EndOfMonth().Unix()
	}

	log.Error("Unknown ceil timestamp type.")
	return 0
}

func unixToHumanTime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

func addPendingIntervalsForProjectByType(builds *[]Build, projectId uint64,
	modelType string, initTimestamp int64, limitTimestamp int64) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectId, "ModelType": modelType,
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

		*builds = append(*builds, Build{ProjectId: projectId, StartTimestamp: startTimestamp,
			EndTimestamp: endTimestamp, ModelType: modelType})
		startTimestamp = endTimestamp + OneSec
	}
}

func addNextIntervalsForProjectByType(builds *[]Build, projectId uint64, modelType string,
	prevBuildEndTime int64, startEventTime int64, endEventTime int64) {

	log.WithFields(log.Fields{"ProjectId": projectId, "ModelType": modelType,
		"PrevBuildEndTime": prevBuildEndTime}).Debug("Adding next intervals to build.")

	if prevBuildEndTime > 0 {
		addPendingIntervalsForProjectByType(builds, projectId, modelType,
			prevBuildEndTime+OneSec, endEventTime)
	} else {
		addPendingIntervalsForProjectByType(builds, projectId, modelType,
			startEventTime, endEventTime)
	}
}

// GetNextBuilds - Gets next batch of intervals by project, for building models.
func GetNextBuilds(db *gorm.DB, cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient) ([]Build, error) {

	if db == nil {
		return nil, fmt.Errorf("db cannot be nil, get build info failed")
	}

	builds := make([]Build, 0, 0)

	projectsMeta, err := PMM.GetProjectsMetadata(cloudManager, etcdClient)
	if err != nil {
		log.Error("Failed to get current project metadata")
		return nil, err
	}

	// All project event timestamp info.
	pEventTimeInfo, errCode := M.GetProjectEventTimeInfo()
	if errCode != http.StatusFound {
		return nil, errors.New("unable to fetch projects")
	}

	// Intervals for existing projects on meta.
	lastBuildOfProjects := makeLastBuildTimestampMap(projectsMeta)
	for pid, buildTimeByType := range *lastBuildOfProjects {
		log.Infof("Last build pid:%d info:%+v", pid, buildTimeByType)
		if (*pEventTimeInfo)[pid] != nil {
			addNextIntervalsForProjectByType(&builds, pid, ModelTypeWeek,
				buildTimeByType[ModelTypeWeek], (*pEventTimeInfo)[pid].First,
				(*pEventTimeInfo)[pid].Last)
			addNextIntervalsForProjectByType(&builds, pid, ModelTypeMonth,
				buildTimeByType[ModelTypeMonth], (*pEventTimeInfo)[pid].First,
				(*pEventTimeInfo)[pid].Last)
		} else {
			log.WithField("ProjectId", pid).Error("No events for a project found on meta.")
		}
	}

	// Intervals for non existing projects on metadata.
	noMetaProjects := make([]uint64, 0, 0)
	for pid := range *pEventTimeInfo {
		if _, exist := (*lastBuildOfProjects)[pid]; !exist {
			noMetaProjects = append(noMetaProjects, pid)
		}
	}

	for _, pid := range noMetaProjects {
		addPendingIntervalsForProjectByType(&builds, pid, ModelTypeWeek,
			(*pEventTimeInfo)[pid].First, (*pEventTimeInfo)[pid].Last)
		addPendingIntervalsForProjectByType(&builds, pid, ModelTypeMonth,
			(*pEventTimeInfo)[pid].First, (*pEventTimeInfo)[pid].Last)
	}

	return builds, nil
}
