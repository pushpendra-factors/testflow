package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model/model"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// temp structs just for this script
type Event struct {
	// Composite primary key with project_id and uuid.
	ID              string  `gorm:"primary_key:true;type:uuid" json:"id"`
	CustomerEventId *string `json:"customer_event_id"`

	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	// (project_id, event_name_id) -> event_names(project_id, id)
	ProjectId int64  `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `json:"user_id"`

	UserProperties *postgres.Jsonb `json:"user_properties"`
	SessionId      *string         `json:session_id`
	EventNameId    int64           `json:"event_name_id"`
	Count          uint64          `json:"count"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties                 postgres.Jsonb `json:"properties,omitempty"`
	PropertiesUpdatedTimestamp int64          `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
	// unix epoch timestamp in seconds.
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type EventName struct {
	// Composite primary key with projectId.
	ID   int64  `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId int64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"	not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var jobMap = map[string][]string{
	"dashboard_caching":          []string{"event_names", "users", "events", "dashboards", "dashboard_units"},
	"add_session":                []string{"event_names", "users", "events"},
	"hubspot_enrich":             []string{"hubspot_documents", "event_names", "project_settings"},
	"salesforce_enrich":          []string{"salesforce_documents", "event_names", "project_settings"},
	"web_analytics_dashboard":    []string{"event_names", "users", "events"},
	"explain_feature_testing_ui": []string{"event_names", "users", "events"},
	"attribution":                []string{"event_names", "users", "events", "project_settings", "adwords_documents", "facebook_documents", "linkedin_documents", "smart_property_rules", "smart_properties"},
	"test":                       []string{"adwords_documents", "facebook_documents", "linkedin_documents"},
}
var projectIdStaging *int64
var folderName *string

// to map eventName id to the generated id
var eventNamesIDMap = make(map[int64]int64)

// to map userId to the generated id
var userIdMap = make(map[string]string)

// to map dashBoardId to the generated dashboard ID
var dashboardIdMap = make(map[int64]int64)

type agentUUID struct {
	AgentUUID string `json:"agent_uuid"`
}
type dashboardID struct {
	DashboardId uint64 `json:"id"`
}
type eventName struct {
	EventName string `json:"name"`
}
type Result struct {
	Id int64 `json:"id"`
}

var numberOfLinesCopied int = 0
var numberOfLinesFailed int = 0
var currDashboardId int64 = 0
var currEventNameId int64 = 0

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")
	// table name to which data is to be pushed
	jobConfig := flag.String("jobConfig", "default", "please enter a job config type")

	// project id to be written in staging table.
	projectIdStaging = flag.Int64("projectIdStaging", 000000, "specify project id to be written in staging")

	folderName = flag.String("folderName", "default", "Enter the folder name which contains records text files")
	flag.Parse()

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	if _, exists := jobMap[string(*jobConfig)]; !exists {
		log.Error("job type " + *jobConfig + " does not exist")
		return
	}
	if *projectIdStaging == 000000 {
		log.Error("please specify the project_id to be inserted in staging")
		return
	}
	if *folderName == "default" {
		log.Error("folder " + *folderName + " does not exist")
		return
	}
	if !(*env == "staging" || *env == "development") {
		log.Error("this can be run only in staging or development")
		return
	}
	currentJobConfig := jobMap[*jobConfig]
	if *env != "development" {
		for _, fileName := range currentJobConfig {
			err := GetFilesFromCloud(fmt.Sprintf("%v/%v/%v/%v", "transfer_data", *projectIdStaging, *folderName, *folderName), fmt.Sprintf("%v.txt", fileName), *folderName)
			if err != nil {
				log.Info(err)
			}
		}
	}
	// generating dashboard id
	row, err := db.Raw("Select MAX(ID) AS id from dashboards").Rows()
	if err != nil {
		log.Error(err)
	}
	for row.Next() {
		var dashboardID Result
		if err := db.ScanRows(row, &dashboardID); err != nil {
			log.Println(err)
		}
		currDashboardId = dashboardID.Id + 1
		// creating new dashboardId by adding 1 to max id present
	}

	//...............................//

	// generating new event names id
	row, err = db.Raw("Select MAX(ID) AS id from event_names").Rows()
	if err != nil {
		log.Error(err)
	}
	for row.Next() {
		var eventNameID Result
		if err := db.ScanRows(row, &eventNameID); err != nil {
			log.Println(err)

		}
		currEventNameId = eventNameID.Id + 1
		// creating new eventNameId by adding 1 to max id present
	}
	//...............................//
	writeData(currentJobConfig, *jobConfig)
}

