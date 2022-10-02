package task

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/apache/beam/sdks/go/pkg/beam"
	beamlog "github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"
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
	ModelType  string `json:"mt"`
}
type UidMap struct {
	Userid string `json:"uid"`
	Index  int64  `json:"idx"`
}

func PullEventsDataDaily(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64,
	modelType string, beamConfig *RunBeamConfig, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Info("Pulling events.")
	per_day_epoch := int64(86400)
	start_time := startTimestampInProjectTimezone
	userIdsMap := make(map[string]int)
	fPath, fName := diskManager.GetModelEventsUnsortedFilePathAndName(projectId, startTimestamp, modelType)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := filepath.Join(fPath, fName)
	startAt := time.Now().UnixNano()

	file_pointer, err := os.Create(tmpEventsFile)
	if err != nil {
		peLog.WithFields(log.Fields{"file": tmpEventsFile, "err": err}).Error("Unable to create file.")
		return err, false
	}

	total_events_count := 0
	total_time := 0
	index := 1
	for start_time <= endTimestampInProjectTimezone {

		end_time_per_day := start_time + per_day_epoch
		peLog.WithFields(log.Fields{"start_time": start_time, "end_time": end_time_per_day, "index": index}).Infof("Pulling events in beam ")

		eventsCount, _, userIdMapPart, err := pullEventsDaily(projectId, start_time, end_time_per_day, tmpEventsFile, file_pointer)
		if err != nil {
			logCtx.WithField("error", err).Error("Pull events failed. Pull and write events failed.")
			status["events-error"] = err.Error()
			return err, false
		}
		timeTakenToPullEvents := (time.Now().UnixNano() - startAt) / 1000000
		total_time += int(timeTakenToPullEvents)
		// Zero events. Returns eventCount as 0.
		if eventsCount == 0 {
			logCtx.Info("No events found.")
			status["events-info"] = "No events found."
			return nil, true
		}

		start_time = end_time_per_day //rewire end time to new start time
		index = index + 1
		for uid, _ := range userIdMapPart {
			userIdsMap[uid] = 1
		}
		total_events_count += eventsCount
		logCtx.Infof("total_events:%d,timeTakenToPullEvents:%d", total_events_count, total_time)

	}
	err = file_pointer.Close()
	if err != nil {
		return fmt.Errorf("unable to close file"), false
	}

	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Write to tmp failed.")
		status["events-error"] = "Failed to pull events. Write to tmp failed."
		return err, false
	}

	cDir, cName := (*cloudManager).GetModelEventsUnsortedFilePathAndName(projectId, startTimestamp, modelType)
	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Upload failed.")
		status["events-error"] = "Failed to pull events. Upload failed."
		return err, false
	}

	//remove downloaded file
	err = os.Remove(tmpEventsFile)
	if err != nil {
		logCtx.Errorf("unable to delete file")
		return err, false
	}
	logCtx.Infof("deleted events file created:%s", tmpEventsFile)

	batch_size := beamConfig.NumWorker
	numsortedEvents, err := pull_events_on_beam(projectId, fPath, cDir, cName, batch_size, userIdsMap,
		beamConfig, cloudManager, diskManager, startTimestamp, modelType)
	if err != nil {
		return err, false
	}

	logCtx.WithFields(log.Fields{
		"EventsCount":           total_events_count,
		"EventsFilePath":        tmpEventsFile,
		"TimeTakenToPullEvents": total_time,
	}).Info("Successfully pulled events and written to file.")

	status["EventsCount"] = total_events_count
	status["SortedEventsCount"] = numsortedEvents
	return nil, true
}

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

func putUserIdsToGCP(projectId int64, userIds map[string]int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	startTime int64, modelType string) error {

	cDir, cName := (diskManager).GetEventsArtificatFilePathAndName(projectId, startTime, modelType)
	path := filepath.Join(cDir, cName)
	err := (diskManager).Create(cDir, cName, bytes.NewReader([]byte("")))
	if err != nil {
		peLog.Errorf("unable to create file %s :%v", path, err)
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	for userId, index := range userIds {
		uidStruct := UidMap{Userid: userId, Index: index}
		ud, err := json.Marshal(uidStruct)
		if err != nil {
			return err
		}
		_, err = file.WriteString(fmt.Sprintf("%s\n", ud))
		if err != nil {
			return err
		}
	}
	err = file.Close()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("unable to open file : %s :%v", path, err)
	}
	r := bufio.NewReader(f)
	gDir, gName := (*cloudManager).GetEventsArtificatFilePathAndName(projectId, startTime, modelType)
	err = (*cloudManager).Create(gDir, gName, r)
	if err != nil {
		return err
	}

	return nil
}

