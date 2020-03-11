package main

import (
	"bufio"
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

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

/*
default to ingest mode
agentUUID = "8d6a92d2-a50d-4450-bd27-b9e99e144efb"
go run run_build_DB_from_model_file.go --file_path=<path-to-denormalized-file> --project_name=<new-project-name> --agent_uuid=<agent-uuid> --mode=
*/

/*
for query mode
gor run run_build_DB_from_model_file.go --mode=query --project_id=<project-id> --start_time= --end_time=
*/

func main() {

	currentTime := time.Now()
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
	mode := flag.String("mode", "ingest", "")
	customStartTime := flag.Int64("start_time", currentTime.AddDate(0, 0, -7).Unix(), "")
	customEndTime := flag.Int64("end_time", currentTime.Unix(), "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")

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

	if *mode == "query" {
		if *customEndTime != 0 && *customStartTime != 0 && *projectIdFlag != 0 {
			testFunnelQuery(*customStartTime, *customEndTime, *projectIdFlag)
		} else {
			log.Error("Not valid parameter for queryMode")
		}
		return
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
		err := eventToDb(event, project)
		if err != http.StatusOK {
			log.Errorf("Failed to create event %+v\n ,  %d", event, err)
		}
	}
}

func getDenormalizedEventsFromFile(file *os.File) ([]denEvent, error) {
	scanner := bufio.NewScanner(file)
	var events []denEvent
	for scanner.Scan() {
		enEvent := scanner.Text()
		var decEvent denEvent
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

func eventToDb(event denEvent, project *M.Project) int {
	user, err := M.GetSegmentUser(project.ID, "", event.UserId)
	if err != http.StatusCreated && err != http.StatusOK {
		log.Errorf("Failed to GetSegmentUser status: %d", err)
		return err
	}

	userPropertiesId, errCode := M.UpdateUserProperties(project.ID, user.ID, event.UserProperties)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return errCode
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
		UserPropertiesId: userPropertiesId,
	})
	if errCode != http.StatusFound && errCode != http.StatusCreated {
		log.Errorf("failed to create event, errCode: %+v\n", errCode)
		return errCode
	}
	return http.StatusOK
}

func testFunnelQuery(startTime int64, endTime int64, projectId uint64) error {
	var queries = []map[string]string{
		{"entity": "event", "type": "categorical", "property": "category", "operator": "equals", "value": "Games", "logicalOp": "AND"},
		{"entity": "event", "type": "categorical", "property": "category", "operator": "equals", "value": "Fashion", "logicalOp": "AND"},
		{"entity": "user", "type": "categorical", "property": "$day_of_first_event", "operator": "equals", "value": "Wednesday", "logicalOp": "AND"},
		{"entity": "user", "type": "categorical", "property": "gender", "operator": "equals", "value": "M", "logicalOp": "AND"},
	}

	var results []M.QueryResult
	for _, queryParams := range queries {
		query := M.Query{
			From: startTime,
			To:   endTime,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "View Project",
					Properties: []M.QueryProperty{
						M.QueryProperty{
							Entity:    queryParams["entity"],
							Property:  queryParams["property"],
							Operator:  queryParams["operator"],
							Value:     queryParams["value"],
							Type:      queryParams["type"],
							LogicalOp: queryParams["logicalOp"],
						},
					},
				},
			},
			Class:           M.QueryClassFunnel,
			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAnyGivenEvent,
		}

		result, errCode, _ := M.Analyze(projectId, query)
		if errCode != http.StatusOK {
			return errors.New("Failed to analyze query")
		}

		results = append(results, *result)
	}

	log.Info("Results:")
	for _, result := range results {
		if len(result.Headers) == 2 {
			log.Info(fmt.Sprintf("Query: %+v\n%s:%d\n%s:%s", result.Meta.Query, result.Headers[0], result.Rows[0][0], result.Headers[1], result.Rows[0][1]))
		} else {
			log.Info(fmt.Sprintf("Query: %+v\n%s:%d", result.Meta.Query, result.Headers[0], result.Rows[0][0]))
		}
	}

	return nil
}
