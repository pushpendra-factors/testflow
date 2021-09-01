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
	"path/filepath"
	"strings"
	"time"

	"github.com/apache/beam/sdks/go/pkg/beam"
	beamlog "github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
	log "github.com/sirupsen/logrus"
)

type CPatternsBeam struct {
	ProjectID         uint64     `json:"projectID"`
	ModelId           uint64     `json:"modelID"`
	PtEnames          [][]string `json:"penames"`
	ModEventsFilepath string     `json:"fpath"`
	CountOccurence    bool       `json:"countOccurence"`
	// UserEventsinfo    *P.UserAndEventsInfo `json:"uEinfo"`
	// ChunkDir   string `json:"chunkDir"`
	BucketName string `json:"bucketName"`
	ObjectName string `json:"objectName"`
	ScopeName  string `json:"ScopeName"`
	Env        string `json:"Env"`
}

type RunBeamConfig struct {
	RunOnBeam    bool             `json:"runonbeam"`
	Ctx          context.Context  `json:"ctx"`
	Scp          beam.Scope       `json:"scp"`
	Pipe         *beam.Pipeline   `json:"pipeLine"`
	Env          string           `json:"Env"`
	DriverConfig *C.Configuration `json:"driverconfig"`
}

func InitConfBeam(config *C.Configuration) {
	log.Info("initializing conf from startbundle")
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)

	C.InitConf(config)
	C.SetIsBeamPipeline()
	C.InitRedisConnection(config.RedisHost, config.RedisPort, true, 20, 0)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
}

func readPatternFromFile(partFilesDir string, cloudManager *filestore.FileManager, projectId, modelId uint64) ([]*P.Pattern, error) {
	mineLog.Infof("Reading from part files Directory : %s", partFilesDir)
	listFiles := (*cloudManager).ListFiles(partFilesDir)
	patterns := make([]*P.Pattern, 0)
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		mineLog.Infof("Reading part file : %s", partFileName)
		file, err := (*cloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 10 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)
		for scanner.Scan() {
			line := scanner.Text()
			var pattern P.Pattern
			if err := json.Unmarshal([]byte(line), &pattern); err != nil {
				log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
				return nil, err
			}
			patterns = append(patterns, &pattern)
		}

		if err := file.Close(); err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Error closing file")
		}

	}
	mineLog.Infof("Total number of patterns mined : %d", len(patterns))
	return patterns, nil
}

func countPatternsWorkerBeam(ctx context.Context, projectID uint64, filepath string,
	patterns []*P.Pattern, countOccurence bool) {

	log.Infof("Reading file from :%s", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		mineLog.WithField("filePath", filepath).Error("Failure on count pattern workers.")
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)
	log.Infof("Number of pattens to CounPatterns :%d", len(patterns))
	P.CountPatterns(projectID, scanner, patterns, countOccurence)
	file.Close()
}

func writeFileToGCP(projectId, modelId uint64, name string, fpath string,
	cloudManager *filestore.FileManager) error {

	path := (*cloudManager).GetProjectModelDir(projectId, modelId)
	mineLog.Infof("Reading file from (toGCP)   : %s ", fpath)
	r, err := os.OpenFile(fpath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("unable to open file : %s :%v", fpath, err)
	}
	mineLog.Infof("Writing file to GCP : %s , %s ", path, name)
	err = (*cloudManager).Create(path, name, r)
	if err != nil {
		mineLog.Error("Failed to create event file to cloud.")
		return err
	}
	return nil
}

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func deleteFile(path string) error {
	// delete file
	var err = os.Remove(path)
	if err != nil {
		return err
	}

	log.Infof("File Deleted : %s", path)
	return nil
}

