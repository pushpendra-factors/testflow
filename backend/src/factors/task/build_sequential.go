package task

import (
	"factors/filestore"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	"factors/util"
	"math/rand"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const taskID = "Task#BuildSequential"
const oneDayInSecs = 24 * 60 * 60

var bsLog = taskLog.WithField("prefix", taskID)

const (
	ModelTypeMonth   = "m"
	ModelTypeWeek    = "w"
	ModelTypeQuarter = "q"
)

// BuildSequential - runs model building sequenitally for all project intervals.
func BuildSequential(projectId uint64, configs map[string]interface{}) (map[string]interface{}, bool) {

	env := configs["env"].(string)
	db := configs["db"].(*gorm.DB)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	etcdClient := configs["etcdClient"].(*serviceEtcd.EtcdClient)
	diskManger := configs["diskManger"].(*serviceDisk.DiskDriver)
	bucketName := configs["bucketName"].(string)
	noOfPatternWorkers := configs["noOfPatternWorkers"].(int)
	projectIdsToSkip := configs["projectIdsToSkip"].(map[uint64]bool)
	maxModelSize := configs["maxModelSize"].(int64)
	modelType := configs["modelType"].(string)
	countOccurence := configs["countOccurence"].(bool)
	numCampaignsLimit := configs["numCampaignsLimit"].(int)
	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	beamConfig := configs["beamConfig"].(*RunBeamConfig)
	countsVersion := configs["countsVersion"].(int)
	hmineSupport := configs["hmineSupport"].(float32)
	hmine_persist := configs["hminePersist"].(int)

	createMetadata := configs["create_metadata"].(bool)
	status := make(map[string]interface{})
	defer util.NotifyOnPanic(taskID, env)

	if _, ok := projectIdsToSkip[projectId]; ok {
		bsLog.WithField("ProjectId", projectId).Info("Skipping build for the project.")
		status["error"] = "Skipping build for the project."
		return status, false
	}

	logCtx := bsLog.WithFields(log.Fields{"ProjectId": projectId,
		"StartTime": startTimestamp, "EndTime": endTimestamp,
		"UnitType": modelType})

	// Prefix timestamp with randomAlphanumeric(5).
	curTimeInMilliSecs := time.Now().UnixNano() / 1000000
	// modelId = time in millisecs + random number upto 3 digits.
	modelId := uint64(curTimeInMilliSecs + rand.Int63n(999))

	// Patten mine
	startAt := time.Now().UnixNano()

	var count_algo_props P.CountAlgoProperties
	count_algo_props.Counting_version = countsVersion
	count_algo_props.Hmine_persist = hmine_persist
	count_algo_props.Hmine_support = hmineSupport

	numChunks, err := PatternMine(db, etcdClient, cloudManager, diskManger,
		bucketName, noOfPatternWorkers, projectId, modelId, modelType,
		startTimestamp, endTimestamp, maxModelSize, countOccurence, numCampaignsLimit,
		beamConfig, createMetadata, count_algo_props)
	if err != nil {
		logCtx.WithError(err).Error("Failed to mine patterns.")
		status["error"] = "Failed to mine patterns."
		return status, false
	}
	timeTakenToMinePatterns := (time.Now().UnixNano() - startAt) / 1000000
	logCtx = logCtx.WithField("TimeTakenToMinePatternsInMS", timeTakenToMinePatterns)

	status["modelId"] = modelId
	status["numChunks"] = numChunks
	status["TimeTakenToMinePatternsInMS"] = timeTakenToMinePatterns
	return status, true
}
