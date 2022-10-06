package postgres

import (
	"factors/model/model"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) fetchOTPSessions(projectID uint64, offlineTouchPointEventNameId string, query *model.AttributionQuery, logCtx log.Entry) (map[string]map[string]model.UserSessionData, []string, error) {

	effectiveFrom := model.LookbackAdjustedFrom(query.From, query.LookbackDays)
	effectiveTo := query.To
	// extend the campaign window for engagement based attribution
	if query.QueryType == model.AttributionQueryTypeEngagementBased {
		effectiveFrom = model.LookbackAdjustedFrom(query.From, query.LookbackDays)
		effectiveTo = model.LookbackAdjustedTo(query.To, query.LookbackDays, U.TimeZoneString(query.Timezone))
	}

	attributionEventKey, err := model.GetAttributionKeyForOffline(query.AttributionKey)
	if err != nil {
		return nil, nil, err
	}

	caseSelectStmt := "CASE WHEN sessions.properties->>? IS NULL THEN ? " +
		" WHEN sessions.properties->>? = '' THEN ? ELSE sessions.properties->>? END"

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
		U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL, model.PropertyValueNone, U.EP_CHANNEL,
		U.EP_TYPE, model.PropertyValueNone, U.EP_TYPE, model.PropertyValueNone, U.EP_TYPE,
		attributionEventKey, model.PropertyValueNone, attributionEventKey, model.PropertyValueNone, attributionEventKey,
		projectID, offlineTouchPointEventNameId, effectiveFrom, effectiveTo)

	rows, tx, err := pg.ExecQueryWithContext(queryUserOTPsessions, qParams)
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	return model.ProcessOTPEventRows(rows, query, logCtx)

}
