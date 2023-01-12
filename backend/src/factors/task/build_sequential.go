package task

import (
	"factors/filestore"
	"factors/merge"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	"factors/util"
	U "factors/util"
	"math/rand"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const taskID = "Task#BuildSequential"

var bsLog = taskLog.WithField("prefix", taskID)

const (
	ModelTypeMonth   = "m"
	ModelTypeWeek    = "w"
	ModelTypeDay     = "d"
	ModelTypeQuarter = "q"
)

// BuildSequential - runs model building sequenitally for all project intervals.
func BuildSequential(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	env := configs["env"].(string)
	db := configs["db"].(*gorm.DB)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	sortedCloudManager := configs["sortedCloudManager"].(*filestore.FileManager)
	etcdClient := configs["etcdClient"].(*serviceEtcd.EtcdClient)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	noOfPatternWorkers := configs["noOfPatternWorkers"].(int)
	projectIdsToSkip := configs["projectIdsToSkip"].(map[int64]bool)
	maxModelSize := configs["maxModelSize"].(int64)
	modelType := configs["modelType"].(string)
	countOccurence := configs["countOccurence"].(bool)
	numCampaignsLimit := configs["numCampaignsLimit"].(int)
	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	beamConfig := configs["beamConfig"].(*merge.RunBeamConfig)
	countsVersion := configs["countsVersion"].(int)
	hmineSupport := configs["hmineSupport"].(float32)
	hmine_persist := configs["hminePersist"].(int)
	hardPull := configs["hardPull"].(bool)
	useBucketV2 := configs["useBucketV2"].(bool)

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

	numChunks, err := PatternMine(db, etcdClient, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, diskManager,
		noOfPatternWorkers, projectId, modelId, modelType,
		startTimestamp, endTimestamp, maxModelSize, countOccurence, numCampaignsLimit,
		beamConfig, createMetadata, count_algo_props, hardPull, useBucketV2)
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

// BuildSequential - runs model building sequenitally for all project intervals.
func BuildSequentialV2(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	env := configs["env"].(string)
	db := configs["db"].(*gorm.DB)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	tmpCloudManager := configs["tmpCloudManager"].(*filestore.FileManager)
	sortedCloudManager := configs["sortedCloudManager"].(*filestore.FileManager)

	// cloudManager := configs["cloudManager"].(*filestore.FileManager)
	etcdClient := configs["etcdClient"].(*serviceEtcd.EtcdClient)
	diskManger := configs["diskManger"].(*serviceDisk.DiskDriver)
	noOfPatternWorkers := configs["noOfPatternWorkers"].(int)
	projectIdsToSkip := configs["projectIdsToSkip"].(map[int64]bool)
	maxModelSize := configs["maxModelSize"].(int64)
	modelType := configs["modelType"].(string)
	countOccurence := configs["countOccurence"].(bool)
	numCampaignsLimit := configs["numCampaignsLimit"].(int)
	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	beamConfig := configs["beamConfig"].(*merge.RunBeamConfig)
	hmineSupport := configs["hmineSupport"].(float32)
	hmine_persist := configs["hminePersist"].(int)
	hardPull := configs["hardPull"].(bool)
	useBucketV2 := configs["useBucketV2"].(bool)

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
	queries, _ := store.GetStore().GetAllSavedExplainV2EntityByProject(projectId)
	for _, query := range queries {

		store.GetStore().UpdateExplainV2EntityStatus(projectId, query.ID, model.BUILDING, 0)
		var actualQuery model.ExplainV2Query
		U.DecodePostgresJsonbToStructType(query.ExplainV2Query, &actualQuery)
		mineLog.Infof("Actual query :%v", actualQuery)

		var count_algo_props P.CountAlgoProperties
		count_algo_props.Counting_version = 4
		count_algo_props.Hmine_persist = hmine_persist
		count_algo_props.Hmine_support = hmineSupport

		var jb model.ExplainV2Query
		err := U.DecodePostgresJsonbToStructType(query.ExplainV2Query, &jb)
		if err != nil {
			status["error"] = "unable to decodequery"
			log.Panic("unable to decodequery")
			return status, false
		}
		mineLog.Info("Job to execute 1 :%v", jb)

		if jb.Query.StartEvent == "" && jb.Query.EndEvent == "" {
			status["error"] = "unable to run explain v2 as both start and events are empty"
			log.Panic("unable to run explain v2 as both start and end events are empty")
			return status, false
		}
		startTimestamp = jb.StartTimestamp
		endTimestamp = jb.EndTimestamp
		count_algo_props.Job = jb
		count_algo_props.JobId = query.ID
		mineLog.Info("Job to execute: %v", jb)
		mineLog.Info("Job to execute start timestamp: %d", startTimestamp)
		mineLog.Info("Job to execute end timestamp : %d", endTimestamp)
		mineLog.Info("Job to execute start evnet : %s", jb.Query.StartEvent)
		mineLog.Info("Job to execute end event : %s", jb.Query.EndEvent)

		// numChunks, err := PatternMine(db, etcdClient, cloudManager, cloudManager, cloudManager, cloudManager, diskManger,
		// 	noOfPatternWorkers, projectId, modelId, modelType,
		// 	startTimestamp, endTimestamp, maxModelSize, countOccurence, numCampaignsLimit,
		// 	beamConfig, createMetadata, count_algo_props, hardPull, useBucketV2)

		numChunks, err := PatternMine(db, etcdClient, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager, diskManger,
			noOfPatternWorkers, projectId, modelId, modelType,
			startTimestamp, endTimestamp, maxModelSize, countOccurence, numCampaignsLimit,
			beamConfig, createMetadata, count_algo_props, hardPull, useBucketV2)
		if err != nil {
			logCtx.WithError(err).Error("Failed to mine patterns.")
			status["error"] = "Failed to mine patterns."
			return status, false
		}
		timeTakenToMinePatterns := (time.Now().UnixNano() - startAt) / 1000000
		logCtx = logCtx.WithField("TimeTakenToMinePatternsInMS", timeTakenToMinePatterns)

		store.GetStore().UpdateExplainV2EntityStatus(projectId, query.ID, model.ACTIVE, modelId)
		status["modelId"] = modelId
		status["numChunks"] = numChunks
		status["TimeTakenToMinePatternsInMS"] = timeTakenToMinePatterns

	}
	return status, true
}