func writeData(tables []string, jobType string) {
	db := C.GetServices().Db
	defer db.Close()
	for _, tableNameValue := range tables {
		numberOfLinesCopied = 0
		numberOfLinesFailed = 0
		path := *folderName + "/" + tableNameValue + ".txt"

		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return
		}
		sc := bufio.NewScanner(file)
		for sc.Scan() {
			recordToBeInserted := sc.Text() // GET the line string
			var err error
			if tableNameValue == "event_names" {
				eventNameObj := getEventNamesObj(recordToBeInserted)
				if jobType == "add_session" {
					if eventNameObj.Name == "$session" {
						continue
					}
				}
				if jobType == "salesforce_enrich" {
					if !strings.HasPrefix(eventNameObj.Name, ("$sf")) {
						continue
					}
				}
				if jobType == "hubspot_enrich" {
					if !strings.HasPrefix(eventNameObj.Name, ("$hubspot")) {
						continue
					}
				}
				err = writeDataEventNames(eventNameObj, db)
			} else if tableNameValue == "users" {
				usersObj := getUsersObj(recordToBeInserted)
				err = writeDataUsers(usersObj, db)
			} else if tableNameValue == "events" {
				eventObj := getEventObj(recordToBeInserted)
				err = writeDataEvents(eventObj, db)
			} else if tableNameValue == "project_settings" {
				projectSettingsObj := getProjectSettingsObj(recordToBeInserted)
				err = writeDataProjectSettings(projectSettingsObj, db)
			} else if tableNameValue == "adwords_documents" {
				adwordsObj := getAdwordsObj(recordToBeInserted)
				err = writeDataAdwords(adwordsObj, db)
			} else if tableNameValue == "facebook_documents" {
				facebookObj := getFacebookObj(recordToBeInserted)
				err = writeDataFacebook(facebookObj, db)
			} else if tableNameValue == "linkedin_documents" {
				linkedinObj := getLinkedinObj(recordToBeInserted)
				err = writeDataLinkedin(linkedinObj, db)
			} else if tableNameValue == "smart_property_rules" {
				smartPropertyRulesObj := getSmartPropertyRuleObj(recordToBeInserted)
				err = writeDataSmartPropertyRules(smartPropertyRulesObj, db)
			} else if tableNameValue == "smart_properties" {
				smartPropertyObj := getSmartPropertyObj(recordToBeInserted)
				err = writeDataSmartProperties(smartPropertyObj, db)
			} else if tableNameValue == "salesforce_documents" {
				salesforceObj := getSalesforceObj(recordToBeInserted)
				err = writeDataSalesforce(salesforceObj, db)
			} else if tableNameValue == "hubspot_documents" {
				hubspotObj := getHubspotObj(recordToBeInserted)
				err = writeDataHubspot(hubspotObj, db)
			} else if tableNameValue == "dashboards" {
				dashboardObj := getDashBoardObj(recordToBeInserted, db)
				err = writeDataDashboards(dashboardObj, db)
			} else if tableNameValue == "dashboard_units" {
				dashboardUnitsObj := getDashBoardUnitObj(recordToBeInserted)
				err = writeDataDashboardUnits(dashboardUnitsObj, db)
			}
			if err != nil {
				numberOfLinesFailed++
			} else {
				numberOfLinesCopied++
			}
		}
		log.Println(
			"succesfully inserted data into " + tableNameValue + " total number of records copied : " + strconv.Itoa(numberOfLinesCopied) + " total number of records failed : " + strconv.Itoa(numberOfLinesFailed))

	}
}

