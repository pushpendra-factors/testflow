package main

import (
	C "factors/config"

	"factors/util"
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

var o_convert_timestamp_key = "ocots"
var o_close_timestamp_key = "oclts"
var o_group_user_id_key = "ouid"

var a_opportunity_id_key = "aoid"
var a_group_user_id_key = "aguid"
var a_user_ids_list_key = "auidl"
var a_cust_user_ids_set_key = "acuids"
var a_num_convert_touchpoints_key = "ancot"
var a_num_close_touchpoints_key = "anclt"

// Fetch List of all Groups with the given group conversion event.
func getOpportunitesForConversion(
	projectID int64,
	opportunityGroupTypeID, createdDateProperty, closeDateProperty string,
	startTime, endTime int64) (map[string]map[string]interface{}, error) {

	db := C.GetServices().Db
	convertedOpportunities := make(map[string]map[string]interface{})

	rawQuery := "select id, " + opportunityGroupTypeID + ", JSON_EXTRACT_STRING(properties, ?), JSON_EXTRACT_STRING(properties, ?) " +
		"from users where project_id=? AND is_group_user=1 AND " +
		opportunityGroupTypeID + " IS NOT NULL AND JSON_EXTRACT_STRING(properties, ?) between ? AND ?;"
	rows, err := db.Raw(rawQuery, createdDateProperty, closeDateProperty, projectID, createdDateProperty, startTime, endTime).Rows()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "query": rawQuery}).Error("Failed getting converted groups")
		return convertedOpportunities, err
	}

	for rows.Next() {
		var opportunityUserID, groupID string
		var convertTimestamp, closeTimestamp int64
		err = rows.Scan(&opportunityUserID, &groupID, &convertTimestamp, &closeTimestamp)
		if err != nil {
			log.WithError(err).Error("Error while scanning row for getting converted groups")
			continue
		}
		if _, ok := convertedOpportunities[groupID]; ok {
			log.Info("Skipping Group " + groupID + ". Already present in map.")
			continue
		}
		convertedOpportunities[groupID] = make(map[string]interface{})
		convertedOpportunities[groupID][o_group_user_id_key] = opportunityUserID
		convertedOpportunities[groupID][o_convert_timestamp_key] = convertTimestamp
		convertedOpportunities[groupID][o_close_timestamp_key] = closeTimestamp
	}

	err = rows.Err()
	if err != nil {
		log.WithError(err).Error("Error while scanning row at the end of fetching converted groups.")
		return convertedOpportunities, err
	}
	return convertedOpportunities, nil
}

func getConvertedAccountsForOpportunities(projectID int64, accountGroupTypeID string,
	convertedOpportunities map[string]map[string]interface{}) (
	map[string]map[string]interface{}, error) {
	db := C.GetServices().Db

	convertedAccounts := make(map[string]map[string]interface{})
	for oppID, value := range convertedOpportunities {
		opportunityUserID, _ := value[o_group_user_id_key]
		var accountGroupID, accountGroupUserID string
		rawQuery := "select right_group_user_id, users." + accountGroupTypeID +
			" from group_relationships JOIN users ON users.group_1_user_id=group_relationships.right_group_user_id where left_group_user_id = ? and right_group_name_id=1 and left_group_name_id=2 and group_relationships.project_id = ? AND users.project_id = ?;"
		row := db.Raw(rawQuery, opportunityUserID.(string), projectID, projectID).Row()
		err := row.Scan(&accountGroupUserID, &accountGroupID)
		if err != nil {
			//log.WithFields(log.Fields{"error": err, "oppUserID": opportunityUserID, "query": rawQuery}).Error("Failed getting accountID for oppID")
			continue
		}
		convertedAccounts[accountGroupID] = value
		convertedAccounts[accountGroupID][a_opportunity_id_key] = oppID
		convertedAccounts[accountGroupID][a_group_user_id_key] = accountGroupUserID
	}
	return convertedAccounts, nil
}

