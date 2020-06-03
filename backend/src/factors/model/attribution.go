package model

import (
	"database/sql"
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type AttributionQuery struct {
	Class                  string   `json:"cl"`
	CampaignMetrics        []string `json:"cm"`
	ConversionEvent        string   `json:"ce"`
	LinkedEvents           []string `json:"lfe"`
	AttributionKey         string   `json:"attribution_key"`
	AttributionMethodology string   `json:"attribution_methodology"`
	LoopbackDays           int      `json:"lbw"`
	From                   int64    `json:"from"`
	To                     int64    `json:"to"`
}

const (
	ATTRIBUTION_METHOD_FIRST_TOUCH = "First_Touch"
	ATTRIBUTION_METHOD_LAST_TOUCH  = "Last_Touch"

	ATTRIBUTION_KEY_CAMPAIGN = "Campaign"
	ATTRIBUTION_KEY_SOURCE   = "Source"
)

var ATTRIBUTION_FIXED_HEADERS = []string{"Impressions", "Clicks", "Spend", "Website Visitors"}

type AttributionData struct {
	Name                 string
	Impressions          int
	Clicks               int
	Spend                float64
	WebsiteVisitors      int64
	ConversionEventCount int64
	LinkedEventsCount    []int64
}

//GetQueryUserProperty returns Touch based on attribution source and method.
func GetQueryUserProperty(query *AttributionQuery) (string, error) {
	if query.AttributionKey == ATTRIBUTION_KEY_CAMPAIGN {
		if query.AttributionMethodology == ATTRIBUTION_METHOD_FIRST_TOUCH {
			return U.UP_INITIAL_CAMPAIGN, nil
		} else if query.AttributionMethodology == ATTRIBUTION_METHOD_LAST_TOUCH {
			return U.UP_LATEST_CAMPAIGN, nil
		}
	} else if query.AttributionKey == ATTRIBUTION_KEY_SOURCE {
		if query.AttributionMethodology == ATTRIBUTION_METHOD_FIRST_TOUCH {
			return U.UP_INITIAL_SOURCE, nil
		} else if query.AttributionMethodology == ATTRIBUTION_METHOD_LAST_TOUCH {
			return U.UP_LATEST_SOURCE, nil
		}
	}
	return "", errors.New("invalid query properties")
}

//GetUniqueUserIntersectionCount adds all common users in conversion event and linked event group by campagn id.
func GetUniqueUserIntersectionCount(linkedEvents []string, attributionData map[string]*AttributionData, userCampaignData map[string]string, linkedUserCampaignData map[string]map[string]string) error {
	for position, linkedEvent := range linkedEvents {
		for userId, userCampaignId := range userCampaignData {
			campaignData := attributionData[userCampaignId]
			if len(campaignData.LinkedEventsCount) != position+1 {
				campaignData.LinkedEventsCount = append(campaignData.LinkedEventsCount, 0) //for all campaign id, irrespective of the match found or could lead to broken position
			}

			if linkedUserCampaignId, exist := linkedUserCampaignData[linkedEvent][userId]; exist {
				if userCampaignId == linkedUserCampaignId {
					campaignData.LinkedEventsCount[position] += 1
				}
			}
		}
	}

	return nil
}

//addHeadersByAttributionKey builds column headers based on query and linked events length
func addHeadersByAttributionKey(result *QueryResult, query *AttributionQuery) {
	attributionKey := query.AttributionKey
	result.Headers = append(append(result.Headers, attributionKey), ATTRIBUTION_FIXED_HEADERS...)
	conversionEventUsers := fmt.Sprintf("%s - Users", query.ConversionEvent)
	result.Headers = append(result.Headers, conversionEventUsers)
	if len(query.LinkedEvents) > 0 {
		for _, event := range query.LinkedEvents {
			result.Headers = append(result.Headers, fmt.Sprintf("%s - Users", event))
		}
	}
}

//ExecuteAttributionQuery - starting point for building attribution data
func ExecuteAttributionQuery(projectId uint64, query *AttributionQuery) (*QueryResult, string, error) {
	result := &QueryResult{}
	attributionData := make(map[string]*AttributionData)

	projectSetting, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return nil, "", errors.New("failed to get project Settings")
	}

	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		return nil, "", errors.New("execute attribution query failed. No customer account id.")
	}

	addHeadersByAttributionKey(result, query)
	userProperty, err := GetQueryUserProperty(query)
	if err != nil {
		return nil, "", err
	}

	err = GetUniqueUserByAttributionKeyAndLookbackWindow(projectId, attributionData, query, userProperty)
	if err != nil {
		return nil, "", err
	}

	err = AddWebsiteVisitorsByEventName(projectId, attributionData, query.From, query.To, userProperty)
	if err != nil {
		return nil, "", err
	}

	currency, err := AddPerformanceReportByCampaign(projectId, attributionData, query.From, query.To, projectSetting.IntAdwordsCustomerAccountId)
	if err != nil {
		return nil, "", err
	}

	result.Rows = getRowsByMaps(attributionData, query)
	return result, currency, nil
}

