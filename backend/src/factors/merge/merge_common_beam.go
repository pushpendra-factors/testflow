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
	ProjectID     int64  `json:"pid"`
	LowNum        int64  `json:"ln"`
	HighNum       int64  `json:"hn"`
	FileTimeIndex int    `json:"fti"`
	BucketName    string `json:"bnm"`
	ScopeName     string `json:"ScopeName"`
	Env           string `json:"Env"`
	StartTime     int64  `json:"ts"`
	EndTime       int64  `json:"end"`
	Channel       string `json:"ch"`
	Group         int    `json:"gr"`
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

type fileNameSortInfo struct {
	fileName    string
	sortIndices []int64
}

const MAX_FILESIZE_PER_WORKER_FOR_BEAM = 2 * 1024 * 1024 * 1024
const MIN_FILESIZE_FOR_BEAM = 512 * 1024 * 1024

func SortDataOnBeam(projectId int64, dataType string, channel string, startIndexToFileInfoMap map[int][]*partFileInfo, indexMap map[string]int64, beamConfig *RunBeamConfig, sorted_cloud_dir, sorted_cloud_name string,
	tmpCloudManager, sortedCloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTime, endTime int64, sortOnGroup int) (int, error) {
	var err error
	var numLines int
	if dataType == U.DataTypeEvent {
		err := putUserIdsToGCP(projectId, indexMap, tmpCloudManager, diskManager, startTime, endTime, sortOnGroup)
		if err != nil {
			return 0, err
		}

		err = Pull_events_beam_controller(projectId, startIndexToFileInfoMap, beamConfig, tmpCloudManager, diskManager, startTime, endTime, sortOnGroup)
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
		err = Pull_channel_beam_controller(projectId, channel, startIndexToFileInfoMap, beamConfig, tmpCloudManager, diskManager, startTime, endTime)
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

// read daily data files from cloud between startTimestamp and endTimestamp, merge all and create merged part files in cloud to be distributed to each worker on beam, also
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
	archiveCloudManager, tmpCloudManager *filestore.FileManager, sortOnGroup int, numberOfUsersPerFile int) (map[int][]*partFileInfo, map[string]int64, int, error) {

	var countLines int
	var startIndexToFileInfoMap = make(map[int][]*partFileInfo)
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
		partFilesDir, _ := pull.GetDailyArchiveFilePathAndName(archiveCloudManager, dataType, fileNamePrefix, projectId, timestamp, 0, 0)
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
				return startIndexToFileInfoMap, indexMap, 0, err
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
						return startIndexToFileInfoMap, indexMap, countLines, err
					}
					userID := GetAptId(event, sortOnGroup)
					if _, ok := indexMap[userID]; !ok {
						indexMap[userID] = keyCount
						keyCount++
					}
					uidIndex := int(indexMap[userID])
					startIndex := (uidIndex / numberOfUsersPerFile) * numberOfUsersPerFile
					if _, ok := startIndexToFileInfoMap[startIndex]; !ok {
						endIndex := startIndex + numberOfUsersPerFile - 1
						cDir, cName := (*tmpCloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, startIndex, endIndex, sortOnGroup, 0)
						cloudWriter, err := (*tmpCloudManager).GetWriter(cDir, cName)
						if err != nil {
							log.WithFields(log.Fields{"fileDir": cDir, "fileName": cName}).Error("unable to get writer for file")
							return startIndexToFileInfoMap, indexMap, 0, err
						}
						startIndexToFileInfoMap[startIndex] = make([]*partFileInfo, 0)
						fileInfo := partFileInfo{name: cName, dir: cDir, startUserIndex: startIndex, endUserIndex: endIndex, fileTimestampIndex: 0, writer: &cloudWriter}
						startIndexToFileInfoMap[startIndex] = append(startIndexToFileInfoMap[startIndex], &fileInfo)
					}
					reqWriter = startIndexToFileInfoMap[startIndex][0].writer
				}

				if getDoctypes {
					var doc *pull.CounterCampaignFormat
					err = json.Unmarshal([]byte(line), &doc)
					if err != nil {
						return startIndexToFileInfoMap, indexMap, countLines, err
					}
					docType := doc.Doctype
					docTypeStr := fmt.Sprintf("%d", docType)
					if _, ok := indexMap[docTypeStr]; !ok {
						indexMap[docTypeStr] = keyCount
						keyCount++
					}
					docTypeIndex := int(indexMap[docTypeStr])
					if _, ok := startIndexToFileInfoMap[docTypeIndex]; !ok {
						cDir, cName := (*tmpCloudManager).GetChannelPartFilePathAndName(fileNamePrefix, projectId, startTimestamp, endTimestamp, false, docTypeIndex, 0)
						cloudWriter, err := (*tmpCloudManager).GetWriter(cDir, cName)
						if err != nil {
							log.WithFields(log.Fields{"fileDir": cDir, "fileName": cName}).Error("unable to get writer for file")
							return startIndexToFileInfoMap, indexMap, 0, err
						}
						startIndexToFileInfoMap[docTypeIndex] = make([]*partFileInfo, 0)
						fileInfo := partFileInfo{name: cName, dir: cDir, startUserIndex: docTypeIndex, endUserIndex: docTypeIndex, fileTimestampIndex: 0, writer: &cloudWriter}
						startIndexToFileInfoMap[docTypeIndex] = append(startIndexToFileInfoMap[docTypeIndex], &fileInfo)
					}
					reqWriter = startIndexToFileInfoMap[docTypeIndex][0].writer
				}

				_, err = io.WriteString(*reqWriter, line+"\n")
				if err != nil {
					log.Errorf("Unable to write to file :%v", err)
					return startIndexToFileInfoMap, indexMap, countLines, err
				}
				countLines += 1
			}
		}
		timestamp += U.Per_day_epoch
	}
	for _, fileInfoList := range startIndexToFileInfoMap {
		for _, fileInfo := range fileInfoList {
			writer := fileInfo.writer
			err := (*writer).Close()
			if err != nil {
				log.WithError(err).Errorf("Unable to close writer for: %s", fileInfo.name)
				return startIndexToFileInfoMap, indexMap, countLines, err
			}
		}
	}
	if dataType == U.DataTypeEvent {
		filesBroken, fileAdded, err := resizeUnsortedPartFilesForBeamByUid(projectId, startTimestamp, endTimestamp, sortOnGroup, startIndexToFileInfoMap, indexMap, tmpCloudManager)
		if err != nil {
			log.WithError(err).Error("error resizeUnsortedPartFilesForBeamByUid")
			return startIndexToFileInfoMap, indexMap, countLines, err
		}
		log.Infof("%d large files split into %d small files by user id", filesBroken, fileAdded)
		filesBroken, fileAdded, err = resizeUnsortedPartFilesForBeamByTimestamp(projectId, startTimestamp, endTimestamp, sortOnGroup, startIndexToFileInfoMap, tmpCloudManager)
		if err != nil {
			log.WithError(err).Error("error resizeUnsortedPartFilesForBeamByTimestamp")
			return startIndexToFileInfoMap, indexMap, countLines, err
		}
		log.Infof("%d large files split into %d small files by timestamp", filesBroken, fileAdded)
	}

	return startIndexToFileInfoMap, indexMap, countLines, nil
}

