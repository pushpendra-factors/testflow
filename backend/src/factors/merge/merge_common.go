package merge

import (
	"bufio"
	"encoding/json"
	"factors/filestore"
	"factors/model/store"
	P "factors/pattern"
	"factors/pull"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const mergeDependencyTaskName = "PullEventsDaily"
const mergeDependencyOffset = 2

type sortedFileInfo struct {
	name       string
	dir        string
	startDay   time.Time
	endDay     time.Time
	daysToSkip int
}

type partFileInfo struct {
	name               string
	dir                string
	startUserIndex     int
	endUserIndex       int
	fileTimestampIndex int
	writer             *io.WriteCloser
}

func createDataFileFromSortedFiles(projectId int64, fileNamePrefix string, sorted_cloud_dir, sorted_cloud_name string, startTimestamp, endTimestamp int64, cloudManager *filestore.FileManager) (int64, bool, error) {
	var numLines int64
	// var fileCreated bool
	var reqFile *sortedFileInfo
	{
		tmp := 1
		if strings.HasSuffix(sorted_cloud_dir, "/") {
			tmp++
		}
		sortedCloudDirSplit := strings.Split(sorted_cloud_dir, "/")
		dataTypeDir := strings.Join(sortedCloudDirSplit[0:len(sortedCloudDirSplit)-tmp], "/") + "/"
		listAllFiles := (*cloudManager).ListFiles(dataTypeDir)
		log.Infof("listDirs: %v", listAllFiles)
		endDateStr := U.GetDateOnlyFromTimestampZ(endTimestamp)
		endDate, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, endDateStr)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"endDateStr": endDateStr}).Error("error parsing endDate")
			return 0, false, err
		}
		startDateStr := U.GetDateOnlyFromTimestampZ(startTimestamp)
		startDate, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, startDateStr)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"startDateStr": startDateStr}).Error("error parsing startDate")
			return 0, false, err
		}

		for _, filePath := range listAllFiles {
			fNamelist := strings.Split(filePath, "/")
			fileName := fNamelist[len(fNamelist)-1]
			fileDir := strings.Join(fNamelist[0:len(fNamelist)-1], "/") + "/"
			if !strings.HasPrefix(fileName, fileNamePrefix) {
				continue
			}
			fileNameSplit := strings.Split(fileName, "_")
			startAndEndDateTxt := fileNameSplit[len(fileNameSplit)-1]
			startAndEndDate := strings.Replace(startAndEndDateTxt, ".txt", "", 1)
			startAndEndDateSplit := strings.Split(startAndEndDate, "-")
			fileStartDateStr := startAndEndDateSplit[0]
			fileStartDate, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, fileStartDateStr)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"startDateStr": fileStartDateStr, "fileName": fileName}).Error("error parsing file startDate")
				continue
			}
			fileEndDateStr := startAndEndDateSplit[1]
			fileEndDate, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, fileEndDateStr)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"endDateStr": fileEndDateStr, "fileName": fileName}).Error("error parsing file endDate")
				continue
			}
			if !(fileStartDate.After(startDate) || fileEndDate.Before(endDate)) {
				filesDaysToSkip := (int(fileEndDate.Sub(endDate).Hours()) / 24) + (int(startDate.Sub(fileStartDate).Hours()) / 24)
				if (reqFile == nil || reqFile.daysToSkip > filesDaysToSkip) && filesDaysToSkip > 0 {
					fileInfo := sortedFileInfo{fileName, fileDir, fileStartDate, fileEndDate, filesDaysToSkip}
					reqFile = &fileInfo
				}
			}
		}
	}

	if reqFile != nil {
		cloudWriter, err := (*cloudManager).GetWriter(sorted_cloud_dir, sorted_cloud_name)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": sorted_cloud_dir, "fileName": sorted_cloud_name}).Error("unable to get writer for file")
			return 0, false, err
		}
		projectDetails, _ := store.GetStore().GetProject(projectId)
		startTimestampInProjectTimezone := startTimestamp
		endTimestampInProjectTimezone := endTimestamp
		if projectDetails.TimeZone != "" {
			// Input time is in UTC. We need the same time in the other timezone
			// if 2021-08-30 00:00:00 is UTC then we need the epoch equivalent in 2021-08-30 00:00:00 IST(project time zone)
			offset := U.FindOffsetInUTC(U.TimeZoneString(projectDetails.TimeZone))
			startTimestampInProjectTimezone = startTimestamp - int64(offset)
			endTimestampInProjectTimezone = endTimestamp - int64(offset)
		}
		log.WithFields(log.Fields{"fileDir": reqFile.dir, "fileName": reqFile.name}).Info("Reading file")
		reader, err := (*cloudManager).Get(reqFile.dir, reqFile.name)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": reqFile.dir, "fileName": reqFile.name}).Error("error reading file")
			return 0, false, err
		}
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		const maxCapacity = 30 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)
		for scanner.Scan() {

			line := scanner.Text()

			var eventDetails P.CounterEventFormat
			if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
				log.WithFields(log.Fields{"line": line, "err": err}).Error("Unmarshal failed")
				return numLines, false, err
			}

			if eventDetails.EventTimestamp < startTimestampInProjectTimezone || eventDetails.EventTimestamp > endTimestampInProjectTimezone {
				continue
			}

			_, err = io.WriteString(cloudWriter, line+"\n")
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"line": line, "fileDir": sorted_cloud_dir, "fileName": sorted_cloud_dir}).Error("Unable to write to file")
				return numLines, false, err
			}
			numLines++
		}
		err = cloudWriter.Close()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"fileDir": sorted_cloud_dir, "fileName": sorted_cloud_dir}).Error("Unable to close writer")
			return numLines, false, err
		}
		return numLines, true, nil
	}
	return 0, false, nil
}