//getRowsByMaps converts maps to rows. For campaign id having no matching campaign name, it is appended to "non" row.
func getRowsByMaps(attributionData map[string]*AttributionData, query *AttributionQuery) [][]interface{} {

	rows := make([][]interface{}, 0)
	nonMatchingRow := []interface{}{"none", 0, 0, float64(0), int64(0), int64(0)}
	for i := 0; i < len(query.LinkedEvents); i++ {
		nonMatchingRow = append(nonMatchingRow, int64(0))
	}

	for _, campaignData := range attributionData {

		if campaignData.Name != "" { //matching campaing name section
			row := make([]interface{}, 6+len(query.LinkedEvents)) // 6 - fixed column header
			row[0] = campaignData.Name
			row[1] = campaignData.Impressions
			row[2] = campaignData.Clicks
			row[3] = campaignData.Spend
			row[4] = campaignData.WebsiteVisitors
			row[5] = campaignData.ConversionEventCount
			for index := 0; index < len(query.LinkedEvents); index++ {
				if index < len(campaignData.LinkedEventsCount) {
					row[5+index+1] = campaignData.LinkedEventsCount[index] //linked event index positioning
					continue
				}
				row[5+index+1] = 0
			}
			rows = append(rows, row)
		} else { //non matching campaign name section
			nonMatchingRow[1] = nonMatchingRow[1].(int) + campaignData.Impressions
			nonMatchingRow[2] = nonMatchingRow[2].(int) + campaignData.Clicks
			nonMatchingRow[3] = nonMatchingRow[3].(float64) + campaignData.Spend
			nonMatchingRow[4] = nonMatchingRow[4].(int64) + campaignData.WebsiteVisitors
			nonMatchingRow[5] = nonMatchingRow[5].(int64) + campaignData.ConversionEventCount
			for index := 0; index < len(query.LinkedEvents); index++ {
				if index < len(campaignData.LinkedEventsCount) {
					nonMatchingRow[5+index+1] = nonMatchingRow[5+index+1].(int64) + campaignData.LinkedEventsCount[index]
				}
			}
		}
	}
	rows = append(rows, nonMatchingRow)

	return rows
}

//getUniqueUserCount groups all unique user by campaign id and adds it to attributionData
func getUniqueUserCount(attributionData map[string]*AttributionData, users interface{}) error {
	for _, campaignId := range users.(map[string]string) {
		if _, exists := attributionData[campaignId]; !exists {
			attributionData[campaignId] = &AttributionData{}
		}
		attributionData[campaignId].ConversionEventCount += 1
	}
	return nil
}

