package merge

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	P "factors/pattern"
	"factors/pull"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// read daily data files from archiveCloudManager between startTimestamp and endTimestamp, merge all and create merged file in tmpCloudManager, then sort the merged file and create sorted file in cloudManager
//
// dataType - events, ad_reports or users,
// sortOnGroup - decides which id to get for events dataType(0:uid, i:group_i_id),
// channelOrDatefield - channel name for ad_reports dataType and dateField for users dataType,
// hardPull - whether to create file if already present
// beamConfig - pass beam configs to sort using beam
func MergeAndWriteSortedFile(projectId int64, dataType, channelOrDatefield string, startTimestamp, endTimestamp int64,
	archiveCloudManager, tmpCloudManager, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, beamConfig *RunBeamConfig, hardPull bool, sortOnGroup int) error {

	if dataType != U.DataTypeEvent && channelOrDatefield == "" {
		return fmt.Errorf("empty channelOrDatefield")
	}
	cloud_dir, cloud_name := GetSortedFilePathAndName(*cloudManager, dataType, channelOrDatefield, projectId, startTimestamp, endTimestamp, sortOnGroup)
	if !hardPull {
		if yes, _ := U.CheckFileExists(cloudManager, cloud_dir, cloud_name); yes {
			log.Info(cloud_name + " already exists")
			return nil
		}
	}

	var fileNamePrefix string
	if dataType == U.DataTypeEvent {
		fileNamePrefix = "events"
	} else {
		fileNamePrefix = channelOrDatefield
	}

	unsorted_cloud_dir, _ := GetSortedFilePathAndName(*tmpCloudManager, dataType, channelOrDatefield, projectId, startTimestamp, endTimestamp, sortOnGroup)
	unsortedCloudName := "unsorted_" + cloud_name

	useBeam := beamConfig.RunOnBeam
	userIdsMap, linesCount, err := ReadAndMergeDailyFiles(projectId, dataType, fileNamePrefix, unsorted_cloud_dir, unsortedCloudName, startTimestamp, endTimestamp, archiveCloudManager, tmpCloudManager, useBeam, sortOnGroup)
	if err != nil {
		return err
	}
	if linesCount == 0 {
		return nil
	}
	log.Info("merged")

	fileSize, err := (*tmpCloudManager).GetObjectSize(unsorted_cloud_dir, unsortedCloudName)
	if err != nil {
		log.Error("Failed to get file size.")
		return err
	}
	log.Infof("merged file size: %d", fileSize)
	var minSizeForBeam int64 = 500 * 1024 * 1024
	if (fileSize > minSizeForBeam && useBeam) && (dataType == U.DataTypeEvent || dataType == U.DataTypeAdReport) {
		batch_size := beamConfig.NumWorker
		var numLines int64
		if dataType == U.DataTypeEvent {
			err = Pull_events_beam_controller(projectId, unsorted_cloud_dir, unsortedCloudName, batch_size, userIdsMap,
				beamConfig, tmpCloudManager, diskManager, startTimestamp, endTimestamp, sortOnGroup)
			if err != nil {
				return err
			}

			numLines, err = ReadAndMergeEventPartFiles(projectId, unsorted_cloud_dir, cloud_dir, cloud_name, tmpCloudManager, cloudManager)
			if err != nil {
				log.Errorf("unable to read part files and merge :%v", err)
				return err
			}
		} else if dataType == U.DataTypeAdReport {
			err = Pull_channel_beam_controller(projectId, channelOrDatefield, unsorted_cloud_dir, unsortedCloudName, batch_size, userIdsMap,
				beamConfig, tmpCloudManager, diskManager, startTimestamp, endTimestamp)
			if err != nil {
				return err
			}

			numLines, err = ReadAndMergeChannelPartFiles(projectId, unsorted_cloud_dir, cloud_dir, cloud_name, tmpCloudManager, cloudManager)
			if err != nil {
				log.Errorf("unable to read part files and merge :%v", err)
				return err
			}
		}
		log.Info("numLines: ", numLines)
	} else {
		if err := SortUnsortedFile(tmpCloudManager, cloudManager, dataType, unsorted_cloud_dir, unsortedCloudName, cloud_dir, cloud_name, sortOnGroup); err != nil {
			return err
		}
	}
	log.Info("sorted")

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
func ReadAndMergeDailyFiles(projectId int64, dataType, fileNamePrefix string, unsortedDir, unsortedFileName string, startTimestamp, endTimestamp int64,
	archiveCloudManager, tmpCloudManager *filestore.FileManager, getIdMap bool, sortOnGroup int) (map[string]int, int64, error) {
	userIdMap := make(map[string]int)
	var countLines int64
	cloudWriter, err := (*tmpCloudManager).GetWriter(unsortedDir, unsortedFileName)
	if err != nil {
		log.Errorf("unable to get writer for file:%s", unsortedDir+unsortedFileName)
		return userIdMap, 0, err
	}

	log.Infof("Merging dailyFiles:%d", projectId)
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
				return userIdMap, 0, err
			}

			scanner := bufio.NewScanner(file)
			const maxCapacity = 30 * 1024 * 1024
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			var getUids bool
			var getDoctypes bool
			if getIdMap {
				if dataType == U.DataTypeEvent {
					getUids = true
				} else if dataType == U.DataTypeAdReport {
					getDoctypes = true
				}
			}
			for scanner.Scan() {
				line := scanner.Text()

				if getUids {
					var event *P.CounterEventFormat
					err = json.Unmarshal([]byte(line), &event)
					if err != nil {
						return userIdMap, countLines, err
					}
					userID := getAptId(event, sortOnGroup)
					if _, ok := userIdMap[userID]; !ok {
						userIdMap[userID] = 1
					}
				}

				if getDoctypes {
					var doc *pull.CounterCampaignFormat
					err = json.Unmarshal([]byte(line), &doc)
					if err != nil {
						return userIdMap, countLines, err
					}
					userID := fmt.Sprintf("%d", doc.Doctype)
					if _, ok := userIdMap[userID]; !ok {
						userIdMap[userID] = 1
					}
				}

				_, err = io.WriteString(cloudWriter, line+"\n")
				if err != nil {
					log.Errorf("Unable to write to file :%v", err)
					return userIdMap, countLines, err
				}
				countLines += 1
			}
		}
		timestamp += U.Per_day_epoch
	}
	err = cloudWriter.Close()

	return userIdMap, countLines, err
}

