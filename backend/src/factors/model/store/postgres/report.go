package postgres

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type ReportExplanation struct {
	Percentage   float64
	Effect       string
	CurValue     float64
	PrevValue    float64
	Diff         float64
	Type         string
	GroupName    string
	GroupValue   string
	TimestampStr string
}

var dashBoardUnitPresentationsToIncludeInReport = []string{
	model.PresentationLine,
	model.PresentationBar,
	model.PresentationTable,
	model.PresentationCard,
	model.PresentationFunnel,
}

const effectIncrease = "increase"
const effectDecrease = "decrease"
const effectEqual = "equal"
const explanationsLimit = 3

func (pg *Postgres) CreateReport(report *model.Report) (*model.Report, int) {
	dbReport, err := model.TranslateReportToDBReport(report)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	if err := db.Create(dbReport).Error; err != nil {
		log.WithError(err).Error("Failed to create dbReport")
		return nil, http.StatusInternalServerError
	}

	createdReport, err := model.TranslateDBReportToReport(dbReport)
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	return createdReport, http.StatusCreated
}

// DeleteReportByDashboardID Delete a report used in a particular dashboard.
func (pg *Postgres) DeleteReportByDashboardID(projectID, dashboardID uint64) int {
	db := C.GetServices().Db

	if err := db.Model(&model.DBReport{}).Where("dashboard_id = ? AND project_id = ?", dashboardID, projectID).
		Update(map[string]interface{}{"is_deleted": true}).Error; err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (pg *Postgres) GetReportByID(id uint64) (*model.Report, int) {
	if id == 0 {
		log.Error("GetReportByID Failed. ID not provided.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	dbReport := model.DBReport{}

	if err := db.Limit(1).Where("id = ? AND is_deleted = ?", id, false).Find(&dbReport).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("GetReportByID Failed.")
		return nil, http.StatusInternalServerError
	}

	report, err := model.TranslateDBReportToReport(&dbReport)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	if report.Invalid {
		return nil, http.StatusNotFound
	}

	return report, http.StatusFound
}

func (pg *Postgres) GetReportsByProjectID(projectID uint64) ([]*model.Report, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	dbReports := make([]*model.DBReport, 0, 0)

	db := C.GetServices().Db
	if err := db.Limit(10).Where("project_id = ? AND is_deleted = ?", projectID, false).
		Find(&dbReports).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(dbReports) == 0 {
		return nil, http.StatusNotFound
	}

	reports, err := model.TranslateDBReportsToReports(dbReports)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	return reports, http.StatusFound
}

func (pg *Postgres) GetValidReportsListAgentHasAccessTo(projectID uint64,
	agentUUID string) ([]*model.ReportDescription, int) {

	if projectID == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	dashboards, errCode := pg.GetDashboards(projectID, agentUUID)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	if len(dashboards) == 0 {
		return nil, http.StatusNotFound
	}

	dashboardIDs := make([]uint64, len(dashboards), len(dashboards))
	for i := 0; i < len(dashboards); i++ {
		dashboardIDs[i] = dashboards[i].ID
	}

	dbReportDecs := make([]*model.ReportDescription, 0, 0)

	db := C.GetServices().Db
	if err := db.Order("end_time DESC").Limit(100).Where("project_id = ? AND is_deleted = ?",
		projectID, false).Where("dashboard_id IN (?)", dashboardIDs).Where("invalid = ?",
		false).Find(&dbReportDecs).Error; err != nil {

		return nil, http.StatusInternalServerError
	}

	if len(dbReportDecs) == 0 {
		return nil, http.StatusNotFound
	}

	return dbReportDecs, http.StatusFound
}

func (pg *Postgres) GenerateReport(projectID, dashboardID uint64, dashboardName string, reportType string,
	intervalBeforeThat, interval model.Interval) (*model.Report, int) {

	logCtx := log.WithField("project_id", projectID).WithField("dashboard_id", dashboardID)

	dashboardUnits, errCode := pg.GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(
		projectID, dashboardID, dashBoardUnitPresentationsToIncludeInReport)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	reportUnits := make([]model.DashboardUnitReport, 0, 0)
	for _, dashboardUnit := range dashboardUnits {
		dashboardUnitReport, errCode := pg.getDashboardUnitReport(projectID,
			dashboardUnit, intervalBeforeThat, interval)

		if errCode != http.StatusOK {
			// Do not break the loop after one failure for a dashboard.
			logCtx.Error("Failed to generate report unit for dashboard.")
			continue
		}

		reportUnits = append(reportUnits, *dashboardUnitReport)
	}

	report := &model.Report{
		ProjectID:     projectID,
		DashboardID:   dashboardID,
		DashboardName: dashboardName,
		Type:          reportType,
		StartTime:     interval.StartTime,
		EndTime:       interval.EndTime,
		Units:         reportUnits,
	}

	if err := pg.addExplanationsAndOrderReportUnits(report); err != nil {
		return nil, http.StatusInternalServerError
	}

	return report, http.StatusOK
}

func overrideDateTimePropertyValue(property *model.QueryProperty, interval *model.Interval) error {
	dateTimeValue, err := model.DecodeDateTimePropertyValue(property.Value)
	if err != nil {
		return err
	}

	dateTimeValue.From = interval.StartTime
	dateTimeValue.To = interval.EndTime

	newDateTimeBytes, err := json.Marshal(dateTimeValue)
	if err != nil {
		return err
	}

	property.Value = string(newDateTimeBytes)
	return nil
}

func (pg *Postgres) getInsightsReportUnit(projectID uint64, query model.Query, interval model.Interval) (*model.ReportUnit, int) {
	query.From = interval.StartTime
	query.To = interval.EndTime

	for _, ewp := range query.EventsWithProperties {
		for i := range ewp.Properties {
			// override user join time with report unit interval.
			if ewp.Properties[i].Type == U.PropertyTypeDateTime &&
				ewp.Properties[i].Property == U.UP_JOIN_TIME {

				err := overrideDateTimePropertyValue(&ewp.Properties[i], &interval)
				if err != nil {
					log.WithField("property", ewp.Properties[i]).Error(
						"Failed overriding user join time by interval.")
					return nil, http.StatusInternalServerError
				}
			}
		}
	}

	queryResult, errCode, errMsg := pg.Analyze(projectID, query)
	if errCode != http.StatusOK {
		log.Errorf("Error creating ReportUnit, ErrMsg: %s", errMsg)
		return nil, http.StatusInternalServerError
	}

	reportUnit := model.ReportUnit{
		StartTime:   interval.StartTime,
		EndTime:     interval.EndTime,
		QueryResult: *queryResult,
	}

	return &reportUnit, http.StatusOK
}

func (pg *Postgres) getChannelReportUnit(projectID uint64, channelQueryUnit model.ChannelQueryUnit,
	presentation string, interval model.Interval) (*model.ReportUnit, int) {

	logCtx := log.WithField("project_id", projectID)
	channelQueryResult, errCode := pg.ExecuteChannelQuery(projectID, channelQueryUnit.Query)
	if errCode != http.StatusOK {
		logCtx.Error("Failed creating channel report unit")
		return nil, http.StatusInternalServerError
	}

	var queryResult model.QueryResult
	if channelQueryResult.Metrics != nil && presentation == model.PresentationCard {
		metricName, exists := (*channelQueryUnit.Meta)["metric"]
		if !exists {
			logCtx.Error("Metric name doesn't exist on dashboard unit for creating report unit.")
			return nil, http.StatusInternalServerError
		}

		metric := metricName.(string)
		queryResult.Headers = []string{"count"}

		row := make([]interface{}, 0, 0)
		value := (*channelQueryResult.Metrics)[metric]
		if value == nil {
			value = 0
		}
		row = append(row, value)
		queryResult.Rows = [][]interface{}{row}
	} else if channelQueryResult.MetricsBreakdown != nil {
		logCtx.Error("Metric breakdown not supported on getChannelReportUnit.")
		return nil, http.StatusInternalServerError
	}

	reportUnit := model.ReportUnit{
		StartTime:   interval.StartTime,
		EndTime:     interval.EndTime,
		QueryResult: queryResult,
	}

	return &reportUnit, http.StatusOK
}

func getPercentageChange(prevCount float64, curCount float64) (float64, string) {
	var percentChange float64

	if prevCount == 0 && curCount == 0 {
		percentChange = 0
	} else if prevCount == 0 && curCount > 0 {
		percentChange = curCount
	} else if curCount == 0 && prevCount > 0 {
		percentChange = prevCount * -1
	} else {
		percentChange = ((curCount - prevCount) / prevCount) * 100
	}

	var effect string
	if percentChange > 0 {
		effect = effectIncrease
	} else {
		effect = effectDecrease
	}

	if percentChange < 0 {
		percentChange = percentChange * -1
	}

	return percentChange, effect
}

func getPositiveDiff(prevCount, curCount float64) float64 {
	diff := curCount - prevCount

	if diff < 0 {
		diff = diff * -1
	}

	return diff
}

func explainTotalChange(percentage float64, effect, title, from, to, reportType string) string {
	if reportType == model.ReportTypeWeekly {
		return fmt.Sprintf("%0.0f%% %s in '%s' from week %s to week %s.", percentage,
			effect, title, from, to)
	}

	if reportType == model.ReportTypeMonthly {
		return fmt.Sprintf("%0.0f%% %s in '%s' from month %s to %s.", percentage,
			effect, title, from, to)
	}

	return ""
}

func explainChange(exp *ReportExplanation) string {
	expStr := fmt.Sprintf("- No.of %s", exp.Type)

	if exp.GroupName != "" && exp.GroupValue != "" {
		expStr = expStr + " " + fmt.Sprintf("with %s as %s", exp.GroupName, exp.GroupValue)
	}

	expStr = expStr + " " + fmt.Sprintf("%sd from %0.0f to %0.0f (%0.0f%%)",
		exp.Effect, exp.PrevValue, exp.CurValue, exp.Percentage)

	if exp.TimestampStr != "" {
		expStr = expStr + " " + fmt.Sprintf("%s", exp.TimestampStr)
	}

	return expStr + "."
}

func addExplanationsForPresentationCard(duReport *model.DashboardUnitReport, reportType string) error {
	prevCount, _ := getAggrAsFloat64(duReport.Results[0].QueryResult.Rows[0][0])
	curCount, _ := getAggrAsFloat64(duReport.Results[1].QueryResult.Rows[0][0])

	fromPeriod := getReadableIntervalByType(duReport.Results[0].StartTime,
		duReport.Results[0].EndTime, reportType)
	toPeriod := getReadableIntervalByType(duReport.Results[1].StartTime,
		duReport.Results[1].EndTime, reportType)

	percentChange, effect := getPercentageChange(prevCount, curCount)
	duReport.Explanations = []string{explainTotalChange(percentChange, effect,
		duReport.Title, fromPeriod, toPeriod, reportType)}
	duReport.ChangeInPercentage = percentChange
	return nil
}

func getAggrByGroup(queryResult *model.QueryResult,
	uniqueGroupsSet *map[string]bool) (map[string]float64, float64, string, error) {
	var totalCount float64
	var groupHeader string
	aggrByGroupMap := make(map[string]float64)
	aggrIndex, _, _ := GetTimstampAndAggregateIndexOnQueryResult(queryResult.Headers)
	if aggrIndex == -1 {
		return aggrByGroupMap, totalCount, groupHeader, fmt.Errorf("invalid index on GetTimstampAndAggregateIndexOnQueryResult for queryResult %+v", queryResult)
	}

	// len should be aggr + 1, if group exist.
	hasGroup := len(queryResult.Headers) > 1
	var groupIndex int
	if hasGroup && aggrIndex == 0 {
		groupIndex = 1
	}

	for _, row := range queryResult.Rows {
		var group string
		if hasGroup {
			group = row[groupIndex].(string)
		}

		aggr, _ := getAggrAsFloat64(row[aggrIndex])
		aggrByGroupMap[group] = aggr
		totalCount = totalCount + aggr
		(*uniqueGroupsSet)[group] = true
	}

	if hasGroup {
		groupHeader = queryResult.Headers[groupIndex]
	}

	return aggrByGroupMap, totalCount, groupHeader, nil
}

func sortAndLimitExplanations(explanations []ReportExplanation) []ReportExplanation {
	sort.SliceStable(explanations, func(i, j int) bool {
		return explanations[i].Diff > explanations[j].Diff
	})

	if len(explanations) < explanationsLimit {
		return explanations
	}

	return explanations[:explanationsLimit]
}

func getEntityFromQueryType(queryType string) string {
	if queryType == model.QueryTypeEventsOccurrence {
		return "occurrences"
	}

	if queryType == model.QueryTypeUniqueUsers {
		return "users"
	}

	return ""
}

func addExplanationsForPresentationBar(duReport *model.DashboardUnitReport, reportType string) error {
	prevResult := duReport.Results[0].QueryResult
	curResult := duReport.Results[1].QueryResult
	resultEntity := getEntityFromQueryType(duReport.Results[1].QueryResult.Meta.Query.Type)

	fromPeriod := getReadableIntervalByType(duReport.Results[0].StartTime,
		duReport.Results[0].EndTime, reportType)
	toPeriod := getReadableIntervalByType(duReport.Results[1].StartTime,
		duReport.Results[1].EndTime, reportType)

	uniqueGroupsSet := make(map[string]bool)
	prevAggrByGroup, prevResultTotal, prevResultGroupName, err := getAggrByGroup(&prevResult, &uniqueGroupsSet)
	if err != nil {
		log.WithError(err).Error("Failed to getAggrByGroup")
		return err
	}
	curAggrByGroup, curResultTotal, curResultGroupName, err := getAggrByGroup(&curResult, &uniqueGroupsSet)
	if err != nil {
		log.WithError(err).Error("Failed to getAggrByGroup")
		return err
	}

	if prevResultGroupName != curResultGroupName {
		log.WithFields(log.Fields{"prev_group_name": prevResultGroupName,
			"cur_group_name": curResultGroupName}).Error("Group name on reports query results are not mathcing.")
	}

	explanations := make([]string, 0, 0)
	percentChange, totalEffect := getPercentageChange(prevResultTotal, curResultTotal)
	duReport.ChangeInPercentage = percentChange

	explanations = append(explanations, explainTotalChange(percentChange, totalEffect,
		duReport.Title, fromPeriod, toPeriod, reportType))

	secExplanations := make([]ReportExplanation, 0, 0)
	for group := range uniqueGroupsSet {
		var prevAggr, curAggr float64

		if _, exists := prevAggrByGroup[group]; exists {
			prevAggr = prevAggrByGroup[group]
		}

		if _, exists := curAggrByGroup[group]; exists {
			curAggr = curAggrByGroup[group]
		}

		percentChange, effect := getPercentageChange(prevAggr, curAggr)
		if percentChange >= 5.0 && effect == totalEffect {
			secExplanations = append(secExplanations,
				ReportExplanation{Type: resultEntity, Percentage: percentChange, Effect: effect,
					GroupName: curResultGroupName, GroupValue: group, Diff: getPositiveDiff(prevAggr, curAggr),
					CurValue: curAggr, PrevValue: prevAggr})
		}
	}

	secExplanations = sortAndLimitExplanations(secExplanations)
	for _, explanation := range secExplanations {
		explanations = append(explanations, explainChange(&explanation))
	}

	duReport.Explanations = explanations
	return nil
}

func addExplanationsForPresentationTable(duReport *model.DashboardUnitReport, reportType string) error {
	prevResult := duReport.Results[0].QueryResult
	curResult := duReport.Results[1].QueryResult
	resultEntity := getEntityFromQueryType(duReport.Results[1].QueryResult.Meta.Query.Type)

	fromPeriod := getReadableIntervalByType(duReport.Results[0].StartTime,
		duReport.Results[0].EndTime, reportType)
	toPeriod := getReadableIntervalByType(duReport.Results[1].StartTime,
		duReport.Results[1].EndTime, reportType)

	uniqueGroupsSet := make(map[string]bool)
	prevAggrByGroup, prevResultTotal, prevResultGroupName, err := getAggrByGroup(&prevResult, &uniqueGroupsSet)
	if err != nil {
		log.WithError(err).Error("Failed to getAggrByGroup")
		return err
	}
	curAggrByGroup, curResultTotal, curResultGroupName, err := getAggrByGroup(&curResult, &uniqueGroupsSet)
	if err != nil {
		log.WithError(err).Error("Failed to getAggrByGroup")
		return err
	}

	if prevResultGroupName != curResultGroupName {
		log.WithFields(log.Fields{"prev_group_name": prevResultGroupName,
			"cur_group_name": curResultGroupName}).Error("Group name on reports query results are not mathcing.")
	}

	explanations := make([]string, 0, 0)
	percentChange, totalEffect := getPercentageChange(prevResultTotal, curResultTotal)
	duReport.ChangeInPercentage = percentChange

	explanations = append(explanations, explainTotalChange(percentChange, totalEffect,
		duReport.Title, fromPeriod, toPeriod, reportType))

	secExplanations := make([]ReportExplanation, 0, 0)
	for group := range uniqueGroupsSet {
		var prevAggr, curAggr float64

		if _, exists := prevAggrByGroup[group]; exists {
			prevAggr = prevAggrByGroup[group]
		}

		if _, exists := curAggrByGroup[group]; exists {
			curAggr = curAggrByGroup[group]
		}

		percentChange, effect := getPercentageChange(prevAggr, curAggr)
		if percentChange >= 5.0 && effect == totalEffect {
			secExplanations = append(secExplanations,
				ReportExplanation{Type: resultEntity, Percentage: percentChange, Effect: effect,
					GroupName: curResultGroupName, GroupValue: group, Diff: getPositiveDiff(prevAggr, curAggr),
					CurValue: curAggr, PrevValue: prevAggr})
		}
	}

	secExplanations = sortAndLimitExplanations(secExplanations)
	for _, explanation := range secExplanations {
		explanations = append(explanations, explainChange(&explanation))
	}

	duReport.Explanations = explanations
	return nil
}

func getAggrAsFloat64(aggr interface{}) (float64, error) {
	switch aggr.(type) {
	case int:
		return float64(aggr.(int)), nil
	case int64:
		return float64(aggr.(int64)), nil
	case float32:
		return float64(aggr.(float32)), nil
	case float64:
		return aggr.(float64), nil
	case string:
		aggrInt, err := strconv.ParseInt(aggr.(string), 10, 64)
		return float64(aggrInt), err
	default:
		return float64(0), errors.New("invalid aggregate value type")
	}
}

func getTimestampAsString(timestampInt interface{}) (string, error) {
	switch timestampInt.(type) {
	case time.Time:
		return (timestampInt.(time.Time)).Format(time.RFC3339), nil
	case string:
		return timestampInt.(string), nil
	default:
		return "", errors.New("invalid timestamp value type")
	}
}

func getAggrByTimestampAndGroup(queryResult *model.QueryResult,
	uniqueGroupsSet *map[string]string) (map[string]map[string]float64, []string, float64) {

	var totalAggr float64
	aggrIndex, timestampIndex, _ := GetTimstampAndAggregateIndexOnQueryResult(queryResult.Headers)
	aggrByTimestampAndGroup := make(map[string]map[string]float64, 0)
	timestamps := make([]string, 0, 0)

	for _, row := range queryResult.Rows {
		timestamp, _ := getTimestampAsString(row[timestampIndex])
		aggr, _ := getAggrAsFloat64(row[aggrIndex])

		if _, tsExists := aggrByTimestampAndGroup[timestamp]; !tsExists {
			aggrByTimestampAndGroup[timestamp] = make(map[string]float64, 0)
			// list of ordered timestamps.
			timestamps = append(timestamps, timestamp)
		}

		var groupValue string
		var displayGroupValue string
		for i, col := range row {
			colValue := fmt.Sprintf("%s", col)

			if i != aggrIndex && i != timestampIndex {
				encGroupValueKey := fmt.Sprintf("c%d:%s", i, colValue)
				if groupValue == "" {
					groupValue = encGroupValueKey
				} else {
					groupValue = groupValue + "_" + encGroupValueKey
				}

				if displayGroupValue == "" {
					displayGroupValue = colValue
				} else {
					displayGroupValue = displayGroupValue + " / " + colValue
				}
			}
		}

		(*uniqueGroupsSet)[groupValue] = displayGroupValue
		aggrByTimestampAndGroup[timestamp][groupValue] = aggr
		totalAggr = totalAggr + aggr
	}

	return aggrByTimestampAndGroup, timestamps, totalAggr
}

func getDayOfTimestamp(timestampStr string) string {
	timestamp, _ := time.Parse(time.RFC3339, timestampStr)
	return timestamp.Weekday().String()
}

func getReadableTimestamp(timestampStr string) string {
	timestamp, _ := time.Parse(time.RFC3339, timestampStr)
	return timestamp.Format("Jan 02")
}

// getTotalAggrForUniqueUsersQuery - Runs same query without group by timestamp.
func (pg *Postgres) getTotalAggrForUniqueUsersQuery(projectId uint64, uuQuery model.Query) float64 {
	uuQuery.GroupByTimestamp = ""

	queryResult, _, _ := pg.Analyze(projectId, uuQuery)

	aggrIndex, _, _ := GetTimstampAndAggregateIndexOnQueryResult(queryResult.Headers)
	var total float64
	for _, row := range queryResult.Rows {
		total = total + float64(row[aggrIndex].(int64))
	}

	return total
}

func getGroupNameForPresentationLine(query model.Query) string {
	var groupName string
	for i, group := range query.GroupByProperties {
		if i > 0 {
			groupName = groupName + " / "
		}

		groupName = groupName + group.Property
	}

	return groupName
}

func (pg *Postgres) addExplanationsForPresentationLine(duReport *model.DashboardUnitReport, reportType string) error {
	prevResult := duReport.Results[0].QueryResult
	curResult := duReport.Results[1].QueryResult
	resultEntity := getEntityFromQueryType(duReport.Results[1].QueryResult.Meta.Query.Type)
	groupName := getGroupNameForPresentationLine(duReport.Results[1].QueryResult.Meta.Query)

	fromPeriod := getReadableIntervalByType(duReport.Results[0].StartTime,
		duReport.Results[0].EndTime, reportType)
	toPeriod := getReadableIntervalByType(duReport.Results[1].StartTime,
		duReport.Results[1].EndTime, reportType)

	uniqueGroupsSet := make(map[string]string)
	prevAggrMap, prevTimestamps, prevTotal := getAggrByTimestampAndGroup(&prevResult, &uniqueGroupsSet)
	curAggrMap, curTimestamps, curTotal := getAggrByTimestampAndGroup(&curResult, &uniqueGroupsSet)

	if curResult.Meta.Query.Type == model.QueryTypeUniqueUsers {
		prevTotal = pg.getTotalAggrForUniqueUsersQuery(duReport.ProjectID, prevResult.Meta.Query)
		curTotal = pg.getTotalAggrForUniqueUsersQuery(duReport.ProjectID, curResult.Meta.Query)
	}

	explanations := make([]string, 0, 0)
	percentChange, totalEffect := getPercentageChange(prevTotal, curTotal)
	duReport.ChangeInPercentage = percentChange

	explanations = append(explanations, explainTotalChange(percentChange, totalEffect,
		duReport.Title, fromPeriod, toPeriod, reportType))

	secExplanations := make([]ReportExplanation, 0, 0)

	var comparisonLen int
	if len(prevTimestamps) < len(curTimestamps) {
		comparisonLen = len(prevTimestamps)
	} else {
		comparisonLen = len(curTimestamps)
	}

	// compare based on array position.
	for i := 0; i < comparisonLen; i++ {
		curTimestamp := curTimestamps[i]
		prevTimestamp := prevTimestamps[i]

		for group, displayGroup := range uniqueGroupsSet {
			var prevAggr, curAggr float64

			if _, exists := prevAggrMap[prevTimestamp][group]; exists {
				prevAggr = prevAggrMap[prevTimestamp][group]
			}

			if _, exists := curAggrMap[curTimestamp][group]; exists {
				curAggr = curAggrMap[curTimestamp][group]
			}

			// atleast one should have an aggr greater than 0.
			if prevAggr == 0 && curAggr == 0 {
				continue
			}

			percentChange, effect := getPercentageChange(prevAggr, curAggr)
			if percentChange >= 5.0 && effect == totalEffect {
				timestampStr := fmt.Sprintf("between %s and %s", getReadableTimestamp(prevTimestamp),
					getReadableTimestamp(curTimestamp))
				if reportType == model.ReportTypeWeekly {
					timestampStr = fmt.Sprintf("on %s (%s)", getDayOfTimestamp(curTimestamp), timestampStr)
				}

				secExplanations = append(secExplanations,
					ReportExplanation{Type: resultEntity, Percentage: percentChange, Effect: effect,
						Diff: getPositiveDiff(prevAggr, curAggr), CurValue: curAggr, PrevValue: prevAggr, GroupName: groupName,
						GroupValue: displayGroup, TimestampStr: timestampStr})
			}
		}
	}

	secExplanations = sortAndLimitExplanations(secExplanations)
	for _, explanation := range secExplanations {
		explanations = append(explanations, explainChange(&explanation))
	}

	duReport.Explanations = explanations
	return nil
}

func getFunnelConversionsFromResult(queryResult *model.QueryResult) ([]float64, float64) {
	conversionIndexes := make([]int, 0, 0)
	var overallIndex int
	for i, col := range queryResult.Headers {
		if strings.HasPrefix(col, model.FunnelConversionPrefix) {
			if strings.HasSuffix(col, "overall") {
				overallIndex = i
			} else {
				conversionIndexes = append(conversionIndexes, i)
			}
		}
	}

	conversions := make([]float64, 0, 0)
	for _, i := range conversionIndexes {
		conversion, _ := getAggrAsFloat64(queryResult.Rows[0][i])
		conversions = append(conversions, conversion)
	}
	total, _ := getAggrAsFloat64(queryResult.Rows[0][overallIndex])

	return conversions, total
}

func getEffect(prev float64, curr float64) string {
	diffTotal := curr - prev
	if diffTotal > 0 {
		return effectIncrease
	} else if diffTotal < 0 {
		return effectDecrease
	}

	return effectEqual
}

func addExplanationsForPresentationFunnel(duReport *model.DashboardUnitReport, reportType string) error {
	prevResult := duReport.Results[0].QueryResult
	curResult := duReport.Results[1].QueryResult

	prevConversions, prevTotal := getFunnelConversionsFromResult(&prevResult)
	curConversions, curTotal := getFunnelConversionsFromResult(&curResult)

	percentageChange, _ := getPercentageChange(prevTotal, curTotal)
	duReport.ChangeInPercentage = percentageChange

	totalEffect := getEffect(prevTotal, curTotal)
	if totalEffect == effectEqual {
		return nil
	}

	explanations := make([]string, 0, 0)
	explanations = append(explanations,
		fmt.Sprintf("Total conversion %sd from %0.0f%% to %0.0f%%.", totalEffect, prevTotal, curTotal))

	// one conversion is equal to total conversion.
	if len(curConversions) == 1 || len(prevConversions) != len(curConversions) {
		duReport.Explanations = explanations
		return nil
	}

	steps := curResult.Meta.Query.EventsWithProperties
	for i := range curConversions {
		convEffect := getEffect(prevConversions[i], curConversions[i])
		if convEffect != totalEffect {
			continue
		}

		explanations = append(explanations,
			fmt.Sprintf("- '%s' to '%s' conversion %sd from %0.0f%% to %0.0f%%.",
				steps[i].Name, steps[i+1].Name, convEffect, prevConversions[i], curConversions[i]))
	}

	duReport.Explanations = explanations
	return nil
}

func (pg *Postgres) addExplanationsByPresentation(duReport model.DashboardUnitReport, reportType string) (model.DashboardUnitReport, error) {
	if duReport.Presentation == "" || len(duReport.Results) < 2 {
		return duReport, nil
	}

	var err error
	switch duReport.Presentation {
	case model.PresentationCard:
		err = addExplanationsForPresentationCard(&duReport, reportType)
	case model.PresentationBar:
		err = addExplanationsForPresentationBar(&duReport, reportType)
	case model.PresentationTable:
		err = addExplanationsForPresentationTable(&duReport, reportType)
	case model.PresentationLine:
		err = pg.addExplanationsForPresentationLine(&duReport, reportType)
	case model.PresentationFunnel:
		err = addExplanationsForPresentationFunnel(&duReport, reportType)
	}

	return duReport, err
}

func (pg *Postgres) addExplanationsAndOrderReportUnits(report *model.Report) (err error) {
	dashboardUnitReports := make([]model.DashboardUnitReport, 0, 0)
	for _, dashboardUnitReport := range report.Units {

		// Temp fix to skip add explanations for Channel query units.
		// Todo: Fix QueryResult on ReportUnit which only supports analytics queries,
		// empty struct being added for channel queries.
		queryClass := dashboardUnitReport.Results[0].QueryResult.Meta.Query.Class
		if queryClass != model.QueryClassFunnel && queryClass != model.QueryClassInsights {
			continue
		}

		report, err := pg.addExplanationsByPresentation(dashboardUnitReport, report.Type)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to addExplanationsByPresentation for project_id: %d", report.ProjectID))
			return err
		}
		dashboardUnitReports = append(dashboardUnitReports, report)
	}

	// orders units by percentage change.
	sort.SliceStable(dashboardUnitReports, func(i, j int) bool {
		return dashboardUnitReports[i].ChangeInPercentage > dashboardUnitReports[j].ChangeInPercentage
	})

	report.Units = dashboardUnitReports
	return nil
}

func (pg *Postgres) getDashboardUnitReport(projectID uint64, dashboardUnit model.DashboardUnit, intervalBeforeThat,
	interval model.Interval) (*model.DashboardUnitReport, int) {

	query := model.Query{}
	err := json.Unmarshal(dashboardUnit.Query.RawMessage, &query)
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	var dashboardUnitReport *model.DashboardUnitReport
	if query.Class == model.QueryClassChannel {
		channelQueryUnit := model.ChannelQueryUnit{}
		err := json.Unmarshal(dashboardUnit.Query.RawMessage, &channelQueryUnit)
		if err != nil {
			return nil, http.StatusInternalServerError
		}

		intervalBeforeReportUnit, errCode := pg.getChannelReportUnit(projectID,
			channelQueryUnit, dashboardUnit.Presentation, intervalBeforeThat)
		if errCode != http.StatusOK {
			return nil, errCode
		}

		intervalReportUnit, errCode := pg.getChannelReportUnit(projectID,
			channelQueryUnit, dashboardUnit.Presentation, interval)
		if errCode != http.StatusOK {
			return nil, errCode
		}

		dashboardUnitReport = &model.DashboardUnitReport{
			ProjectID:    projectID,
			Title:        dashboardUnit.Title,
			Presentation: dashboardUnit.Presentation,
			Results:      []model.ReportUnit{*intervalBeforeReportUnit, *intervalReportUnit},
		}
	} else {
		intervalBeforeReportUnit, errCode := pg.getInsightsReportUnit(projectID, query, intervalBeforeThat)
		if errCode != http.StatusOK {
			return nil, errCode
		}

		intervalReportUnit, errCode := pg.getInsightsReportUnit(projectID, query, interval)
		if errCode != http.StatusOK {
			return nil, errCode
		}

		dashboardUnitReport = &model.DashboardUnitReport{
			ProjectID:    projectID,
			Title:        dashboardUnit.Title,
			Presentation: dashboardUnit.Presentation,
			Results:      []model.ReportUnit{*intervalBeforeReportUnit, *intervalReportUnit},
		}
	}

	return dashboardUnitReport, http.StatusOK
}

func unixToReadableDate(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("Jan 02")
}

func getReadableIntervalByType(from, to int64, typ string) string {
	if typ == model.ReportTypeMonthly {
		return time.Unix(from, 0).UTC().Format("January")
	}

	return unixToReadableDate(from) + "-" + unixToReadableDate(to)
}
