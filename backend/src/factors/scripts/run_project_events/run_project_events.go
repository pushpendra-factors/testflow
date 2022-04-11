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
	"io"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const Session = "$session"
const Hubspot_created = "$hubspot_contact_created"
const Hubspot_updated = "$hubspot_contact_updated"
const Sf_created = "$sf_contact_created"
const Sf_updated = "$sf_contact_updated"
const NoEvent = "NoEvent"
const ReportName = "report.txt"
const DetailedReportName = "detailed_report.txt"

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
	var cloudManager filestore.FileManager

	//Taking flag inputs
	date := time.Now().AddDate(0, 0, -7)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	fromDate := flag.String("date", date.Format(U.DATETIME_FORMAT_YYYYMMDD), "start date of events")
	projectIdFlag := flag.String("project_ids", "", "Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	modelType := flag.String("modelType", "w", "Model Type of events")
	flag.Parse()

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
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

	projectIdsArray := make([]uint64, 0)
	for projectId := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}
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

	//fileDir and fileName
	startTimeStamp := fromTime.Unix()
	for _, project_id := range projectIdsArray {

		fileDir, fileName := (cloudManager).GetModelEventsFilePathAndName(uint64(project_id), startTimeStamp, *modelType)
		reader, err := (cloudManager).Get(fileDir, fileName)
		if err != nil {
			log.Fatal("Failed to get data from file path. error = ", err)
		}
		dataFile, _ := createFile(fileName)
		//downloading file
		if _, err := io.Copy(dataFile, reader); err != nil {
			log.Fatal("Failed to download data at io.Copy. error = ", err)
		}
		_ = closeFile(dataFile)
		log.Info("data download sucessful...")
		//initizalizing validator and validationMap
		file, _ := openFile(fileName)
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		validate := validator.New()
		initialize()
		var reportMap = make(map[string]map[string]bool)
		var daterange = make(map[string]map[string]bool)
		var dates = make([]string, 0, 10)
		All_Events := []string{Session, Hubspot_created, Hubspot_updated, Sf_created, Sf_updated}
		//Scanning the file row by row
		for scanner.Scan() {
			row := []byte(scanner.Text())
			var data P.CounterEventFormat
			json.Unmarshal(row, &data)
			time := time.Unix(data.EventTimestamp, 0)
			weekday := time.Weekday().String()[0:3]
			dateKey := time.AddDate(0, 0, 0).Format(U.DATETIME_FORMAT_YYYYMMDD)
			if dateKey > toDate {
				continue
			}
			mapKey := fmt.Sprintf("Date:%s, Day:%s ", dateKey, weekday)
			if daterange[mapKey] == nil {
				daterange[mapKey] = make(map[string]bool)
				for _, event := range All_Events {
					daterange[mapKey][event] = false
				}
				dates = append(dates, mapKey)
			}
			eventName := data.EventName
			if validations[eventName] == nil {
				eventName = "NoEvent"
			} else {
				daterange[mapKey][eventName] = true
			}
			r := reflect.ValueOf(data)
			for key, value := range validations[eventName] {
				if reportMap[key] == nil {
					reportMap[key] = make(map[string]bool)
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
					reportMap[key][data.UserId] = true
				}
			}
		}
		//end of data file

		// Creating Reports
		var report, detailed_report *os.File
		report, _ = createFile(ReportName)
		detailed_report, _ = createFile(DetailedReportName)
		report.WriteString(fmt.Sprintf("Report from %s to %s.\n\n", *fromDate, toDate))
		detailed_report.WriteString(fmt.Sprintf("Report from %s to %s.\n\n", *fromDate, toDate))

		for key, value := range reportMap {
			report.WriteString(fmt.Sprintf("Total invalid %s : %s \n", key, strconv.Itoa(int(len(value)))))
			if len(value) != 0 {
				detailed_report.WriteString("Invalid " + key + ":\n")
				for uid := range value {
					detailed_report.WriteString(uid + "\n")
				}
				detailed_report.WriteString("\n")
			}
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

		//uploading reports to cloud
		_ = closeFile(report)
		_ = closeFile(detailed_report)
		report, _ = openFile(ReportName)
		detailed_report, _ = openFile(DetailedReportName)
		err = (cloudManager).Create(fileDir, ReportName, report)
		if err != nil {
			log.Fatal("Failed to upload report in cloud. error = ", err)
		}
		err = (cloudManager).Create(fileDir, DetailedReportName, detailed_report)
		if err != nil {
			log.Fatal("Failed to upload detailed_report in cloud. error = ", err)
		}
		//reports uploaded
		//closing reports
		_ = closeFile(report)
		_ = closeFile(detailed_report)
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
	validations[Session] = make(map[string]validationRule)
	validations[Session]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[Session]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[Session]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[Session]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

	validations[Session]["$page_count"] = validationRule{"$page_count", "EventProperties", true, "required,min=1"}
	validations[Session]["$country"] = validationRule{"$country", "EventProperties", true, "required,min=1"}
	validations[Session]["$initial_page_scroll_percent"] = validationRule{"$initial_page_scroll_percent", "EventProperties", true, "required,gt=0"}
	validations[Session]["$session_spent_time"] = validationRule{"$session_spent_time", "EventProperties", true, "required,gt=0"}
}
func initializeNoEvent() {
	validations[NoEvent] = make(map[string]validationRule)
	validations[NoEvent]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[NoEvent]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[NoEvent]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[NoEvent]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

	validations[NoEvent]["$page_spent_time"] = validationRule{"$page_spent_time", "EventProperties", false, "required,gt=0"}
	validations[NoEvent]["$session_count"] = validationRule{"$session_count", "UserProperties", true, "required,gt=0"}
}
func initializeHubspot_Created() {
	validations[Hubspot_created] = make(map[string]validationRule)
	validations[Hubspot_created]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[Hubspot_created]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[Hubspot_created]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[Hubspot_created]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func initializeHubspot_Updated() {
	validations[Hubspot_updated] = make(map[string]validationRule)
	validations[Hubspot_updated]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[Hubspot_updated]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[Hubspot_updated]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[Hubspot_updated]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func initializeSalesforce_Created() {
	validations[Sf_created] = make(map[string]validationRule)
	validations[Sf_created]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[Sf_created]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[Sf_created]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[Sf_created]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

}
func initializeSalesforce_Updated() {
	validations[Sf_updated] = make(map[string]validationRule)
	validations[Sf_updated]["UserId"] = validationRule{"UserId", "", true, "required"}
	validations[Sf_updated]["EventName"] = validationRule{"EventName", "", true, "required"}
	validations[Sf_updated]["UserJoinTimestamp"] = validationRule{"UserJoinTimestamp", "", true, "required,gt=0"}
	validations[Sf_updated]["EventTimestamp"] = validationRule{"EventTimestamp", "", true, "required,gt=0"}

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
