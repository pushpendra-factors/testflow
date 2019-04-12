package main

import (
	"encoding/json"
	"errors"
	"factors/filestore"
	"factors/pattern"
	"factors/pattern_model_meta"
	"factors/pattern_server/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", "development", "")
	bucketNameV1 := flag.String("bucket_name_v1", "/usr/local/var/factors/cloud_storage", "")
	bucketNameV2 := flag.String("bucket_name_v2", "/usr/local/var/factors/cloud_storage_v2", "")
	metadataVersion := flag.String("metadata_version", "", "Metadata version to read from")

	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	if *metadataVersion == "" {
		panic(errors.New("metadata version not given"))
	}

	var cloudManagerV1 filestore.FileManager
	var err error
	if *envFlag == "development" {
		cloudManagerV1 = serviceDisk.New(*bucketNameV1)
	} else {
		cloudManagerV1, err = serviceGCS.New(*bucketNameV1)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client V1")
			panic(err)
		}
	}

	var cloudManagerV2 filestore.FileManager
	if *envFlag == "development" {
		cloudManagerV2 = serviceDisk.New(*bucketNameV2)
	} else {
		cloudManagerV2, err = serviceGCS.New(*bucketNameV2)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client V2")
			panic(err)
		}
	}

	path, name := cloudManagerV1.GetProjectsDataFilePathAndName(*metadataVersion)
	log.WithFields(log.Fields{"path": path,
		"name": name}).Info("Getting project model meta using cloud manager.")

	versionFile, err := cloudManagerV1.Get(path, name)
	if err != nil {
		log.WithFields(log.Fields{"path": path, "name": name,
			"err": err}).Error("Failed to read current version file.")
		panic(err)
	}

	projectMetadata, err := pattern_model_meta.ParseProjectsDataFile(versionFile)
	if err != nil {
		log.Error("Parsing project metadata failed.")
		panic(err)
	}

	log.Infof("No of metadata available to process : %d", len(projectMetadata))

	projectMetadataV2 := make([]pattern_model_meta.ProjectData, 0, 0)

	for _, pm := range projectMetadata {
		// New model id. Using Microseconds as some model gets
		// finished within seconds.
		modelID := uint64(time.Now().UnixNano()) / 1000
		logCtx := log.WithFields(log.Fields{"project_id": pm.ID, "model_id": pm.ModelID})

		for _, chunkId := range pm.Chunks {
			logCtx := log.WithFields(log.Fields{"project_id": pm.ID, "model_id": pm.ModelID, "chunkId": chunkId})
			path, name := cloudManagerV1.GetPatternChunkFilePathAndName(pm.ID, pm.ModelID, chunkId)
			creader, err := cloudManagerV1.Get(path, name)
			if err != nil {
				logCtx.Error("Failed reading chunk")
				creader.Close()
				panic(err)
			}

			patternsWithMeta := make([]*store.PatternWithMeta, 0, 0)

			cscanner := store.CreateScannerFromReader(creader)
			for cscanner.Scan() {
				line := cscanner.Bytes()
				var pattern pattern.Pattern
				if err := json.Unmarshal(line, &pattern); err != nil {
					logCtx.Error("Failed unmarshalling pattern")
					panic(err)
				}

				var pwm store.PatternWithMeta
				pwm.PatternEvents = pattern.EventNames
				pwm.RawPattern = json.RawMessage(string(line))
				patternsWithMeta = append(patternsWithMeta, &pwm)
			}

			creader.Close()

			pwmReader, err := store.CreateReaderFromPatternsWithMeta(patternsWithMeta)
			if err != nil {
				logCtx.Error("Failed creating reader from patterns with meta")
				panic(err)
			}

			pwmPath, pwmName := cloudManagerV2.GetPatternChunkFilePathAndName(pm.ID, modelID, chunkId)
			err = cloudManagerV2.Create(pwmPath, pwmName, pwmReader)
			if err != nil {
				logCtx.Error("Failed writing pattern with meta to cloud")
				panic(err)
			}
		}

		log.WithFields(log.Fields{"project_id": pm.ID, "model_id": modelID, "no_of_chunks": len(pm.Chunks)}).Info("Created all model_v2 chunks.")

		// build metadata for transformed models.
		pmdv2 := pattern_model_meta.ProjectData{ID: pm.ID, ModelType: pm.ModelType, ModelID: modelID, StartTimestamp: pm.StartTimestamp, EndTimestamp: pm.EndTimestamp, Chunks: pm.Chunks}
		projectMetadataV2 = append(projectMetadataV2, pmdv2)

		// copy event info file.
		eiPath, eiName := cloudManagerV1.GetModelEventInfoFilePathAndName(pm.ID, pm.ModelID)
		eiReader, err := cloudManagerV1.Get(eiPath, eiName)
		if err != nil {
			eiReader.Close()
			panic(err)
		}
		eiPathV2, eiNameV2 := cloudManagerV2.GetModelEventInfoFilePathAndName(pm.ID, modelID)
		cloudManagerV2.Create(eiPathV2, eiNameV2, eiReader)
		eiReader.Close()
		logCtx.Info("Copied event info file.")

		// copy events file.
		ePath, eName := cloudManagerV1.GetModelEventsFilePathAndName(pm.ID, pm.ModelID)
		eReader, err := cloudManagerV1.Get(ePath, eName)
		if err != nil {
			eiReader.Close()
			panic(err)
		}
		ePathV2, eNameV2 := cloudManagerV2.GetModelEventsFilePathAndName(pm.ID, modelID)
		cloudManagerV2.Create(ePathV2, eNameV2, eReader)
		eReader.Close()
		logCtx.Info("Copied events file.")
	}

	err = pattern_model_meta.WriteProjectDataFile("metadata_v2", projectMetadataV2, &cloudManagerV2)
	if err != nil {
		log.Error("Failed writing project data file for transformed model version 2")
	}

	log.Info("Generated metadata v2: ")
	for _, meta := range projectMetadataV2 {
		jsonMeta, err := json.Marshal(meta)
		if err != nil {
			log.Error("Failed marshalling metadata for model v2")
			panic(err)
		}
		fmt.Println(string(jsonMeta))
	}
}
