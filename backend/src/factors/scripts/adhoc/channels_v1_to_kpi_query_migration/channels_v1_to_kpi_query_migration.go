package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// what about channel queries.
// Need to take backup.
func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")
	normalRun := flag.Bool("normal_run", false, "")

	projectIDsFlag := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDsFlag := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	flag.Parse()

	model.SetSmartPropertiesReservedNames()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	}

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDsFlag, *disabledProjectIDsFlag)

	log.Info("Starting to initialize database.")
	appName := "channels_v1_to_kpi_query_migration"

	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	dbQueries := []model.Queries{}
	if allProjects {
		err = db.Table("queries").Where("JSON_EXTRACT_STRING(query, 'cl') = 'channel_v1'").Find(&dbQueries).Error
		if err != nil {
			log.Warn("Failed during query fetch")
			os.Exit(0)
		}
	} else {
		projectIdsArray := make([]uint64, 0)
		for projectId, _ := range projectIdsToRun {
			projectIdsArray = append(projectIdsArray, projectId)
		}
		err = db.Table("queries").Where("JSON_EXTRACT_STRING(query, 'cl') = 'channel_v1' AND project_id IN (?)", projectIdsArray).Find(&dbQueries).Error
		if err != nil {
			log.Warn("Failed during query fetch")
			os.Exit(0)
		}
	}

	count := 0
	for _, dbQuery := range dbQueries {
		queryClass, errMsg := store.GetStore().GetQueryClassFromQueries(dbQuery)
		var sourceFormatOfQuery model.ChannelGroupQueryV1

		if errMsg != "" {
			log.WithField("dbQuery", dbQuery).Warn("failed in GetQueryClassFromQueries. Hence skipping.")
			continue
		}
		if queryClass == model.QueryClassChannelV1 {
			err = U.DecodePostgresJsonbToStructType(&dbQuery.Query, &sourceFormatOfQuery)
			if err != nil {
				log.WithField("dbQuery", dbQuery).Warn("failed in DecodePostgresJsonbToStructType. Hence skipping.")
				continue
			}
		}
		finalResultantKPIQuery := model.TransformChannelsV1QueryToKPIQueryGroup(sourceFormatOfQuery)

		if *normalRun == false {
			result, _ := store.GetStore().ExecuteKPIQueryGroup(dbQuery.ProjectID, "", finalResultantKPIQuery, true)
			if result[0].Headers == nil || result[0].Rows == nil {
				log.WithField("time", U.TimeNowZ()).WithField("dbQuery", dbQuery).WithField("finalResultantKPIQuery", finalResultantKPIQuery).Warn("Failed in transforming channel v1 to kpi query.")
			} else {
				log.WithField("dbQuery", dbQuery).WithField("finalResultantKPIQuery", finalResultantKPIQuery).Warn("Successfully transformed channel v1 to kpi query.")
			}
		} else {
			KpiQueryInPostgresFormat, err := U.EncodeStructTypeToPostgresJsonb(finalResultantKPIQuery)
			if err != nil {
				log.WithField("dbQuery", dbQuery).Warn("failed in EncodeStructTypeToPostgresJsonb. Hence skipping.")
				continue
			}
			dbQuery.Query = *KpiQueryInPostgresFormat

			if err = db.Table("queries").Save(dbQuery).Error; err != nil {
				log.WithField("err", err).WithField("dbQuery", dbQuery).Warn("failed in saving transformed Query json Marshal. Hence skipping.")
				continue
			}
		}
		count += 1
	}
	log.WithField("count", count).Warn("Completed with count")
}
