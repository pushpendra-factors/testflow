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

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// denormalized event struct
type denEvent struct {
	UserId            string          `json:"uid"`
	UserJoinTimestamp int64           `json:"ujt"`
	EventName         string          `json:"en"`
	EventTime         int64           `json:"et"`
	EventCount        uint64          `json:"ecd"`
	EventProperties   *postgres.Jsonb `json:"epr"`
	UserProperties    *postgres.Jsonb `json:"upr"`
}

//agentUUID = "8d6a92d2-a50d-4450-bd27-b9e99e144efb"
// go run run_build_DB_from_model_file.go --file_path=<path-to-denormalized-file> --project_name=<new-project-name> --agent_uuid=<agent-uuid>

func main() {

	env := flag.String("env", "development", "")

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
		RedisHost: *redisHost,
		RedisPort: *redisPort,
	}

	// Setup.
	// Initialize configs and connections.
	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	C.InitRedis(config.RedisHost, config.RedisPort)

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
		ProjectId:      project.ID,
		Properties:     *event.UserProperties,
		CustomerUserId: event.UserId,
		JoinTimestamp:  event.UserJoinTimestamp,
	})
	if err != http.StatusCreated {
		log.Errorf("Failed to create user status: %d", err)
		return err
	}

	eventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: event.EventName})
	if errCode != http.StatusConflict && errCode != http.StatusCreated {
		log.Error("failed to create CreateOrGetUserCreatedEventName")
		return errCode
	}

	_, errCode = M.CreateEvent(&M.Event{
		ProjectId:        project.ID,
		EventNameId:      eventName.ID,
		UserId:           user.ID,
		Timestamp:        event.EventTime,
		Count:            event.EventCount,
		Properties:       *event.EventProperties,
		UserPropertiesId: user.PropertiesId,
	})
	if errCode != http.StatusCreated {
		log.Error("failed to create event")
		return errCode
	}
	return errCode
}
