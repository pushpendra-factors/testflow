package reports

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	modelStore "factors/model/store"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var reportMailerLog = baseLog.WithField("prefix", "Task#MailReport")

// Acts as local cache for info needed to mail reports
type store struct {
	dashboardIDToDashboard          map[uint64]*model.Dashboard
	projectIDToProjectAgentsMapping map[uint64][]model.ProjectAgentMapping
	agentUUIDToAgentInfo            map[string]*model.AgentInfo
}

func newStore(dashboards []*model.Dashboard) *store {

	dashboardIDToDashboard := make(map[uint64]*model.Dashboard)
	for _, dashboard := range dashboards {
		dashboardIDToDashboard[dashboard.ID] = dashboard
	}

	projectIDToProjectAgentsMapping := make(map[uint64][]model.ProjectAgentMapping)

	agentUUIDToAgentInfo := make(map[string]*model.AgentInfo)

	return &store{
		dashboardIDToDashboard:          dashboardIDToDashboard,
		projectIDToProjectAgentsMapping: projectIDToProjectAgentsMapping,
		agentUUIDToAgentInfo:            agentUUIDToAgentInfo,
	}
}

func (s *store) getDashboard(id uint64) (*model.Dashboard, int) {
	dashboard, exists := s.dashboardIDToDashboard[id]
	if exists {
		return dashboard, http.StatusFound
	}
	return nil, http.StatusNotFound
}

func (s *store) getAllProjectAgentMappings(projectID uint64) ([]model.ProjectAgentMapping, int) {

	if pam, exists := s.projectIDToProjectAgentsMapping[projectID]; exists {
		return pam, http.StatusFound
	}

	pam, errCode := modelStore.GetStore().GetProjectAgentMappingsByProjectId(projectID)
	if errCode != http.StatusFound {
		return []model.ProjectAgentMapping{}, errCode
	}

	s.projectIDToProjectAgentsMapping[projectID] = pam

	return pam, http.StatusFound
}

func (s *store) getAgentInfo(agentUUID string) (*model.AgentInfo, int) {
	if agentInfo, exists := s.agentUUIDToAgentInfo[agentUUID]; exists {
		return agentInfo, http.StatusFound
	}

	agentInfo, errCode := modelStore.GetStore().GetAgentInfo(agentUUID)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	s.agentUUIDToAgentInfo[agentUUID] = agentInfo

	return agentInfo, http.StatusFound
}

func isProjectAgent(projectAgentMappings []model.ProjectAgentMapping, agentUUID string) bool {
	for _, pam := range projectAgentMappings {
		if pam.AgentUUID == agentUUID {
			return true
		}
	}
	return false
}

func sendReportEmails(report *model.Report, store *store) error {

	reportMailerLog.Infof("sendReportEmails Sending Emails for ReportID: %d", report.ID)

	dashboard, errCode := store.getDashboard(report.DashboardID)
	if errCode != http.StatusFound {
		err := errors.New("DashboardNotFound")
		reportMailerLog.WithError(err).Errorf("sendReportEmails ReportID: %d, Dashboard not found", report.ID)
		return err
	}

	pams, errCode := store.getAllProjectAgentMappings(dashboard.ProjectId)
	if errCode != http.StatusFound {
		err := errors.New("ProjectAgentMappingsNotFound")
		reportMailerLog.WithError(err).Errorf("sendReportEmails ReportID: %d, projectagentmappings not found", report.ID)
		return err
	}

	agentsUUIDs := make([]string, 0, 0)
	if dashboard.Type == model.DashboardTypePrivate && isProjectAgent(pams, dashboard.AgentUUID) {
		agentsUUIDs = append(agentsUUIDs, dashboard.AgentUUID)
	} else if dashboard.Type == model.DashboardTypeProjectVisible {
		for _, pam := range pams {
			agentsUUIDs = append(agentsUUIDs, pam.AgentUUID)
		}
	}

	err := mailReportToAgents(report, agentsUUIDs, store)

	return err
}

func mailReportToAgents(report *model.Report, agentUUIDs []string, store *store) error {

	subject := "Report"
	html := report.GetHTMLContent()
	text := report.GetTextContent()
	mailer := C.GetServices().Mailer
	senderEmail := C.GetFactorsSenderEmail()

	for _, agentUUID := range agentUUIDs {

		reportMailerLog.WithFields(log.Fields{
			"AgentUUID": agentUUID,
			"ReportID":  report.ID,
		}).Infoln("mailReportToAgents Sending mail")

		agentInfo, errCode := store.getAgentInfo(agentUUID)
		if errCode != http.StatusFound {
			reportMailerLog.WithFields(log.Fields{
				"AgentUUID": agentUUID,
				"ReportID":  report.ID,
			}).Errorln("Agent Not Found")
			continue
		}

		err := mailer.SendMail(agentInfo.Email, senderEmail, subject, html, text)
		if err != nil {
			reportMailerLog.WithError(err).WithFields(log.Fields{
				"AgentUUID": agentUUID,
				"ReportID":  report.ID,
			}).Errorln("mailReportToAgents Failed to send email")
			continue
		}
	}
	return nil
}

func sendReportsEmail(store *store, reports []*model.Report) {
	for _, report := range reports {
		err := sendReportEmails(report, store)
		if err != nil {

		}
	}
}
