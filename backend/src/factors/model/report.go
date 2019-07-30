package model

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	ReportTypeWeekly = "w"
)

var dashBoardUnitTypesToIncludeInReport = []string{PresentationLine, PresentationBar, PresentationCard}

// DBReport represents structure to be used for storing report in database
type DBReport struct {
	ID            uint64 `gorm:"primary_key:true;"`
	ProjectID     uint64
	DashboardID   uint64
	DashboardName string
	CreatedAt     time.Time
	Type          string
	StartTime     int64
	EndTime       int64
	Invalid       bool

	Contents postgres.Jsonb
}

// TableName returns tablename to be used by GORM
func (DBReport) TableName() string {
	return "reports"
}

type Report struct {
	ID            uint64    `json:"id"`
	ProjectID     uint64    `json:"project_id"`
	DashboardID   uint64    `json:"dashboard_id"`
	DashboardName string    `json:"dashboard_name"`
	CreatedAt     time.Time `json:"created_at"`
	Type          string    `json:"type"`
	StartTime     int64     `json:"start_time"`
	EndTime       int64     `json:"end_time"`
	Invalid       bool      `json:"invalid"`

	Contents ReportContent `json:"contents"`
}

type ReportContent struct {
	DashboardUnitIDToDashboardUnitReport map[uint64]DashboardUnitReport `json:"dashboardunitid_to_dashboardunitreport"`
}

type DashboardUnitReport struct {
	ProjectID    uint64       `json:"pid"`
	Title        string       `json:"t"`
	Presentation string       `json:"p"`
	Results      []ReportUnit `json:"r"`
	Explanations []string     `json:"e"`
}

type ReportUnit struct {
	StartTime   int64       `json:"st"`
	EndTime     int64       `json:"et"`
	QueryResult QueryResult `json:"qr"`
}

type Interval struct {
	StartTime int64
	EndTime   int64
}

type ReportExplanation struct {
	Percentage  float64
	Effect      string
	Attribution string
}

const effectIncrease = "increase"
const effectDecrease = "decrease"
const explanationsLimit = 5
const primaryExplanationPrefix = "Total"
const secondaryExplanationPrefix = "-"

func TranslateDBReportToReport(dbReport *DBReport) (*Report, error) {

	contents := ReportContent{}

	err := json.Unmarshal(dbReport.Contents.RawMessage, &contents)
	if err != nil {
		return nil, err
	}

	report := Report{
		ID:            dbReport.ID,
		ProjectID:     dbReport.ProjectID,
		DashboardID:   dbReport.DashboardID,
		DashboardName: dbReport.DashboardName,
		CreatedAt:     dbReport.CreatedAt,
		Type:          dbReport.Type,
		StartTime:     dbReport.StartTime,
		EndTime:       dbReport.EndTime,
		Invalid:       dbReport.Invalid,
		Contents:      contents,
	}

	return &report, nil
}

func TranslateReportToDBReport(report *Report) (*DBReport, error) {

	contentJSONBytes, err := json.Marshal(report.Contents)
	if err != nil {
		return nil, err
	}

	postgresJson := postgres.Jsonb{RawMessage: contentJSONBytes}

	dbReport := DBReport{
		ID:            report.ID,
		ProjectID:     report.ProjectID,
		DashboardID:   report.DashboardID,
		DashboardName: report.DashboardName,
		CreatedAt:     report.CreatedAt,
		Type:          report.Type,
		StartTime:     report.StartTime,
		EndTime:       report.EndTime,
		Invalid:       report.Invalid,
		Contents:      postgresJson,
	}

	return &dbReport, nil
}

func TranslateDBReportsToReports(dbReports []*DBReport) ([]*Report, error) {
	reports := make([]*Report, len(dbReports), len(dbReports))
	for i, dbReport := range dbReports {
		report, err := TranslateDBReportToReport(dbReport)
		if err != nil {
			return nil, err
		}
		reports[i] = report
	}
	return reports, nil
}

func CreateReport(report *Report) (*Report, int) {

	dbReport, err := TranslateReportToDBReport(report)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	if err := db.Create(dbReport).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	createdReport, err := TranslateDBReportToReport(dbReport)
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	return createdReport, http.StatusCreated
}

