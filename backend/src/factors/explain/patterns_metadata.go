package explain

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/model/store"
	T "factors/task"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
)

var eventsKeyMap map[string]bool


func GetChunksMetaData(projectId, modelId uint64) (metadata []T.ChunkMetaData, errmsg error) {
	path, name := C.GetConfig().CloudManager.GetChunksMetaDataFilePathAndName(projectId, modelId)
	fmt.Println(path)
	reader, err := C.GetCloudManager().Get(path, name)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error retrieving file from cloud")
		log.Info("trying to fetch for latest modelID")
		modelMetadata, errCode, msg := store.GetStore().GetProjectModelMetadata(projectId)
		if errCode != http.StatusFound {
			log.Error(msg, projectId)
			return nil, err
		}
		sort.Slice(modelMetadata, func(i, j int) bool {
			return modelMetadata[i].ModelId > modelMetadata[j].ModelId
		})
		if len(modelMetadata) > 0 {
			latestModelID := modelMetadata[0].ModelId
			path, name := C.GetConfig().CloudManager.GetChunksMetaDataFilePathAndName(projectId, latestModelID)
			fmt.Println(path)
			reader, err = C.GetCloudManager().Get(path, name)
			if err != nil {
				log.WithError(err).Error("Error retreiving file from cloud")
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	eventsKeyMap = make(map[string]bool)
	scanner := bufio.NewScanner(reader)
	Metadata := []T.ChunkMetaData{}
	for scanner.Scan() {
		var metadataObj T.ChunkMetaData
		data := scanner.Bytes()
		err = json.Unmarshal(data, &metadataObj)
		if err != nil {
			log.WithError(err).Error("Error unmarshalling response")
			return nil, err
		}
		Metadata = MergeMetaData(Metadata, metadataObj)
	}
	response := DedupProperties(Metadata)
	return response, nil
}
func MergeMetaData(result []T.ChunkMetaData, new T.ChunkMetaData) []T.ChunkMetaData {
	if len(result) == 0 {
		result = append(result, new)
		return result
	}
	// for events
	for _, event := range new.Events {
		if _, exists := eventsKeyMap[event]; !exists {
			eventsKeyMap[event] = true
			result[0].Events = append(result[0].Events, event)
		}
	}
	// for properties
	result[0].UserProperties.NumericalProperties = append(result[0].UserProperties.NumericalProperties, new.UserProperties.NumericalProperties...)
	for key, value := range new.UserProperties.CategoricalProperties {
		result[0].UserProperties.CategoricalProperties[key] = append(result[0].UserProperties.CategoricalProperties[key], value...)
	}
	result[0].EventProperties.NumericalProperties = append(result[0].EventProperties.NumericalProperties, new.EventProperties.NumericalProperties...)
	for key, value := range new.EventProperties.CategoricalProperties {
		result[0].EventProperties.CategoricalProperties[key] = append(result[0].EventProperties.CategoricalProperties[key], value...)
	}
	return result
}
func DedupProperties(result []T.ChunkMetaData) []T.ChunkMetaData {
	metadataNew := T.ChunkMetaData{}
	metadata := result[0]
	metadataNew.Events = metadata.Events

	res := DedupArray(metadata.EventProperties.NumericalProperties)
	metadataNew.EventProperties.NumericalProperties = res

	metadataNew.EventProperties.CategoricalProperties = make(map[string][]string)
	for key, val := range metadata.EventProperties.CategoricalProperties {
		res := DedupArray(val)
		metadataNew.EventProperties.CategoricalProperties[key] = res
	}

	res = DedupArray(metadata.UserProperties.NumericalProperties)
	metadataNew.UserProperties.NumericalProperties = res

	metadataNew.UserProperties.CategoricalProperties = make(map[string][]string)
	for key, val := range metadata.UserProperties.CategoricalProperties {
		res := DedupArray(val)
		metadataNew.UserProperties.CategoricalProperties[key] = res
	}
	return []T.ChunkMetaData{metadataNew}
}

func DedupArray(input []string) []string {
	allKeys := make(map[string]bool)
	values := []string{}
	for _, item := range input {
		if _, exists := allKeys[item]; !exists {
			allKeys[item] = true
			values = append(values, item)
		}
	}
	return values
}
