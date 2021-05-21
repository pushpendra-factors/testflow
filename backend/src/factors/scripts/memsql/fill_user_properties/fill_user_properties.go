package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	U "factors/util"

	log "github.com/sirupsen/logrus"

	"factors/util"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm/dialects/postgres"
)

var memSQLDB *gorm.DB

// Event - Using copy of the struct to add new column user_properties only for memsql table.
type Event struct {
	ID                         string         `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	CustomerEventId            *string        `json:"customer_event_id"`
	ProjectId                  uint64         `gorm:"primary_key:true;" json:"project_id"`
	UserId                     string         `json:"user_id"`
	UserPropertiesId           string         `json:"user_properties_id"`
	SessionId                  *string        `json:session_id`
	EventNameId                uint64         `json:"event_name_id"`
	Count                      uint64         `json:"count"`
	Properties                 postgres.Jsonb `json:"properties,omitempty"`
	PropertiesUpdatedTimestamp int64          `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
	UserProperties             postgres.Jsonb `json:"user_properties"`
	Timestamp                  int64          `json:"timestamp"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
}

// User - Using copy of the struct to enable column properties only for memsql table.
type User struct {
	ID                 string         `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	ProjectId          uint64         `gorm:"primary_key:true;" json:"project_id"`
	PropertiesId       string         `json:"properties_id"`
	Properties         postgres.Jsonb `gorm:"properties" json:"properties"`
	SegmentAnonymousId string         `gorm:"type:varchar(200);default:null" json:"seg_aid"`
	AMPUserId          string         `gorm:"default:null";json:"amp_user_id"`
	CustomerUserId     string         `gorm:"type:varchar(255);default:null" json:"c_uid"`
	JoinTimestamp      int64          `json:"join_timestamp"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type UserProperties struct {
	// Composite primary key with project_id, user_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `gorm:"primary_key:true;" json:"user_id"`

	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties       postgres.Jsonb `json:"properties"`
	UpdatedTimestamp int64          `gorm:"not null;default:0" json:"updated_timestamp"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func initMemSQLDB(env, dsn string) {
	var err error
	memSQLDB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.WithError(err).Fatal("Failed connecting to memsql.")
	}

	if env == "development" {
		memSQLDB.LogMode(true)
	} else {
		memSQLDB.LogMode(false)
	}
}

func getEventsByTimerange(projectID uint64, startTimestamp int64, endTimestamp int64) ([]Event, int) {
	var events []Event
	err := memSQLDB.Where("project_id = ? AND timestamp BETWEEN ? AND ? AND user_properties IS NULL",
		projectID, startTimestamp, endTimestamp).Find(&events).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return events, http.StatusNotFound
		}

		log.WithField("project_id", projectID).WithError(err).Error("Failed to get events.")
		return nil, http.StatusInternalServerError
	}

	return events, http.StatusFound
}

func sanitizeJsonb(jsonb *postgres.Jsonb) {
	var unknownCharacters = regexp.MustCompile(`[^[:alnum:][:blank:][:punct:]]`)
	quoteEscaped := strings.ReplaceAll(string(jsonb.RawMessage), `'`, `''`)
	(*jsonb).RawMessage = []byte(unknownCharacters.ReplaceAllString(quoteEscaped, ``))
}