func GetReportByID(id uint64) (*Report, int) {
	if id == 0 {
		log.Error("GetReportByID Failed. ID not provided.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	dbReport := DBReport{}

	if err := db.Limit(1).Where("id = ?", id).Find(&dbReport).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("GetReportByID Failed.")
		return nil, http.StatusInternalServerError
	}

	report, err := TranslateDBReportToReport(&dbReport)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	if report.Invalid {
		return nil, http.StatusNotFound
	}

	return report, http.StatusFound
}

func GetReportsByProjectID(projectID uint64) ([]*Report, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	dbReports := make([]*DBReport, 0, 0)

	if err := db.Limit(10).Where("project_id = ?", projectID).Find(&dbReports).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(dbReports) == 0 {
		return nil, http.StatusNotFound
	}

	reports, err := TranslateDBReportsToReports(dbReports)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	return reports, http.StatusFound
}

type ReportDescription struct {
	ID            uint64    `json:"id"`
	ProjectID     uint64    `json:"project_id"`
	DashboardID   uint64    `json:"dashboard_id"`
	DashboardName string    `json:"dashboard_name"`
	CreatedAt     time.Time `json:"created_at"`
	Type          string    `json:"type"`
	StartTime     int64     `json:"start_time"`
	EndTime       int64     `json:"end_time"`
	Invalid       bool      `json:"invalid"`
}

// TableName returns tablename to be used by GORM
func (ReportDescription) TableName() string {
	return "reports"
}

func GetValidReportsListAgentHasAccessTo(projectID uint64, agentUUID string) ([]*ReportDescription, int) {
	if projectID == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	dashboards, errCode := GetDashboards(projectID, agentUUID)
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

	dbReportDecs := make([]*ReportDescription, 0, 0)

	db := C.GetServices().Db
	if err := db.Limit(100).Where("project_id = ?", projectID).Where("dashboard_id IN (?)",
		dashboardIDs).Where("invalid = ?", false).Find(&dbReportDecs).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(dbReportDecs) == 0 {
		return nil, http.StatusNotFound
	}

	return dbReportDecs, http.StatusFound
}

func GenerateReport(projectID, dashboardID uint64, dashboardName string,
	intervalBeforeThat, interval Interval) (*Report, int) {

	dashboardUnits, errCode := GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID,
		dashboardID, dashBoardUnitTypesToIncludeInReport)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	reportContents := ReportContent{
		DashboardUnitIDToDashboardUnitReport: make(map[uint64]DashboardUnitReport),
	}

	for _, dashboardUnit := range dashboardUnits {
		dashboardUnitReport, errCode := getDashboardUnitReport(projectID, dashboardUnit,
			intervalBeforeThat, interval)
		if errCode != http.StatusOK {
			return nil, errCode
		}
		reportContents.DashboardUnitIDToDashboardUnitReport[dashboardUnit.ID] = *dashboardUnitReport
	}

	report := &Report{
		ProjectID:     projectID,
		DashboardID:   dashboardID,
		DashboardName: dashboardName,
		Type:          ReportTypeWeekly,
		StartTime:     interval.StartTime,
		EndTime:       interval.EndTime,
		Contents:      reportContents,
	}

	addExplanationsToReportUnits(report)

	return report, http.StatusOK
}

func getReportUnit(projectID uint64, query Query, interval Interval) (*ReportUnit, int) {

	query.From = interval.StartTime
	query.To = interval.EndTime

	queryResult, errCode, errMsg := Analyze(projectID, query)
	if errCode != http.StatusOK {
		log.Errorf("Error creating ReportUnit, ErrMsg: %s", errMsg)
		return nil, http.StatusInternalServerError
	}
	reportUnit := ReportUnit{
		StartTime:   interval.StartTime,
		EndTime:     interval.EndTime,
		QueryResult: *queryResult,
	}

	return &reportUnit, http.StatusOK
}