//getAttributedSession finds the timestamp when attributionKey was initialy set for the user
func getAttributedSession(projectId uint64, attributionKey string, from int64, to int64) (map[string]map[string]int64, error) {
	var aggrigateColumn string
	var timeRangeStmnt string
	db := C.GetServices().Db
	userInitialSession := make(map[string]map[string]int64)

	logCtx := log.WithFields(log.Fields{"projectId": projectId})
	queryParams := []interface{}{attributionKey, projectId, U.EVENT_NAME_SESSION, projectId, attributionKey}

	if attributionKey == U.UP_LATEST_CAMPAIGN {
		timeRangeStmnt = "AND sessions.timestamp BETWEEN ? AND ? "
		aggrigateColumn = "max(sessions.timestamp) "
		queryParams = append(queryParams, from, to)
	} else {
		aggrigateColumn = "min(sessions.timestamp) "
	}

	queryStmnt := "SELECT coalesce(users.customer_user_id,users.id ) AS coal_user_id, user_properties.properties->>? as campaing_id," + aggrigateColumn +
		"from events as sessions left join user_properties on user_properties.id = sessions.user_properties_id left join users on sessions.user_id = users.id where " +
		"sessions.project_id = ? and sessions.event_name_id = (select id from event_names where name=? and " +
		"project_id =? limit 1) and user_properties.properties->>? is not null " + timeRangeStmnt + "group by coal_user_id,campaing_id "
	rows, err := db.Raw(queryStmnt, queryParams...).Rows()

	defer rows.Close()

	if err != nil {
		logCtx.WithFields(log.Fields{"err": err}).Error("SQL Query failed for getUserInitialSession")
		return userInitialSession, err
	}

	for rows.Next() {
		var userId string
		var timestamp int64
		var campaign string
		if err = rows.Scan(&userId, &campaign, &timestamp); err != nil {
			logCtx.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			continue
		}

		if _, exists := userInitialSession[userId][campaign]; !exists {
			userInitialSession[userId] = make(map[string]int64)
			userInitialSession[userId][campaign] = timestamp
			continue
		}
	}
	return userInitialSession, nil
}

/*
GetUniqueUserByAttributionKeyAndLookbackWindow finds all users campaign id, group them by campaign id.
If lookback window is provided then events are pruned before inserting users. Valid for both conversion event and linked event
*/
func GetUniqueUserByAttributionKeyAndLookbackWindow(projectId uint64, attributionData map[string]*AttributionData, query *AttributionQuery, attributionKey string) error {
	logctx := log.WithFields(log.Fields{"projectId": projectId})
	userCampaignData, linkedUserCampaignData, err := getUserCampaignIdByAttributionKey(projectId, query, attributionKey)
	if err != nil {
		logctx.WithError(err).Error("Failed to getUserCampaignIdByAttributionKey.")
		return err
	}

	getUniqueUserCount(attributionData, userCampaignData)
	GetUniqueUserIntersectionCount(query.LinkedEvents, attributionData, userCampaignData, linkedUserCampaignData)
	return nil
}

func buildEventNamesPlaceholder(query *AttributionQuery) (string, []interface{}) {
	var eventNamesStmnt string
	eventNamesParamas := make([]interface{}, 0)

	//append conversition event
	eventNamesStmnt = eventNamesStmnt + "?"
	eventNamesParamas = append(eventNamesParamas, query.ConversionEvent)

	//append linked events
	for _, eventName := range query.LinkedEvents {
		eventNamesStmnt = eventNamesStmnt + ",?"
		eventNamesParamas = append(eventNamesParamas, eventName)
	}

	return eventNamesStmnt, eventNamesParamas
}

