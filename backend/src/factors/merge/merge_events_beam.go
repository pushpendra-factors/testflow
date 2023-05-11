package merge

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
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/apache/beam/sdks/go/pkg/beam"
	beamlog "github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"
	log "github.com/sirupsen/logrus"
)

// create local file, write userIds map to it and upload to cloud
func putUserIdsToGCP(projectId int64, userIds map[string]int64, tmpCloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	startTime, endTime int64, sortOnGroup int) error {

	cDir, cName := (diskManager).GetEventsArtifactFilePathAndName(projectId, startTime, endTime, sortOnGroup)
	path := filepath.Join(cDir, cName)
	err := (diskManager).Create(cDir, cName, bytes.NewReader([]byte("")))
	if err != nil {
		log.Errorf("unable to create file %s :%v", path, err)
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
	gDir, gName := (*tmpCloudManager).GetEventsArtifactFilePathAndName(projectId, startTime, endTime, sortOnGroup)
	err = (*tmpCloudManager).Create(gDir, gName, r)
	if err != nil {
		return err
	}

	return nil
}

// get userIds pertaining to index between high and low from userIds file
func getUserIdsFromFile(ctx context.Context, projectId int64, cloudManager *filestore.FileManager,
	startTime, endTime int64, low, high int64, sortOnGroup int) ([]string, error) {

	beamlog.Infof(ctx, "getting users file :%d,%d,%d", projectId, low, high)
	gDir, gName := (*cloudManager).GetEventsArtifactFilePathAndName(projectId, startTime, endTime, sortOnGroup)
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

// read unsorted events file from cloud, pass some uids to each worker, workers sort events for the alloted uids and upload sorted part files
func Pull_events_beam_controller(projectId int64,
	startIndexToFileInfoMap map[int][]*partFileInfo, beamStruct *RunBeamConfig,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTime, endTime int64, sortOnGroup int) error {

	log.Infof("Sorting pull events in beam Project Id :%d", projectId)
	UIdsBeamString := make([]string, 0)

	bucketName := (*cloudManager).GetBucketName()

	for low, fileInfoList := range startIndexToFileInfoMap {
		for _, fileInfo := range fileInfoList {
			// Each worker gets a slice of patterns to count.
			tmp := CUserIdsBeam{projectId, int64(low), int64(fileInfo.endUserIndex), fileInfo.fileTimestampIndex, bucketName, "", beamStruct.Env, startTime, endTime, "", sortOnGroup}

			t, err := json.Marshal(tmp)
			if err != nil {
				return fmt.Errorf("unable to encode string : %v", err)
			}
			UIdsBeamString = append(UIdsBeamString, string(t))
		}
	}

	err := SortUsersExecutor(beamStruct, UIdsBeamString)
	if err != nil {
		return err
	}
	return nil
}

func SortUsersExecutor(beamStruct *RunBeamConfig, cPatternsString []string) error {

	s := beamStruct.Scp
	s = s.Scope("merge_events")

	ctx := beamStruct.Ctx
	p := beamStruct.Pipe

	if beam.Initialized() {
		log.Info("Initialized beam")
	} else {
		log.Fatal("Unable to init beam")
	}

	if s.IsValid() {
		log.Info("Scope is Valid")
	} else {
		log.Fatal("Scope is not valid")
	}
	tmp, _ := json.Marshal(beamStruct)
	log.Infof("beam struct: %s", string(tmp))

	config := beamStruct.DriverConfig
	cPatternsPcol := beam.CreateList(s, cPatternsString)
	cPatternsPcolReshuffled := beam.Reshuffle(s, cPatternsPcol)
	beam.ParDo0(s, &SortUsDoFn{Config: config}, cPatternsPcolReshuffled)
	err := beamx.Run(ctx, p)
	if err != nil {
		log.Errorf("unable to run beam pipeline :  %s", err)
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
	fileTimeIndex := up.FileTimeIndex
	startTime := up.StartTime
	endTime := up.EndTime
	bucketName := up.BucketName
	sortOnGroup := up.Group
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

	err = downloadAndSortEventsFile(ctx, projectId, &cloudManager, startTime, endTime, lowNum, highNum, sortOnGroup, fileTimeIndex)
	if err != nil {
		return err
	}
	return nil
}

// read unsorted file, get events with given uids, sort and upload sorted part file to cloud
func downloadAndSortEventsFile(ctx context.Context, projectId int64, cloudManager *filestore.FileManager, startTime, endTime int64,
	lowNum, highNum int64, sortOnGroup int, fileTimeIndex int) error {

	beamlog.Infof(ctx, "Downloading and sorting events file")
	userIdEventsMap := make(map[string][]*P.CounterEventFormat)

	startIndex := int(lowNum)
	endIndex := int(highNum)
	cDir, cName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTime, endTime, false, startIndex, endIndex, sortOnGroup, fileTimeIndex)
	rc, err := (*cloudManager).Get(cDir, cName)
	if err != nil {
		return fmt.Errorf("unable to get file from cloud :%v", err)
	}
	beamlog.Infof(ctx, "Reading events from :%s%s", cDir, cName)

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

		uid = getAptId(event, sortOnGroup)

		userIdEventsMap[uid] = append(userIdEventsMap[uid], event)
	}

	err = rc.Close()
	if err != nil {
		beamlog.Errorf(ctx, "unable to close events file from gcp :%v", err)
	}

	cDir, cName = (*cloudManager).GetEventsPartFilePathAndName(projectId, startTime, endTime, true, startIndex, endIndex, sortOnGroup, fileTimeIndex)

	cloudWriter, err := (*cloudManager).GetWriter(cDir, cName)
	if err != nil {
		beamlog.Error(ctx, "error getting cloud writer")
		return err
	}
	err = WriteSortedEventsWithUserIds(ctx, userIdEventsMap, &cloudWriter)
	if err != nil {
		log.WithError(err).Error("error sorting events and writing")
		return err
	}
	err = cloudWriter.Close()
	if err != nil {
		log.WithError(err).Error("error closing cloud writer")
		return err
	}
	return nil
}

// sort the list of events for each uid based on EventTimestamp (ascending)
func WriteSortedEventsWithUserIds(ctx context.Context, userIdEventsMap map[string][]*P.CounterEventFormat, cloudWriter *io.WriteCloser) error {

	beamlog.Debugf(ctx, "sort events :%d", len(userIdEventsMap))

	for _, eventList := range userIdEventsMap {

		sort.Slice(eventList[:], func(i, j int) bool {
			return eventList[i].EventTimestamp < eventList[j].EventTimestamp
		})

		for _, el := range eventList {

			eventBytes, err := json.Marshal(el)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to marshal event.")
				return err
			}
			eString := string(eventBytes)
			eString = eString + "\n"
			_, err = io.WriteString(*cloudWriter, eString)
			if err != nil {
				log.Errorf("unable to write event to tmp file :%v ", eString)
				return err
			}

		}
	}

	return nil
}