func updateUserPropertiesToEvents(projectID uint64, ids []string,
	userID string, propertiesJsonb *postgres.Jsonb) int {

	sanitizeJsonb(propertiesJsonb)
	if err := memSQLDB.Model(&Event{}).
		Where("project_id = ? AND id IN (?) AND user_id = ?", projectID, ids, userID).
		Update("user_properties", propertiesJsonb).Error; err != nil {

		log.WithFields(log.Fields{"project_id": projectID, "id": ids}).
			WithError(err).Error("Failed to add user_properties to the event.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func updateUserPropertiesToUser(projectID uint64, id string, propertiesJsonb *postgres.Jsonb) int {
	sanitizeJsonb(propertiesJsonb)
	if err := memSQLDB.Model(&User{}).
		Where("project_id = ? AND id = ?", projectID, id).
		Update("properties", propertiesJsonb).Error; err != nil {

		log.WithFields(log.Fields{"project_id": projectID, "id": id}).
			WithError(err).Error("Failed to add user_properties to the user.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func getUser(projectId uint64, id string) (*User, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": id})

	var user User
	if err := memSQLDB.Limit(1).Where("project_id = ?", projectId).
		Where("id = ?", id).Select("id, properties_id").Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get user using user_id")
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

func getUserPropertiesByUserAndCache(projectID uint64, userPropertiesIDs []string,
	userPropertiesCacheMap *map[string]postgres.Jsonb) int {

	logCtx := log.WithField("project_id", projectID)

	var userProperties []UserProperties
	if err := memSQLDB.Where("project_id = ?", projectID).
		Where("id IN (?)", userPropertiesIDs).Select("id, user_id, properties").
		Find(&userProperties).Error; err != nil {

		logCtx.WithError(err).Error("Failure on user properties download.")

		// Proceed with error for the existing user_properties.
		if len(userProperties) == 0 {
			return http.StatusInternalServerError
		}
	}

	for i := range userProperties {
		(*userPropertiesCacheMap)[fmt.Sprintf("%s:%s", userProperties[i].UserId, userProperties[i].ID)] = userProperties[i].Properties
	}

	return http.StatusFound
}

// Pull and cache given user_properties by ids as 1000 on each call.
func getUserPropertiesByUserAndCachePaginated(projectID uint64, userPropertiesIDs []string,
	userPropertiesCacheMap *map[string]postgres.Jsonb) int {

	userPropertiesIDsList := U.GetStringListAsBatch(userPropertiesIDs, 1000)
	for i := range userPropertiesIDsList {
		errCode := getUserPropertiesByUserAndCache(projectID, userPropertiesIDsList[i], userPropertiesCacheMap)
		if errCode == http.StatusInternalServerError {
			return errCode
		}
	}

	return http.StatusFound
}

func fillUserPropertiesToEventAndUsersTable(projectID uint64, startTimestamp, endTimestamp int64) int {
	logCtx := log.WithField("project_id", projectID).
		WithField("start_timestamp", startTimestamp).
		WithField("end_timestamp", endTimestamp)

	events, errCode := getEventsByTimerange(projectID, startTimestamp, endTimestamp)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get events by timerange.")
		return errCode
	}
	logCtx.WithField("no_of_events", len(events)).Info("Downloaded events.")

	// eventUserPropertiesMap[user_id:user_properties_id] = [event_ids ...]
	eventUserPropertiesMap := map[string][]string{}
	// uniqueUserPropertiesMap[user_properties_id] = true
	uniqueUserPropertiesMap := map[string]bool{}

	uniqueUserIDs := map[string]bool{}
	for i := range events {
		uniqueUserIDs[events[i].UserId] = true
		if !util.IsEmptyPostgresJsonb(&events[i].UserProperties) {
			continue
		}

		if errCode == http.StatusNotFound {
			logCtx.WithField("user_properties_id", events[i].UserPropertiesId).
				Error("User properties on event not found.")
			continue
		}

		// TODO: Remove redundant key generation.
		eventUserPropertiesKey := fmt.Sprintf("%s:%s", events[i].UserId, events[i].UserPropertiesId)
		if _, exists := eventUserPropertiesMap[eventUserPropertiesKey]; !exists {
			eventUserPropertiesMap[eventUserPropertiesKey] = make([]string, 0, 0)
		}
		eventUserPropertiesMap[eventUserPropertiesKey] = append(
			eventUserPropertiesMap[eventUserPropertiesKey],
			events[i].ID,
		)
		uniqueUserPropertiesMap[events[i].UserPropertiesId] = true
	}

	// Download and prepare all user_properties.
	userPropertiesMap := map[string]postgres.Jsonb{}
	userPropertiesIDs := make([]string, 0, 0)
	for userPropertiesID := range uniqueUserPropertiesMap {
		userPropertiesIDs = append(userPropertiesIDs, userPropertiesID)
	}
	errCode = getUserPropertiesByUserAndCachePaginated(projectID, userPropertiesIDs, &userPropertiesMap)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to download and cache user properties in batch.")
		return http.StatusInternalServerError
	}

	// Update user properties on events grouped by user_id, user_properties_id.
	for k, eventIDs := range eventUserPropertiesMap {
		keys := strings.Split(k, ":")
		userID, userPropertiesID := keys[0], keys[1]

		userPropertiesKey := fmt.Sprintf("%s:%s", userID, userPropertiesID)
		userProperties, exists := userPropertiesMap[userPropertiesKey]
		if !exists {
			logCtx.Error("User properties not found on cache.")
			continue
		}

		errCode = updateUserPropertiesToEvents(projectID, eventIDs, userID, &userProperties)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to update user properties to events.")
			continue
		}
	}

	// Prepare missing user_properties and keys for updating
	// latest user_properties to users.
	missingUniqueUserPropertiesKeys := map[string]bool{}
	latestUserProperties := make([]string, 0, 0)
	for userID := range uniqueUserIDs {
		logCtx = logCtx.WithField("user_id", userID)
		user, errCode := getUser(projectID, userID)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get the user on event.")
			return http.StatusInternalServerError
		}

		if !util.IsEmptyPostgresJsonb(&user.Properties) {
			continue
		}

		userPropertiesKey := fmt.Sprintf("%s:%s", userID, user.PropertiesId)
		if _, exists := userPropertiesMap[userPropertiesKey]; !exists {
			missingUniqueUserPropertiesKeys[userPropertiesKey] = true
		}

		latestUserProperties = append(latestUserProperties, userPropertiesKey)
	}

	missingUserPropertiesIDs := make([]string, 0, 0)
	for k := range missingUniqueUserPropertiesKeys {
		keys := strings.Split(k, ":")
		_, userPropertiesID := keys[0], keys[1]
		missingUserPropertiesIDs = append(missingUserPropertiesIDs, userPropertiesID)
	}

	// Download and add the missing user_properties to cache
	// for updating latest user_properties to users.
	errCode = getUserPropertiesByUserAndCachePaginated(projectID, missingUserPropertiesIDs, &userPropertiesMap)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to download and cache user properties in batch.")
		return http.StatusInternalServerError
	}

	// Update latest user properties.
	for _, k := range latestUserProperties {
		keys := strings.Split(k, ":")
		userID := keys[0]

		userProperties, exists := userPropertiesMap[k]
		if !exists {
			logCtx.Error("User properties not found on cache.")
			continue
		}

		errCode = updateUserPropertiesToUser(projectID, userID, &userProperties)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to updated user properties on user.")
			continue
		}
	}

	return http.StatusOK
}

