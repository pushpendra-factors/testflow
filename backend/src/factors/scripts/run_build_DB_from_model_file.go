package main

import (
	"bufio"
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	P "factors/pattern"
	"flag"
	"fmt"
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

/*
default to ingest mode
go run run_build_DB_from_model_file.go --file_path=<path-to-denormalized-file> --project_name=<new-project-name> --agent_uuid=<agent-uuid> --mode=
*/

/*
for query mode
go run run_build_DB_from_model_file.go --mode=query --project_id=<project-id> --start_time= --end_time= --file_path=
*/

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
	mode := flag.String("mode", "ingest", "")
	customStartTime := flag.Int64("start_time", 0, "")
	customEndTime := flag.Int64("end_time", 0, "")
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

	if filePath == nil || *filePath == "" {
		log.Error("No file path provided")
		os.Exit(1)
	}

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if *mode == "query" {
		if *projectIdFlag != 0 {
			err := testData(*customStartTime, *customEndTime, *projectIdFlag, file)
			if err != nil {
				log.Error("Failed to testData ", err)
			}
		} else {
			log.Error("Not valid parameter for queryMode")
		}
		return
	}

	defer file.Close()

	var project *M.Project

	events, _, _, err := getDenormalizedEventsFromFile(file)
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

func getDenormalizedEventsFromFile(file *os.File) ([]denEvent, int64, int64, error) {
	startTime := int64(0)
	endTime := int64(0)
	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, 0, 0, err
	}

	scanner := bufio.NewScanner(file)
	var events []denEvent
	for scanner.Scan() {
		enEvent := scanner.Text()
		var decEvent denEvent
		err := json.Unmarshal([]byte(enEvent), &decEvent)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal")
			return nil, 0, 0, errors.New("failed to unmarshal")
		}

		if startTime == int64(0) && endTime == int64(0) {
			startTime = decEvent.EventTime
			endTime = decEvent.EventTime
		} else {
			if startTime > decEvent.EventTime {
				startTime = decEvent.EventTime
			}
			if endTime < decEvent.EventTime {
				endTime = decEvent.EventTime
			}
		}
		events = append(events, decEvent)
	}

	fmt.Println("length of events : ", len(events))
	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading from file:", err)
		return nil, 0, 0, err
	}

	return events, startTime, endTime, nil
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