// Add to each converted group the list of user ids (coalesced id) that belong to that group.
func addAccountsUsersInfo(
	projectID int64, accountGroupTypeID string, convertedAccounts map[string]map[string]interface{}) error {
	db := C.GetServices().Db

	allGroupIDs := make([]string, len(convertedAccounts))
	i := 0
	for k := range convertedAccounts {
		allGroupIDs[i] = k
		i++
	}

	groupIDsInBatches := util.GetStringListAsBatch(allGroupIDs, 1000)
	for _, groupIDsBatch := range groupIDsInBatches {
		arrayPlaceHolder := util.GetValuePlaceHolder(len(groupIDsBatch))
		arrayPlaceHolderValue := util.GetInterfaceList(groupIDsBatch)
		rawQuery := "select users.id, COALESCE(users.customer_user_id,users.id), " + accountGroupTypeID + " from users where " + accountGroupTypeID + " IN (" +
			arrayPlaceHolder + ") AND is_group_user=0 AND project_id=?;"
		arrayPlaceHolderValue = append(arrayPlaceHolderValue, projectID)
		rows, err := db.Raw(rawQuery, arrayPlaceHolderValue...).Rows()
		if err != nil {
			log.WithFields(log.Fields{"error": err, "query": rawQuery}).Error("Failed getting group users")
			return err
		}

		for rows.Next() {
			var userID, custUserID, groupID string
			err = rows.Scan(&userID, &custUserID, &groupID)
			if err != nil {
				log.WithError(err).Error("Error while scanning row for fetching group users")
				continue
			}

			var custUserIDs map[string]bool
			custUserIdsInterface, ok := convertedAccounts[groupID][a_cust_user_ids_set_key]
			if !ok {
				custUserIDs = make(map[string]bool)
			} else {
				custUserIDs = custUserIdsInterface.(map[string]bool)
			}

			if _, ok := custUserIDs[custUserID]; !ok {
				custUserIDs[custUserID] = true
			}
			convertedAccounts[groupID][a_cust_user_ids_set_key] = custUserIDs

			var userIDsList []string
			userIDsListInterface, ok := convertedAccounts[groupID][a_user_ids_list_key]
			if !ok {
				userIDsList = []string{}
			} else {
				userIDsList = userIDsListInterface.([]string)
			}
			userIDsList = append(userIDsList, userID)
			convertedAccounts[groupID][a_user_ids_list_key] = userIDsList
		}

		err = rows.Err()
		if err != nil {
			log.WithError(err).Error("Error while scanning row at the end of fetching group users.")
			return err
		}
	}
	return nil
}

func expandUserIdsFromcustUsrIds(projectID int64, custUserIDs map[string]bool, userIds []string) []string {
	db := C.GetServices().Db

	userIdsSet := map[string]bool{}
	for _, userId := range userIds {
		userIdsSet[userId] = true
	}

	custUserIdsArray := []string{}
	for cid, _ := range custUserIDs {
		custUserIdsArray = append(custUserIdsArray, cid)
	}

	arrayPlaceHolder := util.GetValuePlaceHolder(len(custUserIdsArray))
	arrayPlaceHolderValue := util.GetInterfaceList(custUserIdsArray)

	queryStr := "select id from users where project_id=? AND customer_user_id IN (" + arrayPlaceHolder + ")"
	arrayPlaceHolderValue = append([]interface{}{projectID}, arrayPlaceHolderValue...)
	rows, err := db.Raw(queryStr, arrayPlaceHolderValue...).Rows()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "query": queryStr}).Error("Failed expandUserIdsFromcustUsrIds")
		return userIds
	}

	for rows.Next() {
		var userID string
		err = rows.Scan(&userID)
		if err != nil {
			log.WithError(err).Error("Error while scanning row for expandUserIdsFromcustUsrIds.")
			continue
		}
		if _, ok := userIdsSet[userID]; !ok {
			userIdsSet[userID] = true
			userIds = append(userIds, userID)
		}
	}

	err = rows.Err()
	if err != nil {
		log.WithError(err).Error("Error while scanning row at the end of expandUserIdsFromcustUsrIds.")
	}
	return userIds
}

// Add to each converted group the number of given touchpoints by all the users in the group.
func addNumTouchPointsBeforeAndAfterConversion(projectID int64, touchPointEventId string,
	touchPointEventPropertiesFilterName string, touchPointEventPropertiesFilterValue string,
	convertedAccounts map[string]map[string]interface{}) error {
	db := C.GetServices().Db

	for groupID, groupInfo := range convertedAccounts {
		userIdsInterface, ok := groupInfo[a_user_ids_list_key]
		if !ok {
			continue
		}
		userIds := userIdsInterface.([]string)

		custUserIDsInterface, ok := groupInfo[a_cust_user_ids_set_key]
		if !ok {
			continue
		}
		custUserIDs := custUserIDsInterface.((map[string]bool))

		userIds = expandUserIdsFromcustUsrIds(projectID, custUserIDs, userIds)
		groupInfo[a_user_ids_list_key] = userIds
		if len(userIds) == 0 {
			continue
		}

		conversionTimestampInterface, ok := groupInfo[o_convert_timestamp_key]
		if !ok {
			continue
		}
		conversionTimestamp := conversionTimestampInterface.(int64) + 7200 // +2hr buffer.

		closeTimestampInterface, ok := groupInfo[o_close_timestamp_key]
		if !ok {
			continue
		}
		closeTimestamp := closeTimestampInterface.(int64) + 7200 // +2hr buffer.

		var numConvertTouchpoints, numCloseTouchpoints int64
		arrayPlaceHolder := util.GetValuePlaceHolder(len(userIds))
		arrayPlaceHolderValue := util.GetInterfaceList(userIds)

		queryStr := "SELECT COUNT(events.id) FROM events WHERE events.project_id = ? AND events.event_name_id = ? AND JSON_EXTRACT_STRING(events.properties, ? ) = ? AND events.timestamp < ? AND events.user_id IN (" + arrayPlaceHolder + ")"
		convertArrayPlaceHolderValue := append([]interface{}{projectID, touchPointEventId,
			touchPointEventPropertiesFilterName, touchPointEventPropertiesFilterValue,
			conversionTimestamp}, arrayPlaceHolderValue...)
		row := db.Raw(queryStr, convertArrayPlaceHolderValue...).Row()
		err := row.Scan(&numConvertTouchpoints)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "query": queryStr}).Error("Failed getting num convert touchpoints")
			return err
		}
		convertedAccounts[groupID][a_num_convert_touchpoints_key] = numConvertTouchpoints

		queryStr = "SELECT COUNT(events.id) FROM events WHERE events.project_id = ? AND events.event_name_id = ? AND JSON_EXTRACT_STRING(events.properties, ? ) = ? AND events.timestamp > ? AND events.timestamp < ? AND events.user_id IN (" + arrayPlaceHolder + ")"
		closeArrayPlaceHolderValue := append([]interface{}{projectID, touchPointEventId,
			touchPointEventPropertiesFilterName, touchPointEventPropertiesFilterValue,
			conversionTimestamp, closeTimestamp}, arrayPlaceHolderValue...)
		row = db.Raw(queryStr, closeArrayPlaceHolderValue...).Row()
		err = row.Scan(&numCloseTouchpoints)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "query": queryStr}).Error("Failed getting num close touchpoints")
			return err
		}
		convertedAccounts[groupID][a_num_close_touchpoints_key] = numCloseTouchpoints
	}

	return nil
}