func resizeUnsortedPartFilesForBeamByUid(projectId, startTimestamp, endTimestamp int64, sortOnGroup int, startIndexToFileInfoMap map[int][]*partFileInfo, indexMap map[string]int64, cloudManager *filestore.FileManager) (int, int, error) {
	noOfFilesBroken := 0
	noOfFilesAdded := 0
	addedStartToEndEndex := make(map[int][]*partFileInfo)
	for start, fileInfoList := range startIndexToFileInfoMap {
		if len(fileInfoList) != 1 {
			continue
		}
		fileInfo := fileInfoList[0]
		end := fileInfo.endUserIndex
		unsortedPartFilesDir, unsortedPartFilesName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, start, end, sortOnGroup, 0)
		fileSize, err := (*cloudManager).GetObjectSize(unsortedPartFilesDir, unsortedPartFilesName)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": unsortedPartFilesDir, "fileName": unsortedPartFilesName}).Error("failed to get file size")
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
				cDir, cName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, startIndex, endIndex, sortOnGroup, 0)
				cloudWriter, err := (*cloudManager).GetWriter(cDir, cName)
				if err != nil {
					log.WithFields(log.Fields{"fileDir": cDir, "fileName": cName}).Error("unable to get writer for file")
					return noOfFilesBroken, noOfFilesAdded, err
				}
				fileInfo := partFileInfo{name: cName, dir: cDir, startUserIndex: startIndex, endUserIndex: endIndex, fileTimestampIndex: fileInfo.fileTimestampIndex, writer: &cloudWriter}
				addedStartToEndEndex[startIndex] = make([]*partFileInfo, 0)
				addedStartToEndEndex[startIndex] = append(addedStartToEndEndex[startIndex], &fileInfo)
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
				userID := GetAptId(event, sortOnGroup)
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
		broken, added, err := resizeUnsortedPartFilesForBeamByUid(projectId, startTimestamp, endTimestamp, sortOnGroup, addedStartToEndEndex, indexMap, cloudManager)
		noOfFilesAdded = noOfFilesAdded + added - broken
		if err != nil {
			log.WithError(err).Errorf("error resizeUnsortedPartFilesForBeam")
			return noOfFilesBroken, noOfFilesAdded, err
		}
		for start, fileInfoList := range addedStartToEndEndex {
			startIndexToFileInfoMap[start] = fileInfoList
		}
	}
	return noOfFilesBroken, noOfFilesAdded, nil
}