func getApproxFileSize(dataType, fileNamePrefix string, projectId, startTimestamp, endTimestamp int64, cloudManager *filestore.FileManager) (int, error) {
	var totalFileSize int
	timestamp := startTimestamp
	for timestamp <= endTimestamp {
		partFilesDir, _ := pull.GetDailyArchiveFilePathAndName(cloudManager, dataType, fileNamePrefix, projectId, timestamp, 0, 0)
		listFiles := (*cloudManager).ListFiles(partFilesDir)
		for _, partFileFullName := range listFiles {
			partFNamelist := strings.Split(partFileFullName, "/")
			partFileName := partFNamelist[len(partFNamelist)-1]
			if !strings.HasPrefix(partFileName, fileNamePrefix) {
				continue
			}

			fileSize, err := (*cloudManager).GetObjectSize(partFilesDir, partFileName)
			if err != nil {
				log.Errorf("Failed to get file size: %s %s", partFilesDir, partFileName)
				return totalFileSize, err
			}
			totalFileSize += int(fileSize)
		}
		timestamp += U.Per_day_epoch
	}
	return totalFileSize, nil
}

func checkDependencyForEventsFile(projectId int64, startTimestamp, endTimestamp int64) (bool, error) {
	taskDetails, _, _ := store.GetStore().GetTaskDetailsByName(mergeDependencyTaskName)
	startTime := time.Unix(startTimestamp, 0)
	endTime := time.Unix(endTimestamp, 0)
	endTimeWithOffset := endTime.AddDate(0, 0, mergeDependencyOffset)
	lookback := int(endTimeWithOffset.Sub(startTime).Hours() / 24)
	configuredDeltas, status, errStr := store.GetStore().GetAllDeltasByConfiguration(taskDetails.TaskID, lookback, &endTimeWithOffset)
	if status != http.StatusOK {
		log.Errorf("GetAllDeltasByConfiguration failed with status:%d and error:%s", status, errStr)
		return false, fmt.Errorf("%s", errStr)
	}
	processedDeltas, status, errStr := store.GetStore().GetAllProcessedIntervals(taskDetails.TaskID, projectId, lookback, &endTimeWithOffset)
	if status != http.StatusOK {
		log.Errorf("GetAllProcessedIntervals failed with status:%d and error:%s", status, errStr)
		return false, fmt.Errorf("%s", errStr)
	}
	for _, delta := range configuredDeltas {
		isDone := U.ContainsUint64InArray(processedDeltas, delta)
		if !isDone {
			log.Infof("%s not yet done for %d", mergeDependencyTaskName, delta)
			return false, nil
		}
	}
	return true, nil
}