// returns the record object from string fetched from the text file.
func getEventNamesObj(recordToBeInserted string) EventName {
	var eventnamesObj EventName
	json.Unmarshal([]byte(recordToBeInserted), &eventnamesObj)
	eventnamesObj.ProjectId = *projectIdStaging
	eventNamesIDMap[eventnamesObj.ID] = currEventNameId
	eventnamesObj.ID = currEventNameId
	currEventNameId++
	return eventnamesObj
}
func getUsersObj(recordToBeInserted string) M.User {
	var usersObj M.User
	json.Unmarshal([]byte(recordToBeInserted), &usersObj)
	usersObj.ProjectId = *projectIdStaging
	generatedid := uuid.New().String()
	userIdMap[usersObj.ID] = generatedid
	usersObj.ID = generatedid
	return usersObj
}
func getEventObj(recordToBeInserted string) Event {
	var eventObj Event
	json.Unmarshal([]byte(recordToBeInserted), &eventObj)
	eventObj.ID = uuid.New().String()
	eventObj.ProjectId = *projectIdStaging
	eventObj.EventNameId = eventNamesIDMap[eventObj.EventNameId]
	eventObj.UserId = userIdMap[eventObj.UserId]
	return eventObj
}
func getDashBoardObj(recordToBeInserted string, db *gorm.DB) M.Dashboard {
	var dashboardObj M.Dashboard
	json.Unmarshal([]byte(recordToBeInserted), &dashboardObj)
	row, err := db.Raw("SELECT agent_uuid from project_agent_mappings WHERE project_id = " + strconv.Itoa(int(dashboardObj.ProjectId)) + " LIMIT 1").Rows()
	if err != nil {
		log.Error(err)
	}
	var agentUuid agentUUID
	for row.Next() {
		if err := db.ScanRows(row, &agentUuid); err != nil {
			log.Println(err)
		}
	}
	dashboardObj.ProjectId = *projectIdStaging
	dashboardObj.AgentUUID = agentUuid.AgentUUID
	dashboardIdMap[dashboardObj.ID] = currDashboardId
	dashboardObj.ID = currDashboardId
	currDashboardId++
	return dashboardObj
}
func getDashBoardUnitObj(recordToBeInserted string) M.DashboardUnit {
	var dashboardUnitsObj M.DashboardUnit
	json.Unmarshal([]byte(recordToBeInserted), &dashboardUnitsObj)
	dashboardUnitsObj.ProjectID = *projectIdStaging
	dashboardUnitsObj.ID = dashboardIdMap[dashboardUnitsObj.ID]
	return dashboardUnitsObj
}
func getHubspotObj(recordToBeInserted string) M.HubspotDocument {
	var hubspotObj M.HubspotDocument
	json.Unmarshal([]byte(recordToBeInserted), &hubspotObj)
	hubspotObj.ProjectId = *projectIdStaging
	hubspotObj.Synced = false
	return hubspotObj
}
func getSalesforceObj(recordToBeInserted string) M.SalesforceDocument {
	var salesforceObj M.SalesforceDocument
	json.Unmarshal([]byte(recordToBeInserted), &salesforceObj)
	salesforceObj.Synced = false
	salesforceObj.ProjectID = *projectIdStaging
	return salesforceObj
}
func getProjectSettingsObj(recordToBeInserted string) M.ProjectSetting {
	var projectSettingObj M.ProjectSetting
	json.Unmarshal([]byte(recordToBeInserted), &projectSettingObj)
	projectSettingObj.ProjectId = *projectIdStaging
	return projectSettingObj
}
func getAdwordsObj(recordToBeInserted string) M.AdwordsDocument {
	var adwordsObj M.AdwordsDocument
	json.Unmarshal([]byte(recordToBeInserted), &adwordsObj)
	adwordsObj.ProjectID = *projectIdStaging
	return adwordsObj
}
func getFacebookObj(recordToBeInserted string) M.FacebookDocument {
	var facebookObj M.FacebookDocument
	json.Unmarshal([]byte(recordToBeInserted), &facebookObj)
	facebookObj.ProjectID = *projectIdStaging
	return facebookObj
}
func getLinkedinObj(recordToBeInserted string) M.LinkedinDocument {
	var linkedinObj M.LinkedinDocument
	json.Unmarshal([]byte(recordToBeInserted), &linkedinObj)
	linkedinObj.ProjectID = *projectIdStaging
	return linkedinObj
}
func getSmartPropertyRuleObj(recordToBeInserted string) M.SmartPropertyRules {
	var smartPropertyRulesObj M.SmartPropertyRules
	json.Unmarshal([]byte(recordToBeInserted), &smartPropertyRulesObj)
	smartPropertyRulesObj.ProjectID = *projectIdStaging
	return smartPropertyRulesObj
}
func getSmartPropertyObj(recordToBeInserted string) M.SmartProperties {
	var smartPropertyObj M.SmartProperties
	json.Unmarshal([]byte(recordToBeInserted), &smartPropertyObj)
	smartPropertyObj.ProjectID = *projectIdStaging
	return smartPropertyObj
}