/*
Count number of touchpoints (for a given touchpoint event) for each group, prior to a given group conversion event.
go run run_count_account_touchpoints.go  --project_id=<projectId>
*/
func main() {

	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	projectID := flag.Int64("project_id", 659, "Project Id.")
	startTimestamp := flag.Int64("start_timestamp", 1571565438, "Start event timestamp.")
	endTimestamp := flag.Int64("end_timestamp", 1666259840, "End event timestamp.")

	flag.Parse()
	defer util.NotifyOnPanic("Task#run_count_account_touchpoints.", *env)

	taskID := "run_count_account_touchpoints"

	if *projectID == 0 {
		log.Error("projectId not provided")
		os.Exit(1)
	}
	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:           *sentryDSN,
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	logCtx := log.WithFields(log.Fields{"project_id": *projectID})

	// conversionEventName := "$salesforce_opportunity_created"
	// conversionEventID := "9927162b-49f4-430e-8e95-b256dbe48c09"
	opportunityGroupTypeID := "group_2_id"
	accountGroupTypeID := "group_1_id"
	createdDateProperty := "$salesforce_opportunity_createddate"
	closeDateProperty := "$salesforce_opportunity_closedate"
	convertedOpportunities, err := getOpportunitesForConversion(*projectID,
		opportunityGroupTypeID, createdDateProperty, closeDateProperty, *startTimestamp, *endTimestamp)
	if err != nil {
		logCtx.WithFields(log.Fields{"error": err}).Error("Failed getting converted opportnities")
		return
	}

	convertedAccounts, err := getConvertedAccountsForOpportunities(*projectID, accountGroupTypeID, convertedOpportunities)
	if err != nil {
		logCtx.WithFields(log.Fields{"error": err}).Error("Failed getting converted accounts")
		return
	}

	if err := addAccountsUsersInfo(*projectID, accountGroupTypeID, convertedAccounts); err != nil {
		logCtx.WithFields(log.Fields{"error": err}).Error("Failed getting groups info")
		return
	}

	// touchPointEventName := "$sf_campaign_member_updated"
	touchPointEventId := "2404de5c-47eb-4ad7-8ea9-85b1cf9f1ce3"
	touchPointEventPropertiesFilterName := "$salesforce_campaignmember_status"
	touchPointEventPropertiesFilterValue := "Responded"
	if err := addNumTouchPointsBeforeAndAfterConversion(*projectID, touchPointEventId,
		touchPointEventPropertiesFilterName, touchPointEventPropertiesFilterValue, convertedAccounts); err != nil {
		logCtx.WithFields(log.Fields{"error": err}).Error("Failed getting num touchpoints")
		return
	}

	// Log the results
	for accountID, groupInfo := range convertedAccounts {
		uniqUsers := groupInfo[a_cust_user_ids_set_key]
		log.WithFields(log.Fields{
			"accountID":             accountID,
			"opportunityID":         groupInfo[a_opportunity_id_key],
			"numUniqUsers":          len(uniqUsers.(map[string]bool)),
			"numConvertTouchpoints": groupInfo[a_num_convert_touchpoints_key],
			"numCloseTouchpoints":   groupInfo[a_num_close_touchpoints_key],
			"uniqUsers":             uniqUsers,
		}).Info("")
	}
}
