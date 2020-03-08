package main

import (
	"bufio"
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// denormalized event struct
type denEvent struct {
	Uid string          `json:uid`
	Ujt int64           `json:ujt`
	En  string          `json:en`
	Et  int64           `json:et`
	Ecd uint64          `json:ecd`
	Epr *postgres.Jsonb `json:epr`
	Upr *postgres.Jsonb `json:upr`
}

//agentUUID = "8d6a92d2-a50d-4450-bd27-b9e99e144efb"
// go run run_build_DB_from_model_file.go --file_parth=<path-to-denormalized-file> --project_name=<new-project-name> --agent_uuid=<agent-uuid>

func main() {

	env := flag.String("env", "development", "")
	port := flag.Int("port", 8100, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	filePath := flag.String("file_path", "", "")
	projectName := flag.String("project_name", "", "")
	agentUUID := flag.String("agent_uuid", "", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")

	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	flag.Parse()

	config := &C.Configuration{
		Env:           *env,
		Port:          *port,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:              *redisHost,
		RedisPort:              *redisPort,
		GeolocationFile:        *geoLocFilePath,
		APIDomain:              *apiDomain,
		APPDomain:              *appDomain,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
	}

	// Setup.
	// Initialize configs and connections.
	if err := C.Init(config); err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}

	if filePath == nil || *filePath == "" {
		log.Error("No file path provided")
		os.Exit(1)
	}

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer file.Close()

	var project *M.Project

	events, err := getDenormalizedEventsFromFile(file)
	if err != nil {
		log.WithError(err).Error("Failed to parse file")
		os.Exit(1)
	}

	if projectName == nil || *projectName == "" && agentUUID == nil || *agentUUID == "" {
		log.Error("No project name or uuid provided ")
		os.Exit(1)
	}

	project, err = dbCreateAndGetProjectWithAgentUUID(*projectName, *agentUUID)
	if err != nil {
		log.WithError(err).Error("Failed to create project for agent")
		os.Exit(1)
	}

	normalizeEventsToDB(events, project)
}

func normalizeEventsToDB(events []denEvent, project *M.Project) {

	for _, event := range events {
		err := EventToDb(event, project)
		if err != http.StatusCreated {
			log.Errorf("Failed to create event %+v\n", event)
		}
	}
}

func getDenormalizedEventsFromFile(file *os.File) ([]denEvent, error) {
	var decEvent denEvent
	scanner := bufio.NewScanner(file)
	var events []denEvent
	for scanner.Scan() {
		enEvent := scanner.Text()
		err := json.Unmarshal([]byte(enEvent), &decEvent)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal")
			return nil, errors.New("failed to unmarshal")
		}
		events = append(events, decEvent)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading from file:", err)
		return nil, err
	}

	return events, nil
}

func dbCreateAndGetProjectWithAgentUUID(projectName string, agentUUID string) (*M.Project, error) {

	billingAcc, errCode := M.GetBillingAccountByAgentUUID(agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("CreateProject Failed, billing account error")
		return nil, errors.New("Failed to create billingaccount")
	}

	cproject, errCode := M.CreateProjectWithDependencies(&M.Project{Name: projectName}, agentUUID, 0, billingAcc.ID)
	if errCode != http.StatusCreated {
		log.Error("failed to CreateProjectWithDependencies")
		return nil, errors.New("failed to CreateProjectWithDependencies")
	}
	return cproject, nil
}

func EventToDb(event denEvent, project *M.Project) int {
	user, err := M.CreateUser(&M.User{
		ProjectId:  project.ID,
		Properties: *event.Upr,
	})
	if err != http.StatusCreated {
		log.Errorf("Failed to create user status: %d", err)
		return err
	}

	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: event.En})
	if errCode != http.StatusConflict && errCode != http.StatusCreated {
		log.Error("failed to create CreateOrGetUserCreatedEventName")
		return errCode
	}

	_, errCode = M.CreateEvent(&M.Event{
		ProjectId:        project.ID,
		EventNameId:      eventName.ID,
		UserId:           user.ID,
		Timestamp:        event.Et,
		Count:            event.Ecd,
		Properties:       *event.Epr,
		UserPropertiesId: user.PropertiesId,
	})
	if errCode != http.StatusCreated {
		log.Error("failed to create event")
		return errCode
	}
	return errCode
}
