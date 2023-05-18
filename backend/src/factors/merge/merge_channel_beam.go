package merge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/pull"
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

// create local file, write docTypes map to it and upload to cloud
func putDoctypesToGCP(channel string, projectId int64, docTypes map[string]int64, tmpCloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	startTime, endTime int64) error {

	cDir, cName := (diskManager).GetChannelArtifactFilePathAndName(channel, projectId, startTime, endTime)
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

	for docType, index := range docTypes {
		uidStruct := UidMap{Userid: docType, Index: index}
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
	gDir, gName := (*tmpCloudManager).GetChannelArtifactFilePathAndName(channel, projectId, startTime, endTime)
	err = (*tmpCloudManager).Create(gDir, gName, r)
	if err != nil {
		return err
	}

	return nil
}

// get doctype pertaining to index between high and low from docTypes file
func getDoctypeFromFile(channel string, ctx context.Context, projectId int64, cloudManager *filestore.FileManager,
	startTime, endTime int64, low, high int64) (int, error) {

	beamlog.Infof(ctx, "getting users file :%d,%d,%d", projectId, low, high)
	gDir, gName := (*cloudManager).GetChannelArtifactFilePathAndName(channel, projectId, startTime, endTime)
	r, err := (*cloudManager).Get(gDir, gName)
	if err != nil {
		return 0, err
	}
	var docType int

	scannerEvents := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scannerEvents.Buffer(buf, 1024*1024)

	for scannerEvents.Scan() {
		line := scannerEvents.Text()
		var ud UidMap
		if err := json.Unmarshal([]byte(line), &ud); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return 0, err
		}

		if (ud.Index >= low) && (ud.Index <= high) {
			num, _ := strconv.ParseInt(ud.Userid, 10, 64)
			docType = int(num)
		}

	}
	beamlog.Infof(ctx, "processing docType :%d,%d", projectId, docType)

	return docType, nil
}

// read unsorted channel file from cloud, pass some docTypes to a worker, workers sort ad reports for the alloted docType and upload sorted part files
func Pull_channel_beam_controller(projectId int64, channel string, startIndexToFileInfoMap map[int][]*partFileInfo, beamStruct *RunBeamConfig,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTime, endTime int64) error {

	log.Infof(" Sorting pull %s in beam Project Id :%d", channel, projectId)
	docTypesBeamString := make([]string, 0)

	bucketName := (*cloudManager).GetBucketName()

	for low, fileInfoList := range startIndexToFileInfoMap {
		for _, fileInfo := range fileInfoList {
			// Each worker gets a docType to sort.
			tmp := CUserIdsBeam{projectId, int64(low), int64(fileInfo.endUserIndex), fileInfo.fileTimestampIndex, bucketName, "", beamStruct.Env, startTime, endTime, channel, 0}

			t, err := json.Marshal(tmp)
			if err != nil {
				return fmt.Errorf("unable to encode string : %v", err)
			}
			docTypesBeamString = append(docTypesBeamString, string(t))
		}
	}
	err := SortDoctypesExecutor(beamStruct, docTypesBeamString)
	if err != nil {
		return err
	}
	return nil
}

func SortDoctypesExecutor(beamStruct *RunBeamConfig, cPatternsString []string) error {

	s := beamStruct.Scp
	s = s.Scope("merge_ad_reports")

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
	beam.ParDo0(s, &SortAdDoFn{Config: config}, cPatternsPcolReshuffled)
	err := beamx.Run(ctx, p)
	if err != nil {
		log.Errorf("unable to run beam pipeline :  %s", err)
		return err
	}
	return nil
}

type SortAdDoFn struct {
	Config *C.Configuration
}

func (f *SortAdDoFn) StartBundle(ctx context.Context) {
	beamlog.Info(ctx, "Initializing conf from sortDoFn")
	initConf(f.Config)
}

func (f *SortAdDoFn) FinishBundle(ctx context.Context) {
	beamlog.Info(ctx, "Closing DB Connection from FinishBundle sortDoFn")
	C.SafeFlushAllCollectors()
}

