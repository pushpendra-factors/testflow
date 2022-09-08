package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	"factors/util"
	U "factors/util"
	"fmt"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	// Composite primary key with project_id and uuid.
	ID              string  `gorm:"primary_key:true;type:uuid" json:"id"`
	CustomerEventId *string `json:"customer_event_id"`

	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	// (project_id, event_name_id) -> event_names(project_id, id)
	ProjectId int64  `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `json:"user_id"`

	UserProperties *postgres.Jsonb `json:"user_properties"`
	SessionId      *string         `json:session_id`
	EventNameId    string          `json:"event_name_id"`
	Count          uint64          `json:"count"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties                 postgres.Jsonb `json:"properties,omitempty"`
	PropertiesUpdatedTimestamp int64          `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
	// unix epoch timestamp in seconds.
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CacheEvent struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"ts"`
}

type EventWithProperties struct {
	ID            string          `json:"id"`
	Name          string          `json:"event_name"`
	PropertiesMap U.PropertiesMap `json:"properties_map"`
}

type EventPropertiesWithCount struct {
	Count      int
	Properties U.PropertiesMap
}

type UpdateEventPropertiesParams struct {
	ProjectID                     int64
	EventID                       string
	UserID                        string
	SessionProperties             *util.PropertiesMap
	SessionEventTimestamp         int64
	NewSessionEventUserProperties *postgres.Jsonb
	EventsOfSession               []*Event
}

const cacheIndexUserLastEvent = "user_last_event"
const tableName = "events"

const NewUserSessionInactivityInSeconds int64 = ThirtyMinutesInSeconds
const ThirtyMinutesInSeconds int64 = 30 * 60
const EventsPullLimit = 100000000
const AdwordsPullLimit = 100000000
const FacebookPullLimit = 100000000
const BingPullLimit = 100000000
const LinkedInPullLimit = 100000000
const GoggleOrganicPullLimit = 100000000
const UsersPullLimit = 100000000

func SetCacheUserLastEvent(projectId int64, userId string, cacheEvent *CacheEvent) error {
	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)
	if projectId == 0 || userId == "" {
		logCtx.Error("Invalid project or user id on addToCacheUserLastEventTimestamp")
		return errors.New("invalid project or user id")
	}

	if cacheEvent == nil {
		logCtx.Error("Nil cache event on setCacheUserLastEvent")
		return errors.New("nil cache event")
	}

	cacheEventJson, err := json.Marshal(cacheEvent)
	if err != nil {
		logCtx.Error("Failed cache event json marshal.")
		return err
	}

	key, err := getUserLastEventCacheKey(projectId, userId)
	if err != nil {
		return err
	}

	var additionalExpiryTime int64 = 5 * 60 // 5 mins
	cacheExpiry := NewUserSessionInactivityInSeconds + additionalExpiryTime
	err = cacheRedis.Set(key, string(cacheEventJson), float64(cacheExpiry))
	if err != nil {
		logCtx.WithError(err).Error("Failed to setCacheUserLastEvent.")
	}

	return err
}

func GetCacheUserLastEvent(projectId int64, userId string) (*CacheEvent, error) {
	key, err := getUserLastEventCacheKey(projectId, userId)
	if err != nil {
		return nil, err
	}

	cacheEventJson, err := cacheRedis.Get(key)
	if err != nil {
		return nil, err
	}

	var cacheEvent CacheEvent
	err = json.Unmarshal([]byte(cacheEventJson), &cacheEvent)
	if err != nil {
		return nil, err
	}

	return &cacheEvent, nil
}

func getUserLastEventCacheKey(projectId int64, userId string) (*cacheRedis.Key, error) {
	suffix := fmt.Sprintf("uid:%s", userId)
	prefix := fmt.Sprintf("%s:%s", tableName, cacheIndexUserLastEvent)
	return cacheRedis.NewKey(projectId, prefix, suffix)
}

// AreMarketingPropertiesMatching This method compares given event's marketing props with another event conservatively.
// If new props exist in 2nd event, return false.
func AreMarketingPropertiesMatching(event1 Event, event2 Event) bool {

	eventProp, err := U.DecodePostgresJsonb(&event1.Properties)
	// In case of error, return not matched.
	if err != nil {
		return false
	}
	lastSessionProp, err := U.DecodePostgresJsonb(&event2.Properties)
	if err != nil {
		return false
	}

	for _, marketingProperty := range U.DEFINED_MARKETING_PROPERTIES {
		val1, exists1 := (*eventProp)[marketingProperty]
		val2, exists2 := (*lastSessionProp)[marketingProperty]
		// Treat empty value as absence of property.
		if val1 == "" {
			exists1 = false
		}
		if val2 == "" {
			exists2 = false
		}
		// 2nd event has additional property.
		if exists2 && !exists1 {
			return false
		}
		// Exists but a different property.
		if exists1 && exists2 && val1 != val2 {
			return false
		}
	}
	return true
}

func GetChannelGroup(project Project, sessionPropertiesMap U.PropertiesMap) (string, string) {

	var channelGroupRules []ChannelPropertyRule

	if !U.IsEmptyPostgresJsonb(&project.ChannelGroupRules) {
		err := U.DecodePostgresJsonbToStructType(&project.ChannelGroupRules, &channelGroupRules)
		if err != nil {
			return "", "Failed to decode channel group rules from project"
		}
	} else {
		channelGroupRules = DefaultChannelPropertyRules
	}

	return EvaluateChannelPropertyRules(channelGroupRules, sessionPropertiesMap, project.ID), ""
}

// GetEventListAsBatch - Returns list of events as batches of events list.
func GetEventListAsBatch(list []*Event, batchSize int) [][]*Event {
	batchList := make([][]*Event, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func GetEventsMinMaxTimestampsAndEventnameIds(events []*Event) (int64, int64, []string, []string) {
	fromTimestamp := int64(0)
	toTimestamp := int64(0)

	eventIds := make([]string, 0, 0)
	uniqueEventNameIDs := make(map[string]bool, 0)
	for i := range events {
		event := *events[i]
		eventIds = append(eventIds, event.ID)
		uniqueEventNameIDs[event.EventNameId] = true

		if toTimestamp == 0 || event.Timestamp > toTimestamp {
			toTimestamp = event.Timestamp
		}

		if fromTimestamp == 0 || event.Timestamp < fromTimestamp {
			fromTimestamp = event.Timestamp
		}
	}

	eventNameIds := make([]string, 0, 0)
	for k := range uniqueEventNameIDs {
		eventNameIds = append(eventNameIds, k)
	}

	return fromTimestamp, toTimestamp, eventIds, eventNameIds
}

func GetUpdateEventPropertiesParamsAsBatch(list []UpdateEventPropertiesParams, batchSize int) [][]UpdateEventPropertiesParams {
	batchList := make([][]UpdateEventPropertiesParams, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}
