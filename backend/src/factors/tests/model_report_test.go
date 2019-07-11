package tests

import (
	M "factors/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTranslateReport(t *testing.T) {

	r := make([][]interface{}, 0, 0)
	r = append(r, []interface{}{float64(1)})

	dashboardUnitIDToDashboardReports := make(map[uint64]M.DashboardUnitReport)
	dashboardUnitIDToDashboardReports[1] = M.DashboardUnitReport{
		Title:        "Test",
		Presentation: M.PresentationCard,
		Results: []M.ReportUnit{M.ReportUnit{
			StartTime: time.Now().UTC().Unix(),
			EndTime:   time.Now().UTC().Unix(),
			QueryResult: M.QueryResult{
				Headers: []string{"count"},
				Rows:    r,
			},
		}},
	}

	report := &M.Report{
		ID:   uint64(1),
		Type: M.ReportTypeWeekly,
		Contents: M.ReportContent{
			DashboardUnitIDToDashboardUnitReport: dashboardUnitIDToDashboardReports,
		},
	}
	dbReport, err := M.TranslateReportToDBReport(report)
	assert.Nil(t, err)

	resultReport, err := M.TranslateDBReportToReport(dbReport)
	assert.Nil(t, err)

	assert.Equal(t, report, resultReport)
}
