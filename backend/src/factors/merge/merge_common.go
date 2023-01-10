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
	"os"
	"path"
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

	sorted_dir, sorted_name := GetSortedFilePathAndName(diskManager, dataType, channelOrDatefield, projectId, startTimestamp, endTimestamp, sortOnGroup)
	sortedFilePath := sorted_dir + sorted_name
	unsortedFileName := "unsorted_" + sorted_name
	unsortedFileDir := sorted_dir
	unsortedFilePath := unsortedFileDir + unsortedFileName

	useBeam := beamConfig.RunOnBeam
	userIdsMap, linesCount, err := ReadAndMergeDailyFiles(projectId, dataType, fileNamePrefix, unsortedFileName, startTimestamp, endTimestamp, sorted_dir, archiveCloudManager, diskManager, useBeam, sortOnGroup)
	if err != nil {
		return err
	}
	if linesCount == 0 {
		return nil
	}
	log.Info("merged")

	if useBeam && (dataType == U.DataTypeEvent || dataType == U.DataTypeAdReport) {
		file, err := diskManager.Get(unsortedFileDir, unsortedFileName)
		if err != nil {
			return err
		}
		defer file.Close()

		unsorted_cloud_dir, _ := GetSortedFilePathAndName(*tmpCloudManager, dataType, channelOrDatefield, projectId, startTimestamp, endTimestamp, sortOnGroup)
		unsortedCloudName := "unsorted_" + cloud_name
		err = (*tmpCloudManager).Create(unsorted_cloud_dir, unsortedCloudName, file)
		if err != nil {
			log.Error("Failed to create file to cloud.")
			return err
		}
		log.Info("uploaded")

		err = os.Remove(unsortedFilePath)
		if err != nil {
			return err
		}
		batch_size := beamConfig.NumWorker
		var numLines int64
		if dataType == U.DataTypeEvent {
			err = Pull_events_beam_controller(projectId, unsorted_cloud_dir, unsortedCloudName, batch_size, userIdsMap,
				beamConfig, tmpCloudManager, diskManager, startTimestamp, endTimestamp, sortOnGroup)
			if err != nil {
				return err
			}

			numLines, err = ReadAndMergeEventPartFiles(projectId, cloud_dir, sorted_dir, sorted_name, tmpCloudManager, diskManager)
			if err != nil {
				log.Errorf("unable to read part files and merge :%v", err)
				return err
			}
		} else if dataType == U.DataTypeAdReport {
			err = Pull_channel_beam_controller(projectId, channelOrDatefield, cloud_dir, unsortedCloudName, batch_size, userIdsMap,
				beamConfig, tmpCloudManager, diskManager, startTimestamp, endTimestamp)
			if err != nil {
				return err
			}

			numLines, err = ReadAndMergeChannelPartFiles(projectId, cloud_dir, sorted_dir, sorted_name, tmpCloudManager, diskManager)
			if err != nil {
				log.Errorf("unable to read part files and merge :%v", err)
				return err
			}
		}
		log.Info("numLines: ", numLines)
	} else {
		if err := SortUnsortedFile(dataType, unsortedFilePath, sortedFilePath, sortOnGroup); err != nil {
			return err
		}
	}
	log.Info("sorted")

	file, err := os.Open(sortedFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = (*cloudManager).Create(cloud_dir, cloud_name, file)
	if err != nil {
		log.Error("Failed to create file to cloud.")
		return err
	}
	log.Info("uploaded")

	err = os.Remove(sortedFilePath)
	if err != nil {
		return err
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
func ReadAndMergeDailyFiles(projectId int64, dataType, fileNamePrefix string, unsortedFileName string, startTimestamp, endTimestamp int64,
	local_dir string, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, getIdMap bool, sortOnGroup int) (map[string]int, int64, error) {
	userIdMap := make(map[string]int)
	var countLines int64
	local_merged_file := path.Join(local_dir, unsortedFileName)
	serviceDisk.MkdirAll(local_dir)
	tmpFile, err := os.OpenFile(local_merged_file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("unable to open local sorted file:%s", local_merged_file)
		return userIdMap, 0, err
	}

	log.Infof("Merging dailyFiles:%d", projectId)
	timestamp := startTimestamp
	for timestamp <= endTimestamp {
		partFilesDir, _ := GetDailyArchiveFilePathAndName(*cloudManager, dataType, fileNamePrefix, projectId, timestamp, 0, 0)
		listFiles := (*cloudManager).ListFiles(partFilesDir)
		for _, partFileFullName := range listFiles {
			partFNamelist := strings.Split(partFileFullName, "/")
			partFileName := partFNamelist[len(partFNamelist)-1]
			if !strings.HasPrefix(partFileName, fileNamePrefix) {
				continue
			}

			log.Infof("Reading daily file :%s, %s", partFilesDir, partFileName)
			file, err := (*cloudManager).Get(partFilesDir, partFileName)
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

				_, err = tmpFile.WriteString(line + "\n")
				if err != nil {
					log.Errorf("Unable to write to file :%v", err)
					return userIdMap, countLines, err
				}
				countLines += 1
			}
		}
		timestamp += U.Per_day_epoch
	}
	defer tmpFile.Close()

	return userIdMap, countLines, nil
}

// read events file locally from filePath, sort events based on [sortOnGroup,timestamp] (sortOnGroup - 0:uid, i:group_i_id) and
// create sorted local file at sortedFilePath
func SortUnsortedEventsFile(filePath string, sortedFilePath string, sortOnGroup int) error {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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
	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	sort.Slice(eventsList, func(i, j int) bool {
		if getAptId(eventsList[i], sortOnGroup) == getAptId(eventsList[i], sortOnGroup) {
			return eventsList[i].EventTimestamp < eventsList[j].EventTimestamp
		}
		return getAptId(eventsList[i], sortOnGroup) < getAptId(eventsList[i], sortOnGroup)
	})
	file, err = os.Create(sortedFilePath)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		return err
	}
	defer file.Close()

	for _, event := range eventsList {
		jsonBytes, err := json.Marshal(event)
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error marshalling")
			return err
		}
		_, err = file.WriteString(fmt.Sprintf("%s\n", jsonBytes))
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error writing")
			return err
		}
	}

	return err
}

