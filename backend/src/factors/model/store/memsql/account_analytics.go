package memsql

import (
	"factors/model/model"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// TODO Changes when adding unique function.
func (store *MemSQL) ExecuteAccountAnalyticsQuery(projectID int64, reqID string, query model.AccountAnalyticsQuery) (model.QueryResult, int) {
	rStmnt := ""
	rParams := make([]interface{}, 0)

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return model.QueryResult{}, http.StatusInternalServerError
	}

	filterStmnt, filterParams, err := buildWhereFromProperties(projectID, query.Filters, 0)
	if err != nil {
		log.WithField("projectID", projectID).WithField("reqID", reqID).WithField("query", query).Warn("Failed to build where condition on account analytics")
		return model.QueryResult{}, http.StatusInternalServerError
	}

	sql := "SELECT %s(%s) as %s FROM  users " +
		"WHERE project_id = ? AND JSON_EXTRACT_STRING(associated_segments, ?) IS NOT NULL " +
		"AND is_group_user = 1 AND source = ? AND group_%d_id IS NOT NULL"
	if filterStmnt != "" {
		sql = sql + " AND " + filterStmnt
	}
	rStmnt = fmt.Sprintf(sql, query.AggregateFunction, query.AggregateProperty, query.Metric, domainGroup.ID)
	rParams = append(rParams, projectID, query.SegmentID, model.UserSourceDomains)
	rParams = append(rParams, filterParams...)

	result, err, reqID := store.ExecQuery(rStmnt, rParams)
	reqResult := *result
	return reqResult, http.StatusOK
}
