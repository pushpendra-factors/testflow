package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type QueryClass struct {
	Class string `json:"cl"`
}

type GenericQueryGroup struct {
	Queries []postgres.Jsonb `json:"query_group"`
}

func getQueryClass(queryJson postgres.Jsonb) string {
	var queryClass QueryClass
	err := json.Unmarshal(queryJson.RawMessage, &queryClass)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal to get query class.")
		return ""
	}

	return queryClass.Class
}

// Returns query as postgres.Jsonb, flattened from groups.
func getAllQueriesAsPostgresJsonb(queries []model.Queries) []postgres.Jsonb {
	queriesJsonb := make([]postgres.Jsonb, 0, 0)
	for qi := range queries {
		query := queries[qi]
		logCtx := log.WithField("query_record", query).
			WithField("query_text", string(query.Query.RawMessage))

		queryText := map[string]interface{}{}
		err := json.Unmarshal(query.Query.RawMessage, &queryText)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmashal query text.")
			continue
		}

		if _, isQueryGroup := queryText["query_group"]; isQueryGroup {
			queryGroup := GenericQueryGroup{}
			err := json.Unmarshal(query.Query.RawMessage, &queryGroup)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal query text to query group.")
				continue
			}

			queriesJsonb = append(queriesJsonb, queryGroup.Queries...)
			continue
		}

		queriesJsonb = append(queriesJsonb, query.Query)
	}

	return queriesJsonb
}

// Runs queries for 10 seconds timerange to compile and cache plans for consequtive runs.
func runQueriesForCompilation(projectID int64, queries []model.Queries, logCtx *log.Entry) {
	queriesJsonb := getAllQueriesAsPostgresJsonb(queries)

	for qji := range queriesJsonb {
		queryJsonb := queriesJsonb[qji]
		queryClass := getQueryClass(queriesJsonb[qji])

		logCtx = logCtx.WithField("query_class", queryClass)

		switch queryClass {
		case model.QueryClassInsights, model.QueryClassEvents, model.QueryClassFunnel:
			var query model.Query
			err := json.Unmarshal(queryJsonb.RawMessage, &query)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal.")
				continue
			}

			query.To = time.Now().Unix()
			query.From = query.To - 10

			_, statusCode, statusMsg := store.GetStore().Analyze(projectID, query, true, false)
			if statusCode == http.StatusInternalServerError {
				logCtx.WithField("message", statusMsg).Error("Failed running insights, events or funnel query.")
				continue
			}

		case model.QueryClassProfiles:
			var query model.ProfileQuery
			err := json.Unmarshal(queryJsonb.RawMessage, &query)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal.")
				continue
			}

			query.To = time.Now().Unix()
			query.From = query.To - 10

			_, statusCode, statusMsg := store.GetStore().ExecuteProfilesQuery(projectID, query, true)
			if statusCode == http.StatusInternalServerError {
				logCtx.WithField("message", statusMsg).Error("Failed running profiles query.")
				continue
			}

		case model.QueryClassChannel:
			var query model.ChannelQuery
			err := json.Unmarshal(queryJsonb.RawMessage, &query)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal.")
				continue
			}

			query.To = time.Now().Unix()
			query.From = query.To - 10

			// _, statusCode := store.GetStore().ExecuteChannelQuery(projectID, &query)
			// if statusCode == http.StatusInternalServerError {
			// 	logCtx.Error("Failed running channel query.")
			// 	continue
			// }

		case model.QueryClassChannelV1:
			var query model.ChannelQueryV1
			err := json.Unmarshal(queryJsonb.RawMessage, &query)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal.")
				continue
			}

			query.To = time.Now().Unix()
			query.From = query.To - 10

			_, statusCode := store.GetStore().ExecuteChannelQueryV1(projectID, &query, "")
			if statusCode == http.StatusInternalServerError {
				logCtx.Error("Failed running channel v1 query.")
				continue
			}

		default:
			logCtx.Warn("Unsupported query class")
		}
	}
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	appName := flag.String("app_name", "queries_warmup", "Override default app_name.")
	healthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck_ping_id.")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")

	projectIDs := flag.String("project_ids", "", "")
	flag.Parse()

	defer C.PingHealthcheckForPanic(*appName, *env, *healthcheckPingID)

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		AppName:   *appName,
		Env:       *env,
		SentryDSN: *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     *appName,
		},
		PrimaryDatastore: C.DatastoreTypeMemSQL,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize DB")
	}

	isAllProjects, allowedProjectsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDs, "")
	if len(allowedProjectsMap) == 0 {
		log.WithError(err).Error("Failed to initialize DB")
		os.Exit(0)
	}

	projects, errCode := store.GetStore().GetAllProjectIDs()
	if errCode != http.StatusFound {
		log.Error("No projects found.")
		return
	}

	for pi := range projects {
		projectID := projects[pi]

		_, isProjectAllowed := allowedProjectsMap[projectID]
		isProjectAllowed = isAllProjects || isProjectAllowed
		if !isProjectAllowed {
			continue
		}

		logCtx := log.WithField("project_id", projectID)

		queries, errCode := store.GetStore().GetALLQueriesWithProjectId(projectID)
		if errCode != http.StatusFound {
			logCtx.Warn("No saved queries.")
			continue
		}

		runQueriesForCompilation(projectID, queries, logCtx)
	}
}
