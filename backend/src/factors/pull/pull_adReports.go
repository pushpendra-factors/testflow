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

// pull ad reports data for a channel into cloud files with proper logging
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

// pull ad reports(rows) from db for channel into cloud files
func pullChannelData(channel string, projectID int64, startTimestampTimezone, endTimestampTimezone int64, fileName string, cloudManager *filestore.FileManager, startTimestamp, endTimestamp int64) (int, error) {

	rows, tx, err := channelToPullMap[channel](projectID, startTimestampTimezone, endTimestampTimezone)
	if err != nil {
		log.WithError(err).Error("SQL Query failed.")
		return 0, err
	}
	defer U.CloseReadQuery(rows, tx)

	var writerMap = make(map[int64]*io.WriteCloser)
	nilValue := 0
	nilSmartProps := 0
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
		//get apt writer w.r.t file timestamp
		writer, err := getAptWriterFromMap(projectID, writerMap, cloudManager, fileName, fileTimestamp, U.DataTypeUser, channel)
		if err != nil {
			log.WithError(err).Error("error getting apt writer from file")
			return 0, err
		}

		//value
		valueMap, err := getMapFromPostgresJson(value)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"project_id": projectID}).Error("error getting value")
			return 0, err
		}
		if len(valueMap) == 0 {
			nilValue++
		}

		//smart props
		smartPropsMap, err := getMapFromPostgresJson(smartProps)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{"project_id": projectID}).Error("error getting smart properties")
			return 0, err
		}
		if len(smartPropsMap) == 0 {
			nilSmartProps++
		}

		//create campInfo
		campInfo := CounterCampaignFormat{
			Id:         id,
			Channel:    channel,
			Value:      valueMap,
			Timestamp:  timestamp,
			Doctype:    docType,
			SmartProps: smartPropsMap,
		}

		//write campInfo to file
		lineBytes, err := json.Marshal(campInfo)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to marshal report.")
			return 0, err
		}
		line := string(lineBytes)
		if _, err := io.WriteString(*writer, fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, err
		}
		rowCount++
	}

	if rowCount > M.AdReportsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, fmt.Errorf("ad reports count has exceeded the %d limit", M.AdReportsPullLimit)
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
