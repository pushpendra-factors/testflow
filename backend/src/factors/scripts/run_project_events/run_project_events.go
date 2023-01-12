package main

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	//	"reflect"
	"sort"
	"strconv"

	//"strings"
	"time"

	T "factors/task"

	//	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const NO_EVENT = "NoEvent"
const MetricsReportName = "metrics.txt"
const AlertsReportName = "alerts.txt"

// const ReportName = "report.txt"
const DetailedReportName = "detailed_report.txt"
const LookbackDaysForCacheData = 7

var ALL_EVENTS = [...]string{
	U.EVENT_NAME_SESSION,
	U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	U.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
	U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
}

type validationRule struct {
	feild    string
	property string
	required bool
	validate string
}
type AnalyticsData struct {
	GlobalLevelData         GlobalLevelData         `json:"global_level_data"`
	WebsiteAnalyticsData    WebsiteAnalyticsData    `json:"website_analytics_data"`
	HubspotAnalyticsData    HubspotAnalyticsData    `json:"hubspot_analytics_data"`
	SalesforceAnalyticsData SalesforceAnalyticsData `json:"salesforce_analytics_data"`
	FacebookAnalyticsData   FacebookAnalyticsData   `json:"facebook_analytics_data"`
	LinkedInAnalyticsData   LinkedinAnalyticsData   `json:"linkedin_analytics_data"`
}
type GlobalLevelData struct {
	TotalActiveUsers      uint64 `json:"total_active_users"`
	TotalEvents           uint64 `json:"total_events"`
	TotalRecordsProcessed uint64 `json:"total_records_processed"`
	IntegrationsEnabled   int64  `json:"integrations_enabled"`
	IntegrationsDisabled  int64  `json:"integrations_disabled"`
}
type WebsiteAnalyticsData struct {
	WebsiteDataActive    bool      `json:"website_data_active"`
	TotalSessions        uint64    `json:"total_sessions"`
	TotalPageViews       uint64    `json:"total_page_views"`
	TotalEventsFromWeb   uint64    `json:"total_events_from_web"`
	LastSeenPageViewTime time.Time `json:"last_seen_page_view_time"`
	LastSeenSessionTime  time.Time `json:"last_seen_session_time"`
}
type AdwordsAnalyticsData struct {
	IntegrationActive                 bool      `json:"integration_active"`
	TotalRecordsProcessed             uint64    `json:"total_records_processed"`
	LastSeenTimeOfRecord              time.Time `json:"last_seen_time_of_record"`
	TotalCampaignRecords              uint64    `json:"total_campaign_records"`
	LastSeenTimeOfCampaign            time.Time `json:"last_seen_time_of_campaign"`
	TotalAdgroupRecords               uint64    `json:"total_adgroup_records"`
	LastSeenTimeOfAdgroup             time.Time `json:"last_seen_time_of_adgroup"`
	TotalClickPerfromanceRecords      uint64    `json:"total_click_perfromance_records"`
	LastSeenTimeOfClickPerfromance    time.Time `json:"last_seen_time_of_click_perfromance"`
	TotalAdPerformanceRecords         uint64    `json:"total_ad_performance_records"`
	LastSeenTimeOfAdPerformance       time.Time `json:"last_seen_time_of_ad_performance"`
	TotalAdGroupPerformanceRecords    uint64    `json:"total_ad_group_performance_records"`
	LastSeenTimeOfAdGroupPerformance  time.Time `json:"last_seen_time_of_ad_group_performance"`
	TotalSearchPerformanceRecords     uint64    `json:"total_search_performance_records"`
	LastSeenTimeOfSearchPerformance   time.Time `json:"last_seen_time_of_search_performance"`
	TotalKeywordPerformanceRecords    uint64    `json:"total_keyword_performance_records"`
	LastSeenTimeOfKeywordPerformance  time.Time `json:"last_seen_time_of_keyword_performance"`
	TotalCustomerPerformanceRecords   uint64    `json:"total_customer_performance_records"`
	LastSeenTimeOfCustomerPerformance time.Time `json:"last_seen_time_of_customer_performance"`
}
type HubspotAnalyticsData struct {
	IntegrationActive          bool                       `json:"integration_active"`
	HubspotAnalyticsSyncData   HubspotAnalyticsSyncData   `json:"hubspot_analytics_sync_data"`
	HubspotEventsAnalyticsData HubspotEventsAnalyticsData `json:"hubspot_events_analytics_data"`
}
type HubspotAnalyticsSyncData struct {
	TotalContactRecordsSynced             uint64    `json:"total_contact_records_synced"`
	LastSeenTimeOfContactSync             time.Time `json:"last_seen_time_of_contact_sync"`
	TotalCompanyRecordsSynced             uint64    `json:"total_company_records_synced"`
	LastSeenTimeOfCompanySync             time.Time `json:"last_seen_time_of_company_sync"`
	TotalDealRecordsSynced                uint64    `json:"total_deal_records_synced"`
	LastSeenTimeOfDealSync                time.Time `json:"last_seen_time_of_deal_sync"`
	TotalEngagementRecordsSynced          uint64    `json:"total_engagement_records_synced"`
	LastSeenTimeOfEngagementSync          time.Time `json:"last_seen_time_of_engagement_sync"`
	TotalFormSubmissionRecordsSynced      uint64    `json:"total_form_submission_records_synced"`
	LastSeenTimeOfFormSubmssionRecordSync time.Time `json:"last_seen_time_of_form_submission_record_sync"`
	TotalFormRecordsSynced                uint64    `json:"total_form_records_synced"`
	LastSeenTimeOfFormRecordSync          time.Time `json:"last_seen_time_of_form_record_sync"`
	TotalContactListRecordsSynced         uint64    `json:"total_contact_list_records_synced"`
	LastSeenTimeOfContactListSynced       time.Time `json:"last_seen_time_of_contact_list_record_synced"`
}
type HubspotEventsAnalyticsData struct {
	TotalContactCreatedEvents    uint64    `json:"total_contact_created_events"`
	TotalContactUpdatedEvents    uint64    `json:"total_contact_updated_events"`
	TotalContactDeletedEvents    uint64    `json:"total_contact_deleted_events"`
	ActiveHubspotContacts        uint64    `json:"active_hubspot_contacts"`
	LastSeenTimeOfContactUpdated time.Time `json:"last_seen_time_of_contact_updated"`
	TotalCompanyCreatedEvents    uint64    `json:"total_company_created_events"`
	TotalCompanyUpdatedEvents    uint64    `json:"total_company_updated_events"`
	TotalCompanyDeletedEvents    uint64    `json:"total_company_deleted_events"`
	ActiveHubspotCompanies       uint64    `json:"active_hubspot_companies"`
	LastSeenTimeOfCompanyUpdated time.Time `json:"last_seen_time_of_company_updated"`
	TotalDealCreatedEvents       uint64    `json:"total_deal_created_events"`
	TotalDealUpdatedEvents       uint64    `json:"total_deal_updated_events"`
	TotalDealDeletedEvents       uint64    `json:"total_deal_deleted_events"`
	ActiveHubspotDeals           uint64    `json:"active_hubspot_deals"`
	LastSeenTimeOfDealUpdated    time.Time `json:"last_seen_time_of_deal_updated"`
	TotalEngagementCreatedEvents uint64    `json:"total_engagement_created_events"`
	TotalEngagementUpdatedEvents uint64    `json:"total_engagement_updated_events"`
	TotalEngagementDeletedEvents uint64    `json:"total_engagement_deleted_events"`
	ActiveHubspotEngagementUsers uint64    `json:"active_hubspot_engagement_users"`
	// contact list, form submission, form
	TotalContactListCreatedEvents    uint64 `json:"total_contact_list_created_events"`
	TotalContactListUpdatedEvents    uint64 `json:"total_contact_list_updated_events"`
	TotalContactListDeletedEvents    uint64 `json:"total_contact_list_deleted_events"`
	TotalFormSubmissionCreatedEvents uint64 `json:"total_form_submission_created_events"`
	TotalFormSubmissionUpdatedEvents uint64 `json:"total_form_submission_updated_events"`
	TotalFormSubmissionDeletedEvents uint64 `json:"total_form_submission_deleted_events"`
	TotalFormCreatedEvents           uint64 `json:"total_form_created_events"`
	TotalFormUpdatedEvents           uint64 `json:"total_form_updated_events"`
	TotalFormDeletedEvents           uint64 `json:"total_form_deleted_events"`
}
type HubspotUniqueMetrics struct {
	ActiveHubspotContactsMap  map[string]uint64
	ActiveHubspotCompaniesMap map[string]uint64
	ActiveHubspotDeals        map[string]uint64
	ActiveHubspotEngagements  map[string]uint64
}
type SalesforceAnalyticsData struct {
	IntegrationActive            bool                         `json:"integration_active"`
	SalesforceAnalyticsSyncData  SalesforceEventsSyncData     `json:"salesforce_analytics_sync_data"`
	SalesforceEventAnalyticsData SalesforceEventAnalyticsData `json:"salesforce_event_analytics_data"`
}
type SalesforceEventsSyncData struct {
	TotalContactRecordsSynced         uint64 `json:"total_contact_records_synced"`
	TotalLeadRecordsSynced            uint64 `json:"total_lead_records_synced"`
	TotalAccountRecordsSynced         uint64 `json:"total_account_records_synced"`
	TotalOpportunityRecordsSynced     uint64 `json:"total_opportunity_records_synced"`
	TotalCampaignRecordsSynced        uint64 `json:"total_campaign_records_synced"`
	TotalCampaignMemberRecordsSynced  uint64 `json:"total_campaign_member_records_synced"`
	TotalGroupAccountRecordsSynced    uint64 `json:"total_group_account_records_synced"`
	TotalOpportunityContactRoleSynced uint64 `json:"total_opportunity_contact_role_records_synced"`
}

