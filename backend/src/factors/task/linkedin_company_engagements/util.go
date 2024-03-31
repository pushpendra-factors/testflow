package linkedin_company_engagements

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
linkedin websites, a.b.com, b.b.com, a.c.com, b.c.com, d.com
 1. getGroupingByDomainV2:
    [[a.b.com, b.b.com], [a.c.com, b.c.com], [d.com]]
 2. final return for batch length 2
    [[[a.b.com, b.b.com], [a.c.com, b.c.com]], [[d.com]]]
    `[[a.b.com, b.b.com], [a.c.com, b.c.com]]` is batch 1
 3. rows with same domain will run serially
*/
func getBatchOfDomainDataV2(projectID int64, domainDataSet []model.DomainDataResponse, batchSize int) [][][]model.DomainDataResponse {

	groupedDomainDataSet := getGroupingByDomainV2(projectID, domainDataSet)
	batchedDomainData := make([][][]model.DomainDataResponse, 0)
	for i := 0; i < len(groupedDomainDataSet); i += batchSize {
		end := i + batchSize

		if end > len(groupedDomainDataSet) {
			end = len(groupedDomainDataSet)
		}
		batchedDomainData = append(batchedDomainData, groupedDomainDataSet[i:end])
	}
	return batchedDomainData
}

func getGroupingByDomainV2(projectID int64, domainDataSet []model.DomainDataResponse) [][]model.DomainDataResponse {
	groupedDomainDataSet := make([][]model.DomainDataResponse, 0)
	mapOfDomainToDomainDataSet := make(map[string][]model.DomainDataResponse)
	for _, domainData := range domainDataSet {
		domain := U.GetDomainGroupDomainName(projectID, domainData.Domain)
		if _, exists := mapOfDomainToDomainDataSet[domain]; !exists {
			mapOfDomainToDomainDataSet[domain] = make([]model.DomainDataResponse, 0)
		}
		mapOfDomainToDomainDataSet[domain] = append(mapOfDomainToDomainDataSet[domain], domainData)
	}
	for key, groupedDomainData := range mapOfDomainToDomainDataSet {
		groupedDomainDataSet = append(groupedDomainDataSet, groupedDomainData)
		if len(groupedDomainData) > 10 {
			log.WithFields(log.Fields{"project_id": projectID, "domain": key, "length": len(groupedDomainData)}).Error("Number of rows per domain exceeding 10")
		}
	}
	return groupedDomainDataSet
}
func checkIfEventCreationReqV2(propertyValue int64, domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) (bool, string, string) {
	if propertyValue <= 0 {
		return false, "", ""
	}
	if value, exists := existingEventsWithCampaignData[timestamp][domain][campaignGroupID]; exists {
		return false, value["id"].(string), value["user_id"].(string)
	}

	return true, "", ""
}

func getExistingPropertyValue(domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) float64 {

	return U.SafeConvertToFloat64(existingEventsWithCampaignData[timestamp][domain][campaignGroupID]["p_value"])
}

func getUserIDFromEventsForUpdatingGroupUser(currUserID string, domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) string {
	if _, exists := existingEventsWithCampaignData[timestamp][domain][campaignGroupID]; !exists {
		return currUserID
	}
	return existingEventsWithCampaignData[timestamp][domain][campaignGroupID]["user_id"].(string)
}

type EventCount struct {
	EventCount int64 `json:"event_count"`
}