// read events file locally from filePath, sort events based on [sortOnGroup,timestamp] (sortOnGroup - 0:uid, i:group_i_id) and
// create sorted local file at sortedFilePath
func SortUnsortedEventsFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, unsortedDir, unsortedName string, sortedDir, sortedName string, sortOnGroup int) error {

	reader, err := (*tmpCloudManager).Get(unsortedDir, unsortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}

	scanner := bufio.NewScanner(reader)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var eventsList = make([]*P.CounterEventFormat, 0)

	for scanner.Scan() {
		line := scanner.Text()
		var rawEvent *P.CounterEventFormat
		err := json.Unmarshal([]byte(line), &rawEvent)
		if err != nil {
			return err
		}
		eventsList = append(eventsList, rawEvent)
	}
	err = reader.Close()
	if err != nil {
		return err
	}
	sort.Slice(eventsList, func(i, j int) bool {
		if getAptId(eventsList[i], sortOnGroup) == getAptId(eventsList[i], sortOnGroup) {
			return eventsList[i].EventTimestamp < eventsList[j].EventTimestamp
		}
		return getAptId(eventsList[i], sortOnGroup) < getAptId(eventsList[i], sortOnGroup)
	})
	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}

	for _, event := range eventsList {
		jsonBytes, err := json.Marshal(event)
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error marshalling")
			return err
		}
		_, err = io.WriteString(cloudWriter, fmt.Sprintf("%s\n", jsonBytes))
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error writing")
			return err
		}
	}
	err = cloudWriter.Close()

	return err
}

// read channel file locally from filePath, sort reports based on [docType,timestamp] and
// create sorted local file at sortedFilePath
func SortUnsortedChannelFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, unsortedDir, unsortedName string, sortedDir, sortedName string) error {

	reader, err := (*tmpCloudManager).Get(unsortedDir, unsortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}

	scanner := bufio.NewScanner(reader)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var docsList = make(map[int][]*pull.CounterCampaignFormat)

	for scanner.Scan() {
		line := scanner.Text()
		var docReport *pull.CounterCampaignFormat
		err := json.Unmarshal([]byte(line), &docReport)
		if err != nil {
			return err
		}
		if _, ok := docsList[docReport.Doctype]; !ok {
			docsList[docReport.Doctype] = make([]*pull.CounterCampaignFormat, 0)
		}
		docsList[docReport.Doctype] = append(docsList[docReport.Doctype], docReport)
	}
	err = reader.Close()
	if err != nil {
		return err
	}

	docTypes := make([]int, 0)
	for docType, docList := range docsList {
		docTypes = append(docTypes, docType)
		sort.Slice(docList, func(i, j int) bool {
			return docList[i].Timestamp < docList[j].Timestamp
		})
	}
	sort.Slice(docTypes, func(i, j int) bool {
		return docTypes[i] < docTypes[j]
	})

	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}

	for _, docType := range docTypes {
		docs := docsList[docType]
		for _, doc := range docs {
			jsonBytes, err := json.Marshal(doc)
			if err != nil {
				fmt.Println(err.Error())
				log.WithError(err).Error("Error marshalling")
				return err
			}
			_, err = io.WriteString(cloudWriter, fmt.Sprintf("%s\n", jsonBytes))
			if err != nil {
				fmt.Println(err.Error())
				log.WithError(err).Error("Error writing")
				return err
			}
		}
	}
	err = cloudWriter.Close()

	return err
}

