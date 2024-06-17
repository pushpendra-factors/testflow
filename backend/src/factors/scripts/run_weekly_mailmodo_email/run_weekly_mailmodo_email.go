package main

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const MAILMODO_WEEKLY_EMAILS_CAMPAIGN_ID = "08e62b90-21c4-55c7-b6aa-609e953bdf2d"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	mailmodoTriggerCampaignAPIKey := flag.String("mailmodo_trigger_campaign_api_key", "dummy", "Mailmodo Email Alert API Key")
	projectIdsStringList := flag.String("project_ids", "", "List of projects for which job will run")
	blackListedProjectIdsStringList := flag.String("blacklisted_project_ids", "", "List of projects for which job will not run")

	flag.Parse()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	}

	appName := "weekly_mailmodo_email_job"
	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:              *primaryDatastore,
		MailModoTriggerCampaignAPIKey: *mailmodoTriggerCampaignAPIKey,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
		os.Exit(1)
	}

	// timestamp of last 15 day
	timestamp := U.TimeNowUnix() - 15*24*60*60
	blackListedProjectIds := C.GetTokensFromStringListAsUint64(*blackListedProjectIdsStringList)

	var projectIds []int64
	if *projectIdsStringList != "*" {
		projectIds = C.GetTokensFromStringListAsUint64(*projectIdsStringList)
	} else {
		projectIds, err = store.GetStore().GetActiveProjectByEventsPerformedTimeStamp(timestamp)
		if err != nil {
			log.WithError(err).Fatal("Failed to get projects.")
			return
		}
	}

	errToProjectIDMap := make(map[string][]int64)

	for _, projectId := range projectIds {

		if U.Int64ValueIn(projectId, blackListedProjectIds) {
			log.WithField("project_id", projectId).Warn("Skipping the project.")
			continue
		}

		project, errCode := store.GetStore().GetProject(projectId)
		if errCode != http.StatusFound {
			errMsg := fmt.Sprintf("%v", errCode)
			errToProjectIDMap[errMsg] = append(errToProjectIDMap[errMsg], projectId)
			continue
		}

		// Get email id of agents of the project
		emails, errCode := GetAgentEmailByProjectID(projectId)
		if errCode != http.StatusFound {
			errMsg := fmt.Sprintf("%v", errCode)
			errToProjectIDMap[errMsg] = append(errToProjectIDMap[errMsg], projectId)
			continue
		}

		// Get the timerange and metrics.
		startTimeStamp, endTimeStamp, err := U.GetQueryRangePresetLastWeekIn(U.TimeZoneStringIST)
		if err != nil {
			errMsg := fmt.Sprintf("%v", err)
			errToProjectIDMap[errMsg] = append(errToProjectIDMap[errMsg], projectId)
			continue
		}

		startDate := U.GetDateFromTimestampAndTimezone(startTimeStamp, U.TimeZoneStringIST)
		endDate := U.GetDateFromTimestampAndTimezone(endTimeStamp, U.TimeZoneStringIST)

		metrics, err := store.GetStore().GetWeeklyMailmodoEmailsMetrics(projectId, startTimeStamp, endTimeStamp)
		if err != nil {
			errMsg := fmt.Sprintf("%v", err)
			errToProjectIDMap[errMsg] = append(errToProjectIDMap[errMsg], projectId)
			continue
		}

		metrics.ProjectName = project.Name

		log.WithField("metrics", metrics).Info("Weekly mailmodo metrics")

		for _, email := range emails {

			isEmailAllowed, err := model.IsReceipentAllowedMailmodo(email, "transactional")
			if err != nil {
				log.WithFields(log.Fields{"projectId": projectId, "email": email}).Error("Failed to check if mail is allowed")
				continue
			}

			if !isEmailAllowed {
				continue
			}

			ExecuteWeeklyMailmodoMails(email, metrics, startDate, endDate, &http.Client{})

		}

	}

	if len(errToProjectIDMap) > 0 {
		log.Warn("Error Map: ", errToProjectIDMap)
	}

}