func resizeUnsortedPartFilesForBeamByTimestamp(projectId, startTimestamp, endTimestamp int64, sortOnGroup int, startIndexToFileInfoMap map[int][]*partFileInfo, cloudManager *filestore.FileManager) (int, int, error) {
	noOfFilesBroken := 0
	noOfFilesAdded := 0
	addedStartToEndEndex := make(map[int][]*partFileInfo)
	for start, fileInfoList := range startIndexToFileInfoMap {
		if len(fileInfoList) != 1 {
			continue
		}
		fileInfo := fileInfoList[0]
		end := fileInfo.endUserIndex
		unsortedPartFilesDir, unsortedPartFilesName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, start, end, sortOnGroup, 0)
		fileSize, err := (*cloudManager).GetObjectSize(unsortedPartFilesDir, unsortedPartFilesName)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": unsortedPartFilesDir, "fileName": unsortedPartFilesName}).Error("failed to get file size")
			return noOfFilesBroken, noOfFilesAdded, err
		}
		if fileSize > MAX_FILESIZE_PER_WORKER_FOR_BEAM {

			addedStartToEndEndex[start] = make([]*partFileInfo, 0)

			noOfSplits := (int(fileSize) / MAX_FILESIZE_PER_WORKER_FOR_BEAM) + 1
			rangeAfterSplit := int64(int(endTimestamp+1-startTimestamp) / noOfSplits)

			var splitToFileMap = make(map[int]*io.WriteCloser)
			startTime := startTimestamp
			for i := 0; i < noOfSplits; i++ {
				var endTime int64
				if i == noOfSplits-1 {
					endTime = endTimestamp
				} else {
					endTime = startTime + rangeAfterSplit - 1
				}
				cDir, cName := (*cloudManager).GetEventsPartFilePathAndName(projectId, startTimestamp, endTimestamp, false, start, end, sortOnGroup, i)
				cloudWriter, err := (*cloudManager).GetWriter(cDir, cName)
				if err != nil {
					log.WithFields(log.Fields{"fileDir": cDir, "fileName": cName}).Error("unable to get writer for file")
					return noOfFilesBroken, noOfFilesAdded, err
				}
				partInfo := partFileInfo{name: cName, dir: cDir, startUserIndex: start, endUserIndex: end, fileTimestampIndex: i, writer: &cloudWriter}
				addedStartToEndEndex[start] = append(addedStartToEndEndex[start], &partInfo)
				splitToFileMap[i] = &cloudWriter
				startTime = endTime + 1
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
				splitNum := int(event.EventTimestamp-startTimestamp) / int(rangeAfterSplit)
				if splitNum == noOfSplits {
					splitNum = noOfSplits - 1
				}
				if reqWriter, ok := splitToFileMap[splitNum]; !ok {
					log.WithFields(log.Fields{"eventTimetamp": event.EventTimestamp, "split": splitNum, "numberOfSplits": noOfSplits}).Error("required writer not found")
					continue
				} else {
					_, err = io.WriteString(*reqWriter, line+"\n")
					if err != nil {
						log.Errorf("Unable to write to file :%v", err)
						return noOfFilesBroken, noOfFilesAdded, err
					}
				}
			}
			for i, writer := range splitToFileMap {
				err := (*writer).Close()
				if err != nil {
					log.WithError(err).Errorf("Unable to close writer with index: %d-%d, split-%d", start, end, i)
					return noOfFilesBroken, noOfFilesAdded, err
				}
				noOfFilesAdded++
			}
			noOfFilesBroken++
		}
	}

	if noOfFilesBroken > 0 {
		for start, fileInfoList := range addedStartToEndEndex {
			startIndexToFileInfoMap[start] = fileInfoList
		}
	}
	return noOfFilesBroken, noOfFilesAdded, nil
}