// pushes data to db tables when invoked.
func writeDataEventNames(data EventName, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into event_names, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataUsers(data M.User, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into users, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataEvents(data Event, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into events, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataDashboards(data M.Dashboard, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into dashboards, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataDashboardUnits(data M.DashboardUnit, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into dashboard_units, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataProjectSettings(data M.ProjectSetting, db *gorm.DB) error {
	updateFields := map[string]interface{}{
		"int_adwords_customer_account_id": data.IntAdwordsCustomerAccountId,
		"int_facebook_ad_account":         data.IntFacebookAdAccount,
		"int_linkedin_ad_account":         data.IntLinkedinAdAccount,
	}
	query := db.Model(&M.ProjectSetting{}).Where("project_id = ?", data.ProjectId).Updates(updateFields)
	if err := query.Error; err != nil {
		log.Error(err)
		return errors.New("Failed to write data into project settings, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataHubspot(data M.HubspotDocument, db *gorm.DB) error {
	// function to write data to hubspot_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into hubspot_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil

}
func writeDataSalesforce(data M.SalesforceDocument, db *gorm.DB) error {
	// function to write data to salesforce_documents table

	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into salesforce_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataAdwords(data M.AdwordsDocument, db *gorm.DB) error {
	// function to write data to adwords_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into adwords_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataFacebook(data M.FacebookDocument, db *gorm.DB) error {
	// function to write data to facebook_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into facebook_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataLinkedin(data M.LinkedinDocument, db *gorm.DB) error {
	// function to write data to linked_documents table
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into linkedin_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataGoogleOrganic(data M.GoogleOrganicDocument, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into google_organic_documents, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataSmartPropertyRules(data M.SmartPropertyRules, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into smart_property_rules, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}
func writeDataSmartProperties(data M.SmartProperties, db *gorm.DB) error {
	err := db.Create(&data).Error
	if err != nil {
		log.Error(err)
		return errors.New("Failed to write data into smart_properties, total records inserted till now : " + strconv.Itoa(numberOfLinesCopied))
	}
	return nil
}

// derefernces a string pointer
func DerefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

func GetFilesFromCloud(dir, fileName string, pathLocal string) error {
	_context := context.Background()
	_storageClient, _err := storage.NewClient(_context)
	if _err != nil {
		log.Info("%v", _err)
	}
	if !strings.HasSuffix(dir, "/") {
		// Append / to the end if not present.
		dir = dir + "/"
	}
	log.Info(dir, fileName)
	obj := _storageClient.Bucket("factors-staging-misc").Object(dir + fileName)
	rc, err := obj.NewReader(_context)
	if err != nil {
		log.WithError(err).Errorln("Failed to create dir")
		return err
	}

	log.Info(pathLocal)
	err = os.MkdirAll(pathLocal, 0755)
	if err != nil {
		log.WithError(err).Errorln("Failed to create dir")
		return err
	}

	if !strings.HasSuffix(pathLocal, "/") {
		// Append / to the end if not present.
		pathLocal = pathLocal + "/"
	}
	file, err := os.Create(pathLocal + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, rc)
	return err
}
