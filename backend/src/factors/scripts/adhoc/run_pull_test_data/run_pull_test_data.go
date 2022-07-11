package main

import (
	"context"
	"database/sql"
	"encoding/json"
	C "factors/config"
	M "factors/model/model"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
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
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `json:"user_id"`

	UserProperties *postgres.Jsonb `json:"user_properties"`
	SessionId      *string         `json:session_id`
	EventNameId    uint64          `json:"event_name_id"`
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
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"	not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var recordLimit *int
var jobMap = map[string][]string{
	"dashboard_caching":          []string{"events", "users", "event_names", "dashboard_units", "dashboards"},
	"add_session":                []string{"events", "users", "event_names"},
	"hubspot_enrich":             []string{"hubspot_documents", "event_names", "project_settings"},
	"salesforce_enrich":          []string{"salesforce_documents", "event_names", "project_settings"},
	"web_analytics_dashboard":    []string{"events", "users", "event_names"},
	"explain_feature_testing_ui": []string{"events", "users", "event_names"},
	"attribution":                []string{"events", "users", "event_names", "project_settings", "adwords_documents", "facebook_documents", "linkedin_documents", "smart_property_rules", "smart_properties"},
	"test":                       []string{"adwords_documents"},
}
var userIDMap = make(map[string]bool)
var eventNameIDMap = make(map[uint64]bool)
var dashboardIDMap = make(map[uint64]bool)
var folderName string

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")
	// to take input parameters to fetch from hubspot_documents

	// table name from which data is to be pulled
	jobConfig := flag.String("jobConfig", "default", "please enter a job config type")

	// if user wants to keep a file storing data pulled from table.(accepts true/false)
	createFile := flag.String("createFile", "true", "do you want to keep file storing data ?")

	pageSize := flag.Int("pageSize", 10000, "specify the page size")

	// timestamps between which user want to pull records
	startTime := flag.String("startTime", "none", "specify start time ")
	endTime := flag.String("endTime", "none", "specify end time")

	//	project_id from which data to be retrieved.
	projectId := flag.Int64("projectId", 000000, "specify the project_id")
	// max number of records to be pulled
	recordLimit = flag.Int("recordLimit", 2000, "enter the upper limit of records to be pulled")

	eventNameIds := flag.String("event_name_ids", "", "")
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
	log.Print("time", *startTime)
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	if *jobConfig == "default" {
		log.Error("No job type specified")
		return
	}
	if _, exists := jobMap[string(*jobConfig)]; !exists {
		log.Error("job type " + *jobConfig + " does not exist")
		return
	}
	if *createFile != "true" && *createFile != "false" {
		log.Error("please specify true or false only for creating file flag")
		return
	}
	if *startTime == "none" || *endTime == "none" {
		log.Error("please enter start time and end time ")
		return
	}
	if *projectId == 000000 {
		log.Error("please specify the project_id")
		return
	}
	if *recordLimit > 10000 {
		log.Error("the limit entered is too high , please enter <= 10000")
		return
	}

	currentJobConfig := jobMap[*jobConfig]
	readData(currentJobConfig, db, *startTime, *endTime, *projectId, *pageSize, *jobConfig, *eventNameIds)
	if *env != "development" {
		CopyFilesToCloud(folderName, fmt.Sprintf("transfer_data/%v/%v", *projectId, folderName), "factors-staging-misc")
	}
}

