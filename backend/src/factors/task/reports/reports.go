package reports

import (
	M "factors/model"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

const (
	buildReportTag = "Task#BuildReport"
)

var reportLog = baseLog.WithField("prefix", buildReportTag)

func BuildReports(env string, db *gorm.DB, dashboards []*M.Dashboard, reportTypes []string,
	customStartTime int64, customEndTime int64, mailReports bool) {
	reportLog.Infof("Build reports started.")

	createdReports := make([]*M.Report, 0, 0)
	successList := make([]string, 0, 0)
	noContentList := make([]string, 0, 0)
	failureList := make([]string, 0, 0)

	for _, reportType := range reportTypes {
		reports, success, noContents, failures := buildReportsByType(env, db, dashboards, reportType,
			customStartTime, customEndTime, mailReports)

		createdReports = append(createdReports, reports...)
		successList = append(successList, success...)
		noContentList = append(noContentList, noContents...)
		failureList = append(failureList, failures...)
	}

	if len(createdReports) == 0 {
		reportLog.Infof("No new reports created.")
		notifyStatus(env, buildReportTag, successList, noContentList, failureList)
		return
	}

	/*
		if mailReports {
			reportLog.Infof("Start Emailing Reports")
			sendReportsEmail(store, createdReports)
		}
	*/

	notifyStatus(env, buildReportTag, successList, noContentList, failureList)
}

func buildReportsByType(env string, db *gorm.DB, dashboards []*M.Dashboard, reportType string,
	customStartTime int64, customEndTime int64, mailReports bool) (reports []*M.Report,
	successList, noContentList, failureList []string) {

	defer func() {
		reportLog.Infof("Successfully built reports.")
	}()

	if mailReports {
		reportLog.Infof("Reports mailing enabled.")
	} else {
		reportLog.Infof("Reports mailing disabled.")
	}

	store := newStore(dashboards)

	reportBuilds := make([]*ReportBuild, 0, 0)
	buildReportsDedupe := make(map[string]bool, 0)

	for _, dashboard := range dashboards {
		reportLog.Infof("Finding which reports to Build for dashboard_id %d.", dashboard.ID)
		dashboardReports, errCode := fetchDashboardReportsByType(db, dashboard.ProjectId, dashboard.ID, reportType)
		if errCode == http.StatusInternalServerError {
			continue
		}

		// custom start_time and end_time will be used for every dashboard, if given.
		buildReportsForDashboard, errCode := findWhichReportsToBuild(db,
			dashboard, dashboardReports, reportType, customStartTime, customEndTime, &buildReportsDedupe)
		if errCode != http.StatusOK {
			continue
		}
		reportBuilds = append(reportBuilds, buildReportsForDashboard...)
		reportLog.Infof("Finding which reports to Re-Build for dashboard_id %d.", dashboard.ID)

		rebuildReportsForDashboard := findWhichInvalidReportsToRebuild(dashboardReports, reportType, store, &buildReportsDedupe)
		reportBuilds = append(reportBuilds, rebuildReportsForDashboard...)
	}

	return buildReportsByBuildConfig(reportBuilds)
}

type ReportBuild struct {
	ProjectID          uint64
	DashboardID        uint64
	DashboardName      string
	Type               string
	IntervalBeforeThat M.Interval
	Interval           M.Interval
}

func getDashboardReportDedupeKey(projectId, dashboardId uint64, reportType string,
	startTime int64, endTime int64) string {

	return fmt.Sprintf("%d:%d_%d:%d", projectId, dashboardId, reportType, startTime, endTime)
}

func getWeeklyStartAndEndTime(dashboard *M.Dashboard, reportType string, eReports []*M.Report,
	cStartTimeUnix, cEndTimeUnix int64) (time.Time, time.Time) {

	var startTime time.Time
	if cStartTimeUnix > 0 {
		startTime = now.New(unixToUTCTime(cStartTimeUnix)).BeginningOfWeek()
	} else {
		startTime = findStartTimeForDashboardByType(dashboard, eReports, reportType)
	}

	var endTime time.Time
	if cEndTimeUnix > 0 && cEndTimeUnix > cStartTimeUnix {
		endTime = now.New(unixToUTCTime(cEndTimeUnix)).EndOfWeek()
	} else {
		endTime = getLastWeekEndTime()
	}

	return startTime, endTime
}

func getMonthlyStartAndEndTime(dashboard *M.Dashboard, reportType string, eReports []*M.Report,
	cStartTimeUnix, cEndTimeUnix int64) (time.Time, time.Time) {

	var startTime time.Time
	if cStartTimeUnix > 0 {
		startTime = now.New(unixToUTCTime(cStartTimeUnix)).BeginningOfMonth()
	} else {
		startTime = findStartTimeForDashboardByType(dashboard, eReports, reportType)
	}

	var endTime time.Time
	if cEndTimeUnix > 0 && cEndTimeUnix > cStartTimeUnix {
		endTime = now.New(unixToUTCTime(cEndTimeUnix)).EndOfMonth()
	} else {
		endTime = getLastMonthEndTime()
	}

	return startTime, endTime
}

func findWhichReportsToBuild(db *gorm.DB, dashboard *M.Dashboard, existingReports []*M.Report,
	reportType string, customStartTime, customEndTime int64,
	buildReportsDedupe *map[string]bool) ([]*ReportBuild, int) {

	var startTime, endTime time.Time
	if reportType == M.ReportTypeWeekly {
		startTime, endTime = getWeeklyStartAndEndTime(dashboard, reportType, existingReports,
			customStartTime, customEndTime)
	} else if reportType == M.ReportTypeMonthly {
		startTime, endTime = getMonthlyStartAndEndTime(dashboard, reportType, existingReports,
			customStartTime, customEndTime)
	} else {
		log.Errorf("invalid report type given : %s", reportType)
		return nil, http.StatusBadRequest
	}

	existingReportsLookup := make(map[string]bool, 0)
	for _, eReport := range existingReports {
		encDedupeKey := getDashboardReportDedupeKey(eReport.ProjectID, eReport.DashboardID, reportType,
			eReport.StartTime, eReport.EndTime)
		existingReportsLookup[encDedupeKey] = true
	}

	intervals := getIntervalsByType(startTime, endTime, reportType)
	reportBuilds := make([]*ReportBuild, 0, 0)
	for i := 1; i < len(intervals); i++ {
		key := getDashboardReportDedupeKey(dashboard.ProjectId, dashboard.ID, reportType,
			intervals[i].StartTime, intervals[i].EndTime)

		_, exists := existingReportsLookup[key]
		_, added := (*buildReportsDedupe)[key]
		if exists || added {
			continue
		}

		reportBuild := &ReportBuild{
			ProjectID:          dashboard.ProjectId,
			DashboardID:        dashboard.ID,
			DashboardName:      dashboard.Name,
			IntervalBeforeThat: intervals[i-1],
			Interval:           intervals[i],
			Type:               reportType,
		}
		reportBuilds = append(reportBuilds, reportBuild)
		(*buildReportsDedupe)[key] = true
	}

	return reportBuilds, http.StatusOK
}

func buildReportsByBuildConfig(buildReportsFor []*ReportBuild) (reports []*M.Report,
	successList, noContentList, failureList []string) {

	failureList = make([]string, 0, 0)
	successList = make([]string, 0, 0)
	noContentList = make([]string, 0, 0)
	reports = make([]*M.Report, 0, 0)

	for _, bR := range buildReportsFor {
		reportLog.Infof("Building report for project_id %d dashboard_id %d.", bR.ProjectID, bR.DashboardID)

		report, errCode := M.GenerateReport(bR.ProjectID, bR.DashboardID, bR.DashboardName,
			bR.Type, bR.IntervalBeforeThat, bR.Interval)
		if errCode == http.StatusInternalServerError {
			failureList = append(failureList,
				fmt.Sprintf("Failed to generate report, ProjectID: %d, DashboardID: %d, IntervalStart: %d",
					bR.ProjectID, bR.DashboardID, bR.IntervalBeforeThat.StartTime))
			continue
		} else if errCode == http.StatusNotFound {
			noContentList = append(noContentList,
				fmt.Sprintf("No Content, ProjectID: %d, DashboardID: %d, IntervalStart: %d",
					bR.ProjectID, bR.DashboardID, bR.IntervalBeforeThat.StartTime))
			continue
		}

		report, errCode = M.CreateReport(report)
		if errCode != http.StatusCreated {
			failureList = append(failureList,
				fmt.Sprintf("failed to store report in DB, ProjectID: %d, DashboardID: %d, IntervalStart: %d",
					bR.ProjectID, bR.DashboardID, bR.IntervalBeforeThat.StartTime))
			continue
		}
		reports = append(reports, report)
		successList = append(successList,
			fmt.Sprintf("ReportID: %d, ProjectID: %d, DashboardID: %d, Type: %s, IntervalStart: %d, IntervalEnd: %d",
				report.ID, bR.ProjectID, bR.DashboardID, report.Type, bR.IntervalBeforeThat.StartTime, bR.Interval.EndTime))
	}
	return
}

func findWhichInvalidReportsToRebuild(existingReports []*M.Report, reportType string,
	store *store, buildReportsDedupe *map[string]bool) []*ReportBuild {

	reportBuilds := make([]*ReportBuild, 0, 0)
	for _, report := range existingReports {
		if !report.Invalid {
			continue
		}

		_, errCode := findValidReportBy(existingReports, report.ProjectID, report.DashboardID, report.StartTime, report.EndTime)
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

		interval := getWeekInterval(report.StartTime, report.EndTime)
		key := getDashboardReportDedupeKey(report.ProjectID, report.DashboardID, reportType, interval.StartTime, interval.EndTime)
		if _, added := (*buildReportsDedupe)[key]; added {
			continue
		}

		intervalBefore := getPrevWeekInterval(report.StartTime, report.EndTime)
		reportBuild := &ReportBuild{
			ProjectID:          report.ProjectID,
			DashboardID:        report.DashboardID,
			DashboardName:      dashboard.Name,
			Interval:           interval,
			IntervalBeforeThat: intervalBefore,
			Type:               reportType,
		}
		reportBuilds = append(reportBuilds, reportBuild)
		(*buildReportsDedupe)[key] = true

	}
	return reportBuilds
}