//getUserCampaignIdByAttributionKey returns a map of user to campaign id. If lookback window is provided then event are pruned before inserting to map
func getUserCampaignIdByAttributionKey(projectId uint64, query *AttributionQuery, attributionKey string) (map[string]string, map[string]map[string]string, error) {
	db := C.GetServices().Db
	usersAttribution := make(map[string]string)
	linkedEventUserCampaign := make(map[string]map[string]string)
	isLookbackActive := query.LoopbackDays > 0
	var userInitialSession map[string]map[string]int64
	var err error
	if isLookbackActive {
		userInitialSession, err = getAttributedSession(projectId, attributionKey, query.From, query.To)
		if err != nil {
			return nil, nil, err
		}
	}

	lookbackDaysTimestamp := int64(query.LoopbackDays) * 86400

	logCtx := log.WithFields(log.Fields{"projectId": projectId})

	groupSelect := "CASE WHEN user_properties.properties->>? IS NULL THEN '$none' WHEN user_properties.properties->>? = '' THEN '$none' ELSE user_properties.properties->>? END AS campaign"
	groupSelectParams := []interface{}{attributionKey, attributionKey, attributionKey}
	qparams := append(groupSelectParams, projectId)

	inStmnt, inParams := buildEventNamesPlaceholder(query)
	qparams = append(append(qparams, inParams...), query.From, query.To)

	stmnt := "SELECT coalesce(users.customer_user_id,users.id ) AS user_id," + groupSelect +
		", events.timestamp as event_timestamp, event_names.name FROM events LEFT JOIN users ON users.id = events.user_id LEFT JOIN user_properties" +
		" ON user_properties.id = users.properties_id LEFT JOIN event_names ON event_names.id = events.event_name_id WHERE events.project_id=? AND event_names.name IN (" + inStmnt +
		") AND timestamp >= ? AND timestamp <=?"

	rows, err := db.Raw(stmnt, qparams...).Rows()

	defer rows.Close()

	if err != nil {
		logCtx.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return nil, nil, err
	}

	for rows.Next() {
		var userId sql.NullString
		var campaign sql.NullString
		var timestamp sql.NullInt64
		var eventName string
		if err = rows.Scan(&userId, &campaign, &timestamp, &eventName); err != nil {
			logCtx.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			continue
		}

		if campaign.Valid && userId.Valid && timestamp.Valid {

			if isLookbackActive {
				var timeSinceEvent int64
				if attributionKey == U.UP_LATEST_CAMPAIGN {
					//Since we applied window in Last Touch there could be some session not present for corner event
					if attributedTimestamp, exist := userInitialSession[userId.String][campaign.String]; exist {
						timeSinceEvent = attributedTimestamp - timestamp.Int64
					} else {
						continue
					}

				} else {
					timeSinceEvent = timestamp.Int64 - userInitialSession[userId.String][campaign.String]
				}

				if timeSinceEvent <= lookbackDaysTimestamp && timeSinceEvent >= 0 {

					if eventName == query.ConversionEvent {
						if _, exists := usersAttribution[userId.String]; exists {
							continue
						}

						usersAttribution[userId.String] = campaign.String
					} else {
						if _, exist := linkedEventUserCampaign[eventName]; !exist {
							linkedEventUserCampaign[eventName] = make(map[string]string)
						}
						linkedEventUserCampaign[eventName][userId.String] = campaign.String
					}
				}

				continue
			}

			if eventName == query.ConversionEvent {
				if _, exists := usersAttribution[userId.String]; exists {
					continue
				}

				usersAttribution[userId.String] = campaign.String
			} else {
				if _, exist := linkedEventUserCampaign[eventName]; !exist {
					linkedEventUserCampaign[eventName] = make(map[string]string)
				}
				linkedEventUserCampaign[eventName][userId.String] = campaign.String
			}
		}
	}

	return usersAttribution, linkedEventUserCampaign, nil
}