func readData(jobConfig []string, db *gorm.DB, startTime string, endTime string, projectId int64, pageSize int, jobConfigKey string, eventNameIds string) {
	time := time.Now().Unix()
	log.Info(time)
	folderName = fmt.Sprintf("%v", time)
	err := os.Mkdir(strconv.FormatInt(time, 10), 0755)
	if err != nil {
		log.Error(err)
	}
	for _, tableNameValue := range jobConfig {
		path := strconv.FormatInt(time, 10) + "/" + tableNameValue + ".txt"
		var file, er = os.Create(path)
		if er != nil {
			log.Fatal("Error creating a text file")
			return
		}
		defer file.Close()
		var selectCountString string
		if tableNameValue == "smart_properties" || tableNameValue == "smart_property_rules" || tableNameValue == "project_settings" {
			selectCountString = "SELECT * FROM " + tableNameValue + " WHERE project_id=" + strconv.FormatInt(projectId, 10) + " LIMIT " + strconv.Itoa(pageSize)
		} else if tableNameValue == "users" || tableNameValue == "event_names" || tableNameValue == "dashboards" {
			var IdString string = getIDString(tableNameValue)
			selectCountString = "SELECT * FROM " + tableNameValue + " WHERE project_id = " + strconv.Itoa(int(projectId)) + " AND id IN " + IdString + " LIMIT " + strconv.Itoa(pageSize)
		} else {
			selectCountString = getSelectQuery(tableNameValue, projectId, startTime, endTime, pageSize, eventNameIds)
		}
		log.Info(selectCountString)

		records, err := db.Raw(selectCountString).Rows()
		if err != nil {
			log.Error(err)

		}
		for records.Next() {
			if tableNameValue == "event_names" {
				writeEventNames(db, records, file, jobConfigKey)
			} else if tableNameValue == "users" {
				writeUsers(db, records, file, jobConfigKey)
			} else if tableNameValue == "events" {
				writeEvents(db, records, file, jobConfigKey)
			} else if tableNameValue == "dashboards" {
				writeDashBoards(db, records, file, jobConfigKey)
			} else if tableNameValue == "dashboard_units" {
				writeDashBoardUnits(db, records, file, jobConfigKey)
			} else if tableNameValue == "hubspot_documents" {
				writeHubspotDocuments(db, records, file, jobConfigKey)
			} else if tableNameValue == "salesforce_documents" {
				writeSalesForceDocuments(db, records, file, jobConfigKey)
			} else if tableNameValue == "adwords_documents" {
				writeAdwordDocuments(db, records, file, jobConfigKey)
			} else if tableNameValue == "facebook_documents" {
				writeFacebookDocuments(db, records, file, jobConfigKey)
			} else if tableNameValue == "linkedin_documents" {
				writeLinkedinDocuments(db, records, file, jobConfigKey)
			} else if tableNameValue == "smart_property_rules" {
				writeSmartPropertyRules(db, records, file, jobConfigKey)
			} else if tableNameValue == "smart_properties" {
				writeSmartProperties(db, records, file, jobConfigKey)
			}

		}
	}
	log.Println("All records have been written to file sucessfully")
}

// converts records to string and invokes the writeFile method to write the record into text file

