package main

import (
	"fmt"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"

	"errors"
	"factors/util"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	log "github.com/sirupsen/logrus"
)

/*
Insert into project by Id
go run run_salesforce_lead_create_enrichment.go --file_path=<path-to-denormalized-file> --project_id=<projectId>
*/
func main() {

	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	filePath := flag.String("file_path", "", "")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	sheetName := flag.String("sheet_name", "report1597302806687", " xlsx file") // internal sheet name

	flag.Parse()

	defer util.NotifyOnPanic("Task#run_build_DB_from_salesforce_data", *env)

	config := &C.Configuration{
		Env: *env,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		PrimaryDatastore: *primaryDatastore,
		RedisHost:        *redisHost,
		RedisPort:        *redisPort,
	}

	// Setup.
	// Initialize configs and connections.
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	C.InitRedis(config.RedisHost, config.RedisPort)

	if filePath == nil || *filePath == "" {
		log.Error("No file path provided")
		os.Exit(1)
	}

	file, err := excelize.OpenFile(*filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if projectIdFlag == nil || *projectIdFlag == 0 {
		log.Error("projectId not provided")
		os.Exit(1)
	}

	project, errCode := store.GetStore().GetProject(*projectIdFlag)
	if errCode != http.StatusFound {
		log.Error("Failed to get project")
		return
	}

	eventName := "SF_lead_created"
	err = buildEventFromFile(file, project, eventName, *sheetName)
	if err != nil {
		log.WithError(err).Error("Failed to parse file")
		os.Exit(1)
	}
}

func buildEventFromFile(file *excelize.File, project *model.Project, eventName, sheetName string) error {

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return err
	}

	if len(rows) == 1 {
		return errors.New("No data")
	}

	keys := buildSalesforcePropertyKeys(rows[0])
	for _, row := range rows[1:] {
		properties := buildProperties(keys, row)
		err = buildUserCreatedEvent(project.ID, properties, eventName)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildProperties(keys []string, row []string) map[string]interface{} {
	if len(row) == 0 || len(keys) == 0 {
		return nil
	}

	properties := make(map[string]interface{})
	for colIndex, value := range row {
		properties[keys[colIndex]] = value
	}
	return properties
}

func getJoinTimeStamp(date string) int64 {
	t, err := time.Parse("01-02-06 15:04:05 -07:00", fmt.Sprintf("%s %s", date, "23:59:59 +05:30")) ////move timestamp to end of the day as per IST
	if err != nil {
		return 0
	}
	return t.Unix()
}

func getPhoneNoAsCustomerUserId(phoneNo string) string {
	var newPhoneNo string
	if len(phoneNo) < 2 {
		return phoneNo
	}

	if string(phoneNo[0]) == "+" {
		newPhoneNo = phoneNo[1:]
	} else {
		newPhoneNo = phoneNo
	}

	if len(newPhoneNo) > 10 {
		newPhoneNo = newPhoneNo[len(newPhoneNo)-10:]
	}

	return newPhoneNo
}

//userIndentificationByPhoneNo tries various pattern in specific order for identifying customer_user_id
func userIndentificationByPhoneNo(projectId uint64, phoneNo string) string {

	if len(phoneNo) > 2 {

		var newPhoneNo string
		if string(phoneNo[0]) == "+" {
			newPhoneNo = phoneNo[1:]
		} else {
			newPhoneNo = phoneNo
		}

		//maintain order for matching
		for i := 0; i < 5; i++ {
			switch i {
			case 0:
			case 1:
				if len(newPhoneNo) > 10 {
					newPhoneNo = newPhoneNo[len(newPhoneNo)-10:]
				} else {
					continue
				}
			case 2:
				newPhoneNo = "0" + newPhoneNo
			case 3:
				newPhoneNo = "91" + newPhoneNo[1:]
			default:
				return ""
			}

			userLatest, errCode := store.GetStore().GetUserLatestByCustomerUserId(projectId, newPhoneNo, model.UserSourceSalesforce)
			if errCode == http.StatusFound {
				return userLatest.ID
			}

			userLatest, errCode = store.GetStore().GetUserLatestByCustomerUserId(projectId, "+"+newPhoneNo, model.UserSourceSalesforce)
			if errCode == http.StatusFound {
				return userLatest.ID
			}
		}
	}

	return ""
}

func buildSalesforcePropertyKeys(keys []string) []string {
	var fmtHeaders []string
	for _, key := range keys {
		trimedKey := strings.Split(key, " ")
		fmtHeaders = append(fmtHeaders, "SF_"+strings.Join(trimedKey, "_"))
	}
	return fmtHeaders
}

func buildUserCreatedEvent(projectID uint64, properties map[string]interface{}, eventName string) error {
	trackPayload := &SDK.TrackPayload{
		Name:            eventName,
		ProjectId:       projectID,
		EventProperties: properties,
		UserProperties:  properties,
		Timestamp:       getJoinTimeStamp(properties["SF_Created_Date"].(string)),
		RequestSource:   model.UserSourceSalesforce,
	}

	latestUserId := userIndentificationByPhoneNo(projectID, properties["SF_Mobile"].(string))
	if latestUserId != "" {
		trackPayload.UserId = latestUserId
	}

	status, response := SDK.Track(projectID, trackPayload, true, "Salesforce", "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		return errors.New("Failed to track Salesforce lead created")
	}

	if latestUserId == "" {
		userId := response.UserId
		customerUserId := getPhoneNoAsCustomerUserId(properties["SF_Mobile"].(string))

		status, _ = SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userId, CustomerUserId: customerUserId, RequestSource: model.UserSourceSalesforce}, false)
		if status != http.StatusOK {
			return errors.New(fmt.Sprintf("Failed to identify user %s", customerUserId))
		}
	}

	return nil
}
