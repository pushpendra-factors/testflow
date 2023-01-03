package task

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/filestore"
	M "factors/model/model"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
	beamlog "github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"
	log "github.com/sirupsen/logrus"
)

type CPatternsBeam_ struct {
	ProjectID         int64      `json:"projectID"`
	ModelId           uint64     `json:"modelID"`
	PtEnames          [][]string `json:"penames"`
	ModEventsFilepath string     `json:"fpath"`
	CountOccurence    bool       `json:"countOccurence"`
	CountsVersion     int        `json:"cntvr"`
	HmineSupport      float32    `json:"hmsp"`
	BucketName        string     `json:"bucketName"`
	ObjectName        string     `json:"objectName"`
	ScopeName         string     `json:"ScopeName"`
	Env               string     `json:"Env"`
}

type CPatternsBeam struct {
	ProjectID         int64   `json:"projectID"`
	ModelId           uint64  `json:"modelID"`
	LowNum            int64   `json:"lw"`
	HighNum           int64   `json:"hg"`
	ModEventsFilepath string  `json:"fpath"`
	CountOccurence    bool    `json:"countOccurence"`
	CountsVersion     int     `json:"cntvr"`
	HmineSupport      float32 `json:"hmsp"`
	BucketName        string  `json:"bucketName"`
	ObjectName        string  `json:"objectName"`
	ScopeName         string  `json:"ScopeName"`
	Env               string  `json:"Env"`
}

type PatterNamesIdx struct {
	Id      int64    `json:"id"`
	Pname   []string `json:"pn"`
	Segment int      `json:"sg"`
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

func InitConfBeam(config *C.Configuration) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	C.InitConf(config)
	C.SetIsBeamPipeline()
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 5, 2)
	if err != nil {
		// TODO(prateek): Check how a panic here will effect the pipeline.
		log.WithError(err).Panic("Failed to initalize db.")
	}
	C.InitRedisConnection(config.RedisHost, config.RedisPort, true, 20, 0)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.KillDBQueriesOnExit()
}

func countPatternsWorkerBeam(ctx context.Context, projectID int64, filepath string,
	patterns []*P.Pattern, countOccurence bool, countsVersion int, hmineSupport float32) {

	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = countsVersion
	cAlgoProps.Hmine_support = hmineSupport
	log.Infof("Reading file from :%s,%d,%f", filepath, countsVersion, hmineSupport)
	file, err := os.Open(filepath)
	if err != nil {
		mineLog.WithField("filePath", filepath).Error("Failure on count pattern workers.")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)
	log.Infof("Number of pattens to CounPatterns :%d", len(patterns))
	if countsVersion == 3 {
		log.Infof("Number of pattens to CounPatterns using hmine :%d,%d,%d", len(patterns), countsVersion, hmineSupport)
	}
	P.CountPatterns(projectID, scanner, patterns, countOccurence, cAlgoProps)
}

func ComputeHistogramWorkerBeam(ctx context.Context, projectID int64, filepath string,
	patterns []*P.Pattern, countOccurence bool, countsVersion int, hmineSupport float32) error {

	var cAlgoProps P.CountAlgoProperties
	cAlgoProps.Counting_version = countsVersion
	cAlgoProps.Hmine_support = hmineSupport
	if countsVersion == 3 {
		log.Infof("Number of pattens to CounPatterns using hmine :%d,%d,%d", len(patterns), countsVersion, hmineSupport)
	}

	if err := computeAllUserPropertiesHistogram(projectID, filepath, patterns[0], countsVersion, hmineSupport); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to compute user properties.")
		return err
	}
	return nil
}