func (f *SortAdDoFn) ProcessElement(ctx context.Context, cpString string) error {
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
	channel := up.Channel
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

	err = downloadAndSortChannelFile(channel, ctx, projectId, &cloudManager, startTime, endTime, lowNum, fileTimeIndex)
	if err != nil {
		return err
	}
	return nil
}

// read unsorted file, get ad reports with given docType, sort and upload sorted part file to cloud
func downloadAndSortChannelFile(channel string, ctx context.Context, projectId int64, cloudManager *filestore.FileManager, startTime, endTime int64,
	docTypeIndex int64, fileTimeIndex int) error {

	beamlog.Infof(ctx, "Downloading and sorting events file")
	adReports := make([]*pull.CounterCampaignFormat, 0)

	cDir, cName := (*cloudManager).GetChannelPartFilePathAndName(channel, projectId, startTime, endTime, false, int(docTypeIndex), fileTimeIndex)
	rc, err := (*cloudManager).Get(cDir, cName)
	if err != nil {
		return fmt.Errorf("unable to get file from cloud :%v", err)
	}

	beamlog.Infof(ctx, "Reading ad reports from :%s%s", cDir, cName)
	scannerEvents := bufio.NewScanner(rc)
	buf := make([]byte, 0, 64*1024)
	scannerEvents.Buffer(buf, 1024*1024)

	for scannerEvents.Scan() {
		line := scannerEvents.Text()
		var report *pull.CounterCampaignFormat
		if err := json.Unmarshal([]byte(line), &report); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return err
		}

		adReports = append(adReports, report)

	}

	err = rc.Close()
	if err != nil {
		beamlog.Errorf(ctx, "unable to close events file from gcp :%v", err)
	}

	cDir, cName = (*cloudManager).GetChannelPartFilePathAndName(channel, projectId, startTime, endTime, true, int(docTypeIndex), fileTimeIndex)
	cloudWriter, err := (*cloudManager).GetWriter(cDir, cName)
	if err != nil {
		beamlog.Error(ctx, "Failed to pull events. Upload failed.")
		return err
	}

	err = WriteSortedAdReportsWithDoctype(ctx, adReports, &cloudWriter)
	if err != nil {
		beamlog.Error(ctx, "error sorting ad reports and writing: ", err)
		return err
	}
	err = cloudWriter.Close()
	if err != nil {
		log.WithError(err).Error("error closing cloud writer")
		return err
	}

	return nil
}

// read all sorted part files from cloud and merge to create one sorted file in cloud
func ReadAndMergeChannelPartFiles(projectId int64, partsDir string, sortedDir string, sortedName string,
	tmpCloudManager, sortedCloudManager *filestore.FileManager) (int, error) {
	var countLines int
	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return 0, err
	}

	log.Infof("Merging partFiles for project: %d", projectId)
	listFiles := (*tmpCloudManager).ListFiles(partsDir)
	filesInfo := make([]fileNameSortInfo, 0)
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		fileNameArr := strings.SplitN(partFileName, "_", 2)
		docType, _ := strconv.ParseInt(fileNameArr[0], 10, 64)
		filesInfo = append(filesInfo, fileNameSortInfo{fileName: partFileName, sortIndices: []int64{docType}})
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
		file, err := (*tmpCloudManager).Get(partsDir, partFileName)
		if err != nil {
			return 0, err
		}

		scanner := bufio.NewScanner(file)
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

// sort the given ad reports based on Timestamp (ascending)
func WriteSortedAdReportsWithDoctype(ctx context.Context, adReports []*pull.CounterCampaignFormat, cloudWriter *io.WriteCloser) error {

	beamlog.Debugf(ctx, "sort ad reports :%d", len(adReports))

	sort.Slice(adReports[:], func(i, j int) bool {
		return adReports[i].Timestamp < adReports[j].Timestamp
	})

	for _, ar := range adReports {

		docBytes, err := json.Marshal(ar)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to marshal event.")
			return err
		}
		eString := string(docBytes)
		eString = eString + "\n"
		_, err = io.WriteString(*cloudWriter, eString)
		if err != nil {
			log.Errorf("unable to write event to tmp file :%v ", eString)
			return err
		}

	}

	return nil
}