// read daily data files from archiveCloudManager between startTimestamp and endTimestamp, merge all and create merged file in tmpCloudManager, then sort the merged file and create sorted file in cloudManager
//
// dataType - events, ad_reports or users,
// sortOnGroup - decides which id to get for events dataType(0:uid, i:group_i_id),
// channelOrDatefield - channel name for ad_reports dataType and dateField for users dataType,
// hardPull - whether to create file if already present
// beamConfig - pass beam configs to sort using beam
// timestampsInProjectTimezone - whether timestamps are in utc or project's timezone
func MergeAndWriteSortedFile(projectId int64, dataType, channelOrDatefield string, startTimestamp, endTimestamp int64,
	archiveCloudManager, tmpCloudManager, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, beamConfig *RunBeamConfig, hardPull bool, sortOnGroup int, useSortedFiles, timestampsInProjectTimezone, checkDependency bool) (string, string, error) {

	if dataType != U.DataTypeEvent && channelOrDatefield == "" {
		log.Error("merge failed as channelOrDatefield is empty")
		return "", "", fmt.Errorf("empty channelOrDatefield")
	}

	if timestampsInProjectTimezone {
		projectDetails, _ := store.GetStore().GetProject(projectId)
		if projectDetails.TimeZone != "" {
			log.Infof("Project Timezone not UTC - Converting timestamps to UTC from project timezone(%s)", projectDetails.TimeZone)
			offset := U.FindOffsetInUTC(U.TimeZoneString(projectDetails.TimeZone))
			startTimestamp = startTimestamp + int64(offset)
			endTimestamp = endTimestamp + int64(offset)
		}
	}

	var err error
	startDayTimestamp := U.GetBeginningOfDayTimestamp(startTimestamp)
	endDayTimestamp := U.GetBeginningOfDayTimestamp(endTimestamp)

	if checkDependency && dataType == U.DataTypeEvent {
		if ok, err := checkDependencyForEventsFile(projectId, startDayTimestamp, endDayTimestamp); !ok {
			if err != nil {
				log.WithError(err).Error("checkDependencyForEventsFile failed")
				return "", "", err
			}
			return "", "", fmt.Errorf("%s not completed for given range", mergeDependencyTaskName)
		}
	}

	sorted_cloud_dir, sorted_cloud_name := GetSortedFilePathAndName(*cloudManager, dataType, channelOrDatefield, projectId, startDayTimestamp, endDayTimestamp, sortOnGroup)
	if !hardPull {
		if yes, _ := U.CheckFileExists(cloudManager, sorted_cloud_dir, sorted_cloud_name); yes {
			log.Info(sorted_cloud_name + " already exists")
			return sorted_cloud_dir, sorted_cloud_name, nil
		}
	}

	var fileNamePrefix string
	if dataType == U.DataTypeEvent {
		fileNamePrefix = U.EVENTS_FILENAME_PREFIX
	} else {
		fileNamePrefix = channelOrDatefield
	}

	if useSortedFiles {
		if numLines, ok, err := createDataFileFromSortedFiles(projectId, fileNamePrefix, sorted_cloud_dir, sorted_cloud_name, startDayTimestamp, endDayTimestamp, cloudManager); err != nil {
			log.WithError(err).Error("error in creating file using sorted files")
			return sorted_cloud_dir, sorted_cloud_name, err
		} else if ok {
			log.Info("created file using existing sorted files - numLines: ", numLines)
			return sorted_cloud_dir, sorted_cloud_name, nil
		}
	}

	approxFileSize, err := getApproxFileSize(dataType, fileNamePrefix, projectId, startDayTimestamp, endDayTimestamp, archiveCloudManager)
	if err != nil {
		log.WithError(err).Error("unable to get approx file size")
		return sorted_cloud_dir, sorted_cloud_name, err
	}

	if approxFileSize == 0 {
		log.WithFields(log.Fields{"start": startTimestamp, "end": endTimestamp}).Error("no data found for the given range")
		cloudWriter, err := (*cloudManager).GetWriter(sorted_cloud_dir, sorted_cloud_name)
		if err != nil {
			log.WithFields(log.Fields{"fileDir": sorted_cloud_dir, "fileName": sorted_cloud_name}).Error("unable to get writer for file")
			return sorted_cloud_dir, sorted_cloud_name, err
		}
		err = cloudWriter.Close()
		if err != nil {
			log.WithFields(log.Fields{"fileDir": sorted_cloud_dir, "fileName": sorted_cloud_name}).Error("unable to close writer")
			return sorted_cloud_dir, sorted_cloud_name, err
		}
		return sorted_cloud_dir, sorted_cloud_name, nil
	}

	var useBeam bool
	if beamConfig.RunOnBeam && approxFileSize > MIN_FILESIZE_FOR_BEAM && (dataType == U.DataTypeEvent || dataType == U.DataTypeAdReport) {
		useBeam = true
	}

	var unsortedLinesCount int
	var indexMap map[string]int64
	var startIndexToFileInfoMap = make(map[int][]*partFileInfo)
	var unsorted_cloud_dir, unsorted_cloud_name string
	if useBeam {
		batchSize := beamConfig.NumWorker
		startIndexToFileInfoMap, indexMap, unsortedLinesCount, err = ReadAndMergeDailyFilesForBeam(projectId, dataType, fileNamePrefix, startDayTimestamp, endDayTimestamp, archiveCloudManager, tmpCloudManager, sortOnGroup, batchSize)
		if err != nil {
			log.WithError(err).Error("unable to read daily part files and merge")
			return sorted_cloud_dir, sorted_cloud_name, err
		}
	} else {
		unsorted_cloud_dir, _ = GetSortedFilePathAndName(*tmpCloudManager, dataType, channelOrDatefield, projectId, startDayTimestamp, endDayTimestamp, sortOnGroup)
		unsorted_cloud_name = "unsorted_" + sorted_cloud_name
		unsortedLinesCount, err = ReadAndMergeDailyFiles(projectId, dataType, fileNamePrefix, unsorted_cloud_dir, unsorted_cloud_name, startDayTimestamp, endDayTimestamp, archiveCloudManager, tmpCloudManager)
		if err != nil {
			return sorted_cloud_dir, sorted_cloud_name, err
		}
	}
	log.Infof("merged (beam: %t) - numLines: %d", useBeam, unsortedLinesCount)

	var sortedLinesCount int
	if useBeam {
		sortedLinesCount, err = SortDataOnBeam(projectId, dataType, channelOrDatefield, startIndexToFileInfoMap, indexMap, beamConfig, sorted_cloud_dir, sorted_cloud_name,
			tmpCloudManager, cloudManager, diskManager, startDayTimestamp, endDayTimestamp, sortOnGroup)
		if err != nil {
			log.WithError(err).Error("error sorting (beam)")
			return sorted_cloud_dir, sorted_cloud_name, err
		}
	} else {
		sortedLinesCount, err = SortUnsortedFile(tmpCloudManager, cloudManager, dataType, unsorted_cloud_dir, unsorted_cloud_name, sorted_cloud_dir, sorted_cloud_name, sortOnGroup)
		if err != nil {
			log.WithError(err).Error("error sorting (non-beam)")
			return sorted_cloud_dir, sorted_cloud_name, err
		}
	}
	log.Infof("sorted (beam: %t) - numLines: %d", useBeam, sortedLinesCount)

	if dataType != U.DataTypeUser && unsortedLinesCount != sortedLinesCount {
		err = fmt.Errorf("sorting error(beam: %t) - unsortedLinesCount:%d, sortedLinesCount:%d", useBeam, unsortedLinesCount, sortedLinesCount)
		log.WithError(err).Errorf("error in sorting (beam: %t)", useBeam)
		return sorted_cloud_dir, sorted_cloud_name, err
	}

	return sorted_cloud_dir, sorted_cloud_name, nil
}

