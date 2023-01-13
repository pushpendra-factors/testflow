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
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
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
func Pull_channel_beam_controller(projectId int64, channel string, cloud_dir, cloud_file string,
	batchSize int, docTypeMap map[string]int, beamStruct *RunBeamConfig,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTime, endTime int64) error {

	log.Infof(" Sorting pull %s in beam Project Id :%d", channel, projectId)
	docTypesBeamString := make([]string, 0)
	docTypesList := make([]string, 0)
	docTypesMapList := make(map[string]int64)

	bucketName := (*cloudManager).GetBucketName()
	cloudEventFileName := path.Join(cloud_dir, cloud_file)
	log.Infof("bucket name %s cloud name:%s", bucketName, cloudEventFileName)

	for docType, _ := range docTypeMap {
		docTypesList = append(docTypesList, docType)
	}

	for idx, uid := range docTypesList {
		docTypesMapList[uid] = int64(idx)
	}

	err := putDoctypesToGCP(channel, projectId, docTypesMapList, cloudManager, diskManager, startTime, endTime)
	if err != nil {
		return err
	}

	num_docTypes := len(docTypesList)
	numRoutines := num_docTypes
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a docType to sort.
		low := int(math.Min(float64(i), float64(num_docTypes)))
		high := int(math.Min(float64(i), float64(num_docTypes)))
		tmp := CUserIdsBeam{projectId, int64(low), int64(high), bucketName, "", beamStruct.Env, startTime, endTime, channel, 0}

		t, err := json.Marshal(tmp)
		if err != nil {
			return fmt.Errorf("unable to encode string : %v", err)
		}
		docTypesBeamString = append(docTypesBeamString, string(t))
	}

	err = SortDoctypesExecutor(beamStruct, docTypesBeamString)
	if err != nil {
		return err
	}
	return nil
}

func SortDoctypesExecutor(beamStruct *RunBeamConfig, cPatternsString []string) error {

	s := beamStruct.Scp
	s = s.Scope("pull_events")

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
		log.Fatalf("unable to run beam pipeline :  %s", err)
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

	docType, err := getDoctypeFromFile(channel, ctx, projectId, &cloudManager, startTime, endTime, lowNum, highNum)
	if err != nil {
		return err
	}

	err = downloadAndSortChannelFile(channel, ctx, projectId, &cloudManager, startTime, endTime, docType)
	if err != nil {
		return err
	}
	return nil
}

// read unsorted file, get ad reports with given docType, sort and upload sorted part file to cloud
func downloadAndSortChannelFile(channel string, ctx context.Context, projectId int64, cloudManager *filestore.FileManager, startTime, endTime int64,
	docType int) error {

	beamlog.Infof(ctx, "Downloading and sorting events file")
	adReports := make([]*pull.CounterCampaignFormat, 0)

	gcpPath, gcpFName := (*cloudManager).GetChannelFilePathAndName(channel, projectId, startTime, endTime)
	gcpFName = "unsorted_" + gcpFName
	rc, err := (*cloudManager).Get(gcpPath, gcpFName)
	if err != nil {
		return fmt.Errorf("unable to get file from cloud :%v", err)
	}

	beamlog.Infof(ctx, "Reading events from :%s%s", gcpPath, gcpFName)
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

		if report.Doctype == docType {
			adReports = append(adReports, report)
		}
	}

	err = rc.Close()
	if err != nil {
		beamlog.Errorf(ctx, "unable to close events file from gcp :%v", err)
	}

	eventsTmpFile, err := ioutil.TempFile("", fmt.Sprintf("%d_%s_", docType, channel))
	if err != nil {
		return fmt.Errorf("unable to create tmp file :%v", err)
	} else {
		beamlog.Infof(ctx, "events file created :%s", eventsTmpFile.Name())
	}

	err = sortAdReportsWithDoctype(ctx, adReports, eventsTmpFile)
	if err != nil {
		return err
	}

	err = eventsTmpFile.Close()
	if err != nil {
		beamlog.Error(ctx, "unable to close sorted events file")
	}
	err = writeSortedFileToGCP(ctx, projectId, startTime, endTime, eventsTmpFile.Name(), cloudManager, U.DataTypeAdReport, channel, 0)
	if err != nil {
		beamlog.Error(ctx, "Failed to write sorted events to gcp:%v", err)

	}
	err = os.Remove(eventsTmpFile.Name())
	if err != nil {
		return err
	}
	beamlog.Infof(ctx, "Deleted tmp events File :%s", eventsTmpFile.Name())

	return nil
}

// merge all sorted part files to create one sorted file
func ReadAndMergeChannelPartFiles(projectId int64, cloudDir string,
	local_dir string, fileName string, tmpCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver) (int64, error) {
	var countLines int64
	local_sorted_file := path.Join(local_dir, fileName)
	eventsTmpFile, err := os.OpenFile(local_sorted_file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("unable to open local sorted file:%s", local_sorted_file)
		return 0, err
	}

	log.Infof("Merging partFiles:%d", projectId)
	partFilesDir := path.Join(cloudDir, "parts/"+strings.Replace(fileName, ".txt", "", 1))
	listFiles := (*tmpCloudManager).ListFiles(partFilesDir)
	fileNames := make([]string, 0)
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		fileNames = append(fileNames, partFileName)
		sort.Slice(fileNames, func(i, j int) bool {
			fileNameArr := strings.SplitN(fileNames[i], "_", 2)
			docType_i, _ := strconv.ParseInt(fileNameArr[0], 10, 64)
			fileNameArr = strings.SplitN(fileNames[j], "_", 2)
			docType_j, _ := strconv.ParseInt(fileNameArr[0], 10, 64)
			return docType_i <= docType_j
		})
	}

	for _, partFileName := range fileNames {

		log.Infof("Reading part file :%s, %s", partFilesDir, partFileName)
		file, err := (*tmpCloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			return 0, err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 30 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			line := scanner.Text()

			_, err = eventsTmpFile.WriteString(line + "\n")
			if err != nil {
				log.Errorf("Unable to write to file :%v", err)
				return 0, err
			}
			countLines += 1
		}
	}
	defer eventsTmpFile.Close()

	return countLines, nil
}

// sort the given ad reports based on Timestamp (ascending)
func sortAdReportsWithDoctype(ctx context.Context, adReports []*pull.CounterCampaignFormat, eventsTmpFile *os.File) error {

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
		_, err = eventsTmpFile.WriteString(eString)
		if err != nil {
			log.Errorf("unable to write event to tmp file :%v ", eString)
			return err
		}

	}

	return nil
}
