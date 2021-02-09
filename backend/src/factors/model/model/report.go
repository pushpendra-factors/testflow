package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Report struct {
	ID            uint64                `json:"id"`
	ProjectID     uint64                `json:"project_id"`
	DashboardID   uint64                `json:"dashboard_id"`
	DashboardName string                `json:"dashboard_name"`
	CreatedAt     time.Time             `json:"created_at"`
	Type          string                `json:"type"`
	StartTime     int64                 `json:"start_time"`
	EndTime       int64                 `json:"end_time"`
	Invalid       bool                  `json:"invalid"`
	Units         []DashboardUnitReport `json:"units"`
}

func (r *Report) GetTextContent() string {
	output := new(strings.Builder)

	output.WriteString("Report for")
	output.WriteString(fmt.Sprintf("%s - %s", U.UnixToHumanTime(r.StartTime), U.UnixToHumanTime(r.EndTime)))
	output.WriteString("\n\n")

	output.WriteString("Dashboard name: ")
	output.WriteString(r.DashboardName)
	output.WriteString("\n\n")

	for _, dshBU := range r.Units {
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

func getCount(r ReportUnit) float64 {
	if len(r.QueryResult.Rows) > 0 && len(r.QueryResult.Rows[0]) > 0 {
		count, _ := getAggrAsFloat64(r.QueryResult.Rows[0][0])
		return count
	}

	return 0
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
	output.WriteString(fmt.Sprintf("%s - %s", U.UnixToHumanTime(r.StartTime), U.UnixToHumanTime(r.EndTime)))
	output.WriteString("</p>")

	output.WriteString("<p>Dashboard name: ")
	output.WriteString(r.DashboardName)
	output.WriteString("</p>")

	output.WriteString("<div>")
	for _, dshBU := range r.Units {
		if dshBU.Presentation != PresentationCard {
			continue
		}
		output.WriteString("<p>")
		output.WriteString(dshBU.Title)
		output.WriteString("</p>")
		output.WriteString("<table>")
		output.WriteString("<tr>")
		output.WriteString("<th style='padding-left:15px;'>")
		output.WriteString(fmt.Sprintf("%s- %s", U.UnixToHumanTime(dshBU.Results[0].StartTime),
			U.UnixToHumanTime(dshBU.Results[0].EndTime)))
		output.WriteString("</th>")
		output.WriteString("<th style='padding-left:15px;'>")
		output.WriteString(fmt.Sprintf("%s- %s", U.UnixToHumanTime(dshBU.Results[1].StartTime),
			U.UnixToHumanTime(dshBU.Results[1].EndTime)))
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

type DashboardUnitReport struct {
	ProjectID          uint64       `json:"pid"`
	Title              string       `json:"t"`
	Presentation       string       `json:"p"`
	Results            []ReportUnit `json:"r"`
	Explanations       []string     `json:"e"`
	ChangeInPercentage float64      `json:"ord"`
}

type ReportUnit struct {
	StartTime   int64       `json:"st"`
	EndTime     int64       `json:"et"`
	QueryResult QueryResult `json:"qr"`
}

// DBReport represents structure to be used for storing report in database
type DBReport struct {
	ID            uint64 `gorm:"primary_key:true;"`
	ProjectID     uint64
	DashboardID   uint64
	DashboardName string
	CreatedAt     time.Time
	IsDeleted     bool `gorm:"not null;default:false"`
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

const (
	ReportTypeWeekly  = "w"
	ReportTypeMonthly = "m"
)

func TranslateDBReportToReport(dbReport *DBReport) (*Report, error) {
	units := make([]DashboardUnitReport, 0, 0)

	err := json.Unmarshal(dbReport.Contents.RawMessage, &units)
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
		Units:         units,
	}

	return &report, nil
}

func TranslateReportToDBReport(report *Report) (*DBReport, error) {
	unitsJsonBytes, err := json.Marshal(report.Units)
	if err != nil {
		return nil, err
	}

	postgresJson := postgres.Jsonb{RawMessage: unitsJsonBytes}

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

type Interval struct {
	StartTime int64
	EndTime   int64
}