// read users file locally from filePath, group all users with same id and set JoinTimestamp as the least of all JoinTimestamps,
// sort users based on JoinTimestamp and create sorted local file at sortedFilePath
func SortUnsortedUsersFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, unsortedDir, unsortedName string, sortedDir, sortedName string) error {

	reader, err := (*tmpCloudManager).Get(unsortedDir, unsortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}

	scanner := bufio.NewScanner(reader)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var uniqueUsersList = make([]*pull.CounterUserFormat, 0)
	//grouping logic (get one entry for each user with min joinTimestamp)
	{
		uniqueUsersJoints := make(map[string]int64)
		var usersList = make([]*pull.CounterUserFormat, 0)
		for scanner.Scan() {
			line := scanner.Text()
			var rawUser *pull.CounterUserFormat
			err := json.Unmarshal([]byte(line), &rawUser)
			if err != nil {
				return err
			}
			if jt, ok := uniqueUsersJoints[rawUser.Id]; !ok || rawUser.JoinTimestamp < jt {
				uniqueUsersJoints[rawUser.Id] = rawUser.JoinTimestamp
			}
			usersList = append(usersList, rawUser)
		}
		for _, user := range usersList {
			if uniqueUsersJoints[user.Id] == user.JoinTimestamp {
				uniqueUsersList = append(uniqueUsersList, user)
			}
			delete(uniqueUsersJoints, user.Id)
		}
	}
	err = reader.Close()
	if err != nil {
		return err
	}

	sort.Slice(uniqueUsersList, func(i, j int) bool {
		return uniqueUsersList[i].JoinTimestamp < uniqueUsersList[j].JoinTimestamp
	})
	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}

	for _, user := range uniqueUsersList {
		jsonBytes, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error marshalling")
			return err
		}
		_, err = io.WriteString(cloudWriter, fmt.Sprintf("%s\n", jsonBytes))
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error writing")
			return err
		}
	}
	err = cloudWriter.Close()

	return err
}

// get the corresponding id from events based on num
//
// 0:uid, 1:group_1_id, 2:group_2_id, 3:group_3_id, 4:group_4_id
func getAptId(event *P.CounterEventFormat, num int) string {
	switch num {
	case 1:
		return event.Group1UserId
	case 2:
		return event.Group2UserId
	case 3:
		return event.Group3UserId
	case 4:
		return event.Group4UserId
	default:
		return event.UserId
	}
}

// sort file in unsortedFilePath based on dataType and created sorted file at sortedFilePath
//
// sortOnGroup - (0:uid, i:group_i_id) for events dataType
func SortUnsortedFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, dataType string, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName string, sortOnGroup int) error {
	var err error
	if dataType == U.DataTypeEvent {
		err = SortUnsortedEventsFile(tmpCloudManager, sortedCloudManager, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName, sortOnGroup)
	} else if dataType == U.DataTypeAdReport {
		err = SortUnsortedChannelFile(tmpCloudManager, sortedCloudManager, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName)
	} else if dataType == U.DataTypeUser {
		err = SortUnsortedUsersFile(tmpCloudManager, sortedCloudManager, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName)
	} else {
		err = fmt.Errorf("wrong dataType: %s", dataType)
	}
	return err
}

// get data file (sorted) path and name given fileManager and dataType
//
// channelOrDatefield - channel name for ad_reports dataType and dateField for users dataType,
// sortOnGroup - (0:uid, i:group_i_id) for events dataType
func GetSortedFilePathAndName(fileManager filestore.FileManager, dataType, channelOrDatefield string, projectId int64, startTime, endTime int64, sortOnGroup int) (string, string) {
	if dataType == U.DataTypeEvent {
		return fileManager.GetEventsGroupFilePathAndName(projectId, startTime, endTime, sortOnGroup)
	} else if dataType == U.DataTypeAdReport {
		channel := channelOrDatefield
		return fileManager.GetChannelFilePathAndName(channel, projectId, startTime, endTime)
	} else if dataType == U.DataTypeUser {
		dateField := channelOrDatefield
		return fileManager.GetUsersFilePathAndName(dateField, projectId, startTime, endTime)
	} else {
		log.Errorf("wrong dataType: %s", dataType)
	}
	return "", ""
}

// get data file (daily) path and name given fileManager and dataType
//
// channelOrDatefield - channel name for ad_reports dataType and dateField for users dataType,
// sortOnGroup - (0:uid, i:group_i_id) for events dataType
func GetDailyArchiveFilePathAndName(fileManager filestore.FileManager, dataType, channelOrDatefield string, projectId int64, dataTimestamp, startTime, endTime int64) (string, string) {
	if dataType == U.DataTypeEvent {
		return fileManager.GetDailyEventArchiveFilePathAndName(projectId, dataTimestamp, startTime, endTime)
	} else if dataType == U.DataTypeAdReport {
		return fileManager.GetDailyChannelArchiveFilePathAndName(channelOrDatefield, projectId, dataTimestamp, startTime, endTime)
	} else if dataType == U.DataTypeUser {
		return fileManager.GetDailyUsersArchiveFilePathAndName(channelOrDatefield, projectId, dataTimestamp, startTime, endTime)
	} else {
		log.Errorf("wrong dataType: %s", dataType)
	}
	return "", ""
}