func getUserIdsFile(ctx context.Context, projectId int64, cloudManager *filestore.FileManager,
	startTime int64, modelType string, low, high int64) ([]string, error) {

	beamlog.Infof(ctx, "getting users file :%d,%d,%d", projectId, low, high)
	gDir, gName := (*cloudManager).GetEventsArtificatFilePathAndName(projectId, startTime, modelType)
	r, err := (*cloudManager).Get(gDir, gName)
	if err != nil {
		return nil, err
	}
	userIdList := make([]string, 0)

	scannerEvents := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scannerEvents.Buffer(buf, 1024*1024)

	for scannerEvents.Scan() {
		line := scannerEvents.Text()
		var ud UidMap
		if err := json.Unmarshal([]byte(line), &ud); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return nil, err
		}

		if (ud.Index >= low) && (ud.Index <= high) {
			userIdList = append(userIdList, ud.Userid)
		}

	}
	beamlog.Infof(ctx, "processing users :%d,%d", projectId, len(userIdList))

	return userIdList, nil

}

func createEventsFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		peLog.Errorf("unable to create file %s :%v", path, err)
		return err
	}
	defer file.Close()
	return nil
}

func pull_events_beam_controller(projectId int64, local_dir string, cloud_dir, cloud_file string,
	batchSize int, userIdMap map[string]int, beamStruct *RunBeamConfig,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTime int64, modelType string) error {

	peLog.Infof(" Sorting pull events in beam Project Id :%d", projectId)
	UIdsBeamString := make([]string, 0)
	userIdsList := make([]string, 0)
	userIdMapList := make(map[string]int64)

	bucketName := (*cloudManager).GetBucketName()
	cloudEventFileName := path.Join(cloud_dir, cloud_file)
	peLog.Infof("bucket name %s cloud name:%s", bucketName, cloudEventFileName)

	// local_dir, _ = diskManager.GetModelEventsFilePathAndName(projectId, startTime, modelType)

	//create events_sorted
	events_sorted := path.Join(local_dir, "events_sorted.txt")
	err := createEventsFile(events_sorted)
	if err != nil {
		peLog.Errorf("Unable to create events sorted file:%s", events_sorted)
	}
	peLog.Infof("create events sorted file in :%s", events_sorted)

	for userKey, _ := range userIdMap {
		userIdsList = append(userIdsList, userKey)
	}

	for idx, uid := range userIdsList {
		userIdMapList[uid] = int64(idx)
	}

	err = putUserIdsToGCP(projectId, userIdMapList, cloudManager, diskManager, startTime, modelType)
	if err != nil {
		return err
	}

	num_users := len(userIdsList)
	numRoutines := int(math.Ceil(float64(num_users) / float64(batchSize)))
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int(math.Min(float64(batchSize*i), float64(num_users)))
		high := int(math.Min(float64(batchSize*(i+1)), float64(num_users)))
		tmp := CUserIdsBeam{projectId, int64(low), int64(high), bucketName, "", beamStruct.Env, startTime, modelType}

		t, err := json.Marshal(tmp)
		if err != nil {
			return fmt.Errorf("unable to encode string : %v", err)
		}
		UIdsBeamString = append(UIdsBeamString, string(t))
	}

	err = SortUsersExecutor(beamStruct, UIdsBeamString)
	if err != nil {
		return err
	}
	return nil
}

func SortUsersExecutor(beamStruct *RunBeamConfig, cPatternsString []string) error {

	s := beamStruct.Scp
	s = s.Scope("pull_events")

	ctx := beamStruct.Ctx
	p := beamStruct.Pipe

	if beam.Initialized() {
		peLog.Info("Initialized beam")
	} else {
		peLog.Fatal("Unable to init beam")
	}

	if s.IsValid() {
		peLog.Info("Scope is Valid")
	} else {
		peLog.Fatal("Scope is not valid")
	}

	config := beamStruct.DriverConfig
	cPatternsPcol := beam.CreateList(s, cPatternsString)
	cPatternsPcolReshuffled := beam.Reshuffle(s, cPatternsPcol)
	beam.ParDo0(s, &SortUsDoFn{Config: config}, cPatternsPcolReshuffled)
	err := beamx.Run(ctx, p)
	if err != nil {
		peLog.Fatalf("unable to run beam pipeline :  %s", err)
		return err
	}
	return nil
}

type SortUsDoFn struct {
	Config *C.Configuration
}

func (f *SortUsDoFn) StartBundle(ctx context.Context) {
	beamlog.Info(ctx, "Initializing conf from sortDoFn")
	initConf(f.Config)
}