func GetAgentEmailByProjectID(projectId int64) ([]string, int) {

	projectAgentsMap, errCode := store.GetStore().GetProjectAgentMappingsByProjectId(projectId)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	agentUUIDs := make([]string, 0)
	for _, projectAgent := range projectAgentsMap {
		agentUUIDs = append(agentUUIDs, projectAgent.AgentUUID)
	}

	agents, errCode := store.GetStore().GetAgentsByUUIDs(agentUUIDs)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	agentEmails := make([]string, 0)
	for _, agent := range agents {
		agentEmails = append(agentEmails, agent.Email)
	}

	return agentEmails, http.StatusFound

}

func ExecuteWeeklyMailmodoMails(email string, metrics model.WeeklyMailmodoEmailMetrics, startDate, endDate string, client HTTPClient) error {

	payloadJSON, err := json.Marshal(model.MailmodoTriggerCampaignRequestPayload{ReceiverEmail: email, Data: map[string]interface{}{
		"project_name":         metrics.ProjectName,
		"from_date":            startDate,
		"to_date":              endDate,
		"unique_visitors":      metrics.SessionUniqueUserCount,
		"form_submissions":     metrics.FormSubmittedUniqueUserCount,
		"companies_identified": metrics.IdentifiedCompaniesCount,
		"count_difference":     (metrics.IdentifiedCompaniesCount - metrics.FormSubmittedUniqueUserCount),
		"company_domain_1":     metrics.Company1.Domain, "company_industry_1": metrics.Company1.Industry, "company_country_1": metrics.Company1.Country, "company_revenue_range_1": metrics.Company1.Revenue,
		"company_domain_2": metrics.Company2.Domain, "company_industry_2": metrics.Company2.Industry, "company_country_2": metrics.Company2.Country, "company_revenue_range_2": metrics.Company2.Revenue,
		"company_domain_3": metrics.Company3.Domain, "company_industry_3": metrics.Company3.Industry, "company_country_3": metrics.Company3.Country, "company_revenue_range_3": metrics.Company3.Revenue,
		"company_domain_4": metrics.Company4.Domain, "company_industry_4": metrics.Company4.Industry, "company_country_4": metrics.Company4.Country, "company_revenue_range_4": metrics.Company4.Revenue,
		"company_domain_5": metrics.Company5.Domain, "company_industry_5": metrics.Company5.Industry, "company_country_5": metrics.Company5.Country, "company_revenue_range_5": metrics.Company5.Revenue,
		"total_accounts":                  metrics.TotalIdentifiedCompaniesCount,
		"industry_percent_this_week":      metrics.Industry.TopValuePercent,
		"industry_this_week":              metrics.Industry.TopValue,
		"industry_percent_last_week":      metrics.Industry.TopValuePercentLastWeek,
		"industry_last_week":              metrics.Industry.TopValueLastWeek,
		"employee_size_percent_this_week": metrics.EmployeeRange.TopValuePercent,
		"employee_size_this_week":         metrics.EmployeeRange.TopValue,
		"employee_size_percent_last_week": metrics.EmployeeRange.TopValuePercentLastWeek,
		"employee_size_last_week":         metrics.EmployeeRange.TopValueLastWeek,
		"country_percent_this_week":       metrics.Country.TopValuePercent,
		"country_this_week":               metrics.Country.TopValue,
		"country_percent_last_week":       metrics.Country.TopValuePercentLastWeek,
		"country_last_week":               metrics.Country.TopValueLastWeek,
		"revenue_range_percent_this_week": metrics.RevenueRange.TopValuePercent,
		"revenue_range_this_week":         metrics.RevenueRange.TopValue,
		"revenue_range_percent_last_week": metrics.RevenueRange.TopValuePercentLastWeek,
		"revenue_range_last_week":         metrics.RevenueRange.TopValueLastWeek,
		"total_session_count":             metrics.SessionCount,
		"top_source":                      metrics.TopSource,
		"top_channel":                     metrics.TopChannel,
		"top_campaign":                    metrics.TopCampaign,
	}})
	if err != nil {
		return err
	}

	//campaign id
	request, err := model.FormMailmodoTriggerCampaignRequest(MAILMODO_WEEKLY_EMAILS_CAMPAIGN_ID, payloadJSON)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("failed to send weekly mail")
	}

	return nil

}