/*
1. We are checking partial data based on adgroup id property.
2. If adgroup id is present then that means event is of campaign type.
3. If adgroup id is not present then that means event is of campaign_group type.
*/
func checkIfPartialDataIsPresent(projectID int64, timestamp int64, imprEventNameID, clickEventNameID string) bool {
	db := C.GetServices().Db
	startTimestamp := timestamp
	endTimestamp := startTimestamp + 86399

	var eventCountWithCampaignData EventCount
	err := db.Table("events").Select("count(*) as event_count").
		Where("project_id = ? and event_name_id in (?, ?) and timestamp between ? and ? and JSON_EXTRACT_STRING(properties, ?) is not null",
			projectID, imprEventNameID, clickEventNameID, startTimestamp, endTimestamp, U.EP_ADGROUP_ID).
		Find(&eventCountWithCampaignData).Error
	if err != nil {
		log.WithError(err).Error("Failed running partial data check query")
		return false
	}

	var eventCountWithoutCampaignData EventCount
	err = db.Table("events").Select("count(*) as event_count").
		Where("project_id = ? and event_name_id in (?, ?) and timestamp between ? and ? and JSON_EXTRACT_STRING(properties, ?) is null",
			projectID, imprEventNameID, clickEventNameID, startTimestamp, endTimestamp, U.EP_ADGROUP_ID).
		Find(&eventCountWithoutCampaignData).Error
	if err != nil {
		log.WithError(err).Error("Failed running partial data check query")
		return false
	}
	if eventCountWithCampaignData.EventCount > 0 && eventCountWithoutCampaignData.EventCount > 0 {
		return true
	}
	return false
}

func checkIfIncomingDataHasCampaigns(domainDataSet []model.DomainDataResponse) bool {
	return domainDataSet[0].CampaignID != ""
}

func checkIfEventCreationReqV3(propertyValue int64, domain string, timestamp int64, campaignID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) (bool, string, string) {
	if propertyValue <= 0 {
		return false, "", ""
	}
	if value, exists := existingEventsWithCampaignData[timestamp][domain][campaignID]; exists {
		return false, value["id"].(string), value["user_id"].(string)
	}

	return true, "", ""
}

func buildBatchOfGroupUsersBasedOnUserIDs(projectID int64, userIDsMap map[string]UserInfoForDeleteAndUpdate, batchSize int) ([][]model.User, string) {
	allUserIds := make([]string, 0)
	for key := range userIDsMap {
		allUserIds = append(allUserIds, key)
	}
	batchedUserIDs := U.GetStringListAsBatch(allUserIds, 1000)

	allUsers := make([]model.User, 0)
	for _, batch := range batchedUserIDs {
		users, errMsg, errCode := getGroupUsersWithReqFields(projectID, batch)
		if errCode != http.StatusFound {
			return nil, errMsg
		}
		allUsers = append(allUsers, users...)
	}
	return getUsersAsBatch(allUsers, batchSize), ""
}
func deleteEventsAndUpdateAccountProperties(projectID int64, user model.User, userInfo UserInfoForDeleteAndUpdate, imprEventNameID, clickEventNameID string) (string, int) {
	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY)
	if errCode != http.StatusFound {
		return "Failed to get group.", http.StatusInternalServerError
	}
	groupID, err := model.GetGroupUserGroupID(&user, group.ID)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}

	errMsg, errCode := updateAccountLevelPropertiesForGroupUser(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, groupID, user.ID, -userInfo.Impressions, -userInfo.Clicks)
	if errMsg != "" {
		return errMsg, errCode
	}

	// EventNameId is passed only for logging and not used as a filter condition.
	// Since we are performing deletetion on both events, hence passing both eventNames to be looged for reference
	errCode = store.GetStore().DeleteEventByIDs(projectID, imprEventNameID+"_"+clickEventNameID, userInfo.EventIDs)
	if errCode != http.StatusAccepted {
		return "Failed to delete events", errCode
	}
	return "", http.StatusAccepted
}

func getGroupUsersWithReqFields(projectID int64, userIDs []string) ([]model.User, string, int) {
	db := C.GetServices().Db

	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY)
	if errCode != http.StatusFound {
		return nil, "Failed to get group.", http.StatusInternalServerError
	}
	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)

	var users []model.User
	err := db.Select(fmt.Sprintf("id, group_%d_id, is_group_user", group.ID)).
		Where("project_id = ? and source = ? and id in (?)", projectID, source, userIDs).Limit("100000").Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get group users")
		return nil, "Failed to find group users", http.StatusInternalServerError
	}
	return users, "", http.StatusFound
}

