package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	C "datasets/config"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const EmptyJsonStr = "{}"
const DbLimit = 5000000

const SourceEventsFileName = "src_events.json"
const MaskedEventsFileName = "masked_events.json"

type Event struct {
	ID                string          `json:"id"`
	Name              string          `json:"na"`
	Count             uint64          `json:"co"`
	Properties        *postgres.Jsonb `json:"pr"`
	Timestamp         int64           `json:"ti"`
	UserId            string          `json:"uid"`
	CustomerUserId    *string         `json:"cuid"`
	UserJoinTimestamp int64           `json:"ujt"`
	UserProperties    *postgres.Jsonb `json:"upr"`
}

// TrackableEvent - struct with fields for event tracking.
type TrackableEvent struct {
	CustomerEventId string          `json:"c_event_id"`
	Auto            bool            `json:"auto"`
	Name            string          `json:"event_name"`
	Properties      *postgres.Jsonb `json:"event_properties"`
	UserId          string          `json:"user_id"`
	Timestamp       int64           `json:"timestamp"`
	UserProperties  *postgres.Jsonb `json:"user_properties"`
}

func GetFilePath(dir string, filename string) string {
	var fullpath string = dir
	if !strings.HasSuffix(dir, "/") {
		fullpath = fullpath + "/"
	}
	fullpath = fullpath + filename

	return fullpath
}

func PullEvents(projectId uint64, startTimestamp int64, endTimestamp int64, pullToDir string) (string, error) {
	db := C.GetServices().Db

	// create dir path if not exist.
	err := os.MkdirAll(pullToDir, 0755)
	if err != nil {
		log.WithError(err).Error("Failed creating events file dir : " + pullToDir)
		return "", err
	}

	eventsFilePath := GetFilePath(pullToDir, SourceEventsFileName)
	eventsFile, err := os.Create(eventsFilePath)
	if err != nil {
		log.WithError(err).Error("Failed creating events file : " + eventsFilePath)
		return "", err
	}
	defer eventsFile.Close()

	rows, _ := db.Raw("SELECT events.id, events.user_id, users.customer_user_id, users.join_timestamp, event_names.name, events.count, events.timestamp, events.properties, user_properties.properties FROM events"+
		" "+"LEFT JOIN event_names ON event_names.id = events.event_name_id"+
		" "+"LEFT JOIN users ON users.id = events.user_id"+
		" "+"LEFT JOIN user_properties ON user_properties.id = events.user_properties_id"+
		" "+"WHERE events.project_id = ? AND events.timestamp >= ? AND events.timestamp <= ?"+
		" "+"LIMIT ?", projectId, startTimestamp, endTimestamp, DbLimit).Rows()

	rowNum := 0
	for rows.Next() {
		var id string
		var userId string
		var customerUserId *string
		var userJoinTimestamp int64
		var name string
		var count uint64
		var timestamp int64
		var properties *postgres.Jsonb
		var userProperties *postgres.Jsonb

		if err := rows.Scan(&id, &userId, &customerUserId, &userJoinTimestamp, &name, &count, &timestamp, &properties, &userProperties); err != nil {
			log.WithError(err).Error("Failed to scan rows")
			return "", err
		}

		var eventPropertiesBytes interface{}
		if properties != nil {
			eventPropertiesBytes, err = properties.Value()
			if err != nil {
				log.WithError(err).Error("Failed to read event properties")
				return "", err
			}
		} else {
			eventPropertiesBytes = []byte(EmptyJsonStr)
		}

		var eventPropertiesMap map[string]interface{}
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		if err != nil {
			log.WithError(err).Error("Failed to umarshal event properties")
			return "", err
		}

		var userPropertiesBytes interface{}
		if userProperties != nil {
			userPropertiesBytes, err = userProperties.Value()
			if err != nil {
				log.WithError(err).Error("Failed to read user properties")
				return "", err
			}
		} else {
			userPropertiesBytes = []byte(EmptyJsonStr)
		}

		var userPropertiesMap map[string]interface{}
		err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
		if err != nil {
			log.WithError(err).Error("Failed to umarshal user properties")
			return "", err
		}

		event := &Event{
			ID:                id,
			Name:              name,
			Count:             count,
			Timestamp:         timestamp,
			Properties:        properties,
			UserId:            userId,
			CustomerUserId:    customerUserId,
			UserJoinTimestamp: userJoinTimestamp,
			UserProperties:    userProperties,
		}

		lineBytes, err := json.Marshal(event)
		if err != nil {
			log.WithError(err).Error("Failed to marshal event.")
			return "", err
		}

		line := string(lineBytes)
		if _, err := eventsFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithError(err).Error("Failed writing line to file.")
			return "", err
		}
		rowNum++
	}

	log.Infof("Downloaded events to %s", eventsFilePath)

	return eventsFilePath, nil
}