func testData(startTime int64, endTime int64, projectId uint64, file *os.File) error {

	patterns, err := getPatterns(file)
	if err != nil {
		return err
	}
	if startTime == 0 || endTime == 0 {
		_, startTime, endTime, err = getDenormalizedEventsFromFile(file)
		if err != nil {
			return err
		}
	}

	query := M.Query{
		From: startTime,
		To:   endTime,
	}

	// FOR SINGLE EVENT(FUND PROJECT) COUNT
	log.Info("Comparing Fund Project pattern count")
	// PerOccurrenceCount is calculate using COUNT on EVENT OCCURRENCE in INSIGHT query, since single event pattern is used.
	enQueries := []byte(`{"cl":"insights","ty":"events_occurrence","ec":"any_given_event","ewp":[{"na":"Fund Project","pr":[]}],"gbp":[],"gbt":"","tz":"UTC","ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Faile to unmarshal")
	}
	result, errCode, _ := M.Analyze(projectId, query)
	if errCode != http.StatusOK {
		errors.New("Failed to analyze query")
	}
	if result.Rows[0][0].(int64) == int64(patterns[0].PerOccurrenceCount) {
		log.Info("Success on compare PerOccurrenceCount for single event pattern")
	} else {
		log.Errorf("Failed to compare PerOccurrenceCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[0].PerOccurrenceCount)
	}

	//TotalUserCount is calculated using COUNT on UNIQUE USERS in INSIGHT query
	enQueries = []byte(`{"cl":"insights","ty":"unique_users","ec":"any_given_event","ewp":[{"na":"$session","pr":[]},{"na":"View Project","pr":[]},{"na":"Fund Project","pr":[]}],"gbp":[],"gbt":"","tz":"UTC","ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Faile to unmarshal")
	}
	result, errCode, _ = M.Analyze(projectId, query)
	if errCode != http.StatusOK {
		errors.New("Failed to analyze query")
	}
	if result.Rows[0][0].(int64) == int64(patterns[0].TotalUserCount) {
		log.Info("Success on compare TotalUserCount for single event pattern")
	} else {
		log.Errorf("Failed to compare TotalUserCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[0].TotalUserCount)
	}

	//PerUserCount is calculated using COUNT on UNIQUE USERS in FUNNEL query. Same as COUNT on UNIQUE USERS in INSIGHT query
	enQueries = []byte(`{"cl":"funnel","ty":"unique_users","ec":"any_given_event","ewp":[{"na":"Fund Project","pr":[]}],"gbp":[],"gbt":"","tz":"UTC","ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Faile to unmarshal")
	}
	result, errCode, _ = M.Analyze(projectId, query)
	if errCode != http.StatusOK {
		errors.New("Failed to analyze query")
	}
	if result.Rows[0][0].(int64) == int64(patterns[0].PerUserCount) {
		log.Info("Success on compare PerUserCount for single event pattern\n")
	} else {
		log.Errorf("Failed to compare PerUserCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[0].PerUserCount)
	}

	// FOR DOUBLE EVENTS(SESSION -> FUND PROJECT) COUNT
	log.Info("Comparing $session -> Fund Project pattern count")
	//TotalUserCount and PerUserCount is calculated using COUNT on conversion of UNIQUE USERS from $session to Fund Project using FUNNEL query,
	// where TotalUserCount is step_0 and PerUserCount is step_1
	enQueries = []byte(`{"cl":"funnel","ty":"unique_users","ec":"any_given_event","ewp":[{"na":"$session","pr":[]},{"na":"Fund Project","pr":[]}],"gbp":[],"gbt":"","tz":"UTC","ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Faile to unmarshal")
	}
	result, errCode, _ = M.Analyze(projectId, query)
	if errCode != http.StatusOK {
		errors.New("Failed to analyze query")
	}
	if result.Rows[0][0].(int64) == int64(patterns[1].TotalUserCount) {
		log.Info("Success on compare TotalUserCount for double event pattern")
	} else {
		log.Errorf("Failed to compare TotalUserCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[1].TotalUserCount)
	}
	if result.Rows[0][1].(int64) == int64(patterns[1].PerUserCount) {
		log.Info("Success on compare PerUserCount for double event pattern\n")
	} else {
		log.Errorf("Failed to compare PerUserCount database-result: %d pattern-result: %d ", result.Rows[0][1], patterns[1].PerUserCount)
	}

	//FOR TRIPLE EVENTS($SESSION -> FUND PROJECT -> VIEW PROJECT)
	log.Info("Comparing $session -> Fund Project -> View Project pattern count")
	//TotalUserCount and PerUserCount is calculated using COUNT on conversion of UNIQUE USERS from $session to Fund Project to View Project using FUNNEL query,
	// where TotalUserCount is step_0 and PerUserCount is step_2
	enQueries = []byte(`{"cl":"funnel","ty":"unique_users","ec":"any_given_event","ewp":[{"na":"$session","pr":[]},{"na":"Fund Project","pr":[]},{"na":"View Project","pr":[]}],"gbp":[],"gbt":"","tz":"UTC","ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Faile to unmarshal")
	}
	result, errCode, _ = M.Analyze(projectId, query)
	if errCode != http.StatusOK {
		return errors.New("Failed to analyze query")
	}
	if result.Rows[0][0].(int64) == int64(patterns[2].TotalUserCount) {
		log.Info("Success on compare TotalUserCount for triple event pattern")
	} else {
		log.Errorf("Failed to compare TotalUserCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[2].TotalUserCount)
	}
	if result.Rows[0][2].(int64) == int64(patterns[2].PerUserCount) {
		log.Info("Success on compare PerUserCount for triple event pattern")
	} else {
		log.Errorf("Failed to compare PerUserCount database-result: %d pattern-result: %d ", result.Rows[0][2], patterns[2].PerUserCount)
	}
	return nil
}

func getPatterns(file *os.File) ([]*P.Pattern, error) {

	// A -> $session , B -> Fund Project , C -> View Project
	pB, _ := P.NewPattern([]string{"Fund Project"}, nil)
	pAB, _ := P.NewPattern([]string{"$session", "Fund Project"}, nil)
	pABC, _ := P.NewPattern([]string{"$session", "Fund Project", "View Project"}, nil)
	patterns := []*P.Pattern{pB, pAB, pABC}

	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	err = P.CountPatterns(scanner, patterns)
	if err != nil {
		return nil, err
	}
	return patterns, nil
}
