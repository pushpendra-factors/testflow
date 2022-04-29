package main

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	T "factors/task"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const NO_EVENT = "NoEvent"
const ReportName = "report.txt"
const DetailedReportName = "detailed_report.txt"

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
	modelType := flag.String("modelType", "w", "Model Type of events")
	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	emailString := flag.String("emails", "", "comma separeted list of emails to which report to be sent")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	flag.Parse()
	config := &C.Configuration{
		Env:         *envFlag,
		AWSKey:      *awsAccessKeyId,
		AWSSecret:   *awsSecretAccessKey,
		AWSRegion:   *awsRegion,
		EmailSender: *factorsEmailSender,
	}
	C.InitConf(config)
	C.InitSenderEmail(C.GetFactorsSenderEmail())
	C.InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	emails := strings.Split(*emailString, ",")
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

	for _, project_id := range projectIdsList {

		efCloudPath, efCloudName := (cloudManager).GetModelEventsFilePathAndName(uint64(project_id), fromTime.Unix(), *modelType)
		efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(uint64(project_id), fromTime.Unix(), *modelType)
		log.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
			"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
		eReader, err := (cloudManager).Get(efCloudPath, efCloudName)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
				"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
		}
		err = diskManager.Create(efTmpPath, efTmpName, eReader)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
				"eventFileName": efCloudName}).Error("Failed creating events file in local.")
		}
		tmpEventsFilePath := efTmpPath + efTmpName
		log.Info("Successfuly downloaded events file from cloud.", tmpEventsFilePath, efTmpPath, efTmpName)
		scanner, err := T.OpenEventFileAndGetScanner(tmpEventsFilePath)
		//initizalizing validator and validationMap
		validate := validator.New()
		initialize()
		var reportMap = make(map[string]map[string]bool)
		var daterange = make(map[string]map[string]bool)
		var dates = make([]string, 0, 10)
		//Scanning the file row by row
		for scanner.Scan() {
			row := scanner.Text()
			var data P.CounterEventFormat
			json.Unmarshal([]byte(row), &data)
			eventTimestamp := time.Unix(data.EventTimestamp, 0)
			weekday := eventTimestamp.Weekday().String()[0:3]
			dateKey := eventTimestamp.AddDate(0, 0, 0).Format(U.DATETIME_FORMAT_YYYYMMDD)
			mapKey := fmt.Sprintf("Date:%s, Day:%s ", dateKey, weekday)
			if daterange[mapKey] == nil {
				daterange[mapKey] = make(map[string]bool)
				for _, event := range ALL_EVENTS {
					daterange[mapKey][event] = false
				}
				dates = append(dates, mapKey)
			}
			eventName := data.EventName

			if validations[eventName] == nil {
				eventName = NO_EVENT
			} else {
				daterange[mapKey][eventName] = true
			}

			//validating data
			r := reflect.ValueOf(data)
			for key, value := range validations[eventName] {
				mapKey := fmt.Sprintf("%s-%s", eventName, key)
				if reportMap[mapKey] == nil {
					reportMap[mapKey] = make(map[string]bool)
				}
				var f reflect.Value
				if value.property == "" {
					f = reflect.Indirect(r).FieldByName(value.feild)
				} else if value.property == "UserProperties" || value.property == "EventProperties" {
					f = reflect.Indirect(r).FieldByName(value.property).MapIndex(reflect.ValueOf(value.feild))
				} else {
					log.Info("invalid property")
					continue
				}
				if (!f.IsValid() && value.required) || (f.IsValid() && validate.Var(f.Interface(), value.validate) != nil) {
					reportMap[mapKey][data.UserId] = true
				}
			}
		}
		//end of data file

		// Creating Reports
		writeReport(dates, daterange, reportMap, *fromDate, toDate)
		writeDetailedReport(reportMap, *fromDate, toDate)

		//uploading reports to cloud
		report, _ := openFile(ReportName)
		detailed_report, _ := openFile(DetailedReportName)
		err = (cloudManager).Create(efCloudPath, ReportName, report)
		if err != nil {
			log.Fatal("Failed to upload report in cloud. error = ", err)
		}
		err = (cloudManager).Create(efCloudPath, DetailedReportName, detailed_report)
		if err != nil {
			log.Fatal("Failed to upload detailed_report in cloud. error = ", err)
		}
		_ = closeFile(report)
		_ = closeFile(detailed_report)
		//reports uploaded and closed

		sendReportviaEmail(emails, project_id)
		log.Info(fmt.Sprintf("Report written successfully from %s to %s.", *fromDate, toDate))
		//sucess log
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

func writeReport(dates []string, daterange, reportMap map[string]map[string]bool, from string, to string) {
	report, _ := createFile(ReportName)
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

func getProjectIdsList(projectIdsString string) (projectIdsList []uint64) {
	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(projectIdsString, "")
	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("Failed to get all projects and project_ids set to '*'.")
		}

		projectIdsToRun = make(map[uint64]bool, 0)
		for _, projectID := range projectIDs {
			projectIdsToRun[projectID] = true
		}
	}
	projectIdsList = make([]uint64, 0)
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

func sendReportviaEmail(emails []string, project_id uint64) {
	var success, fail int
	report, err := openFile(ReportName)
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