func countPatternController(beamStruct *RunBeamConfig, projectId uint64, modelId uint64,
	cloudManager *filestore.FileManager,
	filepathString string, patterns []*P.Pattern, numRoutines int,
	userAndEventsInfo *P.UserAndEventsInfo, countOccurence bool,
	efTmpPath string,
	scopeName string) (string, error) {

	numPatterns := len(patterns)
	mineLog.Info(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := 1000 //600 too long , will have issue in len three stage

	numRoutines = int(math.Ceil(float64(numPatterns) / float64(batchSize)))

	patternEnames := make([][]string, 0)
	for _, v := range patterns {
		patternEnames = append(patternEnames, v.EventNames)
	}

	bucketName := (*cloudManager).GetBucketName()
	modelpathDir := (*cloudManager).GetProjectModelDir(projectId, modelId)
	mineLog.Infof("bucketName :%s", bucketName)
	mineLog.Infof("model path :%s", modelpathDir)
	modEventsFile := fmt.Sprintf("events_modified_%d.txt", modelId)

	modelpath := modelpathDir + modEventsFile
	mineLog.Infof("model path of modified file:%s", modelpath)

	cPatterns := make([]string, 0)
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))

		t, err := json.Marshal(CPatternsBeam{
			projectId, modelId, patternEnames[low:high], filepathString, countOccurence,
			bucketName, modelpath, scopeName, beamStruct.Env})
		if err != nil {
			return "", fmt.Errorf("unable to encode string : %v", err)
		}
		cPatterns = append(cPatterns, string(t))

	}

	ptInputPath := filepath.Join(efTmpPath, "pinput", scopeName+".txt")
	mineLog.Infof("File patterns name : %s", ptInputPath)

	pfFile, err := create(ptInputPath)
	mineLog.Info(err)
	if err != nil {
		return "", fmt.Errorf("unable to create file :%v", ptInputPath)
	}
	for _, cp := range cPatterns {
		_, err := pfFile.WriteString(fmt.Sprintf("%s\n", cp))
		if err != nil {
			return "", fmt.Errorf("error writing json to file :%v", err)
		}
	}
	err = pfFile.Close()
	if err != nil {
		return "", fmt.Errorf("error in closing file : %v", err)
	}

	//-------Write events Info file to GCP--------
	userAndEventsInfoBytes, err := json.Marshal(userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to unmarshal events Info.")
		return "", err
	}

	if len(userAndEventsInfoBytes) > 249900000 {
		// Limit is 250MB
		errorString := fmt.Sprintf(
			"Too big properties info, modelId: %d,projectId: %d, numBytes: %d",
			modelId, projectId, len(userAndEventsInfoBytes))
		mineLog.Error(errorString)
		return "", fmt.Errorf(errorString)
	}
	events := bytes.NewReader(userAndEventsInfoBytes)
	path, _ := (*cloudManager).GetModelEventInfoFilePathAndName(projectId, modelId)
	nameUI := scopeName + "_UI.txt"
	err = (*cloudManager).Create(path, nameUI, events)
	if err != nil {
		mineLog.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return "", err
	}

	//-------End Write events Info file to GCP--------

	//-------Write intermediate patterns file to GCP--------
	fileNameIntermediate := scopeName + "_itm_patterns.txt"
	err = writeFileToGCP(projectId, modelId, fileNameIntermediate, ptInputPath, cloudManager)
	if err != nil {
		return "", fmt.Errorf("unable to write to cloud : %s", fileNameIntermediate)
	}
	//-------END of Write intermediate patterns file to GCP--------

	p, s := beam.NewPipelineWithRoot()
	ctx := beamStruct.Ctx
	config_ := beamStruct.DriverConfig
	configJson, err := json.Marshal(config_)
	if err != nil {
		return "", fmt.Errorf("unable to marshall string :%v", err)
	}
	configEncodedString := string(configJson)
	efTmpPath = modelpathDir
	if beam.Initialized() {
		mineLog.Info("Initialized beam")
	} else {
		mineLog.Fatal("Unable to init beam")
	}
	if s.IsValid() {
		mineLog.Info("Scope is Valid")
	} else {
		mineLog.Fatal("Scope is not valid")
	}
	var itrFname string
	if beamStruct.Env == "development" {
		itrFname = filepath.Join(modelpathDir, fileNameIntermediate)
	} else {
		itrFname = `gs://` + bucketName + `/` + modelpathDir + fileNameIntermediate
	}
	countPatternsExecutor(s, scopeName, configEncodedString, itrFname)
	err = beamx.Run(ctx, p)
	if err != nil {
		mineLog.Fatalf("unable to run beam pipeline :  %s", err)
		return "", err
	}
	patternsFpath := filepath.Join(modelpathDir, "patterns_part", scopeName)
	mineLog.Infof("Patterns writtens to GCP : %s ", patternsFpath)
	mineLog.Infof("completed counting patterns in beam")
	return patternsFpath, nil

}

