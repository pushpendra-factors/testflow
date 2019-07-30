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

func fetchDashboardReports(db *gorm.DB, projectID, dashboardID uint64) ([]*M.Report, int) {
	var dbReports []*M.DBReport

	err := db.Where("project_id = ? ", projectID).Where("dashboard_id = ?", dashboardID).Order("ID ASC").Find(&dbReports).Error

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

func fetchMostRecentReport(reports []*M.Report, projectID, dashboardID uint64, reportType string) (*M.Report, int) {
	var mostRecentReport *M.Report
	status := http.StatusNotFound
	mostRecentTs := int64(0)
	for _, report := range reports {
		if report.ProjectID == projectID && report.DashboardID == dashboardID {
			if report.StartTime > mostRecentTs {
				mostRecentReport = report
				mostRecentTs = report.StartTime
				status = http.StatusFound
			}
		}
	}

	return mostRecentReport, status
}

func findStartTimeForDashboard(dashboard *M.Dashboard, existingReports []*M.Report) time.Time {

	mostRecentReport, errCode := fetchMostRecentReport(existingReports, dashboard.ProjectId, dashboard.ID, M.ReportTypeWeekly)
	if errCode == http.StatusNotFound {
		// no reports for this dashboard add intervals after dashboard's creation time
		return dashboard.CreatedAt
	}

	// If most recent report is for weeks W1 and W2,
	// next report should be created for W2 and W3
	// Report creation should happen from starting of W2
	reportEndTimeTs := unixToUTCTime(mostRecentReport.EndTime)
	beginingOfEndWeek := now.New(reportEndTimeTs).BeginningOfWeek()
	return beginingOfEndWeek
}

func notifyStatus(env, tag string, success, noContent, failures []string) {
	buildStatus := map[string]interface{}{
		"success":   success,
		"noContent": noContent,
		"failures":  failures,
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

// startTime is rounded off to begining of the week
// endTime is rounded off to end of the week
func getWeeklyIntervals(startTime, endTime time.Time) []M.Interval {
	intervals := make([]M.Interval, 0, 0)
	startTs := startTime

	endTs := now.New(endTime).EndOfWeek().UTC()

	for startTs.Before(endTs) {

		weekStart := now.New(startTs).BeginningOfWeek()
		weekEnd := now.New(weekStart).EndOfWeek()

		interval := M.Interval{StartTime: weekStart.UTC().Unix(), EndTime: weekEnd.UTC().Unix()}
		intervals = append(intervals, interval)

		startTs = weekEnd.Add(24 * time.Hour)
	}
	return intervals
}

func getWeekInterval(startUnix, endUnix int64) M.Interval {
	startTime := unixToUTCTime(startUnix)
	endTime := unixToUTCTime(endUnix)

	return M.Interval{StartTime: now.New(startTime).BeginningOfWeek().UTC().Unix(), EndTime: now.New(endTime).EndOfWeek().UTC().Unix()}
}

func getPrevWeekInterval(startUnix, endUnix int64) M.Interval {
	startTime := unixToUTCTime(startUnix)

	startOfGivenWeek := now.New(startTime).BeginningOfWeek()
	prevWeek := now.New(startOfGivenWeek.Add(-24 * time.Hour))

	return M.Interval{StartTime: prevWeek.BeginningOfWeek().UTC().Unix(), EndTime: prevWeek.EndOfWeek().UTC().Unix()}
}

func findValidReportBy(reports []*M.Report, projectID, dashboardID uint64, startTime, endTime int64) (*M.Report, int) {
	for _, report := range reports {
		if report.ProjectID == projectID && report.DashboardID == dashboardID && report.StartTime == startTime && report.EndTime == endTime && !report.Invalid {
			return report, http.StatusFound
		}
	}
	return nil, http.StatusNotFound
}