func (f *SortUsDoFn) FinishBundle(ctx context.Context) {
	beamlog.Info(ctx, "Closing DB Connection from FinishBundle sortDoFn")
	C.SafeFlushAllCollectors()
}

func (f *SortUsDoFn) ProcessElement(ctx context.Context, cpString string) error {
	beamlog.Info(ctx, "Initializing process ctx from sortDoFn")

	var up CUserIdsBeam
	err := json.Unmarshal([]byte(cpString), &up)
	if err != nil {
		return fmt.Errorf("unable to unmarshall string in processElement :%s", cpString)
	}

	envFlag := up.Env
	projectId := up.ProjectID
	lowNum := up.LowNum
	highNum := up.HighNum
	startTime := up.StartTime
	modelType := up.ModelType
	bucketName := up.BucketName
	beamlog.Infof(ctx, "Processing Project:%s,%d,%d", projectId, lowNum, highNum)

	// enable client for GCP
	var cloudManager filestore.FileManager

	if envFlag == "development" {
		cloudManager = serviceDisk.New(bucketName)
	} else {
		beamlog.Info(ctx, "Initializing gcs")
		cloudManager, err = serviceGCS.New(bucketName)
		if err != nil {
			beamlog.Errorf(ctx, "Failed to init New GCS Client:%v", err)
			panic(err)
		}
	}
	userIDs, err := getUserIdsFile(ctx, projectId, &cloudManager, startTime, modelType, lowNum, highNum)
	if err != nil {
		return err
	}

	err = downloadAndSortEventsFile(ctx, projectId, &cloudManager, startTime, modelType, userIDs)
	if err != nil {
		return err
	}
	return nil
}

func downloadAndSortEventsFile(ctx context.Context, projectId int64, cloudManager *filestore.FileManager, startTime int64, modelType string,
	userIds []string) error {

	beamlog.Infof(ctx, "Downloading and sorting events file")
	userIDmap := make(map[string]int)
	userIdEventsMap := make(map[string][]*P.CounterEventFormat)
	for _, uid := range userIds {
		userIDmap[uid] += 1
	}

	gcpPath, gcpFName := (*cloudManager).GetModelEventsUnsortedFilePathAndName(projectId, startTime, modelType)
	rc, err := (*cloudManager).Get(gcpPath, gcpFName)
	if err != nil {
		return fmt.Errorf("unable to get file from cloud :%v", err)
	}

	beamlog.Infof(ctx, "Reading events from :%s,%s", gcpPath, gcpFName)
	scannerEvents := bufio.NewScanner(rc)
	buf := make([]byte, 0, 64*1024)
	scannerEvents.Buffer(buf, 1024*1024)

	for scannerEvents.Scan() {
		line := scannerEvents.Text()
		var event *P.CounterEventFormat
		var uid string
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return err
		}

		uid = event.UserId

		if _, ok := userIDmap[uid]; ok {
			userIdEventsMap[uid] = append(userIdEventsMap[uid], event)
		}
	}

	err = rc.Close()
	if err != nil {
		beamlog.Errorf(ctx, "unable to close events file from gcp :%v", err)
	}

	eventsTmpFile, err := ioutil.TempFile("", "events_")
	if err != nil {
		return fmt.Errorf("unable to create tmp file :%v", err)
	} else {
		beamlog.Infof(ctx, "events file created :%s", eventsTmpFile.Name())
	}

	err = sortEventsWithUserIds(ctx, userIdEventsMap, eventsTmpFile)
	if err != nil {
		return err
	}

	err = eventsTmpFile.Close()
	if err != nil {
		beamlog.Error(ctx, "unable to close sorted events file")
	}
	err = writeSorteddEventFileToGCP(ctx, projectId, startTime, modelType, eventsTmpFile.Name(), cloudManager)
	if err != nil {
		beamlog.Error(ctx, "Failed to write sorted events to gcp:%v", err)

	}
	err = deleteFile(eventsTmpFile.Name())
	if err != nil {
		return err
	}
	beamlog.Infof(ctx, "Deleted tmp events File :%s", eventsTmpFile.Name())

	return nil
}

func writeSorteddEventFileToGCP(ctx context.Context, projectId int64, timestamp int64, modelType string, tmpEventsFile string,
	cloudManager *filestore.FileManager) error {

	beamlog.Infof(ctx, "reading and uploading from :%v", tmpEventsFile)
	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		beamlog.Error(ctx, "Failed to pull events. Write to tmp failed : %v", err)
		return err
	}
	defer tmpOutputFile.Close()

	cDir, _ := (*cloudManager).GetModelEventsFilePathAndName(projectId, timestamp, modelType)
	cDir = path.Join(cDir, "parts")
	lastName := strings.Split(tmpOutputFile.Name(), "/")

	if len(lastName) > 0 {
		cName := lastName[len(lastName)-1]
		err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
		if err != nil {
			beamlog.Error(ctx, "Failed to pull events. Upload failed.")
			return err
		}
	}
	return nil
}

