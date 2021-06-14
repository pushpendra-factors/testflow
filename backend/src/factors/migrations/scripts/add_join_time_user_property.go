package main

import (
	"encoding/json"
	C "factors/config"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to run migration. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	/*

		NOTICE: DEPRECATED - GETTING ALL USERS OF ALL PROJECTS IS NOT RECOMMENDED ANY MORE.
		USE LIMITED SET OF USERS BY PROJECT.

		var users []model.User
		err = db.Find(&users).Error
		if err != nil {
			log.WithError(err).Fatal("Failed to fetch unidentified users.")
		}

	*/

	type userJoinTimestamp struct {
		userIds   []string
		timestamp int64
	}

	var projectCustomerUser map[uint64]map[string]*userJoinTimestamp
	projectCustomerUser = make(map[uint64]map[string]*userJoinTimestamp, 0)

	// Update join time for unidentifed users and collect
	// customer_user_id of identified users.
	for index, user := range users {
		if user.CustomerUserId == "" {

			/*

				NOTICE: DEPRECATED - GETTING ALL USER PROPERTIES RECORDS FOR A GIVEN USER ID
				IS NOT RECOMMENDED ANY MORE. FIND A DIFFERENT WAY TO ACHIEVE THIS.

				var userPropertyRecords []model.UserProperties
				err = db.Where("project_id = ? AND user_id = ?", user.ProjectId, user.ID).Find(&userPropertyRecords).Error
				if err != nil {
					log.WithError(err).Fatal("Failed to fetch current users propery records for a user.")
				}

			*/

			for _, userProperties := range userPropertyRecords {
				var existingProperties map[string]interface{}
				if err := json.Unmarshal(userProperties.Properties.RawMessage, &existingProperties); err != nil {
					log.WithError(err).Fatal("Failed to unmarshal exiting properties.")
				}

				if _, exists := existingProperties[U.UP_JOIN_TIME]; exists {
					log.Info("No update required. Join time already exists.")
					continue
				}

				newPropertiesJsonb, err := U.AddToPostgresJsonb(&userProperties.Properties,
					map[string]interface{}{U.UP_JOIN_TIME: user.JoinTimestamp}, true)
				if err != nil {
					log.WithError(err).Fatal("Failed to add join timestamp to properties.")
				}

				errCode := store.GetStore().OverwriteUserProperties(user.ProjectId, user.ID, userProperties.ID, newPropertiesJsonb)
				if errCode == http.StatusInternalServerError {
					log.WithError(err).Fatal("Failed to replace user properties with join time.")
				}

			}

			log.Infof("Updated %d unidentified users.", index+1)
		} else {
			if _, exists := projectCustomerUser[user.ProjectId]; !exists {
				projectCustomerUser[user.ProjectId] = make(map[string]*userJoinTimestamp, 0)
			}

			if _, exists := projectCustomerUser[user.ProjectId][user.CustomerUserId]; !exists {
				userIds := make([]string, 0, 0)
				userIds = append(userIds, user.ID)
				usersMinJointimestamp := &userJoinTimestamp{
					timestamp: user.JoinTimestamp,
					userIds:   userIds,
				}
				projectCustomerUser[user.ProjectId][user.CustomerUserId] = usersMinJointimestamp

			} else {
				projectCustomerUser[user.ProjectId][user.CustomerUserId].userIds = append(
					projectCustomerUser[user.ProjectId][user.CustomerUserId].userIds,
					user.ID,
				)

				// min join timestamp with same customer user_id.
				if user.JoinTimestamp < projectCustomerUser[user.ProjectId][user.CustomerUserId].timestamp {
					projectCustomerUser[user.ProjectId][user.CustomerUserId].timestamp = user.JoinTimestamp
				}
			}
		}
	}

	// Update join time for identified users using an existing method.
	var counter int
	for projectId, customerUsers := range projectCustomerUser {
		for _, usersMinJoinTimestamp := range customerUsers {
			for _, userId := range usersMinJoinTimestamp.userIds {
				errCode := UpdatePropertyOnAllUserPropertyRecordsIfPropertyNotExist(projectId, userId, U.UP_JOIN_TIME, usersMinJoinTimestamp.timestamp)
				if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
					log.Fatal("Failed to update user properties with join time.")
				}
			}
			counter++

			log.Infof("Updated %d customer users.", counter)
		}
	}

}

func UpdatePropertyOnAllUserPropertyRecordsIfPropertyNotExist(projectId uint64, userId string,
	property string, value interface{}) int {

	/* NOTE: GetUserPropertyRecordsByUserId IS DEPRECATED. NOT USED ON PRODUCTION.
	userPropertyRecords, errCode := store.GetStore().GetUserPropertyRecordsByUserId(projectId, userId)
	if errCode == http.StatusInternalServerError {
		return errCode
	} else if errCode == http.StatusNotFound {
		return http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	for _, userProperties := range userPropertyRecords {
		var propertiesMap map[string]interface{}

		if !U.IsEmptyPostgresJsonb(&userProperties.Properties) {
			err := json.Unmarshal(userProperties.Properties.RawMessage, &propertiesMap)
			if err != nil {
				logCtx.Error("Failed to update user property record. JSON unmarshal failed.")
				continue
			}
		} else {
			propertiesMap = make(map[string]interface{}, 0)
		}

		// Script changes: donot update if property key exist.
		if _, exists := propertiesMap[property]; exists {
			log.Info("No update required. Property already exists.")
			continue
		}

		logCtx = logCtx.WithFields(log.Fields{"properties_id": userProperties.ID, "property": property, "value": value})

		propertiesMap[property] = value
		properitesBytes, err := json.Marshal(propertiesMap)
		if err != nil {
			// log and continue update next user property.
			logCtx.Error("Failed to update user property record. JSON marshal failed.")
			continue
		}
		updatedProperties := postgres.Jsonb{RawMessage: json.RawMessage(properitesBytes)}

		// Triggers multiple updates.
		errCode := store.GetStore().OverwriteUserProperties(projectId, userId, userProperties.ID, &updatedProperties)
		if errCode == http.StatusInternalServerError {
			logCtx.WithError(err).Error("Failed to update user property record. DB query failed.")
			continue
		}
	}

	return http.StatusAccepted
	*/

}
