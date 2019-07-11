package reports

import (
	M "factors/model"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const (
	buildReportTag = "Task#BuildReport"
)

var reportLog = baseLog.WithField("prefix", buildReportTag)

func BuildReports(env string, db *gorm.DB, dashboards []*M.Dashboard, mailReports bool) {

	defer func() {
		reportLog.Infof("Finished")
	}()

	if mailReports {
		reportLog.Infof("Reports mailing enabled")
	} else {
		reportLog.Infof("Reports mailing disabled")
	}

	store := newStore(dashboards)

	buildReportsFor := make([]*ReportBuild, 0, 0)

	for _, dashboard := range dashboards {

		reportLog.Infof("Finding which reports to Build for dashboardID: %d", dashboard.ID)
		dashboardReports, errCode := fetchDashboardReports(db, dashboard.ProjectId, dashboard.ID)
		if errCode == http.StatusInternalServerError {
			continue
		}

		buildReportsForDashboard, errCode := findWhichWeeklyReportsToBuildForDashboard(db, dashboard, dashboardReports)
		if errCode != http.StatusOK {
			continue
		}
		buildReportsFor = append(buildReportsFor, buildReportsForDashboard...)
		reportLog.Infof("Finding which reports to Re-Build for dashboardID: %d", dashboard.ID)

		rebuildReports := findWhichInvalidReportsToRebuild(dashboardReports, store)
		buildReportsFor = append(buildReportsFor, rebuildReports...)

	}

	reportLog.Infof("Start Creating Reports")
	createdReports, successList, noContentList, failureList := buildReports(buildReportsFor)
	if len(createdReports) == 0 {
		reportLog.Infof("No New Reports Created")
		notifyStatus(env, buildReportTag, successList, noContentList, failureList)
		return
	}

	if mailReports {
		reportLog.Infof("Start Emailing Reports")
		sendReportsEmail(store, createdReports)
	}

	notifyStatus(env, buildReportTag, successList, noContentList, failureList)

}

type ReportBuild struct {
	ProjectID          uint64
	DashboardID        uint64
	DashboardName      string
	IntervalBeforeThat M.Interval
	Interval           M.Interval
}

func findWhichWeeklyReportsToBuildForDashboard(db *gorm.DB, dashboard *M.Dashboard, existingReports []*M.Report) ([]*ReportBuild, int) {

	reportBuilds := make([]*ReportBuild, 0, 0)
	endTime := getLastWeekEndTime()

	startTime := findStartTimeForDashboard(existingReports, dashboard)
	intervals := getWeeklyIntervals(startTime, endTime)

	for i := 1; i < len(intervals); i++ {
		reportBuild := &ReportBuild{
			ProjectID:          dashboard.ProjectId,
			DashboardID:        dashboard.ID,
			DashboardName:      dashboard.Name,
			IntervalBeforeThat: intervals[i-1],
			Interval:           intervals[i],
		}
		reportBuilds = append(reportBuilds, reportBuild)
	}

	return reportBuilds, http.StatusOK
}

func buildReports(buildReportsFor []*ReportBuild) (reports []*M.Report, successList, noContentList, failureList []string) {
	failureList = make([]string, 0, 0)
	successList = make([]string, 0, 0)
	noContentList = make([]string, 0, 0)
	reports = make([]*M.Report, 0, 0)

	for _, bR := range buildReportsFor {
		report, errCode := M.GenerateReport(bR.ProjectID, bR.DashboardID, bR.DashboardName, bR.IntervalBeforeThat, bR.Interval)
		if errCode == http.StatusInternalServerError {
			failureList = append(failureList, fmt.Sprintf("Failed to generate report, ProjectID: %d, DashboardID: %d, IntervalStart: %d", bR.ProjectID, bR.DashboardID, bR.IntervalBeforeThat.StartTime))
			continue
		} else if errCode == http.StatusNotFound {
			noContentList = append(noContentList, fmt.Sprintf("No Content, ProjectID: %d, DashboardID: %d, IntervalStart: %d", bR.ProjectID, bR.DashboardID, bR.IntervalBeforeThat.StartTime))
			continue
		}
		report, errCode = M.CreateReport(report)
		if errCode != http.StatusCreated {
			failureList = append(failureList, fmt.Sprintf("failed to store report in DB, ProjectID: %d, DashboardID: %d, IntervalStart: %d", bR.ProjectID, bR.DashboardID, bR.IntervalBeforeThat.StartTime))
			continue
		}
		reports = append(reports, report)
		successList = append(successList, fmt.Sprintf("ReportID: %d, ProjectID: %d, DashboardID: %d, Type: %s, IntervalStart: %d, IntervalEnd: %d", report.ID, bR.ProjectID, bR.DashboardID, report.Type, bR.IntervalBeforeThat.StartTime, bR.Interval.EndTime))
	}
	return
}

func findWhichInvalidReportsToRebuild(reports []*M.Report, store *store) []*ReportBuild {
	reportBuilds := make([]*ReportBuild, 0, 0)
	for _, report := range reports {

		if !report.Invalid {
			continue
		}

		_, errCode := findValidReportBy(reports, report.ProjectID, report.DashboardID, report.StartTime, report.EndTime)
		validReportPresentForSameInterval := errCode == http.StatusFound

		if validReportPresentForSameInterval {
			continue
		}

		dashboard, errCode := store.getDashboard(report.DashboardID)
		if errCode != http.StatusFound {
			reportLog.WithFields(log.Fields{
				"ProjectID":   report.ProjectID,
				"DashboardID": report.DashboardID,
				"StartTime":   report.StartTime,
				"EndTime":     report.EndTime,
			}).Errorln("Failed to fetch dashboard")
			continue
		}

		intervals := getWeeklyIntervals(unixToUTCTime(report.StartTime), unixToUTCTime(report.EndTime))

		rb := &ReportBuild{
			ProjectID:          report.ProjectID,
			DashboardID:        report.DashboardID,
			DashboardName:      dashboard.Name,
			IntervalBeforeThat: intervals[0],
			Interval:           intervals[1],
		}
		reportBuilds = append(reportBuilds, rb)

	}
	return reportBuilds
}