// ------------------Pardo Fn - Beam build seq----------------
func countPatternsExecutor(s beam.Scope, scopeName string, config_ string,
	itrFname string) {

	s = s.Scope(scopeName)
	cPatternsPcol := Read(s, itrFname)
	cPatternsPcolReshuffled := beam.Reshuffle(s, cPatternsPcol)
	beam.ParDo0(s, &CpThreadDoFn{Config: config_}, cPatternsPcolReshuffled)
}

// ----------------cpThreadFn DoFn------------------------

type CpThreadDoFn struct {
	Config string
}

func (f *CpThreadDoFn) StartBundle(ctx context.Context) {
	beamlog.Info(ctx, "Initializing conf from CpThreadDoFn")
	var configVar *C.Configuration
	err := json.Unmarshal([]byte(f.Config), &configVar)
	if err != nil {
		log.Infof("Unable to unmarshall :%v", err)
	}
	InitConfBeam(configVar)
}

func (f *CpThreadDoFn) FinishBundle(ctx context.Context) {
	beamlog.Info(ctx, "Closing DB Connection from FinishBundle cpThreadDoFn")
	C.SafeFlushAllCollectors()
}

func (f *CpThreadDoFn) ProcessElement(ctx context.Context, cpString string) error {
	beamlog.Info(ctx, "Inside countPatternsThread")

	var cp CPatternsBeam
	err := json.Unmarshal([]byte(cpString), &cp)
	if err != nil {
		return fmt.Errorf("unable to unmarshall string in processElement :%s", cpString)
	}

	var patterns []*P.Pattern
	var UserAndEventsInfo = P.NewUserAndEventsInfo()
	projectID := cp.ProjectID
	eventsFilePath := cp.ModEventsFilepath
	countOccurence := cp.CountOccurence
	bucketName := cp.BucketName
	objectName := cp.ObjectName
	modelId := cp.ModelId
	scopeName := cp.ScopeName
	envFlag := cp.Env
	log.Infof("Project Id : %d  ", projectID)
	log.Infof("Model Id : %d  ", modelId)
	log.Infof("BucketName: %d  ", bucketName)

	// enable client for GCP
	var cloudManager filestore.FileManager
	if envFlag == "development" {
		cloudManager = serviceDisk.New(bucketName)
	} else {
		cloudManager, err = serviceGCS.New(bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}

	modelpath := cloudManager.GetProjectModelDir(projectID, modelId)
	pathEinfo, _ := cloudManager.GetModelEventInfoFilePathAndName(projectID, modelId)
	einfo, err := (cloudManager).Get(pathEinfo, scopeName+"_UI.txt")
	if err != nil {
		return fmt.Errorf("unable to get events info file :%v", err)
	}

	line, err := ioutil.ReadAll(einfo)
	if err != nil {
		return fmt.Errorf("unable to read string :%v", err)
	}
	log.Infof("Number of bytes read :%d", len(line))
	err = json.Unmarshal(line, &UserAndEventsInfo)
	if err != nil {
		return fmt.Errorf("unable to unmarshall userEventsInfo :%v", err)
	}

	log.Infof("Read user and events Info Model Version : %d", UserAndEventsInfo.ModelVersion)
	log.Infof("number of patterns : %d", len(cp.PtEnames))

	for _, eventNames := range cp.PtEnames {
		p, err := P.NewPattern(eventNames, UserAndEventsInfo)
		if err != nil {
			return nil
		}
		patterns = append(patterns, p)
	}

	// downloading events file
	eventsTmpFile, err := ioutil.TempFile("", "patterns_")
	if err != nil {
		return fmt.Errorf("unable to create tmp file :%v", err)
	}

	log.Infof("bucket : %s, objectName : %s", bucketName, objectName)
	modEventsFile := fmt.Sprintf("events_modified_%d.txt", modelId)
	rc, err := (cloudManager).Get(modelpath, modEventsFile)
	if err != nil {
		return fmt.Errorf("unable to get file from cloud :%v", err)
	}
	scannerEvents := bufio.NewScanner(rc)
	buf := make([]byte, 0, 64*1024)
	scannerEvents.Buffer(buf, 1024*1024)
	countEventLines := 0
	tmparr := []string{"$hubspot_deal_state_changed",
		"MQLs", "Lead Status - Connected", "Lead Status - Demo Completed"}
	for scannerEvents.Scan() {
		line := scannerEvents.Text()
		n, err := eventsTmpFile.WriteString(line + "\n")
		if err != nil {

			for _, v := range tmparr {
				if strings.Contains(v, line) {
					log.Infof("Error reading line :%s,%d", line, n)
				}
			}

			return err
		} else {
			countEventLines++
		}
	}
	log.Infof("Number of Event lines written :%d", countEventLines)
	info, err := os.Stat(eventsTmpFile.Name())
	if err != nil {
		return err
	}
	eventsFileSize := info.Size()

	log.Infof("tmpFile name : %s", eventsTmpFile.Name())
	log.Infof("tmpFile size : %d", eventsFileSize)

	eventsFilePath = eventsTmpFile.Name()
	beamlog.Info(ctx, "calling  countPatternsWorkerBeam")
	countPatternsWorkerBeam(ctx, projectID, eventsFilePath, patterns, countOccurence)
	//-------write as part files---------
	key := time.Now().Nanosecond()
	patternTmpFileLocal, err := ioutil.TempFile("", "patterns_computed_")
	if err != nil {
		return fmt.Errorf("unable to create tmp file :%v", err)
	}
	log.Infof("Writing patterns to local tmp file :%s", patternTmpFileLocal.Name())
	tmpPatternsFile := filepath.Join("patterns_part", scopeName, "part_"+fmt.Sprint(key)+"_UI.txt")
	lineNum := 0
	for _, pat := range patterns {
		tmparr := []string{"$hubspot_deal_state_changed",
			"MQLs", "Lead Status - Connected", "Lead Status - Demo Completed"}
		for _, v := range tmparr {
			if strings.Compare(v, pat.EventNames[0]) == 0 {
				log.Infof(" count pattern  %s is %d : %v", v, pat.PerUserCount, pat)
			}
		}
		pattBytes, _ := json.Marshal(pat)
		patternTmpFileLocal.WriteString(string(pattBytes) + "\n")
		lineNum++
		if math.Mod(float64(lineNum), 100.0) == 0.0 {
			log.Infof("Written %d lines", lineNum)
		}
	}

	if envFlag == "development" {
		fi, err := create(filepath.Join(modelpath, tmpPatternsFile))
		if err != nil {
			return err
		}
		fi.Close()
	}

	err = writeFileToGCP(projectID, modelId, tmpPatternsFile, patternTmpFileLocal.Name(), &cloudManager)
	if err != nil {
		return fmt.Errorf("unable to write part file to GCP : %v", err)
	}

	err = deleteFile(patternTmpFileLocal.Name())
	if err != nil {
		return err
	}
	log.Infof("Delete patterns File :%s", patternTmpFileLocal.Name())

	//--destroy events file---
	err = deleteFile(eventsTmpFile.Name())
	if err != nil {
		return err
	}
	log.Infof("Delete events File :%s", eventsTmpFile.Name())

	return nil
}

// cpReadDoFn - to handle token too long error

func Read(s beam.Scope, glob string) beam.PCollection {
	s = s.Scope("Read")

	filesystem.ValidateScheme(glob)
	return read(s, beam.Create(s, glob))
}
func read(s beam.Scope, col beam.PCollection) beam.PCollection {
	files := beam.ParDo(s, expandFn, col)
	return beam.ParDo(s, readFn, files)
}

func expandFn(ctx context.Context, glob string, emit func(string)) error {
	if strings.TrimSpace(glob) == "" {
		return nil // ignore empty string elements here
	}

	fs, err := filesystem.New(ctx, glob)
	if err != nil {
		return err
	}
	defer fs.Close()

	files, err := fs.List(ctx, glob)
	if err != nil {
		return err
	}
	for _, filename := range files {
		emit(filename)
	}
	return nil
}

func readFn(ctx context.Context, filename string, emit func(string)) error {
	log.Infof("Reading from %v", filename)

	fs, err := filesystem.New(ctx, filename)
	if err != nil {
		return err
	}
	defer fs.Close()

	fd, err := fs.OpenRead(ctx, filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		emit(scanner.Text())
	}
	return scanner.Err()
}