func GetEventUserId(clientUserId string, eventTimestamp int64, clientUserIdToUserIdMap *map[string]string,
	apiHost string, apiToken string) (string, error) {

	userId, found := (*clientUserIdToUserIdMap)[clientUserId]
	if found {
		return userId, nil
	}

	// Create a user.
	userRequestMap := make(map[string]interface{})
	userRequestMap["c_uid"] = clientUserId
	userRequestMap["join_timestamp"] = eventTimestamp

	reqBody, _ := json.Marshal(userRequestMap)
	url := fmt.Sprintf("%s/sdk/user/identify", apiHost)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Add("Authorization", apiToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(fmt.Sprintf(
			"Http Post user creation failed. Url: %s, reqBody: %s, response: %+v, error: %+v",
			url, reqBody, resp, err))
		return "", err
	}
	// always close the response-body, even if content is not required
	defer resp.Body.Close()

	jsonResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to parse http user create response.")
		return "", err
	}
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)

	userId = jsonResponseMap["user_id"].(string)
	(*clientUserIdToUserIdMap)[clientUserId] = userId

	return userId, nil
}

// GetCustomerEventId - Random customerEventId with
// current unix timestamp with a random num (or line num).
func GetCustomerEventId(num int) string {
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	numStr := strconv.Itoa(num)
	return timeStr + numStr
}

func getFileEvents(scanner *bufio.Scanner, batchSize int) ([]string, bool) {
	lineNumber := 1
	fileEvents := []string{}
	for scanner.Scan() && lineNumber <= batchSize {
		line := scanner.Text()
		fileEvents = append(fileEvents, line)
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return fileEvents, lineNumber < batchSize
}

func bulkIngestEvents(events []TrackableEvent, apiHost string, apiToken string) error {
	reqBody, err := json.Marshal(events)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/sdk/event/track/bulk", apiHost)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", apiToken)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(fmt.Sprintf("Http Post event creation failed. Url: %s, reqBody: %s", url, reqBody))
		return err
	}
	defer resp.Body.Close()

	return nil
}

func isAutoTrackedEvent(eventProperties *postgres.Jsonb) bool {
	var properties map[string]interface{}
	json.Unmarshal(eventProperties.RawMessage, &properties)
	_, pageRawUrlExists := properties["$page_raw_url"]
	_, pageRawUrl2Exists := properties["$pageRawURL"]
	_, pageRawUrl3Exists := properties["_$pageRawURL"]
	_, pageRawUrl4Exists := properties["_$page_raw_url"]
	_, rawUrlExists := properties["$rawURL"]
	_, rawUrl2Exists := properties["_$rawURL"]
	return pageRawUrlExists || pageRawUrl2Exists || pageRawUrl3Exists || pageRawUrl4Exists || rawUrlExists || rawUrl2Exists
}

func IsEmptyPostgresJsonb(jsonb *postgres.Jsonb) bool {
	strJson := string((*jsonb).RawMessage)
	return strJson == "" || strJson == "null"
}

func DecodePostgresJsonb(sourceJsonb *postgres.Jsonb) (*map[string]interface{}, error) {
	var sourceMap map[string]interface{}
	if !IsEmptyPostgresJsonb(sourceJsonb) {
		if err := json.Unmarshal((*sourceJsonb).RawMessage, &sourceMap); err != nil {
			return nil, err
		}
	} else {
		sourceMap = make(map[string]interface{}, 0)
	}

	return &sourceMap, nil
}

func EncodeToPostgresJsonb(sourceMap *map[string]interface{}) (*postgres.Jsonb, error) {
	sourceJsonBytes, err := json.Marshal(sourceMap)
	if err != nil {
		return nil, err
	}

	return &postgres.Jsonb{sourceJsonBytes}, nil
}

