package pull

import (
	"database/sql"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io"
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
func PullDataForChannel(channel string, projectId int64, cloudManager *filestore.FileManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Info("Pulling " + channel)
	// Writing adwords data to tmp file before upload.
	_, cName := (*cloudManager).GetDailyChannelArchiveFilePathAndName(channel, projectId, 0, startTimestamp, startTimestamp+U.Per_day_epoch-1)

	startAt := time.Now().UnixNano()
	count, err := pullChannelData(channel, projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, cName, cloudManager, startTimestamp, endTimestamp)
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
	} else {
		logCtx.WithFields(log.Fields{
			channel + "-RowsCount":       count,
			channel + "-TimeTakenToPull": timeTakenToPullEvents,
		}).Info("Successfully pulled " + channel + " and uploaded to files.")
		status[channel+"-RowsCount"] = count
	}

	return nil, true
}

// pull ad reports(rows) from db and generate local files w.r.t timestamp and return map with (key,value) as (timestamp,path)
func pullChannelData(channel string, projectID int64, startTimestampTimezone, endTimestampTimezone int64, cName string, cloudManager *filestore.FileManager, startTimestamp, endTimestamp int64) (int, error) {

	rows, tx, err := channelToPullMap[channel](projectID, startTimestampTimezone, endTimestampTimezone)
	if err != nil {
		log.WithError(err).Error("SQL Query failed.")
		return 0, err
	}
	defer U.CloseReadQuery(rows, tx)

	var writerMap = make(map[int64]*io.WriteCloser)
	rowCount := 0
	for rows.Next() {
		var id string
		var value *postgres.Jsonb
		var timestamp int64
		var docType int
		var smartProps *postgres.Jsonb
		if err = rows.Scan(&id, &value, &timestamp, &docType, &smartProps); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, err
		}

		timestampUnix, err := U.GetBeginningDayTimestampFromDateString(fmt.Sprintf("%d", timestamp))
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("parsing date failed")
			return 0, err
		}
		fileTimestamp := timestampUnix
		writer, ok := writerMap[fileTimestamp]
		if !ok {
			cPath, _ := (*cloudManager).GetDailyChannelArchiveFilePathAndName(channel, projectID, fileTimestamp, 0, 0)
			cloudWriter, err := (*cloudManager).GetWriter(cPath, cName)
			if err != nil {
				log.WithFields(log.Fields{"file": cName, "err": err}).Error("Unable to get cloud file writer")
				return 0, err
			}
			writerMap[fileTimestamp] = &cloudWriter
			writer = &cloudWriter
		}

		var valueMap map[string]interface{}
		if value != nil {
			valueBytes, err := value.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal value")
				return 0, err
			}
			err = json.Unmarshal(valueBytes.([]byte), &valueMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal value")
				return 0, err
			}
		} else {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil value")
		}

		var smartPropsMap map[string]interface{}
		if smartProps != nil {
			smartPropsBytes, err := smartProps.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to get value from smart props json")
				return 0, err
			}
			err = json.Unmarshal(smartPropsBytes.([]byte), &smartPropsMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal smart props")
				return 0, err
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
			return 0, err
		}
		line := string(lineBytes)
		if _, err := io.WriteString(*writer, fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, err
		}
		rowCount++
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, fmt.Errorf("ad reports count has exceeded the %d limit", M.EventsPullLimit)
	}

	for _, writer := range writerMap {
		err := (*writer).Close()
		if err != nil {
			log.WithError(err).Error("Error closing writer")
			return 0, err
		}
	}

	return rowCount, nil
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
