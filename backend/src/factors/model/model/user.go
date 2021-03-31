package model

import (
	"errors"
	"fmt"
	"time"

	cacheRedis "factors/cache/redis"
	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type User struct {
	// Composite primary key with project_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId    uint64 `gorm:"primary_key:true;" json:"project_id"`
	PropertiesId string `json:"properties_id"`
	// Not part of table, but part of json. Stored in UserProperties table.
	Properties         postgres.Jsonb `gorm:"-" json:"properties"`
	SegmentAnonymousId string         `gorm:"type:varchar(200);default:null" json:"seg_aid"`
	AMPUserId          string         `gorm:"default:null";json:"amp_user_id"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string `gorm:"type:varchar(255);default:null" json:"c_uid"`
	// unix epoch timestamp in seconds.
	JoinTimestamp int64     `json:"join_timestamp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	// Increamenting count based on EventNameId, not by EventName.
	if user.JoinTimestamp <= 0 {
		// Default to 60 seconds earlier than now, so that if event is also created simultaneously
		// user join is earlier.
		user.JoinTimestamp = time.Now().Unix() - 60
	}

	// adds join timestamp to user properties.
	newProperties := map[string]interface{}{
		U.UP_JOIN_TIME: user.JoinTimestamp,
	}
	newPropertiesJsonb, err := U.AddToPostgresJsonb(&user.Properties, newProperties, true)
	if err != nil {
		return err
	}
	user.Properties = *newPropertiesJsonb

	return nil
}

func GetIdentifiedUserPropertiesAsJsonb(customerUserId string) (*postgres.Jsonb, error) {
	if customerUserId == "" {
		return nil, errors.New("invalid customer user id")
	}

	properties := map[string]interface{}{
		U.UP_USER_ID: customerUserId,
	}

	if U.IsEmail(customerUserId) {
		properties[U.UP_EMAIL] = customerUserId
	}

	return U.EncodeToPostgresJsonb(&properties)
}

// Today's cache keys
func GetUsersCachedCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:LIST"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetUserPropertiesCategoryByProjectCacheKey(projectId uint64, property string, category string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:PC"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s:%s", dateKey, category, property))

}

func GetValuesByUserPropertyCacheKey(projectId uint64, property_name string, value string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, property_name), fmt.Sprintf("%s:%s", dateKey, value))
}

//sorted sets 
func GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:US:PC"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s", dateKey))

}

func GetValuesByUserPropertyCacheKeySortedSet(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s", prefix), fmt.Sprintf("%s", dateKey))
}

// Rollup cache keys
func GetUserPropertiesCategoryByProjectRollUpCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "RollUp:US:PC"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetValuesByUserPropertyRollUpCacheKey(projectId uint64, property_name string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "RollUp:US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, property_name), dateKey)
}

// Today's cache keys count
func GetUserPropertiesCategoryByProjectCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:US:PC"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

func GetValuesByUserPropertyCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:US:PV"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}
