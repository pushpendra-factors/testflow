package main

import (
	"bufio"
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	P "factors/pattern"
	U "factors/util"
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

/*
for query with filters mode
go run run_build_DB_from_model_file.go --mode=query_filter --project_id=<project-id> --start_time= --end_time= --file_path=
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

	if *mode == "query_filter" {
		if *projectIdFlag != 0 {
			err := testWithEventConstraints(*customStartTime, *customEndTime, *projectIdFlag, file)
			if err != nil {
				log.Error("Failed to testWithEventConstraints ", err)
			}
		} else {
			log.Error("Not valid parameter for queryMode")
		}
		return
	}

	defer file.Close()

	var project *M.Project

	if projectName == nil || *projectName == "" {
		log.Error("No project name or uuid provided ")
		os.Exit(1)
	}

	if agentUUID == nil || *agentUUID == "" {
		*agentUUID, err = createRandomAgentUUID()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

	}

	project, err = dbCreateAndGetProjectWithAgentUUID(*projectName, *agentUUID)
	if err != nil {
		log.WithError(err).Error("Failed to create project for agent")
		os.Exit(1)
	}
	_, _, err = denormalizedEventsFileToProject(file, project)
	if err != nil {
		log.WithError(err).Error("Failed to parse file")
		os.Exit(1)
	}

}

func createRandomAgentUUID() (string, error) {
	email := U.RandomLowerAphaNumString(6) + "@asdfds.local"
	createAgentParams := M.CreateAgentParams{
		Agent:    &M.Agent{Email: email, Phone: "987654321"},
		PlanCode: M.FreePlanCode,
	}
	if agent, err := M.CreateAgentWithDependencies(&createAgentParams); err != http.StatusCreated {
		return "", errors.New("Failed to create agent")
	} else {
		return agent.Agent.UUID, nil
	}

}

func denormalizedEventsFileToProject(file *os.File, project *M.Project) (int64, int64, error) {
	var startTime int64
	var endTime int64
	_, err := file.Seek(0, 0)
	if err != nil {
		return 0, 0, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		enEvent := scanner.Text()
		var decEvent denEvent
		err := json.Unmarshal([]byte(enEvent), &decEvent)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal")
			return 0, 0, errors.New("failed to unmarshal")
		}
		//if project is provided then it writes to DB without creating array
		if project != nil {
			err := eventToDb(decEvent, project)
			if err != http.StatusOK {
				return 0, 0, errors.New(fmt.Sprintf("Failed to create event %+v\n ,  %d", decEvent, err))
			}
		}

		if startTime == 0 && endTime == 0 {
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
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading from file:", err)
		return 0, 0, err
	}

	return startTime, endTime, nil
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
	user, err := M.CreateOrGetUser(project.ID, event.UserId)
	if err != http.StatusCreated && err != http.StatusOK {
		log.Errorf("Failed to GetSegmentUser status: %d", err)
		return err
	}

	userPropertiesId, errCode := M.UpdateUserProperties(project.ID, user.ID, event.UserProperties, event.UserJoinTimestamp)
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
		startTime, endTime, err = denormalizedEventsFileToProject(file, nil)
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
		log.Info("Success on compare PerOccurrenceCount for single event pattern PerOccurrenceCount: ", result.Rows[0][0].(int64))
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
		log.Info("Success on compare TotalUserCount for single event pattern TotalUserCount: ", result.Rows[0][0].(int64))
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
		log.Info("Success on compare PerUserCount for single event pattern PerUserCount: ", result.Rows[0][0].(int64))
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
		log.Info("Success on compare TotalUserCount for double event pattern TotalUserCount: ", result.Rows[0][0].(int64))
	} else {
		log.Errorf("Failed to compare TotalUserCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[1].TotalUserCount)
	}
	if result.Rows[0][1].(int64) == int64(patterns[1].PerUserCount) {
		log.Info("Success on compare PerUserCount for double event pattern PerUserCount: ", result.Rows[0][1].(int64))
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
		log.Info("Success on compare TotalUserCount for triple event pattern TotalUserCount: ", result.Rows[0][0].(int64))
	} else {
		log.Errorf("Failed to compare TotalUserCount database-result: %d pattern-result: %d ", result.Rows[0][0], patterns[2].TotalUserCount)
	}
	if result.Rows[0][3] == int64(patterns[2].PerUserCount) {
		log.Info("Success on compare PerUserCount for triple event pattern PerUserCount: ", result.Rows[0][3])
	} else {
		log.Errorf("Failed to compare PerUserCount database-result: %d pattern-result: %d ", result.Rows[0][3], patterns[2].PerUserCount)
	}

	return nil
}

func getPatterns(file *os.File) ([]*P.Pattern, error) {

	// A -> $session , B -> Fund Project , C -> View Project

	actualEventInfoMap := make(map[string]*P.PropertiesInfo)
	// Initialize.
	for _, eventName := range []string{"Fund Project", "$session", "View Project"} {
		// Initialize info.
		actualEventInfoMap[eventName] = &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		}
	}

	userAndEventsInfo := P.UserAndEventsInfo{
		UserPropertiesInfo: &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		},
		EventPropertiesInfoMap: &actualEventInfoMap,
	}

	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	err = P.CollectPropertiesInfo(scanner, &userAndEventsInfo)
	if err != nil {
		return nil, err
	}

	pB, _ := P.NewPattern([]string{"Fund Project"}, &userAndEventsInfo)
	pAB, _ := P.NewPattern([]string{"$session", "Fund Project"}, &userAndEventsInfo)
	pABC, _ := P.NewPattern([]string{"$session", "Fund Project", "View Project"}, &userAndEventsInfo)
	patterns := []*P.Pattern{pB, pAB, pABC}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	scanner = bufio.NewScanner(file)
	err = P.CountPatterns(scanner, patterns)
	if err != nil {
		return nil, err
	}
	return patterns, nil
}

func testWithEventConstraints(startTime int64, endTime int64, projectId uint64, file *os.File) error {
	//Test with event constraints
	patterns, err := getPatterns(file)
	if err != nil {
		return err
	}
	if startTime == 0 || endTime == 0 {
		startTime, endTime, err = denormalizedEventsFileToProject(file, nil)
		if err != nil {
			return err
		}
	}

	//For Fund Project with $day_of_week equals Friday
	log.Info("For Fund Project event")
	genericPropertiesConstraints := []P.EventConstraints{
		P.EventConstraints{
			EPCategoricalConstraints: []P.CategoricalConstraint{
				P.CategoricalConstraint{
					PropertyName:  "$day_of_week",
					PropertyValue: "Friday",
				},
			},
		},
	}
	perUserCount, err := patterns[0].GetPerUserCount(genericPropertiesConstraints)
	if err != nil {
		return err
	}
	query := M.Query{
		From: startTime,
		To:   endTime,
	}

	enQueries := []byte(`{"cl":"funnel","ty":"unique_users","ec":"any_given_event","ewp":[{"na":"Fund Project","pr":[{"en":"event","ty":"categorical","pr":"$day_of_week","op":"equals","va":"Friday","lop":"AND"}]}],"gbp":[],"gbt":"","tz":"UTC","fr":1393632000,"to":1394236799,"ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Failed to unmarshal")
	}
	result, _, _ := M.Analyze(projectId, query)
	log.Info(fmt.Sprintf("PATTERN RESULT \nperUserCount:%d\n", perUserCount))
	log.Info(fmt.Sprintf("\nQuery Result\nperUserCount:%d\n", result.Rows[0][0]))

	//FOR TRIPLE EVENTS($SESSION -> FUND PROJECT -> VIEW PROJECT)
	// with $SESSION event property $day_of_week equals Friday
	log.Info("EVENTS($SESSION -> FUND PROJECT -> VIEW PROJECT")
	genericPropertiesConstraints = []P.EventConstraints{
		P.EventConstraints{
			EPCategoricalConstraints: []P.CategoricalConstraint{
				P.CategoricalConstraint{
					PropertyName:  "$day_of_week",
					PropertyValue: "Friday",
				},
			},
		},
		P.EventConstraints{},
		P.EventConstraints{},
	}

	perUserCount, _ = patterns[2].GetPerUserCount(genericPropertiesConstraints)

	enQueries = []byte(`{"cl":"funnel","ty":"unique_users","ec":"any_given_event","ewp":[{"na":"$session","pr":[{"en":"event","ty":"categorical","pr":"$day_of_week","op":"equals","va":"Friday","lop":"AND"}]},{"na":"Fund Project","pr":[]},{"na":"View Project","pr":[]}],"gbp":[],"gbt":"","tz":"UTC","ovp":false}`)
	err = json.Unmarshal(enQueries, &query)
	if err != nil {
		return errors.New("Failed to unmarshal")
	}
	result, _, _ = M.Analyze(projectId, query)
	log.Info(fmt.Sprintf("PATTERN RESULT \nperUserCount:%d\n", perUserCount))
	log.Info(fmt.Sprintf("\nQuery Result\nnperUserCount:%d\n", result.Rows[0][3]))
	return nil
}
