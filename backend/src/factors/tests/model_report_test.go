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

	units := make([]M.DashboardUnitReport, 0, 0)
	unit := M.DashboardUnitReport{
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
	units = append(units, unit)

	report := &M.Report{
		ID:    uint64(1),
		Type:  M.ReportTypeWeekly,
		Units: units,
	}
	dbReport, err := M.TranslateReportToDBReport(report)
	assert.Nil(t, err)

	resultReport, err := M.TranslateDBReportToReport(dbReport)
	assert.Nil(t, err)

	assert.Equal(t, report, resultReport)
}