// read daily data files from cloud between startTimestamp and endTimestamp, merge all and create merged file in cloud at unsortedDir+unsortedFileName, also
// get countLines as output
//
// OUTPUTS
// countLines - number of lines in merged file
//
// INPUTS
// dataType - events, ad_reports or users,
// fileNamePrefix - all daily files only with this prefix are read (give "events" for events, channel name for ad_reports and dateField for users),
// getIdMap - whether to generate userIdMap
func ReadAndMergeDailyFiles(projectId int64, dataType, fileNamePrefix string, unsortedDir, unsortedFileName string, startTimestamp, endTimestamp int64,
	archiveCloudManager, tmpCloudManager *filestore.FileManager) (int, error) {
	var countLines int
	cloudWriter, err := (*tmpCloudManager).GetWriter(unsortedDir, unsortedFileName)
	if err != nil {
		log.Errorf("unable to get writer for file:%s", unsortedDir+unsortedFileName)
		return 0, err
	}

	log.Infof("Merging dailyFiles for project:%d", projectId)
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
					return countLines, err
				}
				countLines += 1
			}
		}
		timestamp += U.Per_day_epoch
	}
	err = cloudWriter.Close()

	return countLines, err
}

// read events file locally from filePath, sort events based on [sortOnGroup,timestamp] (sortOnGroup - 0:uid, i:group_i_id) and
// create sorted local file at sortedFilePath
func SortUnsortedEventsFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, unsortedDir, unsortedName string, sortedDir, sortedName string, sortOnGroup int) (int, error) {
	var numLines int
	reader, err := (*tmpCloudManager).Get(unsortedDir, unsortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return 0, err
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
			return 0, err
		}
		eventsList = append(eventsList, rawEvent)
	}
	err = reader.Close()
	if err != nil {
		return 0, err
	}
	sort.Slice(eventsList, func(i, j int) bool {
		if getAptId(eventsList[i], sortOnGroup) == getAptId(eventsList[j], sortOnGroup) {
			return eventsList[i].EventTimestamp < eventsList[j].EventTimestamp
		}
		return getAptId(eventsList[i], sortOnGroup) < getAptId(eventsList[j], sortOnGroup)
	})
	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return 0, err
	}

	for _, event := range eventsList {
		jsonBytes, err := json.Marshal(event)
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error marshalling")
			return numLines, err
		}
		_, err = io.WriteString(cloudWriter, fmt.Sprintf("%s\n", jsonBytes))
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error writing")
			return numLines, err
		}
		numLines++
	}
	err = cloudWriter.Close()

	return numLines, err
}

