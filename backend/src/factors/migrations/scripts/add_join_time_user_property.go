package main

import (
	"encoding/json"
	C "factors/config"
	M "factors/model"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	numRoutines := flag.Int("num_routines", 2, "")

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
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Fatal("Failed to run migration. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	var users []M.User
	err = db.Find(&users).Error
	if err != nil {
		log.WithError(err).Fatal("Failed to fetch unidentified users.")
	}

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
			var userPropertyRecords []M.UserProperties
			err = db.Where("project_id = ? AND user_id = ?", user.ProjectId, user.ID).Find(&userPropertyRecords).Error
			if err != nil {
				log.WithError(err).Fatal("Failed to fetch current users propery records for a user.")
			}

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
					map[string]interface{}{U.UP_JOIN_TIME: user.JoinTimestamp})
				if err != nil {
					log.WithError(err).Fatal("Failed to add join timestamp to properties.")
				}

				errCode := M.OverwriteUserProperties(user.ProjectId, user.ID, userProperties.ID, newPropertiesJsonb)
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
				// min join timestamp with same customer user_id.
				if user.JoinTimestamp < projectCustomerUser[user.ProjectId][user.CustomerUserId].timestamp {
					projectCustomerUser[user.ProjectId][user.CustomerUserId].timestamp = user.JoinTimestamp
					projectCustomerUser[user.ProjectId][user.CustomerUserId].userIds = append(
						projectCustomerUser[user.ProjectId][user.CustomerUserId].userIds,
						user.ID,
					)
				}
			}
		}
	}

	// Update join time for identified users using an existing method.
	joinTimeUpdates := make([][]interface{}, 0, 0)
	for projectId, customerUsers := range projectCustomerUser {
		for _, usersMinJoinTimestamp := range customerUsers {
			update := make([]interface{}, 0, 0)
			for _, userId := range usersMinJoinTimestamp.userIds {
				update = append(update, projectId, userId, U.UP_JOIN_TIME, usersMinJoinTimestamp.timestamp)
			}
			joinTimeUpdates = append(joinTimeUpdates, update)
		}
	}

	// Use go routines.
	joinTimeUpdatesSplit := splitArray(joinTimeUpdates, *numRoutines)
	var wg sync.WaitGroup
	var routineId int
	for _, updates := range joinTimeUpdatesSplit {
		wg.Add(1)
		go updateJoinTimeForIdentifiedUsers(&wg, updates, routineId)
		routineId++
	}

	wg.Wait()
}

func updateJoinTimeForIdentifiedUsers(wg *sync.WaitGroup, joinTimeUpdates [][]interface{}, routineId int) {
	total := len(joinTimeUpdates)
	var updated int
	for _, col := range joinTimeUpdates {
		errCode := M.UpdatePropertyOnAllUserPropertyRecords(col[0].(uint64), col[1].(string), col[2].(string), col[3])
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			log.Error("Failed to update user properties with join time.")
		}
		log.Infof("Routine [%d]: updated %d of %d..", routineId, updated, total)
		updated++
	}

	wg.Done()
}

func splitArray(src [][]interface{}, nos int) [][][]interface{} {
	var divided [][][]interface{}

	chunkSize := (len(src) + nos - 1) / nos
	for i := 0; i < len(src); i += chunkSize {
		end := i + chunkSize

		if end > len(src) {
			end = len(src)
		}

		divided = append(divided, src[i:end])
	}

	return divided
}