func getPercentageChange(prevCount float64, curCount float64) (float64, string) {
	var percentChange float64
	if prevCount > 0 {
		percentChange = ((curCount - prevCount) / prevCount) * 100
	} else {
		percentChange = curCount
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

func explainPercentageChange(percentChange float64, effect string,
	attribution string, prefix string) string {

	return fmt.Sprintf("%s %0.0f%% %s in %s.", prefix, percentChange, effect, attribution)
}

func addExplanationsForPresentationCard(duReport *DashboardUnitReport) {
	prevCount, _ := getAggrAsFloat64(duReport.Results[0].QueryResult.Rows[0][0])
	curCount, _ := getAggrAsFloat64(duReport.Results[1].QueryResult.Rows[0][0])

	percentChange, effect := getPercentageChange(prevCount, curCount)
	duReport.Explanations = []string{explainPercentageChange(percentChange, effect, duReport.Title, primaryExplanationPrefix)}
}

func getAggrByGroup(queryResult *QueryResult,
	uniqueGroupsSet *map[string]bool) (map[string]float64, float64, string) {
	aggrIndex, _, _ := GetTimstampAndAggregateIndexOnQueryResult(queryResult.Headers)
	aggrByGroupMap := make(map[string]float64)

	var groupIndex int
	if aggrIndex == 0 {
		groupIndex = 1
	}

	var totalCount float64
	for _, row := range queryResult.Rows {
		group := row[groupIndex].(string)
		aggr, _ := getAggrAsFloat64(row[aggrIndex])
		aggrByGroupMap[group] = aggr
		totalCount = totalCount + aggr
		(*uniqueGroupsSet)[group] = true
	}

	return aggrByGroupMap, totalCount, queryResult.Headers[groupIndex]
}

func sortAndLimitExplanations(explanations []ReportExplanation) []ReportExplanation {
	sort.SliceStable(explanations, func(i, j int) bool {
		return explanations[i].Percentage > explanations[j].Percentage
	})

	if len(explanations) < explanationsLimit {
		return explanations
	}

	return explanations[:explanationsLimit]
}

func addExplanationsForPresentationBar(duReport *DashboardUnitReport) {
	prevResult := duReport.Results[0].QueryResult
	curResult := duReport.Results[1].QueryResult

	uniqueGroupsSet := make(map[string]bool)
	prevAggrByGroup, prevResultTotal, prevResultGroupName := getAggrByGroup(&prevResult, &uniqueGroupsSet)
	curAggrByGroup, curResultTotal, curResultGroupName := getAggrByGroup(&curResult, &uniqueGroupsSet)

	if prevResultGroupName != curResultGroupName {
		log.WithFields(log.Fields{"prev_group_name": prevResultGroupName,
			"cur_group_name": curResultGroupName}).Error("Group name on reports query results are not mathcing.")
	}

	explanations := make([]string, 0, 0)
	percentChange, totalEffect := getPercentageChange(prevResultTotal, curResultTotal)
	explanations = append(explanations, explainPercentageChange(percentChange, totalEffect,
		duReport.Title, primaryExplanationPrefix))

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
			secExplanations = append(secExplanations, ReportExplanation{Percentage: percentChange, Effect: effect,
				Attribution: fmt.Sprintf("%s %s", curResultGroupName, group)})
		}
	}

	secExplanations = sortAndLimitExplanations(secExplanations)
	for _, explanation := range secExplanations {
		explanations = append(explanations, explainPercentageChange(explanation.Percentage,
			explanation.Effect, explanation.Attribution, secondaryExplanationPrefix))
	}

	duReport.Explanations = explanations
}

func getAggrAsFloat64(aggrInt interface{}) (float64, error) {
	switch aggrInt.(type) {
	case int:
		return float64(aggrInt.(int)), nil
	case int64:
		return float64(aggrInt.(int64)), nil
	case float32:
		return float64(aggrInt.(float32)), nil
	case float64:
		return aggrInt.(float64), nil
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

func getAggrByTimestampAndGroup(queryResult *QueryResult,
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
					displayGroupValue = groupValue + " / " + colValue
				}
			}
		}

		(*uniqueGroupsSet)[groupValue] = displayGroupValue
		aggrByTimestampAndGroup[timestamp][groupValue] = aggr
		totalAggr = totalAggr + aggr
	}

	return aggrByTimestampAndGroup, timestamps, totalAggr
}