// Script for testing the bulk download performance of MemSQL.
func main() {
	env := flag.String("env", "development", "")

	memSQLDSN := flag.String(
		"memsql_dsn",
		"admin:LIuvIgQDHU@tcp(svc-89fe9813-850d-49e1-864b-aa1a8c600f3c-ddl.gcp-mumbai-1.db.memsql.com:3306)/factors?charset=utf8mb4&parseTime=True&loc=Local",
		"",
	)
	projectID := flag.Uint64("project_id", 0, "")
	startTimestamp := flag.Int64("start_timestamp", 0, "")
	endTimestamp := flag.Int64("end_timestamp", 0, "")

	pageSizeInHours := flag.Int64("page_size_in_hours", 12, "")

	flag.Parse()

	if *projectID == 0 || *startTimestamp == 0 || *endTimestamp == 0 || *pageSizeInHours == 0 {
		log.WithFields(log.Fields{"start_timestap": *startTimestamp,
			"end_timestamp": *endTimestamp, "project_id": *projectID,
		}).Fatal("Invalid flags.")
	}

	initMemSQLDB(*env, *memSQLDSN)

	incrementPeriod := *pageSizeInHours * 60 * 60
	for cursorStartTimestamp, cursorEndTimestamp := *startTimestamp, *startTimestamp+incrementPeriod; cursorStartTimestamp < *endTimestamp; {
		errCode := fillUserPropertiesToEventAndUsersTable(*projectID, cursorStartTimestamp, cursorEndTimestamp)
		if errCode != http.StatusOK {
			log.Error("Failed to fill the user properties.")
			return
		}

		cursorStartTimestamp = cursorEndTimestamp + 1
		cursorEndTimestamp = cursorEndTimestamp + incrementPeriod
	}

	log.Info("Succesfully filled the user_properties.")
}