type SalesforceEventAnalyticsData struct {
	TotalContactCreatedEvents                uint64 `json:"total_contact_created_events"`
	TotalContactUpdatedEvents                uint64 `json:"total_contact_updated_events"`
	TotalLeadCreatedEvents                   uint64 `json:"total_lead_created_events"`
	TotalLeadUpdatedEvents                   uint64 `json:"total_lead_updated_events"`
	TotalAccountCreatedEvents                uint64 `json:"total_account_created_events"`
	TotalAccountUpdatedEvents                uint64 `json:"total_account_updated_events"`
	TotalOpportunityCreatedEvents            uint64 `json:"total_opportunity_created_events"`
	TotalOpportunityUpdatedEvents            uint64 `json:"total_opportunity_updated_events"`
	TotalCampaignCreatedEvents               uint64 `json:"total_campaign_created_events"`
	TotalCampaignUpdatedEvents               uint64 `json:"total_campaign_updated_events"`
	TotalCampaignMemberCreatedEvents         uint64 `json:"total_campaign_member_created_events"`
	TotalCampaignMemberUpdatedEvents         uint64 `json:"total_campaign_member_updated_events"`
	TotalGroupAccountCreatedEvents           uint64 `json:"total_group_account_created_events"`
	TotalGroupAccountUpdatedEvents           uint64 `json:"total_group_account_updated_events"`
	TotalOpportunityContactRoleCreatedEvents uint64 `json:"total_opportunity_contact_role_created_events"`
	TotalOpportunityContactRoleUpdatedEvents uint64 `json:"total_opportunity_contact_role_updated_events"`
}
type AlertsData struct {
	WebsiteAlerts WebsiteAlerts `json:"website_alerts"`
}
type WebsiteAlerts struct {
	PageViewAlerts PageViewAlerts `json:"page_view_alerts"`
}
type PageViewAlerts struct {
	PageViewsMissingPageTitle                    bool `json:"pvmpt"`
	PageViewsMissingSessionID                    bool `json:"pvmsid"`
	PageViewsHavingInvalidOrMissingPageDomain    bool `json:"pvhimpd"`
	PageViewsHavingInvalidOrMissingPageRawURL    bool `json:"pvhimprurl"`
	PageViewsHavingInvalidOrMissingPageURL       bool `json:"pvhimpurl"`
	PageViewsHavingQPbutMissingQPData            bool `json:"pvhqpmqp"`
	PageViewsHavingInvalidOrMissingPageLoadTime  bool `json:"pvhimplt"`
	PageViewsHavingInvalidOrMissingPageSpentTime bool `json:"pvhimpst"`
}
type PageViewAlertsData struct {
	TotalPageViews                               uint64
	PageViewsMissingPageTitle                    uint64
	PageViewsMissingSessionID                    uint64
	PageViewsHavingInvalidOrMissingPageDomain    uint64
	PageViewsHavingInvalidOrMissingPageRawURL    uint64
	PageViewsHavingInvalidOrMissingPageURL       uint64
	PageViewsHavingQPbutMissingQPData            uint64
	PageViewsHavingInvalidOrMissingPageLoadTime  uint64
	PageViewsHavingInvalidOrMissingPageSpentTime uint64
}
type FacebookAnalyticsData struct {
	IntegrationActive     bool      `json:"integration_active"`
	TotalRecordsProcessed uint64    `json:"total_records_processed"`
	LastSeenTimeOfRecord  time.Time `json:"last_seen_time_of_record"`
}
type LinkedinAnalyticsData struct {
	IntegrationActive     bool      `json:"integration_active"`
	TotalRecordsProcessed uint64    `json:"total_records_processed"`
	LastSeenTimeOfRecord  time.Time `json:"last_seen_time_of_record"`
}