func writeUsers(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var usersObj M.User
	if err := db.ScanRows(records, &usersObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(usersObj)
	if err != nil {
		log.Error("Error while marshalling users  ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeEventNames(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var eventnamesObj EventName
	if err := db.ScanRows(records, &eventnamesObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(eventnamesObj)
	if err != nil {
		log.Error("Error while marshalling event_names document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeEvents(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var eventObj Event
	if err := db.ScanRows(records, &eventObj); err != nil {
		log.Println(err)
	}
	if _, exsits := userIDMap[eventObj.UserId]; !exsits {
		userIDMap[eventObj.UserId] = true
	}
	if _, exsits := eventNameIDMap[eventObj.EventNameId]; !exsits {
		eventNameIDMap[eventObj.EventNameId] = true
	}

	dataToBeWritten, err := json.Marshal(eventObj)
	if err != nil {
		log.Error("Error while marshalling users  ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeProjectSettings(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var projectSettingObj M.ProjectSetting
	if err := db.ScanRows(records, &projectSettingObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(projectSettingObj)
	if err != nil {
		log.Error("Error while marshalling project settings ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeDashBoards(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var dashboardObj M.Dashboard

	if err := db.ScanRows(records, &dashboardObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(dashboardObj)
	if err != nil {
		log.Error("Error while marshalling dashboards  ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeDashBoardUnits(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var dashboardUnitsObj M.DashboardUnit
	if err := db.ScanRows(records, &dashboardUnitsObj); err != nil {
		log.Println(err)
	}
	if _, exsits := dashboardIDMap[dashboardUnitsObj.ID]; !exsits {
		dashboardIDMap[dashboardUnitsObj.ID] = true
	}
	if err := db.ScanRows(records, &dashboardUnitsObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(dashboardUnitsObj)
	if err != nil {
		log.Error("Error while marshalling dashboard_units  ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeHubspotDocuments(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var hubspotObj M.HubspotDocument
	if err := db.ScanRows(records, &hubspotObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(hubspotObj)
	if err != nil {
		log.Error("Error while marshalling salesforce document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeSmartPropertyRules(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var smartPropertyRulesObj M.SmartPropertyRules
	if err := db.ScanRows(records, &smartPropertyRulesObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(smartPropertyRulesObj)
	if err != nil {
		log.Error("Error while marshalling smart property rules document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeSmartProperties(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var smartPropertyObj M.SmartProperties
	if err := db.ScanRows(records, &smartPropertyObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(smartPropertyObj)
	if err != nil {
		log.Error("Error while marshalling smart property document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeAdwordDocuments(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var adwordsObj M.AdwordsDocument
	if err := db.ScanRows(records, &adwordsObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(adwordsObj)
	if err != nil {
		log.Error("Error while marshalling adwords document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeFacebookDocuments(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var facebookObj M.FacebookDocument
	if err := db.ScanRows(records, &facebookObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(facebookObj)
	if err != nil {
		log.Error("Error while marshalling facebook document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeLinkedinDocuments(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var linkedinObj M.LinkedinDocument
	if err := db.ScanRows(records, &linkedinObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(linkedinObj)
	if err != nil {
		log.Error("Error while marshalling linkedin document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}
func writeSalesForceDocuments(db *gorm.DB, records *sql.Rows, file *os.File, logKey string) {
	var salesforceObj M.SalesforceDocument
	if err := db.ScanRows(records, &salesforceObj); err != nil {
		log.Println(err)
	}

	dataToBeWritten, err := json.Marshal(salesforceObj)
	if err != nil {
		log.Error("Error while marshalling salesforce document ")
		return
	}
	// writing the data into text file
	writefile(file, string(dataToBeWritten)+"\n")
}

// returns sql query to select data
func getSelectQuery(tableName string, ProjectID int64, startTime, endTime string, pageSize int, event_name_ids string) string {
	query := ("SELECT * FROM " + tableName + " WHERE updated_at between '" + string(startTime) + "' AND '" + string(endTime) + "' AND project_id = " + strconv.FormatInt(ProjectID, 10))
	if tableName == "events" && event_name_ids != "" {
		query = query + " AND event_name_id in (" + event_name_ids + ")"
	}
	query = query + "  LIMIT " + strconv.Itoa(pageSize)
	return query
}

// returns event_name_ids from map in (x,y) fashion
func getIDString(tableName string) string {
	str := "( "
	if tableName == "event_names" {
		for id := range eventNameIDMap {
			str += strconv.FormatUint(id, 10) + ", "
		}
		str += "000000 )"
	} else if tableName == "users" {
		for id := range userIDMap {
			str += `'` + id + `'` + ", "
		}
		str += "'0000' )"
	} else if tableName == "dashboards" {
		for id := range dashboardIDMap {
			str += strconv.FormatUint(id, 10) + ", "
		}
		str += "000000 )"
	}

	return str
}

// writes string to text file
func writefile(file *os.File, dataToBeWritten string) {
	_, err := file.WriteString(dataToBeWritten)
	if err != nil {
		log.Fatal("Error writing String", err)
	}

}

func CopyFilesToCloud(folderPath string, bucketPath string, bucketName string) {
	_context := context.Background()
	_storageClient, _err := storage.NewClient(_context)
	if _err != nil {
		log.Info("%v", _err)
	}
	_bucket := _storageClient.Bucket(bucketName)
	files := GetAllFiles(folderPath, "")
	for _, element := range files {
		_file, _err := os.Open(element)
		if _err != nil {
			log.Error("os.Open: %v", _err)
		}
		defer _file.Close()
		_storageWriter := _bucket.Object(fmt.Sprintf("%v/%v", bucketPath, element)).NewWriter(_context)
		if _, _err = io.Copy(_storageWriter, _file); _err != nil {
			log.Error(_err)
		}
		if _err := _storageWriter.Close(); _err != nil {
			log.Error(_err)
		}
	}
}

func GetAllFiles(FileRootPath string, filenameprefix string) []string {

	var files []string
	log.Info("Searching for all the files in path %s", FileRootPath)

	err := filepath.Walk(FileRootPath, func(path string, info os.FileInfo, err error) error {
		if filenameprefix != "" {
			if !strings.HasPrefix(info.Name(), filenameprefix) {
				return nil
			}
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return files
}