func getDayFromTimestampStr(timestampStr string) string {
	timestamp, _ := time.Parse(time.RFC3339, timestampStr)
	return fmt.Sprintf("%s (%s)", timestamp.Weekday().String(), timestamp.Format("Jan 02"))
}

// getTotalAggrForUniqueUsersQuery - Runs same query without group by timestamp.
func getTotalAggrForUniqueUsersQuery(projectId uint64, uuQuery Query) float64 {
	uuQuery.GroupByTimestamp = false

	queryResult, _, _ := Analyze(projectId, uuQuery)

	aggrIndex, _, _ := GetTimstampAndAggregateIndexOnQueryResult(queryResult.Headers)
	var total float64
	for _, row := range queryResult.Rows {
		total = total + float64(row[aggrIndex].(int64))
	}

	return total
}

func addExplanationsForPresentationLine(duReport *DashboardUnitReport) {
	prevResult := duReport.Results[0].QueryResult
	curResult := duReport.Results[1].QueryResult

	uniqueGroupsSet := make(map[string]string)
	prevAggrMap, prevTimestamps, prevTotal := getAggrByTimestampAndGroup(&prevResult, &uniqueGroupsSet)
	curAggrMap, curTimestamps, curTotal := getAggrByTimestampAndGroup(&curResult, &uniqueGroupsSet)

	if curResult.Meta.Query.Type == QueryTypeUniqueUsers {
		prevTotal = getTotalAggrForUniqueUsersQuery(duReport.ProjectID, prevResult.Meta.Query)
		curTotal = getTotalAggrForUniqueUsersQuery(duReport.ProjectID, curResult.Meta.Query)
	}

	explanations := make([]string, 0, 0)
	percentChange, totalEffect := getPercentageChange(prevTotal, curTotal)
	explanations = append(explanations, explainPercentageChange(percentChange, totalEffect,
		duReport.Title, primaryExplanationPrefix))

	secExplanations := make([]ReportExplanation, 0, 0)
	for i := range curTimestamps {
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
				secExplanations = append(secExplanations, ReportExplanation{Percentage: percentChange, Effect: effect,
					Attribution: fmt.Sprintf("%s last %s", displayGroup, getDayFromTimestampStr(curTimestamp))})
			}
		}
	}

	secExplanations = sortAndLimitExplanations(secExplanations)
	for _, explanation := range secExplanations {
		explanations = append(explanations, explainPercentageChange(explanation.Percentage,
			explanation.Effect, explanation.Attribution, secondaryExplanationPrefix))
	}

	duReport.Explanations = explanations
}

func addExplanationsByPresentation(duReport DashboardUnitReport) DashboardUnitReport {
	if duReport.Presentation == "" || len(duReport.Results) < 2 {
		return duReport
	}

	switch duReport.Presentation {
	case PresentationCard:
		addExplanationsForPresentationCard(&duReport)
	case PresentationBar:
		addExplanationsForPresentationBar(&duReport)
	case PresentationLine:
		addExplanationsForPresentationLine(&duReport)
	}

	return duReport
}

func addExplanationsToReportUnits(report *Report) {
	for unitId, dashboardUnitReport := range report.Contents.DashboardUnitIDToDashboardUnitReport {
		report.Contents.DashboardUnitIDToDashboardUnitReport[unitId] = addExplanationsByPresentation(dashboardUnitReport)
	}
}

func getDashboardUnitReport(projectID uint64, dashboardUnit DashboardUnit,
	intervalBeforeThat, interval Interval) (*DashboardUnitReport, int) {

	query := Query{}
	err := json.Unmarshal(dashboardUnit.Query.RawMessage, &query)
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	intervalBeforeReportUnit, errCode := getReportUnit(projectID, query, intervalBeforeThat)
	if errCode != http.StatusOK {
		return nil, errCode
	}
	intervalReportUnit, errCode := getReportUnit(projectID, query, interval)
	if errCode != http.StatusOK {
		return nil, errCode
	}

	dashboardUnitReport := &DashboardUnitReport{
		ProjectID:    projectID,
		Title:        dashboardUnit.Title,
		Presentation: dashboardUnit.Presentation,
		Results:      []ReportUnit{*intervalBeforeReportUnit, *intervalReportUnit},
	}

	return dashboardUnitReport, http.StatusOK
}

