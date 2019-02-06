package pattern_model_meta

import (
	"bufio"
	"bytes"
	"encoding/json"
	encjson "encoding/json"
	"factors/filestore"
	serviceEtcd "factors/services/etcd"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type ProjectData struct {
	ID             uint64   `json:"pid"`
	ModelID        uint64   `json:"mid"`
	ModelType      string   `json:"mt"`
	StartTimestamp int64    `json:"st"`
	EndTimestamp   int64    `json:"et"`
	Chunks         []string `json:"cs"`
}

// GetProjectsMetadata - Get project data for current version of meta.
func GetProjectsMetadata(cloudManager *filestore.FileManager,
	etcdClient *serviceEtcd.EtcdClient) ([]ProjectData, error) {

	curVersion, err := etcdClient.GetProjectVersion()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to fetch current project version.")
		return nil, err
	}

	path, name := (*cloudManager).GetProjectsDataFilePathAndName(curVersion)
	log.WithFields(log.Fields{"path": path,
		"name": name}).Info("Getting project model meta using cloud manager.")

	versionFile, err := (*cloudManager).Get(path, name)
	if err != nil {
		log.WithFields(log.Fields{"path": path, "name": name,
			"err": err}).Error("Failed to read current version file.")
		return nil, err
	}

	projectMetadata, err := ParseProjectsDataFile(versionFile)
	if err != nil {
		return nil, err
	}

	return projectMetadata, nil
}

func ParseProjectsDataFile(reader io.Reader) ([]ProjectData, error) {
	scanner := bufio.NewScanner(reader)
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	projectDatas := make([]ProjectData, 0, 0)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var p ProjectData
		if err := encjson.Unmarshal([]byte(line), &p); err != nil {
			log.WithFields(log.Fields{"line_num": lineNum,
				"err": err}).Error("Unmarshal error. Failed to ParseProjectsDataFile.")
			continue
		}
		projectDatas = append(projectDatas, p)
	}
	err := scanner.Err()
	if err != nil {
		log.WithError(err).Errorln("Scanner error. Failed to ParseProjectsDataFile.")
		return projectDatas, err
	}

	return projectDatas, nil
}

func WriteProjectDataFile(newVersionName string, projectDatas []ProjectData,
	cloudManager *filestore.FileManager) error {

	// Todo(Dinesh): Add distributed lock.
	path, name := (*cloudManager).GetProjectsDataFilePathAndName(newVersionName)

	buffer := bytes.NewBuffer(nil)
	for _, pD := range projectDatas {
		bytes, err := json.Marshal(pD)
		if err != nil {
			return err
		}
		_, err = buffer.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		if err != nil {
			return err
		}
	}
	err := (*cloudManager).Create(path, name, bytes.NewReader(buffer.Bytes()))

	return err
}
