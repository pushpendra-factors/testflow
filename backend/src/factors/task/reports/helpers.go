package reports

import (
	M "factors/model"
	"factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

func fetchDashboardReportsByType(db *gorm.DB, projectID, dashboardID uint64,
	reportType string) ([]*M.Report, int) {

	var dbReports []*M.DBReport

	err := db.Where("project_id = ? ", projectID).Where("dashboard_id = ? AND type = ?",
		dashboardID, reportType).Order("ID ASC").Find(&dbReports).Error

	if err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(dbReports) == 0 {
		return nil, http.StatusNotFound
	}

	reports := make([]*M.Report, len(dbReports), len(dbReports))

	for i, dbReport := range dbReports {
		report, err := M.TranslateDBReportToReport(dbReport)
		if err != nil {
			continue
		}
		reports[i] = report
	}

	return reports, http.StatusFound
}

func fetchMostRecentReportByType(reports []*M.Report, projectID, dashboardID uint64,
	reportType string) (*M.Report, int) {

	var mostRecentReport *M.Report
	status := http.StatusNotFound
	mostRecentTs := int64(0)
	for _, report := range reports {
		if report.ProjectID == projectID &&
			report.DashboardID == dashboardID &&
			report.Type == reportType {

			if report.StartTime > mostRecentTs {
				mostRecentReport = report
				mostRecentTs = report.StartTime
				status = http.StatusFound
			}
		}
	}

	return mostRecentReport, status
}

func findStartTimeForDashboardByType(dashboard *M.Dashboard,
	existingReports []*M.Report, reportType string) time.Time {

	var startTime time.Time
	mostRecentReport, errCode := fetchMostRecentReportByType(existingReports,
		dashboard.ProjectId, dashboard.ID, reportType)
	if errCode == http.StatusNotFound {
		// no reports for this dashboard add intervals after dashboard's creation time
		startTime = dashboard.CreatedAt
	} else {
		// If most recent report is for weeks W1 and W2,
		// next report should be created for W2 and W3
		// Report creation should happen from starting of W2
		startTime = unixToUTCTime(mostRecentReport.EndTime)
	}

	var beginingOfPeriod time.Time
	if reportType == M.ReportTypeWeekly {
		beginingOfPeriod = now.New(startTime).BeginningOfWeek()
	} else if reportType == M.ReportTypeMonthly {
		beginingOfPeriod = now.New(startTime).BeginningOfMonth()
	}

	return beginingOfPeriod
}

func notifyStatus(env, tag string, success, noContent, failures []string) {
	buildStatus := map[string]interface{}{
		"1.no_of_success":                  len(success),
		"2.no_of_failures":                 len(failures),
		"3.no_of_empty_dashboards_skipped": len(noContent),

		"4.success":  success,
		"5.failures": failures,
	}
	if err := util.NotifyThroughSNS(tag, env, buildStatus); err != nil {
		log.WithError(err).Error("Failed to notify report build status.")
	}
}

func unixToUTCTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).UTC()
}

func getLastWeekEndTime() time.Time {
	currentWeekStart := now.BeginningOfWeek().UTC()
	lastWeekEndTs := now.New(currentWeekStart.Add(-24 * time.Hour)).EndOfWeek()
	return lastWeekEndTs
}

func getLastMonthEndTime() time.Time {
	currentMonthStart := now.BeginningOfMonth().UTC()
	lastMonthEndTs := now.New(currentMonthStart.Add(-24 * time.Hour)).EndOfMonth()
	return lastMonthEndTs
}

// startTime is rounded off to begining of the period type
// endTime is rounded off to end of the period type
func getIntervalsByType(startTime, endTime time.Time, reportType string) []M.Interval {
	intervals := make([]M.Interval, 0, 0)
	startTs := startTime.UTC()

	var endTs time.Time
	if reportType == M.ReportTypeWeekly {
		endTs = now.New(endTime).EndOfWeek()
	} else if reportType == M.ReportTypeMonthly {
		endTs = now.New(endTime).EndOfMonth()
	} else {
		log.Fatalf("Invalid report type give on get_intervals: %s", reportType)
	}
	endTs = endTs.UTC()

	for startTs.Before(endTs) {
		var periodStart, periodEnd time.Time
		if reportType == M.ReportTypeWeekly {
			periodStart = now.New(startTs).BeginningOfWeek()
			periodEnd = now.New(periodStart).EndOfWeek()
		} else if reportType == M.ReportTypeMonthly {
			periodStart = now.New(startTs).BeginningOfMonth()
			periodEnd = now.New(periodStart).EndOfMonth()
		} else {
			log.Fatalf("Invalid report type give on get_intervals: %s", reportType)
		}

		interval := M.Interval{StartTime: periodStart.Unix(), EndTime: periodEnd.Unix()}
		intervals = append(intervals, interval)

		startTs = periodEnd.Add(24 * time.Hour)
	}

	return intervals
}

func getWeekInterval(startUnix, endUnix int64) M.Interval {
	startTime := unixToUTCTime(startUnix)
	endTime := unixToUTCTime(endUnix)

	return M.Interval{StartTime: now.New(startTime).BeginningOfWeek().UTC().Unix(),
		EndTime: now.New(endTime).EndOfWeek().UTC().Unix()}
}

func getPrevWeekInterval(startUnix, endUnix int64) M.Interval {
	startTime := unixToUTCTime(startUnix)

	startOfGivenWeek := now.New(startTime).BeginningOfWeek()
	prevWeek := now.New(startOfGivenWeek.Add(-24 * time.Hour))

	return M.Interval{StartTime: prevWeek.BeginningOfWeek().UTC().Unix(),
		EndTime: prevWeek.EndOfWeek().UTC().Unix()}
}

func findValidReportBy(reports []*M.Report, projectID, dashboardID uint64,
	startTime, endTime int64) (*M.Report, int) {

	for _, report := range reports {
		if report.ProjectID == projectID && report.DashboardID == dashboardID &&
			report.StartTime == startTime && report.EndTime == endTime && !report.Invalid {

			return report, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}
