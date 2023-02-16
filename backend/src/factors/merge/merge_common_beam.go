package merge

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"os"
	"path"
	"strings"

	"github.com/apache/beam/sdks/go/pkg/beam"
	beamlog "github.com/apache/beam/sdks/go/pkg/beam/log"
	log "github.com/sirupsen/logrus"
)

type CUserIdsBeam struct {
	ProjectID  int64  `json:"pid"`
	LowNum     int64  `json:"ln"`
	HighNum    int64  `json:"hn"`
	BucketName string `json:"bnm"`
	ScopeName  string `json:"ScopeName"`
	Env        string `json:"Env"`
	StartTime  int64  `json:"ts"`
	EndTime    int64  `json:"end"`
	Channel    string `json:"ch"`
	Group      int    `json:"gr"`
}

type RunBeamConfig struct {
	RunOnBeam    bool             `json:"runonbeam"`
	Ctx          context.Context  `json:"ctx"`
	Scp          beam.Scope       `json:"scp"`
	Pipe         *beam.Pipeline   `json:"pipeLine"`
	Env          string           `json:"Env"`
	DriverConfig *C.Configuration `json:"driverconfig"`
	NumWorker    int              `json:"nw"`
}

type UidMap struct {
	Userid string `json:"uid"`
	Index  int64  `json:"idx"`
}

// initialise configs for beam
func initConf(config *C.Configuration) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.Infof("config :%v", config)
	C.InitConf(config)
	C.SetIsBeamPipeline()
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 5, 2)
	if err != nil {
		log.WithError(err).Panic("Failed to initalize db.")
	}
	C.InitRedisConnection(config.RedisHost, config.RedisPort, true, 20, 0)
	C.KillDBQueriesOnExit()
}

// upload sorted file from tmpEventsFile and upload on cloud
func writeSortedFileToGCP(ctx context.Context, projectId int64, startTime, endTime int64, tmpEventsFile string,
	cloudManager *filestore.FileManager, dataType, channelOrDatefield string, sortOnGroup int) error {

	beamlog.Infof(ctx, "reading and uploading from :%v", tmpEventsFile)
	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		beamlog.Error(ctx, "Failed to pull events. Write to tmp failed : %v", err)
		return err
	}
	defer tmpOutputFile.Close()

	cDir, cName := GetSortedFilePathAndName(*cloudManager, dataType, channelOrDatefield, projectId, startTime, endTime, sortOnGroup)
	cDir = path.Join(cDir, "parts/"+strings.Replace(cName, ".txt", "", 1))
	lastName := strings.Split(tmpOutputFile.Name(), "/")

	if len(lastName) > 0 {
		cName = lastName[len(lastName)-1]
		cNameArr := strings.Split(cName, "_")
		cName = strings.Join(cNameArr[:len(cNameArr)-1], "_")
		err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
		if err != nil {
			beamlog.Error(ctx, "Failed to pull events. Upload failed.")
			return err
		}
	}
	return nil
}
