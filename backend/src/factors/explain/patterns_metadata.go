package explain

import (
	"encoding/json"
	C "factors/config"
	"factors/model/store"
	T "factors/task"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"
)

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
		} else {
			return nil, err
		}
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return nil, err
	}

	var metadataObj []T.ChunkMetaData
	err = json.Unmarshal(data, &metadataObj)
	if err != nil {
		log.WithError(err).Error("Error unmarshalling response")
		return nil, err
	}
	return metadataObj, nil
}
