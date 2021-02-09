package tests

import (
	"factors/model/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTranslateReport(t *testing.T) {

	r := make([][]interface{}, 0, 0)
	r = append(r, []interface{}{float64(1)})

	units := make([]model.DashboardUnitReport, 0, 0)
	unit := model.DashboardUnitReport{
		Title:        "Test",
		Presentation: model.PresentationCard,
		Results: []model.ReportUnit{model.ReportUnit{
			StartTime: time.Now().UTC().Unix(),
			EndTime:   time.Now().UTC().Unix(),
			QueryResult: model.QueryResult{
				Headers: []string{"count"},
				Rows:    r,
			},
		}},
	}
	units = append(units, unit)

	report := &model.Report{
		ID:    uint64(1),
		Type:  model.ReportTypeWeekly,
		Units: units,
	}
	dbReport, err := model.TranslateReportToDBReport(report)
	assert.Nil(t, err)

	resultReport, err := model.TranslateDBReportToReport(dbReport)
	assert.Nil(t, err)

	assert.Equal(t, report, resultReport)
}