func getUsersAsBatch(users []model.User, batchSize int) [][]model.User {
	batchList := make([][]model.User, 0)
	listLen := len(users)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, users[i:next])
		i = next
	}
	return batchList
}

type EventPropertySum struct {
	UserID      string  `json:"user_id"`
	Impressions float64 `json:"impressions"`
	Clicks      float64 `json:"clicks"`
}
type UserInfoForDeleteAndUpdate struct {
	Impressions float64  `json:"impressions"`
	Clicks      float64  `json:"clicks"`
	EventIDs    []string `json:"event_ids"`
}

func getUserInfoDeleteAndUpdate(projectID int64, timestamp int64, imprEventNameID, clickEventNameID string, deleteCampaignType bool) (map[string]UserInfoForDeleteAndUpdate, error) {
	db := C.GetServices().Db
	startTimestamp := timestamp
	endTimestamp := startTimestamp + 86399
	userIDToUserInfoForDeleteAndUpdate := make(map[string]UserInfoForDeleteAndUpdate)

	if !deleteCampaignType {
		propertySumByUserID := make([]EventPropertySum, 0)
		err := db.Table("events").Select("user_id, sum(JSON_EXTRACT_STRING(properties, ?)) as impressions, sum(JSON_EXTRACT_STRING(properties, ?)) as clicks", U.LI_AD_VIEW_COUNT, U.LI_AD_CLICK_COUNT).
			Where("project_id = ? and event_name_id in (?, ?) and timestamp between ? and ? and JSON_EXTRACT_STRING(properties, ?) is null",
				projectID, imprEventNameID, clickEventNameID, startTimestamp, endTimestamp, U.EP_ADGROUP_ID).Group("user_id").
			Find(&propertySumByUserID).Error
		if err != nil {
			log.WithFields(log.Fields{"projectID": projectID, "timestamp": timestamp, "isDeleteCampaignType": deleteCampaignType}).WithError(err).Error("Failed running get property sum query")
			return userIDToUserInfoForDeleteAndUpdate, err
		}

		for _, userInfo := range propertySumByUserID {
			eventIDs := make([]string, 0)
			rows, err := db.Table("events").Select("id").
				Where("project_id = ? and user_id = ? and event_name_id in (?, ?) and timestamp between ? and ? and JSON_EXTRACT_STRING(properties, ?) is null",
					projectID, userInfo.UserID, imprEventNameID, clickEventNameID, startTimestamp, endTimestamp, U.EP_ADGROUP_ID).Rows()
			if err != nil {
				log.WithFields(log.Fields{"projectID": projectID, "timestamp": timestamp, "isDeleteCampaignType": deleteCampaignType}).WithError(err).Error("Failed get eventIDs for each user")
				return userIDToUserInfoForDeleteAndUpdate, err
			}
			defer rows.Close()

			for rows.Next() {
				var eventID string
				if err := rows.Scan(&eventID); err != nil {
					log.WithError(err).Error("Failed to scan event id on getUserInfoDeleteAndUpdate.")
					return userIDToUserInfoForDeleteAndUpdate, err
				}

				eventIDs = append(eventIDs, eventID)
			}
			userIDToUserInfoForDeleteAndUpdate[userInfo.UserID] = UserInfoForDeleteAndUpdate{
				Impressions: userInfo.Impressions,
				Clicks:      userInfo.Clicks,
				EventIDs:    eventIDs,
			}
		}
	} else {
		propertySumByUserID := make([]EventPropertySum, 0)
		err := db.Table("events").Select("user_id, sum(JSON_EXTRACT_STRING(properties, ?)) as impressions, sum(JSON_EXTRACT_STRING(properties, ?)) as clicks", U.LI_AD_VIEW_COUNT, U.LI_AD_CLICK_COUNT).
			Where("project_id = ? and event_name_id in (?, ?) and timestamp between ? and ? and JSON_EXTRACT_STRING(properties, ?) is not null",
				projectID, imprEventNameID, clickEventNameID, startTimestamp, endTimestamp, U.EP_ADGROUP_ID).Group("user_id").
			Find(&propertySumByUserID).Error
		if err != nil {
			log.WithFields(log.Fields{"projectID": projectID, "timestamp": timestamp, "isDeleteCampaignType": deleteCampaignType}).WithError(err).Error("Failed running get property sum query")
			return userIDToUserInfoForDeleteAndUpdate, err
		}
		for _, userInfo := range propertySumByUserID {
			eventIDs := make([]string, 0)
			rows, err := db.Table("events").Select("id").
				Where("project_id = ? and user_id = ? and event_name_id in (?, ?) and timestamp between ? and ? and JSON_EXTRACT_STRING(properties, ?) is not null",
					projectID, userInfo.UserID, imprEventNameID, clickEventNameID, startTimestamp, endTimestamp, U.EP_ADGROUP_ID).Rows()
			if err != nil {
				log.WithFields(log.Fields{"projectID": projectID, "timestamp": timestamp, "isDeleteCampaignType": deleteCampaignType}).WithError(err).Error("Failed get eventIDs for each user")
				return userIDToUserInfoForDeleteAndUpdate, err
			}
			defer rows.Close()

			for rows.Next() {
				var eventID string
				if err := rows.Scan(&eventID); err != nil {
					log.WithError(err).Error("Failed to scan event id on getUserInfoDeleteAndUpdate.")
					return userIDToUserInfoForDeleteAndUpdate, err
				}

				eventIDs = append(eventIDs, eventID)
			}
			userIDToUserInfoForDeleteAndUpdate[userInfo.UserID] = UserInfoForDeleteAndUpdate{
				Impressions: userInfo.Impressions,
				Clicks:      userInfo.Clicks,
				EventIDs:    eventIDs,
			}
		}
	}
	return userIDToUserInfoForDeleteAndUpdate, nil
}