func writeFileToGCP(projectId int64, modelId uint64, name string, fpath string,
	cloudManager *filestore.FileManager, writePath string) error {

	var path string
	if writePath == "" {
		path = (*cloudManager).GetProjectModelDir(projectId, modelId)
	} else {
		path = writePath
	}
	mineLog.Infof("Reading file from (toGCP)   : %s ", fpath)
	f, err := os.OpenFile(fpath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("unable to open file : %s :%v", fpath, err)
	}
	defer f.Close()

	r := struct {
		WriteTo int // to "disable" WriteTo method of reader type
		*bufio.Reader
	}{0, bufio.NewReader(f)}

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

func shufflePatterns(patternNames [][]string) [][]string {
	// shuffle slice of patterns to ensure all heavey patterns are distributed across nodes while computing
	patternEnames_ := make([][]string, 0)
	for _, v := range patternNames {
		patternEnames_ = append(patternEnames_, v)
	}
	patternEnames := make([][]string, len(patternEnames_))
	perm := rand.Perm(len(patternEnames_))
	for i, v := range perm {
		patternEnames[v] = patternEnames_[i]
	}

	return patternEnames

}

func countPatternController(beamStruct *RunBeamConfig, projectId int64, modelId uint64,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	filepathString string, patterns []*P.Pattern, numRoutines int,
	userAndEventsInfo *P.UserAndEventsInfo, countOccurence bool,
	efTmpPath string,
	scopeName string, cAlgoProps P.CountAlgoProperties) (string, error) {

	var countsVersion int = cAlgoProps.Counting_version
	var hmineSupport float32 = cAlgoProps.Hmine_support
	patternEnames := make([]PatterNamesIdx, 0)

	numPatterns := len(patterns)
	mineLog.Info(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := beamStruct.NumWorker
	numRoutines = int(math.Ceil(float64(numPatterns) / float64(batchSize)))

	for isx, v := range patterns {
		patternEnames = append(patternEnames, PatterNamesIdx{int64(isx), v.EventNames, v.Segment})
	}

	bucketName := (*cloudManager).GetBucketName()
	modelpathDir := (*cloudManager).GetProjectModelDir(projectId, modelId)

	err := WritePatternsToGCP(projectId, modelId, scopeName, patterns, cloudManager, diskManager)
	if err != nil {
		return "", fmt.Errorf("unable write Patterns to GCP: %v", err)
	}

	modEventsFile := fmt.Sprintf("events_modified_%d.txt", modelId)
	modelpath := modelpathDir + modEventsFile
	mineLog.Infof("model path of modified file:%s", modelpath)
	cPatterns := make([]string, 0)
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int64(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int64(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))

		t, err := json.Marshal(CPatternsBeam{
			projectId, modelId, low, high, filepathString, countOccurence, countsVersion, hmineSupport,
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
	path = path + "patterns_part/" + scopeName + "/"
	err = (*cloudManager).Create(path, nameUI, events)
	if err != nil {
		mineLog.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return "", err
	}

	//-------End Write events Info file to GCP--------

	//-------Write intermediate patterns file to GCP--------
	fileNameIntermediate := scopeName + "_itm_patterns.txt"
	err = writeFileToGCP(projectId, modelId, fileNameIntermediate, ptInputPath, cloudManager, path)
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
	efTmpPath = efTmpPath + "patterns_part/" + scopeName + "/"
	if beamStruct.Env == "development" {

		itrFname = filepath.Join(efTmpPath, fileNameIntermediate)
	} else {
		itrFname = `gs://` + bucketName + `/` + efTmpPath + fileNameIntermediate
	}
	countPatternsExecutor(s, scopeName, configEncodedString, itrFname, countsVersion)
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

func UserPropertiesHistogramController(beamStruct *RunBeamConfig, projectId int64, modelId uint64,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	mod_events_path string, patterns []*P.Pattern, numRoutines int,
	userAndEventsInfo *P.UserAndEventsInfo, countOccurence bool,
	efTmpPath string,
	scopeName string, cAlgoProps P.CountAlgoProperties) (string, error) {

	var countsVersion int = cAlgoProps.Counting_version
	var hmineSupport float32 = cAlgoProps.Hmine_support
	numPatterns := len(patterns)
	mineLog.Info(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := beamStruct.NumWorker
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

	err := WritePatternsToGCP(projectId, modelId, scopeName, patterns, cloudManager, diskManager)
	if err != nil {
		return "", err
	}

	modelpath := modelpathDir + modEventsFile
	mineLog.Infof("model path of modified file:%s", modelpath)

	cPatterns := make([]string, 0)
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int64(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int64(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))

		t, err := json.Marshal(CPatternsBeam{
			projectId, modelId, low, high, mod_events_path, countOccurence, countsVersion, hmineSupport,
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
	path = path + "patterns_part/" + scopeName + "/"
	err = (*cloudManager).Create(path, nameUI, events)
	if err != nil {
		mineLog.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return "", err
	}

	//-------End Write events Info file to GCP--------

	//-------Write intermediate patterns file to GCP--------
	fileNameIntermediate := scopeName + "_itm_patterns.txt"
	err = writeFileToGCP(projectId, modelId, fileNameIntermediate, ptInputPath, cloudManager, path)
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
	efTmpPath = efTmpPath + "patterns_part/" + scopeName + "/"
	if beamStruct.Env == "development" {

		itrFname = filepath.Join(efTmpPath, fileNameIntermediate)
	} else {
		itrFname = `gs://` + bucketName + `/` + efTmpPath + fileNameIntermediate
	}
	UserPropertiesExecutor(s, scopeName, configEncodedString, itrFname, countsVersion)
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
	itrFname string, countsVersion int) {

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
	var inBeam bool = true
	var debugHmine bool = true

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
	countsVersion := cp.CountsVersion
	hmineSupport := cp.HmineSupport
	log.Infof("Project Id : %d  ", projectID)
	log.Infof("Model Id : %d  ", modelId)
	log.Infof("BucketName: %s  ", bucketName)
	if countsVersion == 3 {
		log.Infof("counting algo : hmine")
	}

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
	pathEinfo = pathEinfo + "patterns_part/" + scopeName + "/"
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

	patterns, err = ReadPatternsGCP(ctx, projectID, modelId, scopeName, &cloudManager, UserAndEventsInfo, cp.LowNum, cp.HighNum)
	if err != nil {
		return fmt.Errorf("Unable to read patterns")
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
	// create a temp folder for artifacts

	if inBeam {
		dname, err := ioutil.TempDir("", "fptree")
		if err != nil {
			return fmt.Errorf("unable to create tmp folder :%v", err)
		}
		for _, p := range patterns {
			p.PropertiesBaseFolder = dname
		}
		log.Infof("creating a base folder for artifacts:%s", dname)

		if countsVersion == 3 {

			baseFolder := dname
			basePathUserProps := path.Join(baseFolder, "fptree", "userProps")
			basePathEventProps := path.Join(baseFolder, "fptree", "eventProps")
			basePathUserPropsRes := path.Join(baseFolder, "fptree", "userPropsRes")
			basePathEventPropsRes := path.Join(baseFolder, "fptree", "eventPropsRes")

			err = os.MkdirAll(basePathUserPropsRes, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp user directory")
			}

			err = os.MkdirAll(basePathEventPropsRes, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp event directory")
			}

			err = os.MkdirAll(basePathUserProps, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp user directory")
			}

			err = os.MkdirAll(basePathEventProps, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp event directory")
			}

			log.Infof("created tmp properties User folder : %s", basePathUserProps)
			log.Infof("created tmp properties Events folder : %s", basePathEventProps)
			log.Infof("created tmp properties User Results folder : %s", basePathUserPropsRes)
			log.Infof("created tmp properties Events Results folder : %s", basePathEventPropsRes)

		}
	}
	countPatternsWorkerBeam(ctx, projectID, eventsFilePath, patterns, countOccurence, countsVersion, hmineSupport)
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

	err = writeFileToGCP(projectID, modelId, tmpPatternsFile, patternTmpFileLocal.Name(), &cloudManager, "")
	if err != nil {
		return fmt.Errorf("unable to write part file to GCP : %v", err)
	}

	err = deleteFile(patternTmpFileLocal.Name())
	if err != nil {
		return err
	}
	log.Infof("Delete patterns File :%s", patternTmpFileLocal.Name())

	if debugHmine {
		// upload all temp files generated if need to debug
		for _, pt := range patterns {

			up_local_path := filepath.Join(pt.PropertiesBaseFolder, "userProps", pt.PropertiesBasePath)
			if _, err := os.Stat(up_local_path); err == nil {
				if envFlag == "development" {
					fi, err := create(filepath.Join(modelpath, up_local_path))
					if err != nil {
						return err
					}
					fi.Close()
				}

				up_cloud_path := filepath.Join("patterns_part", "user_properties", pt.PropertiesBasePath)
				err := writeFileToGCP(projectID, modelId, up_cloud_path, up_local_path, &cloudManager, "")
				if err != nil {
					return fmt.Errorf("unable to upload tmp user properties file :%v", err)
				}

			} else if errors.Is(err, os.ErrNotExist) {
				// path/to/whatever does *not* exist
				log.Infof("userProperties file does not exist: %s", up_local_path)

			}

			ep_local_path := filepath.Join(pt.PropertiesBaseFolder, "eventProps", pt.PropertiesBasePath)

			if _, err := os.Stat(ep_local_path); err == nil {
				if envFlag == "development" {
					fi, err := create(filepath.Join(modelpath, ep_local_path))
					if err != nil {
						return err
					}
					fi.Close()
				}

				ep_cloud_path := filepath.Join("patterns_part", "event_properties", pt.PropertiesBasePath)
				err := writeFileToGCP(projectID, modelId, ep_cloud_path, ep_local_path, &cloudManager, "")
				if err != nil {
					return fmt.Errorf("unable to upload tmp user properties file :%v", err)
				}

			} else if errors.Is(err, os.ErrNotExist) {
				// path/to/whatever does *not* exist
				log.Infof("eventProperties file does not exist: %s", ep_local_path)

			}

		}

	}

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

func ReadFilterAndCompressPatternsFromFile(partFilesDir string, cloudManager *filestore.FileManager, maxTotalBytes int64, totalConsumedBytes int64, currentPatternsLength int, maxPatternsLength int) ([]*P.Pattern, int64, error) { //, cloudManager *filestore.FileManager

	if currentPatternsLength > maxPatternsLength {
		errorString := fmt.Sprintf(
			"Current pattern length greater than max length. currentPatternsLength:%d, maxPatternsLength: %d",
			currentPatternsLength, maxPatternsLength)
		mineLog.Error(errorString)
		return []*P.Pattern{}, 0, fmt.Errorf(errorString)
	}
	if totalConsumedBytes >= maxTotalBytes {
		mineLog.Info(fmt.Sprintf("No quota. totalConsumedBytes: %d, maxTotalBytes: %d",
			totalConsumedBytes, maxTotalBytes))
		return []*P.Pattern{}, 0, nil
	}

	mineLog.Infof("Reading from part files Directory : %s", partFilesDir)
	listFiles := (*cloudManager).ListFiles(partFilesDir)
	patterns := make([]*P.Pattern, 0)
	trim_stage := 0
	var patternsBytes int64
	var cumulativeBytes int64
	var totalFileSize int64
	var readBytes int64
	trimMap := make(map[int]int64)

	// get totalFileSize
	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		if !U.HasPrefixFromList(partFileName, []string{"part_"}) {
			continue
		}
		fileSize, err := (*cloudManager).GetObjectSize(partFilesDir, partFileName)
		if err != nil {
			log.WithFields(log.Fields{"file": partFileName, "fileSize": fileSize, "err": err}).Fatal("Couldn't get part file size")
		}
		mineLog.Infof("part file: %s, fileSize : %d", partFileName, fileSize)
		totalFileSize += fileSize
	}

	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		if !U.HasPrefixFromList(partFileName, []string{"part_"}) {
			continue
		}
		mineLog.Infof("Reading part file : %s", partFileName)
		file, err := (*cloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			return nil, 0, err
		}

		scanner := bufio.NewScanner(file)
		const maxCapacity = 30 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {

			line := scanner.Text()
			var pattern P.Pattern
			if err := json.Unmarshal([]byte(line), &pattern); err != nil {
				log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
				return nil, 0, err
			}

			// If pattern has positive count
			if pattern.PerUserCount > 0 {

				// if size exceeds and further trimming possible, trim
				if trim_stage < 3 && (totalConsumedBytes+patternsBytes) > maxTotalBytes {
					trim_stage++
					totalPatternsSize := totalFileSize * patternsBytes / readBytes // estimating size of ALL +ve count patterns compressed to trim_stage-1
					trimMap[trim_stage] = totalPatternsSize
					log.WithFields(log.Fields{"patternsBytes": patternsBytes, "trim_stage": trim_stage}).Info("BEFORE")
					patterns, patternsBytes, err = compressPatternsList(patterns, maxTotalBytes, trimMap, trim_stage) //compress the patterns got uptil now
					if err != nil {
						return []*P.Pattern{}, 0, err
					}
					log.WithFields(log.Fields{"patternsBytes": patternsBytes, "trim_stage": trim_stage}).Info("AFTER")
				}

				pattern, pbytes, err := cumulativeCompressPattern(&pattern, maxTotalBytes, trimMap, trim_stage) //compress the new pattern to be added
				if err != nil {
					return []*P.Pattern{}, 0, err
				}
				// TODO : write it to a file ,instead of accumulating in memory
				patterns = append(patterns, pattern)
				patternsBytes += pbytes
			}
			readBytes += int64(len([]byte(line)))
		}

		if err := file.Close(); err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Error closing file")
		}
	}

	if (totalConsumedBytes + patternsBytes) <= maxTotalBytes {
		mineLog.WithFields(log.Fields{
			"trim_stage":           trim_stage,
			"numPatterns":          len(patterns),
			"maxTotalBytes":        maxTotalBytes,
			"totalConsumedBytes":   totalConsumedBytes,
			"currentPatternsBytes": patternsBytes,
		}).Info("Returning compressed patterns")
		return patterns, patternsBytes, nil
	}

	// Patterns are added only till it does not go over maxTotalBytes, in
	// decreasing order of count.
	// Sort the patterns in descending order.
	sort.Slice(patterns,
		func(i, j int) bool {
			return patterns[i].PerUserCount > patterns[j].PerUserCount
		})
	cumulativeBytes = 0
	for i, pattern := range patterns {
		pBytes, err := getPatternSize(pattern)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
		if totalConsumedBytes+cumulativeBytes+pBytes > maxTotalBytes {
			mineLog.WithFields(log.Fields{
				"trim_stage":           trim_stage,
				"numPatterns":          i,
				"numDroppedPatterns":   len(patterns) - i,
				"maxTotalBytes":        maxTotalBytes,
				"totalConsumedBytes":   totalConsumedBytes,
				"currentPatternsBytes": cumulativeBytes,
			}).Info("Dropping patterns")
			patterns = patterns[:i]
			break
		}
		cumulativeBytes += pBytes
	}

	mineLog.Infof("Total number of patterns mined : %d", len(patterns))
	return patterns, cumulativeBytes, nil
}

func compressPattern(pattern *P.Pattern, maxBytesSize int64, trimMap map[int]int64, trim_stage int) (*P.Pattern, int64, error) {

	if pattern.PatternVersion == 1 {
		TRIM_MULTIPLIER := 0.8
		trimFraction := float64(maxBytesSize) * TRIM_MULTIPLIER / float64(trimMap[trim_stage])
		switch trim_stage {
		case 1:
			if pattern.PerUserUserCategoricalProperties != nil && pattern.PerUserEventCategoricalProperties != nil {
				(*pattern.PerUserEventCategoricalProperties).TrimByFmapSize(trimFraction)
				(*pattern.PerUserUserCategoricalProperties).TrimByFmapSize(trimFraction)
				(*pattern.PerOccurrenceEventCategoricalProperties).TrimByFmapSize(trimFraction)
				(*pattern.PerOccurrenceUserCategoricalProperties).TrimByFmapSize(trimFraction)
			}
		case 2:
			if pattern.PerUserUserNumericProperties != nil && pattern.PerUserEventNumericProperties != nil {
				(*pattern.PerUserEventNumericProperties).TrimByBinSize(trimFraction)
				(*pattern.PerUserUserNumericProperties).TrimByBinSize(trimFraction)
				(*pattern.PerOccurrenceEventNumericProperties).TrimByBinSize(trimFraction)
				(*pattern.PerOccurrenceUserNumericProperties).TrimByBinSize(trimFraction)
			}
		case 3:
			if pattern.PerUserEventCategoricalProperties != nil && pattern.PerUserUserCategoricalProperties != nil {
				(*pattern.PerUserEventCategoricalProperties).TrimByBinSize(trimFraction)
				(*pattern.PerUserUserCategoricalProperties).TrimByBinSize(trimFraction)
				(*pattern.PerOccurrenceEventCategoricalProperties).TrimByBinSize(trimFraction)
				(*pattern.PerOccurrenceUserCategoricalProperties).TrimByBinSize(trimFraction)
			}
		default:

		}
	}

	b, err := json.Marshal(pattern)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
		return nil, 0, err
	}
	pString := string(b)
	patternTrimBytes := int64(len([]byte(pString)))
	if patternTrimBytes > max_PATTERN_SIZE_IN_BYTES {
		mineLog.Infof("pattern size exceeds max Pattern size :%v", pattern.EventNames)
	}
	return pattern, patternTrimBytes, nil
}

func compressPatternsList(patterns []*P.Pattern, maxBytesSize int64, trimMap map[int]int64, trim_size int) ([]*P.Pattern, int64, error) {
	var totalBytes int64
	for _, pattern := range patterns {
		if _, pbytes, err := compressPattern(pattern, maxBytesSize, trimMap, trim_size); err != nil {
			return nil, 0, err
		} else {
			totalBytes += pbytes
		}
	}
	return patterns, totalBytes, nil
}

func cumulativeCompressPattern(pattern *P.Pattern, maxBytesSize int64, trimMap map[int]int64, trim_stage int) (*P.Pattern, int64, error) {
	var pBytes int64
	for i := 0; i <= trim_stage; i++ {
		_, pbytes, err := compressPattern(pattern, maxBytesSize, trimMap, trim_stage)
		if err != nil {
			return nil, 0, err
		}
		pBytes = pbytes
	}
	return pattern, pBytes, nil
}

func getPatternSize(pattern *P.Pattern) (int64, error) {
	b, err := json.Marshal(pattern)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
		return 0, err
	}
	pString := string(b)
	pBytes := int64(len([]byte(pString)))
	return pBytes, nil
}

// == user properties and all user histogram on beam

func UserPropertiesExecutor(s beam.Scope, scopeName string, config_ string,
	itrFname string, countsVersion int) {

	s = s.Scope(scopeName)
	cPatternsPcol := Read(s, itrFname)
	cPatternsPcolReshuffled := beam.Reshuffle(s, cPatternsPcol)
	beam.ParDo0(s, &UpThreadDoFn{Config: config_}, cPatternsPcolReshuffled)
}

type UpThreadDoFn struct {
	// UserPropertiesHistogram
	Config string
}

func (f *UpThreadDoFn) StartBundle(ctx context.Context) {
	beamlog.Info(ctx, "Initializing conf from CpThreadDoFn")
	var configVar *C.Configuration
	err := json.Unmarshal([]byte(f.Config), &configVar)
	if err != nil {
		log.Infof("Unable to unmarshall :%v", err)
	}
	InitConfBeam(configVar)
}

func (f *UpThreadDoFn) FinishBundle(ctx context.Context) {
	beamlog.Info(ctx, "Closing DB Connection from FinishBundle cpThreadDoFn")
	C.SafeFlushAllCollectors()
}

func (f *UpThreadDoFn) ProcessElement(ctx context.Context, cpString string) error {
	beamlog.Info(ctx, "Inside UserPropertiesHistogramThread")

	var cp CPatternsBeam
	var inBeam bool = true
	var debugHmine bool = true

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
	countsVersion := cp.CountsVersion
	hmineSupport := cp.HmineSupport
	log.Infof("Project Id : %d  ", projectID)
	log.Infof("Model Id : %d  ", modelId)
	log.Infof("BucketName: %s  ", bucketName)
	if countsVersion == 3 {
		log.Infof("counting algo : hmine")
	}

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
	pathEinfo = pathEinfo + "patterns_part/" + scopeName + "/"
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
	patterns, err = ReadPatternsGCP(ctx, projectID, modelId, scopeName, &cloudManager, UserAndEventsInfo, cp.LowNum, cp.HighNum)
	if err != nil {
		return fmt.Errorf("unable to read patterns from gcp")
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
	// create a temp folder for artifacts

	if inBeam {
		dname, err := ioutil.TempDir("", "fptree")
		if err != nil {
			return fmt.Errorf("unable to create tmp folder :%v", err)
		}
		for _, p := range patterns {
			p.PropertiesBaseFolder = dname
		}
		log.Infof("creating a base folder for artifacts:%s", dname)

		if countsVersion == 3 {

			baseFolder := dname
			basePathUserProps := path.Join(baseFolder, "fptree", "userProps")
			basePathEventProps := path.Join(baseFolder, "fptree", "eventProps")
			basePathUserPropsRes := path.Join(baseFolder, "fptree", "userPropsRes")
			basePathEventPropsRes := path.Join(baseFolder, "fptree", "eventPropsRes")

			err = os.MkdirAll(basePathUserPropsRes, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp user directory")
			}

			err = os.MkdirAll(basePathEventPropsRes, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp event directory")
			}

			err = os.MkdirAll(basePathUserProps, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp user directory")
			}

			err = os.MkdirAll(basePathEventProps, os.ModePerm)
			if err != nil {
				log.Fatal("unable to create temp event directory")
			}

			log.Infof("created tmp properties User folder : %s", basePathUserProps)
			log.Infof("created tmp properties Events folder : %s", basePathEventProps)
			log.Infof("created tmp properties User Results folder : %s", basePathUserPropsRes)
			log.Infof("created tmp properties Events Results folder : %s", basePathEventPropsRes)

		}
	}
	ComputeHistogramWorkerBeam(ctx, projectID, eventsFilePath, patterns, countOccurence, countsVersion, hmineSupport)

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

	err = writeFileToGCP(projectID, modelId, tmpPatternsFile, patternTmpFileLocal.Name(), &cloudManager, "")
	if err != nil {
		return fmt.Errorf("unable to write part file to GCP : %v", err)
	}

	err = deleteFile(patternTmpFileLocal.Name())
	if err != nil {
		return err
	}
	log.Infof("Delete patterns File :%s", patternTmpFileLocal.Name())

	if debugHmine {
		// upload all temp files generated if need to debug
		for _, pt := range patterns {

			up_local_path := filepath.Join(pt.PropertiesBaseFolder, "userProps", pt.PropertiesBasePath)
			if _, err := os.Stat(up_local_path); err == nil {
				if envFlag == "development" {
					fi, err := create(filepath.Join(modelpath, up_local_path))
					if err != nil {
						return err
					}
					fi.Close()
				}

				up_cloud_path := filepath.Join("patterns_part", "user_properties", pt.PropertiesBasePath)
				err := writeFileToGCP(projectID, modelId, up_cloud_path, up_local_path, &cloudManager, "")
				if err != nil {
					return fmt.Errorf("unable to upload tmp user properties file :%v", err)
				}

			} else if errors.Is(err, os.ErrNotExist) {
				// path/to/whatever does *not* exist
				log.Infof("userProperties file does not exist: %s", up_local_path)

			}

			ep_local_path := filepath.Join(pt.PropertiesBaseFolder, "eventProps", pt.PropertiesBasePath)

			if _, err := os.Stat(ep_local_path); err == nil {
				if envFlag == "development" {
					fi, err := create(filepath.Join(modelpath, ep_local_path))
					if err != nil {
						return err
					}
					fi.Close()
				}

				ep_cloud_path := filepath.Join("patterns_part", "event_properties", pt.PropertiesBasePath)
				err := writeFileToGCP(projectID, modelId, ep_cloud_path, ep_local_path, &cloudManager, "")
				if err != nil {
					return fmt.Errorf("unable to upload tmp user properties file :%v", err)
				}

			} else if errors.Is(err, os.ErrNotExist) {
				// path/to/whatever does *not* exist
				log.Infof("eventProperties file does not exist: %s", ep_local_path)

			}

		}

	}

	//--destroy events file---
	err = deleteFile(eventsTmpFile.Name())
	if err != nil {
		return err
	}
	log.Infof("Delete events File :%s", eventsTmpFile.Name())

	return nil
}

func WritePatternsBlocksToGCP(projectId int64, modelId uint64, cloudManager *filestore.FileManager,
	numRoutines, batchSize, countsVersion int, hmineSupport float32, countOccurence bool, numPatterns int, beam_env, bucketName,
	modelpath, scopeName, filepathString, efTmpPath string) (string, error) {

	ptInputPath := filepath.Join(efTmpPath, "pinput", scopeName+".txt")
	mineLog.Infof("File patterns name : %s", ptInputPath)
	pfFile, err := create(ptInputPath)
	mineLog.Info(err)
	if err != nil {
		return "", fmt.Errorf("unable to create file :%v", ptInputPath)
	}
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int64(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int64(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))

		t, err := json.Marshal(CPatternsBeam{
			projectId, modelId, low, high,
			filepathString, countOccurence, countsVersion, hmineSupport,
			bucketName, modelpath, scopeName, beam_env})
		if err != nil {
			return "", fmt.Errorf("unable to encode string : %v", err)
		}
		_, err = pfFile.WriteString(fmt.Sprintf("%s\n", t))
		if err != nil {
			return "", fmt.Errorf("error writing json to file :%v", err)
		}

	}

	err = pfFile.Close()
	if err != nil {
		return "", fmt.Errorf("error in closing file : %v", err)
	}

	mpath, _ := (*cloudManager).GetModelEventInfoFilePathAndName(projectId, modelId)
	path := filepath.Join(mpath, "patterns_part", scopeName)
	fileNameIntermediate := strings.Join([]string{scopeName, "itm_patterns.txt"}, "_")
	err = writeFileToGCP(projectId, modelId, fileNameIntermediate, ptInputPath, cloudManager, path)
	if err != nil {
		return "", fmt.Errorf("unable to write to cloud : %s", fileNameIntermediate)
	}

	return fileNameIntermediate, nil
}

func WritePatternsToGCP(project_id int64, model_id uint64, scope string, patterns []*P.Pattern,
	cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver) error {

	artifacts_path := (diskManager).GetModelArtifactsPath(project_id, model_id)
	scope_path := filepath.Join(artifacts_path, scope, "patterns.txt")

	pattern_file, err := create(scope_path)
	mineLog.Info(err)
	if err != nil {
		return fmt.Errorf("unable to create file :%v", scope_path)
	}

	for idx, p := range patterns {

		t, err := json.Marshal(PatterNamesIdx{int64(idx), p.EventNames, p.Segment})
		if err != nil {
			return fmt.Errorf("unable to encode string : %v", err)
		}
		_, err = pattern_file.WriteString(fmt.Sprintf("%s\n", t))
		if err != nil {
			return fmt.Errorf("error writing json to file :%v", err)
		}
	}

	err = pattern_file.Close()
	if err != nil {
		return fmt.Errorf("error in closing file : %v", err)
	}

	artifacts_gcp_path := (*cloudManager).GetModelArtifactsPath(project_id, model_id)
	patterns_gcp_path := filepath.Join(artifacts_gcp_path, scope)
	file_name := "patterns.txt"
	err = writeFileToGCP(project_id, model_id, file_name, scope_path, cloudManager, patterns_gcp_path)
	if err != nil {
		return fmt.Errorf("unable to write to cloud : %s/%s", patterns_gcp_path, file_name)
	}

	return nil
}

func ReadPatternsGCP(ctx context.Context, project_id int64, model_id uint64, scope string,
	cloudManager *filestore.FileManager, userAndEventsInfo *P.UserAndEventsInfo,
	low int64, high int64) ([]*P.Pattern, error) {

	patterns := make([]*P.Pattern, 0)
	artifacts_path := (*cloudManager).GetModelArtifactsPath(project_id, model_id)
	patterns_gcp_path := filepath.Join(artifacts_path, scope)
	rc, err := (*cloudManager).Get(patterns_gcp_path, "patterns.txt")
	if err != nil {
		return nil, err
	}
	beamlog.Debugf(ctx, "Reading patterns from :%s,%s", patterns_gcp_path, "patterns.txt")
	scannerEvents := bufio.NewScanner(rc)
	buf := make([]byte, 0, 64*1024)
	scannerEvents.Buffer(buf, 1024*1024)
	for scannerEvents.Scan() {
		var ptx PatterNamesIdx
		line := scannerEvents.Text()
		err := json.Unmarshal([]byte(line), &ptx)
		if err != nil {
			return nil, err
		}

		if (ptx.Id >= low) && (ptx.Id < high) {
			pt, err := P.NewPattern(ptx.Pname, userAndEventsInfo)
			if err != nil {
				return nil, err
			}
			pt.Segment = ptx.Segment
			patterns = append(patterns, pt)
		}

	}
	err = rc.Close()
	if err != nil {
		return nil, err
	}
	beamlog.Infof(ctx, "-----------------Total number of patterns read ---%d--------------%", len(patterns))
	return patterns, nil

}

func GenLenOneV2(jb M.ExplainV2Query, us *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	job_events_map := make(map[string]bool, 0)
	patterns := make([]*P.Pattern, 0)

	job_events_map[jb.Query.StartEvent] = true
	job_events_map[jb.Query.EndEvent] = true
	for _, e := range jb.Query.Rule.IncludedEvents {
		job_events_map[e] = true
	}
	for k, _ := range job_events_map {
		ev := make([]string, 0)
		ev = append(ev, k)
		p, err := P.NewPattern(ev, us)
		if err != nil {
			return nil, err
		}
		p.Segment = 0
		patterns = append(patterns, p)

	}
	return patterns, nil
}

func CreateCombinationEventsV2(jb M.ExplainV2Query, us *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	patterns := make([]*P.Pattern, 0)
	events_to_count := make([]string, 2)
	events_to_count[0] = jb.Query.StartEvent
	events_to_count[1] = jb.Query.EndEvent
	p, err := P.NewPattern(events_to_count, us)
	if err != nil {
		return nil, err
	}
	p.Segment = 1 // have to set to 1 for explain v2 job
	patterns = append(patterns, p)
	return patterns, nil
}

func GenThreeLenEventsV2(jb M.ExplainV2Query, us *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	patterns := make([]*P.Pattern, 0)
	for _, e := range jb.Query.Rule.IncludedEvents {
		events_to_count := make([]string, 3)
		events_to_count[0] = jb.Query.StartEvent
		events_to_count[1] = e
		events_to_count[2] = jb.Query.EndEvent
		p, err := P.NewPattern(events_to_count, us)
		if err != nil {
			return nil, err
		}
		p.Segment = 0
		patterns = append(patterns, p)
	}
	return patterns, nil
}

func CreateFourLenEventsV2(jb M.ExplainV2Query, us *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	patterns := make([]*P.Pattern, 0)

	for _, e1 := range jb.Query.Rule.IncludedEvents {
		for _, e2 := range jb.Query.Rule.IncludedEvents {
			if e1 != e2 {
				events_to_count := make([]string, 4)
				events_to_count[0] = jb.Query.StartEvent
				events_to_count[1] = e1
				events_to_count[2] = e2
				events_to_count[3] = jb.Query.EndEvent
				p1, err := P.NewPattern(events_to_count, us)
				if err != nil {
					return nil, err
				}
				p1.Segment = 0
				patterns = append(patterns, p1)
			}

		}
	}

	for _, p := range patterns {
		mineLog.Infof("four len patterns :%v", p.EventNames)
	}
	return patterns, nil
}

// func GetPatternsToCompute(cloudManager *filestore.FileManager, cpString string) ([]string, error) {

// 	var cp CPatternsBeam
// 	pattern_names := make([]string, 0)
// 	err := json.Unmarshal([]byte(cpString), &cp)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to unmarshall string in processElement :%s", cpString)
// 	}

// 	return nil
// }