func getCount(r ReportUnit) float64 {
	if len(r.QueryResult.Rows) > 0 && len(r.QueryResult.Rows[0]) > 0 {
		count, _ := getAggrAsFloat64(r.QueryResult.Rows[0][0])
		return count
	}

	return 0
}

func (r *Report) GetTextContent() string {
	output := new(strings.Builder)

	output.WriteString("Report for")
	output.WriteString(fmt.Sprintf("%s - %s", unixToHumanTime(r.StartTime), unixToHumanTime(r.EndTime)))
	output.WriteString("\n\n")

	output.WriteString("Dashboard name: ")
	output.WriteString(r.DashboardName)
	output.WriteString("\n\n")

	for _, dshBU := range r.Contents.DashboardUnitIDToDashboardUnitReport {
		if dshBU.Presentation != PresentationCard {
			continue
		}

		output.WriteString(dshBU.Title)
		output.WriteString("\n\n")
		output.WriteString(fmt.Sprintf("%s - %s", dshBU.Results[0].StartTime, dshBU.Results[0].EndTime))
		output.WriteString(fmt.Sprintf("%s - %s", dshBU.Results[1].StartTime, dshBU.Results[1].EndTime))
		output.WriteString("\n\n")
		output.WriteString(fmt.Sprintf("%f", getCount(dshBU.Results[0])))
		output.WriteString(fmt.Sprintf("%f", getCount(dshBU.Results[1])))
		output.WriteString("\n\n")
	}
	return output.String()
}

/*
<div>
	<p>Report for Jun 2, 2019 - Jun 15, 2019</p>
	<p>Dashboard Name</p>
	<div>
		<p>DashboardUnit Name</p>
		<table>
			<tr>
				<th style="padding:0 15px 0 15px;">Jun 2, 2019 - Jun 8, 2019</th>
				<th style="padding:0 15px 0 15px;">Jun 9, 2019 - Jun 15, 2019</th>
			</tr>
			<tr>
				<td style="padding:0 15px 0 15px; text-align: center;">12225</td>
				<td style="padding:0 15px 0 15px; text-align: center;">6224</td>
			</tr>
		</table>
	</div>
</div>
*/
func (r *Report) GetHTMLContent() string {
	output := new(strings.Builder)

	output.WriteString("<p>Report for ")
	output.WriteString(fmt.Sprintf("%s - %s", unixToHumanTime(r.StartTime), unixToHumanTime(r.EndTime)))
	output.WriteString("</p>")

	output.WriteString("<p>Dashboard name: ")
	output.WriteString(r.DashboardName)
	output.WriteString("</p>")

	output.WriteString("<div>")
	for _, dshBU := range r.Contents.DashboardUnitIDToDashboardUnitReport {
		if dshBU.Presentation != PresentationCard {
			continue
		}
		output.WriteString("<p>")
		output.WriteString(dshBU.Title)
		output.WriteString("</p>")
		output.WriteString("<table>")
		output.WriteString("<tr>")
		output.WriteString("<th style='padding-left:15px;'>")
		output.WriteString(fmt.Sprintf("%s- %s", unixToHumanTime(dshBU.Results[0].StartTime),
			unixToHumanTime(dshBU.Results[0].EndTime)))
		output.WriteString("</th>")
		output.WriteString("<th style='padding-left:15px;'>")
		output.WriteString(fmt.Sprintf("%s- %s", unixToHumanTime(dshBU.Results[1].StartTime),
			unixToHumanTime(dshBU.Results[1].EndTime)))
		output.WriteString("</th>")
		output.WriteString("</tr>")
		output.WriteString("<tr>")
		output.WriteString("<td style='padding-left:15px;'>")
		output.WriteString(fmt.Sprintf("%f", getCount(dshBU.Results[0])))
		output.WriteString("</td>")
		output.WriteString("<td style='padding-left:15px;'>")
		output.WriteString(fmt.Sprintf("%f", getCount(dshBU.Results[1])))
		output.WriteString("</td>")
		output.WriteString("</tr>")
		output.WriteString("</table>")
	}
	output.WriteString("</div>")
	return output.String()
}

func unixToHumanTime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}
