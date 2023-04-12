package merge

import (
	"bufio"
	"context"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	P "factors/pattern"
	"factors/pull"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io"
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

const MAX_FILESIZE_PER_WORKER_FOR_BEAM = 2 * 1024 * 1024 * 1024
const MIN_FILESIZE_FOR_BEAM = 512 * 1024 * 1024

func SortDataOnBeam(projectId int64, dataType string, channel string, startToEndIndex map[int]int, indexMap map[string]int64, beamConfig *RunBeamConfig, sorted_cloud_dir, sorted_cloud_name string,
	tmpCloudManager, sortedCloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTime, endTime int64, sortOnGroup int) (int, error) {
	var err error
	var numLines int
	if dataType == U.DataTypeEvent {
		err := putUserIdsToGCP(projectId, indexMap, tmpCloudManager, diskManager, startTime, endTime, sortOnGroup)
		if err != nil {
			return 0, err
		}

		err = Pull_events_beam_controller(projectId, startToEndIndex, beamConfig, tmpCloudManager, diskManager, startTime, endTime, sortOnGroup)
		if err != nil {
			return 0, err
		}

		partFilesDir := (*tmpCloudManager).GetEventsPartFilesDir(projectId, startTime, endTime, true, sortOnGroup)
		numLines, err = ReadAndMergeEventPartFiles(projectId, partFilesDir, sorted_cloud_dir, sorted_cloud_name, tmpCloudManager, sortedCloudManager)
		if err != nil {
			log.WithError(err).Error("unable to read part files and merge")
			return numLines, err
		}
	} else if dataType == U.DataTypeAdReport {
		err := putDoctypesToGCP(channel, projectId, indexMap, tmpCloudManager, diskManager, startTime, endTime)
		if err != nil {
			return 0, err
		}
		err = Pull_channel_beam_controller(projectId, channel, startToEndIndex, beamConfig, tmpCloudManager, diskManager, startTime, endTime)
		if err != nil {
			return 0, err
		}

		partFilesDir := (*tmpCloudManager).GetChannelPartFilesDir(channel, projectId, startTime, endTime, true)
		numLines, err = ReadAndMergeChannelPartFiles(projectId, partFilesDir, sorted_cloud_dir, sorted_cloud_name, tmpCloudManager, sortedCloudManager)
		if err != nil {
			log.Errorf("unable to read part files and merge :%v", err)
			return numLines, err
		}
	}
	return numLines, err
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

// read daily data files from cloud between startTimestamp and endTimestamp, merge all and create merged file locally at local_dir+unsortedFileName, also
// get userIdMap,countLines as output
//
// OUTPUTS
// userIdMap - (all ids present in merged file as keys and value 1) for use when sorting in beam afterwards (get docTypes as keys for ad_reports)
// countLines - number of lines in merged file
//
// INPUTS
// dataType - events, ad_reports or users,
// sortOnGroup - decides which id to get for events dataType(0:uid, i:group_i_id),
// fileNamePrefix - all daily files only with this prefix are read (give "events" for events, channel name for ad_reports and dateField for users),
// getIdMap - whether to generate userIdMap
func ReadAndMergeDailyFilesForBeam(projectId int64, dataType, fileNamePrefix string, startTimestamp, endTimestamp int64,
	archiveCloudManager, tmpCloudManager *filestore.FileManager, sortOnGroup int, numberOfUsersPerFile int) (map[int]int, map[string]int64, int, error) {

	var countLines int
	var startIndexToWriterMap = make(map[int]*io.WriteCloser)
	var startToEndIndex = make(map[int]int)
	var indexMap = make(map[string]int64)
	var keyCount int64 = 0
	var getUids bool
	var getDoctypes bool

	if dataType == U.DataTypeEvent {
		getUids = true
	} else if dataType == U.DataTypeAdReport {
		getDoctypes = true
	}

	log.Infof("Merging dailyFiles for project: %d", projectId)
	timestamp := startTimestamp
	for timestamp <= endTimestamp {
		partFilesDir, _ := GetDailyArchiveFilePathAndName(*archiveCloudManager, dataType, fileNamePrefix, projectId, timestamp, 0, 0)
		listFiles := (*archiveCloudManager).ListFiles(partFilesDir)
		for _, partFileFullName := range listFiles {
			partFNamelist := strings.Split(partFileFullName, "/")
			partFileName := partFNamelist[len(partFNamelist)-1]
			if !strings.HasPrefix(partFileName, fileNamePrefix) {
				continue
			}

			log.Infof("Reading daily file :%s, %s", partFilesDir, partFileName)
			file, err := (*archiveCloudManager).Get(partFilesDir, partFileName)
			if err != nil {
				return startToEndIndex, indexMap, 0, err
			}

			scanner := bufio.NewScanner(file)
			const maxCapacity = 30 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			for scanner.Scan() {
				line := scanner.Text()
				var reqWriter *io.WriteCloser

				if getUids {
					var event *P.CounterEventFormat
					err = json.Unmarshal([]byte(line), &event)
					if err != nil {
						return startToEndIndex, indexMap, countLines, err
					}
					userID := getAptId(event, sortOnGroup)
					if _, ok := indexMap[userID]; !ok {
						indexMap[userID] = keyCount
						keyCount++
					}
					uidIndex := int(indexMap[userID])
					startIndex := (uidIndex / numberOfUsersPerFile) * numberOfUsersPerFile
					if _, ok := startIndexToWriterMap[startIndex]; !ok {
						endIndex := startIndex + numberOfUsersPerFile - 1
						startToEndIndex[startIndex] = endIndex
						cDir, cName := (*tmpCloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, startIndex, endIndex, sortOnGroup)
						cloudWriter, err := (*tmpCloudManager).GetWriter(cDir, cName)
						if err != nil {
							log.Errorf("unable to get writer for file:%s", cDir+cName)
							return startToEndIndex, indexMap, 0, err
						}
						startIndexToWriterMap[startIndex] = &cloudWriter
					}
					reqWriter = startIndexToWriterMap[startIndex]
				}

				if getDoctypes {
					var doc *pull.CounterCampaignFormat
					err = json.Unmarshal([]byte(line), &doc)
					if err != nil {
						return startToEndIndex, indexMap, countLines, err
					}
					docType := doc.Doctype
					docTypeStr := fmt.Sprintf("%d", docType)
					if _, ok := indexMap[docTypeStr]; !ok {
						indexMap[docTypeStr] = keyCount
						keyCount++
					}
					docTypeIndex := int(indexMap[docTypeStr])
					if _, ok := startIndexToWriterMap[docTypeIndex]; !ok {
						startToEndIndex[docTypeIndex] = docTypeIndex
						cDir, cName := (*tmpCloudManager).GetChannelPartFilePathAndName(fileNamePrefix, projectId, startTimestamp, endTimestamp, false, docTypeIndex)
						cloudWriter, err := (*tmpCloudManager).GetWriter(cDir, cName)
						if err != nil {
							log.Errorf("unable to get writer for file:%s", cDir+cName)
							return startToEndIndex, indexMap, 0, err
						}
						startIndexToWriterMap[docTypeIndex] = &cloudWriter
					}
					reqWriter = startIndexToWriterMap[docTypeIndex]
				}

				_, err = io.WriteString(*reqWriter, line+"\n")
				if err != nil {
					log.Errorf("Unable to write to file :%v", err)
					return startToEndIndex, indexMap, countLines, err
				}
				countLines += 1
			}
		}
		timestamp += U.Per_day_epoch
	}
	for i, writer := range startIndexToWriterMap {
		err := (*writer).Close()
		if err != nil {
			log.WithError(err).Errorf("Unable to close writer with index: %d-%d", i, startToEndIndex[i])
			return startToEndIndex, indexMap, countLines, err
		}
	}
	filesBroken, fileAdded, err := resizeUnsortedPartFilesForBeam(projectId, startTimestamp, endTimestamp, sortOnGroup, startToEndIndex, indexMap, tmpCloudManager)
	if err != nil {
		log.WithError(err).Error("error resizeUnsortedPartFilesForBeam")
		return startToEndIndex, indexMap, countLines, err
	}
	log.Infof("%d large files split into %d small files", filesBroken, fileAdded)

	return startToEndIndex, indexMap, countLines, nil
}

func resizeUnsortedPartFilesForBeam(projectId, startTimestamp, endTimestamp int64, sortOnGroup int, startToEndIndex map[int]int, indexMap map[string]int64, cloudManager *filestore.FileManager) (int, int, error) {
	noOfFilesBroken := 0
	noOfFilesAdded := 0
	addedStartToEndEndex := make(map[int]int)
	for start, end := range startToEndIndex {
		unsortedPartFilesDir, unsortedPartFilesName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, start, end, sortOnGroup)
		fileSize, err := (*cloudManager).GetObjectSize(unsortedPartFilesDir, unsortedPartFilesName)
		if err != nil {
			log.Errorf("Failed to get file size: %s %s", unsortedPartFilesDir, unsortedPartFilesName)
			return noOfFilesBroken, noOfFilesAdded, err
		}
		if fileSize > MAX_FILESIZE_PER_WORKER_FOR_BEAM {

			noOfSplits := (int(fileSize) / MAX_FILESIZE_PER_WORKER_FOR_BEAM) + 1
			noOfUidsAfterSplit := int(end+1-start) / noOfSplits

			if noOfUidsAfterSplit == 0 {
				noOfSplits = int(end + 1 - start)
				noOfUidsAfterSplit = 1
				if noOfSplits == 1 {
					continue
				}
			}

			var splitToFileMap = make(map[int]*io.WriteCloser)
			startIndex := start
			for i := 0; i < noOfSplits; i++ {
				var endIndex int
				if i == noOfSplits-1 {
					endIndex = end
				} else {
					endIndex = startIndex + noOfUidsAfterSplit - 1
				}
				addedStartToEndEndex[startIndex] = endIndex
				cDir, cName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, startIndex, endIndex, sortOnGroup)
				cloudWriter, err := (*cloudManager).GetWriter(cDir, cName)
				if err != nil {
					log.Errorf("unable to get writer for file:%s", cDir+cName)
					return noOfFilesBroken, noOfFilesAdded, err
				}
				splitToFileMap[i] = &cloudWriter
				startIndex = endIndex + 1
			}

			file, err := (*cloudManager).Get(unsortedPartFilesDir, unsortedPartFilesName)
			if err != nil {
				return noOfFilesBroken, noOfFilesAdded, err
			}
			scanner := bufio.NewScanner(file)
			const maxCapacity = 30 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)
			for scanner.Scan() {
				line := scanner.Text()
				var event *P.CounterEventFormat
				err = json.Unmarshal([]byte(line), &event)
				if err != nil {
					return noOfFilesBroken, noOfFilesAdded, err
				}
				userID := getAptId(event, sortOnGroup)
				uidIndex := int(indexMap[userID])
				splitNum := (uidIndex - start) / noOfUidsAfterSplit
				if splitNum == noOfSplits {
					splitNum = noOfSplits - 1
				}
				reqWriter := splitToFileMap[splitNum]
				_, err = io.WriteString(*reqWriter, line+"\n")
				if err != nil {
					log.Errorf("Unable to write to file :%v", err)
					return noOfFilesBroken, noOfFilesAdded, err
				}
			}
			for i, writer := range splitToFileMap {
				err := (*writer).Close()
				if err != nil {
					startIndex := start + i*noOfUidsAfterSplit
					log.WithError(err).Errorf("Unable to close writer with index: %d-%d", startIndex, addedStartToEndEndex[startIndex])
					return noOfFilesBroken, noOfFilesAdded, err
				}
				noOfFilesAdded++
			}
			noOfFilesBroken++
		}
	}
	if noOfFilesBroken > 0 {
		broken, added, err := resizeUnsortedPartFilesForBeam(projectId, startTimestamp, endTimestamp, sortOnGroup, addedStartToEndEndex, indexMap, cloudManager)
		noOfFilesAdded = noOfFilesAdded + added - broken
		if err != nil {
			log.WithError(err).Errorf("error resizeUnsortedPartFilesForBeam")
			return noOfFilesBroken, noOfFilesAdded, err
		}
		for start, end := range addedStartToEndEndex {
			startToEndIndex[start] = end
		}
	}
	return noOfFilesBroken, noOfFilesAdded, nil
}
