package delta

import (
	C "factors/config"
	"factors/model/store"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func MailWeeklyInsights(projectID int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	dashboards := StaticDashboardListForMailer()
	startTimestamp := configs["startTimestamp"].(int64)
	var response interface{}
	for _, dashboard := range dashboards {
		queryId := dashboard.QueryId
		startTimestampTimeFormat := time.Unix(startTimestamp, 0)
		startTimestampMinusOneweekTimeFormat := time.Unix(startTimestamp-(86400*7), 0)
		response, _ = GetWeeklyInsights(projectID, "", queryId, &startTimestampTimeFormat, &startTimestampMinusOneweekTimeFormat, "w", 11, 100, 1)
	}
	projectAgents, _ := store.GetStore().GetProjectAgentMappingsByProjectId(projectID)
	for _, agent := range projectAgents {
		text := ""
		sub := "Weekly Insights Digest"
		html := fmt.Sprintf("%v", response)
		agent, _ := store.GetStore().GetAgentByUUID(agent.AgentUUID)
		err := C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), sub, html, text)
		if err != nil {
			log.WithError(err).Error("failed to send email alert")
			continue
		}
	}
	return result, true
}