// merge all sorted part files to create one sorted file
func ReadAndMergeEventPartFiles(projectId int64, partsDir string, sortedDir string, sortedName string,
	tmpCloudManager, sortedCloudManager *filestore.FileManager) (int, error) {
	var countLines int
	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error getting gcp writer")
		return 0, err
	}

	log.Infof("Merging partFiles for project: %d", projectId)

	listFiles := (*tmpCloudManager).ListFiles(partsDir)
	filesInfo := make([]fileNameSortInfo, 0)
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		fileNameSplit := strings.Split(partFileName, "_")
		startAndEndUid := fileNameSplit[0]
		startAndEndUidSplit := strings.Split(startAndEndUid, "-")
		fileStartUid, _ := strconv.ParseInt(startAndEndUidSplit[0], 10, 64)
		fileEndUid, _ := strconv.ParseInt(startAndEndUidSplit[0], 10, 64)
		fileIndexStr := strings.Replace(fileNameSplit[len(fileNameSplit)-1], ".txt", "", 1)
		fileIndex, _ := strconv.ParseInt(fileIndexStr, 10, 64)
		filesInfo = append(filesInfo, fileNameSortInfo{fileName: partFileName, sortIndices: []int64{fileStartUid, fileEndUid, fileIndex}})
	}
	sort.Slice(filesInfo, func(i, j int) bool {
		for k := range filesInfo[i].sortIndices {
			if filesInfo[i].sortIndices[k] == filesInfo[j].sortIndices[k] {
				continue
			}
			return filesInfo[i].sortIndices[k] < filesInfo[j].sortIndices[k]
		}
		return false
	})

	for _, partFileInfo := range filesInfo {
		partFileName := partFileInfo.fileName

		log.Infof("Reading part file :%s, %s", partsDir, partFileName)
		eReader, err := (*tmpCloudManager).Get(partsDir, partFileName)
		if err != nil {
			return 0, err
		}

		scanner := bufio.NewScanner(eReader)
		const maxCapacity = 30 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			line := scanner.Text()

			_, err = io.WriteString(cloudWriter, line+"\n")
			if err != nil {
				log.Errorf("Unable to write to file :%v", err)
				return 0, err
			}
			countLines += 1
		}
	}
	err = cloudWriter.Close()

	return countLines, err
}
