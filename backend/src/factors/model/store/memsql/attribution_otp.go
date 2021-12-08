package memsql

import (
	"factors/model/model"
	U "factors/util"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) fetchOTPSessions(projectID uint64, offlineTouchPointEventNameId string, query *model.AttributionQuery) (map[string]map[string]model.UserSessionData, []string, error) {

	logCtx := log.WithFields(log.Fields{"ProjectId": projectID})
	effectiveFrom := model.LookbackAdjustedFrom(query.From, query.LookbackDays)
	effectiveTo := query.To
	// extend the campaign window for engagement based attribution
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		effectiveFrom = model.LookbackAdjustedFrom(query.From, query.LookbackDays)
		effectiveTo = model.LookbackAdjustedTo(query.To, query.LookbackDays)
	}

	attributionEventKey, err := model.GetAttributionKeyForOffline(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}

	caseSelectStmt := "CASE WHEN JSON_EXTRACT_STRING(sessions.properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(sessions.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(sessions.properties, ?) END"

	queryUserOTPsessions := "SELECT sessions.user_id, " +
		caseSelectStmt + " AS campaignID, " +
		caseSelectStmt + " AS campaignName, " +
		caseSelectStmt + " AS source, " +
		caseSelectStmt + " AS channel, " +
		caseSelectStmt + " AS type, " +
		caseSelectStmt + " AS attribution_id, " +
		" sessions.timestamp FROM events AS sessions " +
		" WHERE sessions.project_id=? AND sessions.event_name_id=? AND sessions.timestamp BETWEEN ? AND ?"
	var qParams []interface{}

	qParams = append(qParams,
		U.EP_CAMPAIGN_ID, model.PropertyValueNone, U.EP_CAMPAIGN_ID, model.PropertyValueNone, U.EP_CAMPAIGN_ID,
		U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN, model.PropertyValueNone, U.EP_CAMPAIGN,
		U.EP_SOURCE, model.PropertyValueNone, U.EP_SOURCE, model.PropertyValueNone, U.EP_SOURCE,
		U.EP_TYPE, model.PropertyValueNone, U.EP_TYPE, model.PropertyValueNone, U.EP_TYPE,
		U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID, model.PropertyValueNone, U.EP_GCLID,
		attributionEventKey, model.PropertyValueNone, attributionEventKey, model.PropertyValueNone, attributionEventKey,
		projectID, offlineTouchPointEventNameId, effectiveFrom, effectiveTo)

	rows, tx, err := store.ExecQueryWithContext(queryUserOTPsessions, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	return model.ProcessOTPEventRows(rows, query, logCtx)

}