type AnalyticsDataPlaceHolder map[int64]AnalyticsData

var validations = make(map[string]map[string]validationRule)

func main() {

	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	envFlag := flag.String("env", "development", "environment")
	date := time.Now().AddDate(0, 0, -7)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	fromDate := flag.String("date", date.Format(U.DATETIME_FORMAT_YYYYMMDD), "start date of events")
	projectIdFlag := flag.String("project_ids", "", "Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	//	emailString := flag.String("emails", "", "comma separeted list of emails to which report to be sent")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	flag.Parse()
	config := &C.Configuration{
		Env:                 *envFlag,
		AWSKey:              *awsAccessKeyId,
		AWSSecret:           *awsSecretAccessKey,
		AWSRegion:           *awsRegion,
		EmailSender:         *factorsEmailSender,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}
	C.InitConf(config)
	C.InitSenderEmail(C.GetFactorsSenderEmail())
	//C.InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitFilemanager(*bucketName, *envFlag, config)
	//emails := strings.Split(*emailString, ",")
	flag.Parse()

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		var err error
		cloudManager, err = serviceGCS.New(*bucketName)
		if err != nil {
			log.Error("Failed to init New GCS Client")
			panic(err)
		}
	}
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)
	projectIdsList := getProjectIdsList(*projectIdFlag)
	fromTime, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, *fromDate)
	if err != nil {
		log.Fatal("Failed in parsing date. Error = ", err)
	}
	if fromTime.Format(U.DATETIME_FORMAT_YYYYMMDD) > time.Now().AddDate(0, 0, -7).Format(U.DATETIME_FORMAT_YYYYMMDD) {
		log.Fatal("Could not run for given date range.")
	} else {
		for fromTime.Weekday() != time.Sunday {
			fromTime = fromTime.AddDate(0, 0, -1)
		}
		*fromDate = fromTime.Format(U.DATETIME_FORMAT_YYYYMMDD)
	}
	toDate := fromTime.AddDate(0, 0, 6).Format(U.DATETIME_FORMAT_YYYYMMDD)
	analyticsData, err := getAnalyticsDataForAllProjects(projectIdsList)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve cache analytics data")
	}

	for _, project_id := range projectIdsList {

		toTime := fromTime.Unix() + U.Per_week_epoch - 1
		mfCloudPath, mfCloudName := (cloudManager).GetModelMetricsFilePathAndName(int64(project_id), fromTime.Unix(), toTime)
		afCloudPath, afCloudName := (cloudManager).GetModelAlertsFilePathAndName(int64(project_id), fromTime.Unix(), toTime)
		efCloudPath, efCloudName := (cloudManager).GetEventsFilePathAndName(int64(project_id), fromTime.Unix(), toTime)
		efTmpPath, efTmpName := diskManager.GetEventsFilePathAndName(int64(project_id), fromTime.Unix(), toTime)
		log.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
			"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
		eReader, err := (cloudManager).Get(efCloudPath, efCloudName)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
				"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
		}
		log.Info("Successfuly downloaded events file from cloud.", efTmpPath, efTmpName)
		err = diskManager.Create(efTmpPath, efTmpName, eReader)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
				"eventFileName": efCloudName}).Error("Failed creating events file in local.")
		}
		tmpEventsFilePath := efTmpPath + efTmpName

		scanner, err := T.OpenEventFileAndGetScanner(tmpEventsFilePath)
		if err != nil {
			log.WithError(err).Error("Failed to open Events File and Scanner ", tmpEventsFilePath)
		}
		//initizalizing validator and validationMap
		// validate := validator.New()
		// initialize()
		// var reportMap = make(map[string]map[string]bool)
		// var daterange = make(map[string]map[string]bool)
		// var dates = make([]string, 0, 10)
		//	Scanning the file row by row
		var pageViewAlertsData PageViewAlertsData
		var WebsiteData WebsiteAnalyticsData
		var HubspotData HubspotAnalyticsData
		var SalesForceData SalesforceAnalyticsData
		HubspotUniqueMetrics := HubspotUniqueMetrics{
			ActiveHubspotContactsMap:  make(map[string]uint64),
			ActiveHubspotCompaniesMap: make(map[string]uint64),
			ActiveHubspotDeals:        make(map[string]uint64),
			ActiveHubspotEngagements:  make(map[string]uint64),
		}
		for scanner.Scan() {
			row := scanner.Text()
			var data P.CounterEventFormat
			err := json.Unmarshal([]byte(row), &data)
			if err != nil {
				log.WithError(err).Error("Failed to unmarshal event to CounterEventFormat")
			}
			if data.EventName == "$session" {
				UpdateMetricsForWebsiteRelatedData(data, &WebsiteData, &pageViewAlertsData)
			} else if strings.HasPrefix(data.EventName, "$hubspot") {
				UpdateMetricsForHubspotRelatedData(data, &HubspotData, &HubspotUniqueMetrics)
			} else if strings.HasPrefix(data.EventName, "$sf") || strings.HasPrefix(data.EventName, "$sf") {
				UpdateMetricsForSalesforceRelatedData(data, &SalesForceData)
			}

			// eventTimestamp := time.Unix(data.EventTimestamp, 0)
			// weekday := eventTimestamp.Weekday().String()[0:3]
			// dateKey := eventTimestamp.AddDate(0, 0, 0).Format(U.DATETIME_FORMAT_YYYYMMDD)
			// mapKey := fmt.Sprintf("Date:%s, Day:%s ", dateKey, weekday)
			// if daterange[mapKey] == nil {
			// 	daterange[mapKey] = make(map[string]bool)
			// 	for _, event := range ALL_EVENTS {
			// 		daterange[mapKey][event] = false
			// 	}
			// 	dates = append(dates, mapKey)
			// }
			// eventName := data.EventName

			// if validations[eventName] == nil {
			// 	eventName = NO_EVENT
			// } else {
			// 	daterange[mapKey][eventName] = true
			// }

			// //validating data
			// r := reflect.ValueOf(data)
			// for key, value := range validations[eventName] {
			// 	mapKey := fmt.Sprintf("%s-%s", eventName, key)
			// 	if reportMap[mapKey] == nil {
			// 		reportMap[mapKey] = make(map[string]bool)
			// 	}
			// 	var f reflect.Value
			// 	if value.property == "" {
			// 		f = reflect.Indirect(r).FieldByName(value.feild)
			// 	} else if value.property == "UserProperties" || value.property == "EventProperties" {
			// 		f = reflect.Indirect(r).FieldByName(value.property).MapIndex(reflect.ValueOf(value.feild))
			// 	} else {
			// 		log.Info("invalid property")
			// 		continue
			// 	}
			// 	if (!f.IsValid() && value.required) || (f.IsValid() && validate.Var(f.Interface(), value.validate) != nil) {
			// 		reportMap[mapKey][data.UserId] = true
			// 	}
			// }
		}
		//	end of data file

		// Creating Reports
		//writeReport(dates, daterange, reportMap, *fromDate, toDate)
		//writeDetailedReport(reportMap, *fromDate, toDate)
		if anaData, exists := analyticsData[project_id]; exists {
			anaData.WebsiteAnalyticsData = WebsiteData
			pageViewAlertsData.TotalPageViews = anaData.WebsiteAnalyticsData.TotalPageViews
			anaData.HubspotAnalyticsData = HubspotData
			// active hubspot data (UNIQUE)
			anaData.HubspotAnalyticsData.HubspotEventsAnalyticsData.ActiveHubspotContacts = uint64(len(HubspotUniqueMetrics.ActiveHubspotContactsMap))
			anaData.HubspotAnalyticsData.HubspotEventsAnalyticsData.ActiveHubspotCompanies = uint64(len(HubspotUniqueMetrics.ActiveHubspotCompaniesMap))
			anaData.HubspotAnalyticsData.HubspotEventsAnalyticsData.ActiveHubspotDeals = uint64(len(HubspotUniqueMetrics.ActiveHubspotDeals))
			anaData.HubspotAnalyticsData.HubspotEventsAnalyticsData.ActiveHubspotEngagementUsers = uint64(len(HubspotUniqueMetrics.ActiveHubspotEngagements))
			anaData.SalesforceAnalyticsData = SalesForceData
			analyticsData[project_id] = anaData
		}
		var alerts AlertsData
		alerts.WebsiteAlerts.PageViewAlerts = generatePageViewAlertsByData(pageViewAlertsData)
		writeMetricsReport(analyticsData[project_id])
		writeAlertsReport(alerts)
		//uploading reports to cloud
		metricsReport, _ := openFile(MetricsReportName)

		//	detailed_report, _ := openFile(DetailedReportName)
		//log.Info("$$$$cloud path ", efCloudPath)
		err = (cloudManager).Create(mfCloudPath, mfCloudName, metricsReport)
		if err != nil {
			//log.Fatal("Failed to upload report in cloud. error = ", err)
			log.WithError(err).Error("failed to upload metrics report in cloud")
		} else {
			log.Info("success in uploading metrics report to cloud", mfCloudPath, mfCloudName)
		}

		_ = closeFile(metricsReport)
		alertsReport, _ := openFile(AlertsReportName)

		//	detailed_report, _ := openFile(DetailedReportName)
		//log.Info("$$$$cloud path ", efCloudPath)
		err = (cloudManager).Create(afCloudPath, afCloudName, alertsReport)
		if err != nil {
			//log.Fatal("Failed to upload report in cloud. error = ", err)
			log.WithError(err).Error("failed to upload alerts report in cloud")
		} else {
			log.Info("success in uploading alerts report to cloud", mfCloudPath, mfCloudName)
		}
		// err = (cloudManager).Create(efCloudPath, DetailedReportName, detailed_report)
		// if err != nil {
		// 	log.Fatal("Failed to upload detailed_report in cloud. error = ", err)
		// }
		_ = closeFile(alertsReport)
		//	_ = closeFile(detailed_report)
		//reports uploaded and closed

		//sendReportviaEmail(emails, project_id)
		log.Info(fmt.Sprintf("Report written successfully from %s to %s.", *fromDate, toDate))
		//sucess log
	}
}
func UpdateMetricsForWebsiteRelatedData(event P.CounterEventFormat, websiteData *WebsiteAnalyticsData, pageViewAlerts *PageViewAlertsData) {
	websiteData.WebsiteDataActive = true
	websiteData.TotalSessions += 1
	if _, exists := event.EventProperties[U.EP_IS_PAGE_VIEW]; exists {
		if event.EventProperties[U.EP_IS_PAGE_VIEW] == true {
			websiteData.TotalPageViews += 1
			UpdatePageViewAlertsData(event, pageViewAlerts)
		}
	}
	if _, exists := event.EventProperties[U.EP_SOURCE]; exists {
		if event.EventProperties[U.EP_SOURCE] == U.PLATFORM_WEB {
			websiteData.TotalEventsFromWeb += 1
		}
	}

}
func UpdatePageViewAlertsData(event P.CounterEventFormat, pageViewAlertsData *PageViewAlertsData) {
	// page title
	if _, exists := event.EventProperties[U.EP_PAGE_TITLE]; exists {
		if event.EventProperties[U.EP_PAGE_TITLE] == "" {
			pageViewAlertsData.PageViewsMissingPageTitle += 1
		}
	} else {
		pageViewAlertsData.PageViewsMissingPageTitle += 1
	}
	// page domain
	if _, exists := event.EventProperties[U.EP_PAGE_DOMAIN]; exists {
		if event.EventProperties[U.EP_PAGE_DOMAIN] == "" {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageDomain += 1
		}
	} else {
		pageViewAlertsData.PageViewsHavingInvalidOrMissingPageDomain += 1
	}
	// page raw url
	if _, exists := event.EventProperties[U.EP_PAGE_RAW_URL]; exists {
		if event.EventProperties[U.EP_PAGE_RAW_URL] == "" {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageRawURL += 1
		}
	} else {
		pageViewAlertsData.PageViewsHavingInvalidOrMissingPageRawURL += 1
	}
	// page url
	if _, exists := event.EventProperties[U.EP_PAGE_URL]; exists {
		if event.EventProperties[U.EP_PAGE_URL] == "" {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageURL += 1
		}
		//TODO : check if it has $qp

	} else {
		pageViewAlertsData.PageViewsHavingInvalidOrMissingPageURL += 1
	}
	// page load time
	if _, exists := event.EventProperties[U.EP_PAGE_LOAD_TIME]; exists {
		if event.EventProperties[U.EP_PAGE_LOAD_TIME] == "" {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageLoadTime += 1

		} else if event.EventProperties[U.EP_PAGE_LOAD_TIME].(int) <= 0 || event.EventProperties[U.EP_PAGE_LOAD_TIME].(int) > 1800 {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageLoadTime += 1
		}
	} else {
		pageViewAlertsData.PageViewsHavingInvalidOrMissingPageLoadTime += 1
	}
	// page spent time
	if _, exists := event.EventProperties[U.EP_PAGE_SPENT_TIME]; exists {
		if event.EventProperties[U.EP_PAGE_SPENT_TIME] == "" {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageSpentTime += 1

		} else if event.EventProperties[U.EP_PAGE_SPENT_TIME].(int) <= 0 || event.EventProperties[U.EP_PAGE_LOAD_TIME].(int) > 3600 {
			pageViewAlertsData.PageViewsHavingInvalidOrMissingPageSpentTime += 1
		}
	} else {
		pageViewAlertsData.PageViewsHavingInvalidOrMissingPageLoadTime += 1
	}
}
func generatePageViewAlertsByData(pageViewAlertsData PageViewAlertsData) PageViewAlerts {
	var pageViewAlerts PageViewAlerts
	// If more than 10 % of page views has missing $page_title.
	pageViewsThreshold := 0.1 * float64(pageViewAlertsData.TotalPageViews)
	if pageViewAlertsData.PageViewsMissingPageTitle >= uint64(pageViewsThreshold) {
		pageViewAlerts.PageViewsMissingPageTitle = true
	}
	// If more than absolute 10 page view events have invalid or missing $page_domain. (Invalid if it is not the right domain or a full url)
	if pageViewAlertsData.PageViewsHavingInvalidOrMissingPageDomain > 10 {
		pageViewAlerts.PageViewsHavingInvalidOrMissingPageDomain = true
	}
	// If more than absolute 10 page view events have invalid or missing $page_raw_url.
	if pageViewAlertsData.PageViewsHavingInvalidOrMissingPageRawURL > 10 {
		pageViewAlerts.PageViewsHavingInvalidOrMissingPageRawURL = true
	}
	// If more than absolute 10 page view events have invalid or missing $page_url. (Invalid if it is not a url or if the url has query parameters)
	if pageViewAlertsData.PageViewsHavingInvalidOrMissingPageURL > 10 {
		pageViewAlerts.PageViewsHavingInvalidOrMissingPageURL = true
	}
	// If more than absolute 10 page view events have invalid or missing $page_load_time. (Invalid is 0 or negative or greater than 1800)
	if pageViewAlertsData.PageViewsHavingInvalidOrMissingPageLoadTime > 10 {
		pageViewAlerts.PageViewsHavingInvalidOrMissingPageLoadTime = true
	}
	// If more than absolute 10 page view events have invalid or missing $page_spent_time. (Invalid is 0 or negative or greater than 3600)
	if pageViewAlertsData.PageViewsHavingInvalidOrMissingPageSpentTime > 10 {
		pageViewAlerts.PageViewsHavingInvalidOrMissingPageSpentTime = true
	}
	return pageViewAlerts
}
func UpdateMetricsForHubspotRelatedData(event P.CounterEventFormat, hubspotData *HubspotAnalyticsData, HubspotUniqueMetrics *HubspotUniqueMetrics) {
	hubspotData.IntegrationActive = true
	if strings.HasPrefix(event.EventName, "$hubspot_contact_list") {
		hubspotData.HubspotAnalyticsSyncData.TotalContactListRecordsSynced += 1
		// checking suffix to determine action
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalContactListCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalContactListDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalContactListUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$hubspot_contact") {
		hubspotData.HubspotAnalyticsSyncData.TotalContactRecordsSynced += 1
		HubspotUniqueMetrics.ActiveHubspotContactsMap[event.UserId] += 1
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalContactCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalContactDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalContactUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$hubspot_company") {
		hubspotData.HubspotAnalyticsSyncData.TotalCompanyRecordsSynced += 1
		HubspotUniqueMetrics.ActiveHubspotCompaniesMap[event.UserId] += 1
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalCompanyCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalCompanyDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalCompanyUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$hubspot_engagement") {
		hubspotData.HubspotAnalyticsSyncData.TotalEngagementRecordsSynced += 1
		HubspotUniqueMetrics.ActiveHubspotEngagements[event.UserId] += 1
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalEngagementCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalEngagementDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalEngagementUpdatedEvents += 1

		}
	} else if strings.HasPrefix(event.EventName, "$hubspot_form_submission") {
		hubspotData.HubspotAnalyticsSyncData.TotalFormSubmissionRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalFormSubmissionCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalFormSubmissionDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalFormSubmissionUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$hubspot_form") {
		hubspotData.HubspotAnalyticsSyncData.TotalFormRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalFormCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalFormDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalFormUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$hubspot_deal") {
		hubspotData.HubspotAnalyticsSyncData.TotalDealRecordsSynced += 1
		HubspotUniqueMetrics.ActiveHubspotDeals[event.UserId] += 1
		if strings.HasSuffix(event.EventName, "created") {
			hubspotData.HubspotEventsAnalyticsData.TotalDealCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "deleted") {
			hubspotData.HubspotEventsAnalyticsData.TotalDealDeletedEvents += 1
		} else if strings.HasSuffix(event.EventName, "associations_updated") {

		} else if strings.HasSuffix(event.EventName, "updated") {
			hubspotData.HubspotEventsAnalyticsData.TotalDealUpdatedEvents += 1
		}
	}
}
func UpdateMetricsForSalesforceRelatedData(event P.CounterEventFormat, salesforceData *SalesforceAnalyticsData) {
	salesforceData.IntegrationActive = true
	if strings.HasPrefix(event.EventName, "$salesforce_contact") || strings.HasPrefix(event.EventName, "$sf_contact") {
		salesforceData.SalesforceAnalyticsSyncData.TotalContactRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalContactCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalContactUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_lead") || strings.HasPrefix(event.EventName, "$sf_lead") {
		salesforceData.SalesforceAnalyticsSyncData.TotalLeadRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalLeadCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalLeadUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_account") || strings.HasPrefix(event.EventName, "$sf_account") {
		salesforceData.SalesforceAnalyticsSyncData.TotalAccountRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalAccountCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalAccountUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_opportunity_contact_role") || strings.HasPrefix(event.EventName, "$sf_opportunity_contact_role") {
		salesforceData.SalesforceAnalyticsSyncData.TotalOpportunityContactRoleSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalOpportunityContactRoleCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalOpportunityContactRoleUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_opportunity") || strings.HasPrefix(event.EventName, "$sf_opportunity") {
		salesforceData.SalesforceAnalyticsSyncData.TotalOpportunityRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalOpportunityCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalOpportunityUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_campaign_member") || strings.HasPrefix(event.EventName, "$sf_campaign_member") {
		salesforceData.SalesforceAnalyticsSyncData.TotalCampaignMemberRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalCampaignMemberCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalCampaignMemberUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_campaign") || strings.HasPrefix(event.EventName, "$sf_campaign") {
		salesforceData.SalesforceAnalyticsSyncData.TotalCampaignRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalCampaignCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalCampaignUpdatedEvents += 1
		}
	} else if strings.HasPrefix(event.EventName, "$salesforce_group_account") || strings.HasPrefix(event.EventName, "$sf_group_account") {
		salesforceData.SalesforceAnalyticsSyncData.TotalGroupAccountRecordsSynced += 1
		if strings.HasSuffix(event.EventName, "created") {
			salesforceData.SalesforceEventAnalyticsData.TotalGroupAccountCreatedEvents += 1
		} else if strings.HasSuffix(event.EventName, "updated") {
			salesforceData.SalesforceEventAnalyticsData.TotalGroupAccountUpdatedEvents += 1
		}
	}

}
func initialize() {
	initializeSession()
	initializeNoEvent()
	initializeHubspot_Created()
	initializeHubspot_Updated()
	initializeSalesforce_Created()
	initializeSalesforce_Updated()
}
func initializeSession() {
	validations[U.EVENT_NAME_SESSION] = make(map[string]validationRule)
	validations[U.EVENT_NAME_SESSION]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[U.EVENT_NAME_SESSION]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[U.EVENT_NAME_SESSION]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[U.EVENT_NAME_SESSION]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}
	validations[U.EVENT_NAME_SESSION]["$session_count"] = validationRule{"$session_count", "UserProperties", false, "required,gt=0"}

	validations[U.EVENT_NAME_SESSION]["$page_count"] = validationRule{"$page_count", "EventProperties", true, "required,min=1"}
	validations[U.EVENT_NAME_SESSION]["$country"] = validationRule{"$country", "EventProperties", true, "required,min=1"}
	validations[U.EVENT_NAME_SESSION]["$initial_page_scroll_percent"] = validationRule{"$initial_page_scroll_percent", "EventProperties", true, "required,gt=0"}
	validations[U.EVENT_NAME_SESSION]["$session_spent_time"] = validationRule{"$session_spent_time", "EventProperties", true, "required,gt=0"}
}
func initializeNoEvent() {
	validations[NO_EVENT] = make(map[string]validationRule)
	validations[NO_EVENT]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[NO_EVENT]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[NO_EVENT]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[NO_EVENT]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

	validations[NO_EVENT]["$page_spent_time"] = validationRule{"$page_spent_time", "EventProperties", false, "required,gt=0"}
}
func initializeHubspot_Created() {
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_CREATED] = make(map[string]validationRule)
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_CREATED]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_CREATED]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_CREATED]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_CREATED]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func initializeHubspot_Updated() {
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED] = make(map[string]validationRule)
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func initializeSalesforce_Created() {
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_CREATED] = make(map[string]validationRule)
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_CREATED]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_CREATED]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_CREATED]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_CREATED]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func initializeSalesforce_Updated() {
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED] = make(map[string]validationRule)
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func getAnalyticsDataForAllProjects(projectIdsList []int64) (AnalyticsDataPlaceHolder map[int64]AnalyticsData, err error) {
	// data from cache
	analytics, err := store.GetStore().GetEventUserCountsMerged(projectIdsList, LookbackDaysForCacheData, time.Now())
	if err != nil {
		log.WithError(err).Error("Failed to get project analytics data")
	}
	mergedData := make(map[int64]AnalyticsData)
	// merging last 7 days data into one
	for projectID, data := range analytics {
		var analyticsData AnalyticsData
		globalData := GlobalLevelData{
			TotalActiveUsers:      data.TotalUniqueUsers,
			TotalEvents:           data.TotalEvents,
			TotalRecordsProcessed: getTotalRecordsProcessed(data),
		}
		analyticsData.GlobalLevelData = globalData
		mergedData[projectID] = analyticsData
	}
	return mergedData, nil
}
func getTotalRecordsProcessed(data *model.ProjectAnalytics) uint64 {
	return data.AdwordsEvents + data.FacebookEvents + data.HubspotEvents + data.LinkedinEvents + data.SalesforceEvents
}
func writeReport(dates []string, daterange, reportMap map[string]map[string]bool, from string, to string) {
	report, _ := createFile(MetricsReportName)
	report.WriteString(fmt.Sprintf("Report from %s to %s.\n\n", from, to))
	for key, value := range reportMap {
		report.WriteString(fmt.Sprintf("Total invalid %s : %s \n", key, strconv.Itoa(int(len(value)))))
	}
	sort.Strings(dates)
	report.WriteString("\nEvents for Date Ranges :\n")
	for _, key := range dates {
		report.WriteString(key + " : ")
		value := daterange[key]
		s := ""
		flag := false
		for event, present := range value {
			if present {
				flag = true
				s += event
				s += ", "
			}
		}
		if !flag {
			report.WriteString("None\n")
			continue
		}
		report.WriteString(s[0:len(s)-2] + "\n")
	}
	report.WriteString("\nMissing events for Date Ranges :\n")
	for _, key := range dates {
		report.WriteString(key + " : ")
		value := daterange[key]
		s := ""
		flag := false
		for event, present := range value {
			if !present {
				flag = true
				s += event
				s += ", "
			}
		}
		if !flag {
			report.WriteString("None\n")
			continue
		}
		report.WriteString(s[0:len(s)-2] + "\n")
	}
	_ = closeFile(report)
}
func writeMetricsReport(analyticsData AnalyticsData) {
	report, _ := createFile(MetricsReportName)
	analyticsDataJson, err := json.Marshal(analyticsData)
	if err != nil {
		log.WithError(err).Error("Failed to marshal analytics data")
	}
	analyticsDataString := string(analyticsDataJson)
	report.WriteString(analyticsDataString)
}
func writeAlertsReport(alerts AlertsData) {
	report, _ := createFile(AlertsReportName)
	alertsDataJson, err := json.Marshal(alerts)
	if err != nil {
		log.WithError(err).Error("Failed to marshal alerts data")
	}
	alertsDataString := string(alertsDataJson)
	report.WriteString(alertsDataString)
}
func writeDetailedReport(reportMap map[string]map[string]bool, from string, to string) {
	detailed_report, _ := createFile(DetailedReportName)
	detailed_report.WriteString(fmt.Sprintf("Report from %s to %s.\n\n", from, to))
	for key, value := range reportMap {
		if len(value) != 0 {
			detailed_report.WriteString("Invalid " + key + ":\n")
			for uid := range value {
				detailed_report.WriteString(uid + "\n")
			}
			detailed_report.WriteString("\n")
		}
	}
	_ = closeFile(detailed_report)
}