//AddWebsiteVisitorsByEventName counts distinct users session by campaign id and adds it to attributionData
func AddWebsiteVisitorsByEventName(projectId uint64, attributionData map[string]*AttributionData, from, to int64, attributionKey string) error {
	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "range": fmt.Sprintf("%d - %d", from, to)})

	rows, err := db.Raw("select DISTINCT( CASE WHEN user_properties.properties->>? IS NULL THEN '$none' ELSE user_properties.properties->>? END) "+
		"as attributionKey, count(DISTINCT(sessions.user_id)) from events as sessions left join user_properties on sessions.user_properties_id = user_properties.id "+
		"where sessions.project_id=? and event_name_id = (select id from event_names where name=? "+
		"and project_id =? limit 1) AND sessions.timestamp BETWEEN ? AND ? GROUP BY attributionKey",
		attributionKey, attributionKey, projectId, U.EVENT_NAME_SESSION, projectId, from, to).Rows()

	defer rows.Close()
	if err != nil {
		logCtx.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return err
	}

	for rows.Next() {
		var campaign sql.NullString
		var count int64
		if err = rows.Scan(&campaign, &count); err != nil {
			logCtx.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			continue
		}
		if campaign.Valid {
			if _, exists := attributionData[campaign.String]; !exists { // campaign id not inserted in conversion event will be inserted here.
				attributionData[campaign.String] = &AttributionData{}
			}
			attributionData[campaign.String].WebsiteVisitors = count
		}
	}

	return nil
}

//AddPerformanceReportByCampaign adds channel data to attributionData based on campaign id. Campaign id with no matching channel data and left with empty name parameter
func AddPerformanceReportByCampaign(projectId uint64, attributionData map[string]*AttributionData, from, to int64, customerAccountId *string) (string, error) {
	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "range": fmt.Sprintf("%d - %d", from, to)})
	rows, err := db.Raw("select value->>'campaign_id' as campaign_id,  value->>'campaign_name' as campaign_name, "+
		"SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks, "+
		"SUM((value->>'cost')::float)/1000000 as total_cost from adwords_documents "+
		"where project_id = ? and customer_account_id = ? and type = ? and timestamp between ? and ? "+
		"group by value->>'campaign_id', campaign_name", projectId, customerAccountId, 5, getDateOnlyFromTimestamp(from), getDateOnlyFromTimestamp(to)).Rows()

	defer rows.Close()
	if err != nil {
		logCtx.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return "", err
	}

	for rows.Next() {
		var campaignName string
		var campaignId string
		var impressions int
		var clicks int
		var spend float64
		if err = rows.Scan(&campaignId, &campaignName, &impressions, &clicks, &spend); err != nil {
			logCtx.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			continue
		}

		if _, exists := attributionData[campaignId]; exists { // only matching campaign id is filled
			attributionData[campaignId].Name = campaignName
			attributionData[campaignId].Impressions = impressions
			attributionData[campaignId].Clicks = clicks
			attributionData[campaignId].Spend = spend
		}
	}

	currency, err := getAdwordsCurrency(projectId, customerAccountId, from, to)
	if err != nil {
		return "", err
	}
	return currency, nil
}

//getDateOnlyFromTimestamp returns date in YYYYMMDD format
func getDateOnlyFromTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("20060102")
}

//getAdwordsCurrency returns currency used for customer_account_id
func getAdwordsCurrency(projectId uint64, customerAccountId *string, from, to int64) (string, error) {
	stmnt := "SELECT value->>'currency_code' as currency FROM adwords_documents" +
		" " + "WHERE project_id=? AND customer_account_id=? AND type=? AND timestamp BETWEEN ? AND ?" +
		" " + "ORDER BY timestamp DESC LIMIT 1"

	logCtx := log.WithField("project_id", projectId)

	db := C.GetServices().Db
	rows, err := db.Raw(stmnt,
		projectId, customerAccountId, 9, getDateOnlyFromTimestamp(from), getDateOnlyFromTimestamp(to)).Rows()

	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return "", err
	}
	defer rows.Close()

	var currency string
	for rows.Next() {
		rows.Scan(&currency)
	}

	return currency, nil
}