// read channel file locally from filePath, sort reports based on [docType,timestamp] and
// create sorted local file at sortedFilePath
func SortUnsortedChannelFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, unsortedDir, unsortedName string, sortedDir, sortedName string) (int, error) {

	var numLines int
	reader, err := (*tmpCloudManager).Get(unsortedDir, unsortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return 0, err
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
			return 0, err
		}
		if _, ok := docsList[docReport.Doctype]; !ok {
			docsList[docReport.Doctype] = make([]*pull.CounterCampaignFormat, 0)
		}
		docsList[docReport.Doctype] = append(docsList[docReport.Doctype], docReport)
	}
	err = reader.Close()
	if err != nil {
		return 0, err
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
		return 0, err
	}

	for _, docType := range docTypes {
		docs := docsList[docType]
		for _, doc := range docs {
			jsonBytes, err := json.Marshal(doc)
			if err != nil {
				fmt.Println(err.Error())
				log.WithError(err).Error("Error marshalling")
				return numLines, err
			}
			_, err = io.WriteString(cloudWriter, fmt.Sprintf("%s\n", jsonBytes))
			if err != nil {
				fmt.Println(err.Error())
				log.WithError(err).Error("Error writing")
				return numLines, err
			}
			numLines++
		}
	}
	err = cloudWriter.Close()

	return numLines, err
}

// read users file locally from filePath, group all users with same id and set JoinTimestamp as the least of all JoinTimestamps,
// sort users based on JoinTimestamp and create sorted local file at sortedFilePath
func SortUnsortedUsersFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, unsortedDir, unsortedName string, sortedDir, sortedName string) (int, error) {

	var numLines int
	reader, err := (*tmpCloudManager).Get(unsortedDir, unsortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return 0, err
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
				return 0, err
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
		return 0, err
	}

	sort.Slice(uniqueUsersList, func(i, j int) bool {
		return uniqueUsersList[i].JoinTimestamp < uniqueUsersList[j].JoinTimestamp
	})
	cloudWriter, err := (*sortedCloudManager).GetWriter(sortedDir, sortedName)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("Error reading file")
		return 0, err
	}

	for _, user := range uniqueUsersList {
		jsonBytes, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error marshalling")
			return numLines, err
		}
		_, err = io.WriteString(cloudWriter, fmt.Sprintf("%s\n", jsonBytes))
		if err != nil {
			fmt.Println(err.Error())
			log.WithError(err).Error("Error writing")
			return numLines, err
		}
		numLines++
	}
	err = cloudWriter.Close()

	return numLines, err
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
func SortUnsortedFile(tmpCloudManager, sortedCloudManager *filestore.FileManager, dataType string, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName string, sortOnGroup int) (int, error) {
	var numLines int
	var err error
	if dataType == U.DataTypeEvent {
		numLines, err = SortUnsortedEventsFile(tmpCloudManager, sortedCloudManager, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName, sortOnGroup)
	} else if dataType == U.DataTypeAdReport {
		numLines, err = SortUnsortedChannelFile(tmpCloudManager, sortedCloudManager, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName)
	} else if dataType == U.DataTypeUser {
		numLines, err = SortUnsortedUsersFile(tmpCloudManager, sortedCloudManager, unsortedFileDir, unsortedFileName, sortedFileDir, sortedFileName)
	} else {
		err = fmt.Errorf("wrong dataType: %s", dataType)
	}
	return numLines, err
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