func getProjectIdsList(projectIdsString string) (projectIdsList []int64) {
	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(projectIdsString, "")
	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("Failed to get all projects and project_ids set to '*'.")
		}

		projectIdsToRun = make(map[int64]bool, 0)
		for _, projectID := range projectIDs {
			projectIdsToRun[projectID] = true
		}
	}
	projectIdsList = make([]int64, 0)
	for projectId := range projectIdsToRun {
		projectIdsList = append(projectIdsList, projectId)
	}
	return projectIdsList
}

func openFile(fileName string) (*os.File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("failed to open file : ", fileName, err)
	}
	return file, err
}

func closeFile(fileName *os.File) error {
	err := fileName.Close()
	if err != nil {
		log.Fatal("failed to close file : ", fileName, err)
	}
	return err
}

func createFile(fileName string) (*os.File, error) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Failed to create OS file : ", fileName, err)
	}
	return f, err
}

func sendReportviaEmail(emails []string, project_id int64) {
	var success, fail int
	report, err := openFile(MetricsReportName)
	if err != nil {
		log.Error(err)
		return
	}
	scanner := bufio.NewScanner(report)
	// send the report as mail
	var data []string
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	html := getTemplateForReport(data)
	email_subject := fmt.Sprintf("Factors Report for Project Id %v", project_id)
	for _, email := range emails {
		err = C.GetServices().Mailer.SendMail(email, C.GetFactorsSenderEmail(), email_subject, html, "")
		if err != nil {
			fail++
			log.WithError(err).Error("failed to send email for project_events")
			continue
		}
		success++
	}
	defer report.Close()
	log.Info("Report sent successfully successfully to ", success, "emails and failed to ", fail, "emails")
}

func getTemplateForReport(data []string) string {
	var html string
	html = "<html><body>"
	for _, line := range data {
		html += line
		html += "<br>"
	}
	html += "</body></html>"
	return html
}