func ReadAndMergeEventPartFiles(projectId int64, cloudDir string,
	local_dir string, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver) (int64, error) {
	var countLines int64
	local_sorted_file := path.Join(local_dir, "events.txt")
	eventsTmpFile, err := os.OpenFile(local_sorted_file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		peLog.Errorf("unable to open local sorted file:%s", local_sorted_file)
		return 0, err
	}

	peLog.Infof("Merging partFiles:%d", projectId)
	partFilesDir := path.Join(cloudDir, "parts")
	listFiles := (*cloudManager).ListFiles(partFilesDir)
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]

		peLog.Infof("Reading part file :%s, %s", partFilesDir, partFileName)
		file, err := (*cloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			return 0, err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 30 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			line := scanner.Text()
			var rawEvent *P.CounterEventFormat
			var cEvent P.CounterEventFormat
			err = json.Unmarshal([]byte(line), &rawEvent)
			if err != nil {
				return 0, err
			}

			cEvent.UserId = rawEvent.UserId
			cEvent.UserJoinTimestamp = rawEvent.EventTimestamp
			cEvent.EventName = rawEvent.EventName
			cEvent.EventTimestamp = rawEvent.EventTimestamp
			cEvent.EventCardinality = rawEvent.EventCardinality
			cEvent.EventProperties = rawEvent.EventProperties
			cEvent.UserProperties = rawEvent.UserProperties

			eventBytes, err := json.Marshal(cEvent)
			if err != nil {
				return 0, fmt.Errorf("unable to unmarshall userEventsInfo :%v", err)
			}
			_, err = eventsTmpFile.WriteString(string(eventBytes) + "\n")
			if err != nil {
				peLog.Errorf("Unable to write to file :%v", err)
				return 0, err
			}
			countLines += 1
		}
	}
	defer eventsTmpFile.Close()

	return countLines, nil

}
func sortEventsWithUserIds(ctx context.Context, userIdMap map[string][]*P.CounterEventFormat, eventsTmpFile *os.File) error {

	beamlog.Debugf(ctx, "sort events :%d", len(userIdMap))

	for _, eventList := range userIdMap {

		sort.Slice(eventList[:], func(i, j int) bool {
			return eventList[i].EventTimestamp < eventList[j].EventTimestamp
		})

		for _, el := range eventList {

			eventBytes, err := json.Marshal(el)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to marshal event.")
				return err
			}
			eString := string(eventBytes)
			eString = eString + "\n"
			_, err = eventsTmpFile.WriteString(eString)
			if err != nil {
				peLog.Errorf("unable to write event to tmp file :%v ", eString)
				return err
			}

		}
	}

	return nil
}

func pull_events_on_beam(projectId int64, local_dir string, cloud_dir, cloud_file string,
	batchSize int, userIdMap map[string]int, beamStruct *RunBeamConfig,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	startTime int64, modelType string) (int64, error) {

	peLog.Infof("sorting on beam")
	local_events_dir := (diskManager).GetProjectEventFileDir(projectId, startTime, modelType)

	cloud_events_dir, cloud_name := (*cloudManager).GetModelEventsFilePathAndName(projectId, startTime, modelType)

	err := pull_events_beam_controller(projectId, local_events_dir, cloud_events_dir, cloud_name, batchSize, userIdMap, beamStruct, cloudManager, diskManager, startTime, modelType)
	if err != nil {
		return 0, err
	}

	numLines, err := ReadAndMergeEventPartFiles(projectId, cloud_events_dir, local_events_dir, cloudManager, diskManager)
	if err != nil {
		peLog.Errorf("unable to read part files and merge :%v", err)
		return 0, err
	}
	peLog.Infof("num of events sorted and written to file :%d", numLines)
	file, err := (diskManager).Get(local_events_dir, "events.txt")
	if err != nil {
		peLog.Errorf("unable to get sorted events files :%v", err)
		return 0, err
	}
	err = (*cloudManager).Create(cloud_events_dir, "events.txt", file)
	if err != nil {
		peLog.Errorf("unable to create sorted events files to gcp:%v", err)
		return 0, err
	}

	eventsFile := filepath.Join(local_events_dir, "events.txt")
	//remove downloaded file
	err = os.Remove(eventsFile)
	if err != nil {
		peLog.Errorf("unable to delete file")
		return 0, err
	}
	peLog.Infof("deleted events file created:%s", eventsFile)

	peLog.Infof("Done sorting events and writing it to file")
	return numLines, nil
}