func getUserIDsWithDataMismatch(projectID int64) (string, string, int) {
	userIDs := ""
	db := C.GetServices().Db

	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY)
	if errCode != http.StatusFound {
		return "", "Failed to get group.", http.StatusInternalServerError
	}
	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)

	imprEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get impr event name")
		return "", "Failed to get impression eventname", http.StatusInternalServerError
	}
	clickEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get clicks event name")
		return "", "Failed to get click eventname", http.StatusInternalServerError
	}

	queryStr := "With users_0 as (SELECT id, group_%d_id, is_group_user, JSON_EXTRACT_STRING(properties, '$li_total_ad_view_count') as user_impressions, " +
		"JSON_EXTRACT_STRING(properties, '$li_total_ad_click_count') as user_clicks from users where project_id = ? and source = ?), " +
		"events_0 as (SELECT user_id, sum(JSON_EXTRACT_STRING(properties, '$li_ad_view_count')) as event_impressions, " +
		"Case when sum(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) is null then 0 else " +
		"sum(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) END as event_clicks from events where project_id = ? and event_name_id in (?,?)" +
		"group by user_id order by user_id) SELECT id, group_%d_id, is_group_user from users_0 join events_0 on id=user_id " +
		"where user_impressions != event_impressions or user_clicks != event_clicks"
	queryStr = fmt.Sprintf(queryStr, group.ID, group.ID)
	var users []model.User
	err = db.Raw(queryStr, projectID, source, projectID, imprEventName.ID, clickEventName.ID).Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get group users")
		return "", "Failed to find group users", http.StatusInternalServerError
	}
	userIDsArr := make([]string, 0)
	for _, user := range users {
		userIDsArr = append(userIDsArr, user.ID)
	}
	userIDs = strings.Join(userIDsArr, ",")
	return userIDs, "", http.StatusOK
}