// read channel file locally from filePath, sort reports based on [docType,timestamp] and
// create sorted local file at sortedFilePath
func SortUnsortedChannelFile(filePath string, sortedFilePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var docsList = make(map[int][]*pull.CounterCampaignFormat)

	for scanner.Scan() {
		line := scanner.Text()
		var rawEvent *pull.CounterCampaignFormat
		err := json.Unmarshal([]byte(line), &rawEvent)
		if err != nil {
			return err
		}
		if _, ok := docsList[rawEvent.Doctype]; !ok {
			docsList[rawEvent.Doctype] = make([]*pull.CounterCampaignFormat, 0)
		}
		docsList[rawEvent.Doctype] = append(docsList[rawEvent.Doctype], rawEvent)
	}
	err = os.Remove(filePath)
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
	file, err = os.Create(sortedFilePath)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		return err
	}
	defer file.Close()

	for _, docType := range docTypes {
		for _, doc := range docsList[docType] {
			jsonBytes, err := json.Marshal(doc)
			if err != nil {
				fmt.Println(err.Error())
				log.WithError(err).Error("Error marshalling")
				return err
			}
			_, err = file.WriteString(fmt.Sprintf("%s\n", jsonBytes))
			if err != nil {
				fmt.Println(err.Error())
				log.WithError(err).Error("Error writing")
				return err
			}
		}
	}

	return err
}

// read users file locally from filePath, group all users with same id and set JoinTimestamp as the least of all JoinTimestamps,
// sort users based on JoinTimestamp and create sorted local file at sortedFilePath
func SortUnsortedUsersFile(filePath string, sortedFilePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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

	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	sort.Slice(uniqueUsersList, func(i, j int) bool {
		return uniqueUsersList[i].JoinTimestamp < uniqueUsersList[j].JoinTimestamp
	})
	file, err = os.Create(sortedFilePath)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		return err
	}
	defer file.Close()

	for _, user := range uniqueUsersList {
		jsonBytes, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error marshalling")
			return err
		}
		_, err = file.WriteString(fmt.Sprintf("%s\n", jsonBytes))
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error writing")
			return err
		}
	}

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
func SortUnsortedFile(dataType string, unsortedFilePath, sortedFilePath string, sortOnGroup int) error {
	var err error
	if dataType == "events" {
		err = SortUnsortedEventsFile(unsortedFilePath, sortedFilePath, sortOnGroup)
	} else if dataType == "ad_reports" {
		err = SortUnsortedChannelFile(unsortedFilePath, sortedFilePath)
	} else if dataType == "users" {
		err = SortUnsortedUsersFile(unsortedFilePath, sortedFilePath)
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