func renameProperties(src *map[string]interface{},
	renameMap *map[string]string) *map[string]interface{} {

	dest := make(map[string]interface{}, 0)
	for k, v := range *src {
		if _, exists := (*renameMap)[k]; exists {
			dest[(*renameMap)[k]] = v
		} else {
			dest[k] = v
		}
	}

	return &dest
}

func convEventAsTrackable(eventJson string, clientUserIdToUserIdMap *map[string]string,
	apiHost string, apiToken string, eventPropertiesRenameMap,
	userPropertiesRenameMap *map[string]string) (*TrackableEvent, error) {

	var event Event
	err := json.Unmarshal([]byte(eventJson), &event)
	if err != nil {
		return nil, err
	}

	var eventProperties *postgres.Jsonb
	if eventPropertiesRenameMap != nil {
		eventPropertiesMap, err := DecodePostgresJsonb(event.Properties)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal event properties on rename.")
			return nil, err
		}

		eventPropertiesRenamed := renameProperties(eventPropertiesMap, eventPropertiesRenameMap)
		eventProperties, err = EncodeToPostgresJsonb(eventPropertiesRenamed)
		if err != nil {
			log.WithError(err).Error("Failed to marshal event properties after rename.")
			return nil, err
		}
	} else {
		eventProperties = event.Properties
	}

	var userProperties *postgres.Jsonb
	if userPropertiesRenameMap != nil {
		userPropertiesMap, err := DecodePostgresJsonb(event.UserProperties)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal user properties on rename.")
			return nil, err
		}

		userPropertiesRenamed := renameProperties(userPropertiesMap, userPropertiesRenameMap)
		userProperties, err = EncodeToPostgresJsonb(userPropertiesRenamed)
		if err != nil {
			log.WithError(err).Error("Failed to marshal user properties after rename.")
			return nil, err
		}
	} else {
		userProperties = event.UserProperties
	}

	var trackEvent TrackableEvent
	trackEvent.Name = event.Name
	trackEvent.Properties = eventProperties
	trackEvent.UserProperties = userProperties
	trackEvent.Timestamp = event.Timestamp
	trackEvent.Auto = isAutoTrackedEvent(eventProperties)

	// using src event's id as customer_event_id.
	trackEvent.CustomerEventId = event.ID

	var clientUserId string
	if event.CustomerUserId != nil {
		clientUserId = *event.CustomerUserId
	} else {
		// using user_id of source event as c_uid
		// for grouping events done by the same user.
		// to be discussed.
		clientUserId = event.UserId
	}

	cUserId, err := GetEventUserId(clientUserId, event.UserJoinTimestamp,
		clientUserIdToUserIdMap, apiHost, apiToken)
	if err != nil {
		return nil, err
	}
	trackEvent.UserId = cUserId

	return &trackEvent, nil
}

func IngestEventsFromFile(filepath string, apiHost string, apiToken string,
	clientUserIdToUserIdMap *map[string]string, excludeEventNamePrefixes []string,
	eventPropertiesRenameMap, userPropertiesRenameMap *map[string]string) error {

	file, err := os.Open(filepath)
	if err != nil {
		log.WithError(err).Error("Failed to ingest events from file.")
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	maxBatchSize := 1000
	batchNumber := 1
	eof := false
	for !eof {
		log.WithFields(log.Fields{
			"start": (batchNumber - 1) * maxBatchSize,
			"end":   batchNumber * maxBatchSize,
			"batch": batchNumber,
		}).Info("Ingesting Batch")

		eventLines, tmpEof := getFileEvents(scanner, maxBatchSize)
		translatedEvents := make([]TrackableEvent, 0, 0)
		for _, eventJson := range eventLines {
			trEvent, err := convEventAsTrackable(eventJson, clientUserIdToUserIdMap, apiHost, apiToken,
				eventPropertiesRenameMap, userPropertiesRenameMap)
			if err == nil {
				exclude := false
				for _, exName := range excludeEventNamePrefixes {
					if strings.HasPrefix(trEvent.Name, exName) {
						exclude = true
					}
				}

				if !exclude {
					translatedEvents = append(translatedEvents, *trEvent)
				}
			} else {
				log.WithError(err).Error("Failed to translate event into trackable. Skipped.")
			}
		}

		err := bulkIngestEvents(translatedEvents, apiHost, apiToken)
		if err != nil {
			log.WithError(err).Error("Failed to ingest events from file.")
			return err
		}

		eof = tmpEof
		batchNumber++
	}

	return nil
}
