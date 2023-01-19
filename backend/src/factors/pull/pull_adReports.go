package pull

import (
	"database/sql"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type CounterCampaignFormat struct {
	Id         string                 `json:"id"`
	Channel    string                 `json:"source"`
	Doctype    int                    `json:"type"`
	Timestamp  int64                  `json:"timestamp"`
	Value      map[string]interface{} `json:"value"`
	SmartProps map[string]interface{} `json:"sp"`
}

var channelToPullMap = map[string]func(int64, int64, int64) (*sql.Rows, *sql.Tx, error){
	M.ADWORDS:        store.GetStore().PullAdwordsRowsV2,
	M.FACEBOOK:       store.GetStore().PullFacebookRowsV2,
	M.BINGADS:        store.GetStore().PullBingAdsRowsV2,
	M.LINKEDIN:       store.GetStore().PullLinkedInRowsV2,
	M.GOOGLE_ORGANIC: store.GetStore().PullGoogleOrganicRowsV2,
}

// pull channel data into files, and upload each local file to its proper cloud location
func PullDataForChannel(channel string, projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Info("Pulling " + channel)
	// Writing adwords data to tmp file before upload.
	_, fName := diskManager.GetDailyChannelArchiveFilePathAndName(channel, projectId, 0, startTimestamp, startTimestamp+U.Per_day_epoch-1)

	startAt := time.Now().UnixNano()
	count, filePaths, err := pullChannelData(channel, projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, fName, diskManager, startTimestamp, endTimestamp)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull " + channel + " failed. Pull and write failed.")
		status[channel+"-error"] = err.Error()
		return err, false
	}
	timeTakenToPullEvents := (time.Now().UnixNano() - startAt) / 1000000
	// Zero events. Returns eventCount as 0.
	if count == 0 {
		logCtx.Info("No " + channel + " data found.")
		status[channel+"-info"] = "No " + channel + " data found."
		return err, true
	}

	_, cName := (*cloudManager).GetDailyChannelArchiveFilePathAndName(channel, projectId, 0, startTimestamp, startTimestamp+U.Per_day_epoch-1)
	for timestamp, path := range filePaths {
		tmpOutputFile, err := os.Open(path)
		if err != nil {
			logCtx.WithField("error", err).Error("Failed to pull " + channel + ". Write to tmp failed.")
			status[channel+"-error"] = "Write to tmp failed for path " + path
			return err, false
		}

		cDir, _ := (*cloudManager).GetDailyChannelArchiveFilePathAndName(channel, projectId, timestamp, 0, 0)
		err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
		if err != nil {
			logCtx.WithField("error", err).Errorf("Failed to pull "+channel+". Upload failed.", channel)
			status[channel+"-error"] = "Upload failed for " + cDir + cName
			return err, false
		}

		err = os.Remove(path)
		if err != nil {
			logCtx.Error("unable to delete file")
			return err, false
		}
	}

	logCtx.WithFields(log.Fields{
		channel + "-RowsCount":       count,
		channel + "-TimeTakenToPull": timeTakenToPullEvents,
	}).Info("Successfully pulled " + channel + " and written to file.")

	status[channel+"-RowsCount"] = count

	return nil, true
}

// pull ad reports(rows) from db and generate local files w.r.t timestamp and return map with (key,value) as (timestamp,path)
func pullChannelData(channel string, projectID int64, startTimestampTimezone, endTimestampTimezone int64, fName string, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp int64) (int, map[int64]string, error) {

	rows, tx, err := channelToPullMap[channel](projectID, startTimestampTimezone, endTimestampTimezone)
	if err != nil {
		log.WithError(err).Error("SQL Query failed.")
		return 0, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	var fileMap = make(map[int64]*os.File)
	var pathMap = make(map[int64]string)
	rowCount := 0
	for rows.Next() {
		var id string
		var value *postgres.Jsonb
		var timestamp int64
		var docType int
		var smartProps *postgres.Jsonb
		if err = rows.Scan(&id, &value, &timestamp, &docType, &smartProps); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, nil, err
		}

		timestampUnix, err := U.GetBeginningDayTimestampFromDateString(fmt.Sprintf("%d", timestamp))
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("parsing date failed")
			return 0, nil, err
		}
		fileTimestamp := timestampUnix
		file, ok := fileMap[fileTimestamp]
		if !ok {
			fPath, _ := diskManager.GetDailyChannelArchiveFilePathAndName(channel, projectID, fileTimestamp, 0, 0)
			serviceDisk.MkdirAll(fPath) // create dir if not exist.
			tmpEventsFile := fPath + fName
			fileNew, err := os.Create(tmpEventsFile)
			if err != nil {
				log.WithFields(log.Fields{"file": fName, "err": err}).Error("Unable to create file.")
				return 0, nil, err
			}
			defer file.Close()
			pathMap[fileTimestamp] = tmpEventsFile
			fileMap[fileTimestamp] = fileNew
			file = fileNew
		}

		var valueMap map[string]interface{}
		if value != nil {
			valueBytes, err := value.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal value")
				return 0, nil, err
			}
			err = json.Unmarshal(valueBytes.([]byte), &valueMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal value")
				return 0, nil, err
			}
		} else {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil value")
		}

		var smartPropsMap map[string]interface{}
		if smartProps != nil {
			smartPropsBytes, err := smartProps.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to get value from smart props json")
				return 0, nil, err
			}
			err = json.Unmarshal(smartPropsBytes.([]byte), &smartPropsMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal smart props")
				return 0, nil, err
			}
		}

		doc := CounterCampaignFormat{
			Id:         id,
			Channel:    channel,
			Value:      valueMap,
			Timestamp:  timestamp,
			Doctype:    docType,
			SmartProps: smartPropsMap,
		}

		lineBytes, err := json.Marshal(doc)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to marshal document.")
			return 0, nil, err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, nil, err
		}
		rowCount++
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, pathMap, fmt.Errorf("ad reports count has exceeded the %d limit", M.EventsPullLimit)
	}

	return rowCount, pathMap, nil
}

// check if channel data file has been generated already for given start and end timestamp, in dataTimestamp's day and
// its previous day's folder
func CheckChannelFileExists(channel string, cloudManager *filestore.FileManager, projectId int64, dataTimestamp, startTimestamp, endTimestamp int64) (bool, error) {
	path, name := (*cloudManager).GetDailyChannelArchiveFilePathAndName(channel, projectId, dataTimestamp, startTimestamp, endTimestamp)
	if yes, _ := U.CheckFileExists(cloudManager, path, name); yes {
		return true, nil
	}
	path, _ = (*cloudManager).GetDailyChannelArchiveFilePathAndName(channel, projectId, dataTimestamp-U.Per_day_epoch, startTimestamp, endTimestamp)
	if yes, _ := U.CheckFileExists(cloudManager, path, name); yes {
		return true, nil
	}
	return false, fmt.Errorf("file not found in cloud: %s", name)
}
