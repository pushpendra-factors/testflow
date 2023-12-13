package main

import (
	"fmt"
	"flag"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	H "factors/handler/helpers"
	U "factors/util"
	"net/http"
	"os"
	"time"
	
	log "github.com/sirupsen/logrus"
)

/*
Plan of execution ------ BACKUP BACKUP BACKUP BACKUP BACKUP TODO ---
Release default data creation on production - appserver for new projects.
Step 1 - Delete the custom KPI, derived KPI and dashboard query - Closed Won
Step 2 - Delete the derived kpi - Deal Created -> Demo Completed.
Step 2 - Run Update queries for Pipeline and Deals.
Step 2 - Run migration script for Pipeline and Deals. Migrate existing saved queries of kpi and attribution to support new custom KPI name.
Step 3 - Run Create new queries.

- TODO check - Old query when comes from UI - dashboard and saved. It wont fail.

*/

/*
Step 1 queries
DELETE FROM queries where id = 2012641;
DELETE from dashboard_units where query_id = 2012641;
DELETE from from custom_metrics where project_id =  AND name IN ('HubSpot Overall Closed Won Deals (Close Date)', 'HubSpot Overall Closed Won Deals (Deal Create Date)', 'HubSpot Overall Closed Won Revenue (Close Date)', 'HubSpot Overall Closed Won Revenue (Deal Create Date)'); 
select * from queries where query like '%HubSpot Overall Closed Won%'; 

Step 2 Queries
UPDATE custom_metrics SET name = 'Pipeline', description = 'Pipeline' where project_id =  AND name = 'HubSpot Overall Pipeline';
UPDATE custom_metrics SET name = 'Deals', description = 'Deals' where project_id =  AND name = 'HubSpot Overall Deals';

Step3 Create new Queries
INSERT INTO custom_metrics
SELECT projects.id, UUID(), 'Revenue', 'Revenue', 1, 'hubspot_deals', '{"agFn":"sum","agPr":"$hubspot_deal_amount","agPrTy":"numerical","daFie":"$hubspot_deal_closedate","fil":[{"co":"equals","en":"user","lOp":"AND","objTy":"","prDaTy":"categorical","prNa":"$hubspot_deal_dealstage","va":"closedwon"}]}', NOW(), NOW()  
from projects


Backup Queries for Step 2

CREATE ROWSTORE TABLE `queries2` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `project_id` bigint(20) NOT NULL DEFAULT '0',
  `title` text CHARACTER SET utf8 COLLATE utf8_general_ci,
  `query` JSON COLLATE utf8_bin,
  `settings` JSON COLLATE utf8_bin,
  `type` int(11) DEFAULT NULL,
  `is_deleted` tinyint(1) NOT NULL DEFAULT 0,
  `created_by` text CHARACTER SET utf8 COLLATE utf8_general_ci,
  `created_at` timestamp(6) NOT NULL,
  `updated_at` timestamp(6) NOT NULL,
  `id_text` text CHARACTER SET utf8 COLLATE utf8_general_ci,
  `converted` tinyint(1) DEFAULT 0,
  `locked_for_cache_invalidation` tinyint(1) DEFAULT 0,
  SHARD KEY `project_id` (`project_id`),
  PRIMARY KEY (`project_id`,`id`),
  UNIQUE KEY `project_id_2` (`project_id`,`type`,`id`),
  KEY `updated_at` (`updated_at`)
) AUTOSTATS_CARDINALITY_MODE=PERIODIC AUTOSTATS_HISTOGRAM_MODE=CREATE SQL_MODE='STRICT_ALL_TABLES';

INSERT INTO queries2(id, project_id, title, query, settings, type, is_deleted, created_by, created_at, updated_at, id_text, converted, locked_for_cache_invalidation) 

SELECT * FROM queries where project_id = 51 AND (JSON_EXTRACT_STRING(query, 'cl') = 'insights'  OR JSON_EXTRACT_STRING(query, 'cl') = 'funnel') limit 50;
SELECT * FROM queries where project_id = 51 AND JSON_EXTRACT_STRING(query, 'query_group', 0, 'cl') = 'events' limit 50;
SELECT * FROM queries where project_id = 51 AND JSON_EXTRACT_STRING(query, 'cl') = 'kpi'  AND (query like '%HubSpot Overall Pipeline%' OR query like '%HubSpot Overall Deals%') limit 50;
SELECT * FROM queries where project_id = 51 AND JSON_EXTRACT_STRING(query, 'cl') = 'attribution' AND (query like '%HubSpot Overall Pipeline%' OR query like '%HubSpot Overall Deals%') limit 50;

*/

// Todo remove limit.
func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	projectIDsFlag := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDsFlag := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	// runUserObjectMigration := flag.Bool("run_user_object_migration", false, "")
	runSleep := flag.Bool("run_sleep", false, "")

	flag.Parse()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	}

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDsFlag, *disabledProjectIDsFlag)
	
	log.Info("Starting to initialize database.")
	appName := "run_saved_queries_migrate"
	
	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
		os.Exit(1)
	}
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	db := C.GetServices().Db
	defer db.Close()
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	projectIdsArray := make([]int64, 0)
	if !allProjects {
		for projectId, _ := range projectIdsToRun {
			projectIdsArray = append(projectIdsArray, projectId)
		}
	}

	updateSavedQueriesForKPI(allProjects, projectIdsArray, *runSleep)
	updateSavedQueriesForAttribution(allProjects, projectIdsArray, *runSleep)

}

func updateSavedQueriesForKPI(allProjects bool, projectIdsArray []int64, runSleep bool) {
	savedQueries := []model.Queries{}
	db := C.GetServices().Db

	query1 := db.Table("queries").Where("type = 2 AND JSON_EXTRACT_STRING(query, 'cl') = 'kpi' AND (query like '%HubSpot Overall Deals%' OR query like '%HubSpot Overall Pipeline%') AND is_deleted = 0 ")
	if !allProjects {
		query1 = query1.Where("project_id IN (?)", projectIdsArray)
	}

	if err := query1.Find(&savedQueries).Error; err !=nil {
		log.WithField("err", err.Error()).Warn("Failed during query fetch of saved queries - KPI")
		return
	}
	count := 0
	for _, dbQuery := range savedQueries {
		queryClass, errMsg := store.GetStore().GetQueryClassFromQueries(dbQuery)
		var internalQuery model.KPIQueryGroup
		if errMsg != "" {
			log.WithField("dbQuery", dbQuery).Warn("Failed in GetQueryClassFromQueries. Hence skipping - KPI")
			continue
		}

		if queryClass != model.QueryClassKPI {
			log.WithField("dbQuery", dbQuery).WithField("queryClass", queryClass).Warn("Failed 1 - Non KPI")
			continue
		} else {
			err := U.DecodePostgresJsonbToStructType(&dbQuery.Query, &internalQuery)
			if err != nil {
				log.WithField("dbQuery", dbQuery).Warn("Failed in DecodePostgresJsonbToStructType. Hence skipping - KPI")
				continue
			}
		}

		for index, query := range internalQuery.Queries {
			if query.Metrics[0] == "HubSpot Overall Deals" {
				internalQuery.Queries[index].Metrics[0] = "Deals"
			} else if query.Metrics[0] == "HubSpot Overall Pipeline" {
				internalQuery.Queries[index].Metrics[0] = "Pipeline"
			}
		}

		KpiQueryInPostgresFormat, err := U.EncodeStructTypeToPostgresJsonb(internalQuery)
			if err != nil {
				log.WithField("dbQuery", dbQuery).Warn("Failed in EncodeStructTypeToPostgresJsonb. Hence skipping - KPI")
				continue
			}
		dbQuery.Query = *KpiQueryInPostgresFormat

		if err := db.Table("queries").Save(&dbQuery).Error; err != nil {
			log.WithField("err", err).WithField("dbQuery", dbQuery).Warn("Failed in saving transformed Query json Marshal. Hence skipping - KPI")
			continue
		}
		statusCode := H.InValidateSavedQueryCache(&dbQuery)
		if statusCode != http.StatusOK {
			log.WithField("query_id", dbQuery.ID).Error("Failed in invalidating saved query cache - KPI")
		}

		count += 1

		if runSleep && count%5 == 0 {
			time.Sleep(30 * time.Second)
		}
	}
	log.WithField("count", count).Warn("Completed with count - KPI Query")

}

// Type 2 attribution can be ignored.
func updateSavedQueriesForAttribution(allProjects bool, projectIdsArray []int64, runSleep bool) {
	savedQueries := []model.Queries{}
	db := C.GetServices().Db

	query1 := db.Table("queries").Where("type = 3 AND JSON_EXTRACT_STRING(query, 'cl') = 'attribution' AND (query like '%HubSpot Overall Deals%' OR query like '%HubSpot Overall Pipeline%') AND is_deleted = 0 ")
	if !allProjects {
		query1 = query1.Where("project_id IN (?)", projectIdsArray)
	}
	if err := query1.Find(&savedQueries).Error; err !=nil {
		log.WithField("err", err.Error()).Warn("Failed during query fetch of saved queries - attr")
		return
	}

	count := 0
	for _, dbQuery := range savedQueries {
		queryClass, errMsg := store.GetStore().GetQueryClassFromQueries(dbQuery)
		var internalQuery model.AttributionQueryUnitV1
		if errMsg != "" {
			log.WithField("dbQuery", dbQuery).Warn("Failed in GetQueryClassFromQueries. Hence skipping - attr")
			continue
		}

		if queryClass != model.QueryClassAttribution {
			log.WithField("dbQuery", dbQuery).WithField("queryClass", queryClass).Warn("Failed - No attr")
			continue
		} else {
			err := U.DecodePostgresJsonbToStructType(&dbQuery.Query, &internalQuery)
			if err != nil {
				log.WithField("dbQuery", dbQuery).Warn("Failed in DecodePostgresJsonbToStructType. Hence skipping - attr")
				continue
			}
		}

		for attrQueryIndex, attrQuery := range internalQuery.Query.KPIQueries {
			for kpiIndex, kpiQuery := range attrQuery.KPI.Queries {
				if kpiQuery.Metrics[0] == "HubSpot Overall Deals" {
					internalQuery.Query.KPIQueries[attrQueryIndex].KPI.Queries[kpiIndex].Metrics[0] = "Deals"
				} else if kpiQuery.Metrics[0] == "HubSpot Overall Pipeline" {
					internalQuery.Query.KPIQueries[attrQueryIndex].KPI.Queries[kpiIndex].Metrics[0] = "Pipeline"
				}
			}
		}
		internalQuery.Query.ConversionEvent.Properties = make([]model.QueryProperty, 0) 
		internalQuery.Query.ConversionEventCompare.Properties = make([]model.QueryProperty, 0) 

		queryInPostgresFormat, err := U.EncodeStructTypeToPostgresJsonb(internalQuery)
		if err != nil {
			log.WithField("dbQuery", dbQuery).Warn("Failed in EncodeStructTypeToPostgresJsonb. Hence skipping - attr.")
			continue
		}
		dbQuery.Query = *queryInPostgresFormat

		if err := db.Table("queries").Save(&dbQuery).Error; err != nil {
			log.WithField("err", err).WithField("dbQuery", dbQuery).Warn("failed in saving transformed Query json Marshal. Hence skipping - attr.")
			continue
		}

		store.GetStore().DeleteAttributionDBResult(dbQuery.ProjectID, dbQuery.ID)

		count += 1

		if runSleep && count%5 == 0 {
			time.Sleep(30 * time.Second)
		}

	}
	log.WithField("count", count).Warn("Completed with count - Attribution")
}

// // Check if we need to give a sleep because it can bring down sdk processing - persistent cache.
// // We are going ahead with other approach of changing in saved query response.
// func updateGroupAnalysisForSavedEventQueries(allProjects bool, projectIdsArray []int64, runSleep bool) {
// 	savedQueries := []model.Queries{}
// 	db := C.GetServices().Db

// 	query1 := db.Table("queries").Where("type = 2 AND (JSON_EXTRACT_STRING(query, 'cl') = 'insights'  OR JSON_EXTRACT_STRING(query, 'cl') = 'funnel') AND is_deleted = 0 ")
// 	if !allProjects {
// 		query1 = query1.Where("project_id IN (?)", projectIdsArray)
// 	}
// 	query1 = query1.Limit(1000)
// 	if err := query1.Find(&savedQueries).Error; err !=nil {
// 		log.WithField("err", err.Error()).Warn("Failed during query fetch of saved queries")
// 		return
// 	}

// 	count := 0
// 	for _, dbQuery := range savedQueries {
// 		queryClass, errMsg := store.GetStore().GetQueryClassFromQueries(dbQuery)
// 		var internalQuery model.Query
// 		if errMsg != "" {
// 			log.WithField("dbQuery", dbQuery).Warn("Failed in GetQueryClassFromQueries. Hence skipping.")
// 			continue
// 		}

// 		if !(queryClass == model.QueryClassInsights || queryClass == model.QueryClassFunnel) {
// 			log.WithField("queryClass", queryClass).Warn("Failed - insights")
// 		} else {
// 			err := U.DecodePostgresJsonbToStructType(&dbQuery.Query, &internalQuery)
// 			if err != nil {
// 				log.WithField("dbQuery", dbQuery).Warn("Failed in DecodePostgresJsonbToStructType. Hence skipping - insights")
// 				continue
// 			}
// 		}

// 		if internalQuery.GroupAnalysis != "" {
// 			continue
// 		}

// 		if (internalQuery.Type == model.QueryTypeUniqueUsers ) {
// 			internalQuery.GroupAnalysis = "users"
// 		} else if (internalQuery.Type == model.QueryTypeEventsOccurrence) {
// 			internalQuery.GroupAnalysis = "events"
// 		} else {
// 			internalQuery.GroupAnalysis = "events"
// 		}

// 		queryInPostgresFormat, err := U.EncodeStructTypeToPostgresJsonb(internalQuery)
// 		if err != nil {
// 			log.WithField("dbQuery", dbQuery).Warn("Failed in EncodeStructTypeToPostgresJsonb. Hence skipping - insights")
// 			continue
// 		}
// 		dbQuery.Query = *queryInPostgresFormat

// 		if err := db.Table("queries").Save(&dbQuery).Error; err != nil {
// 			log.WithField("err", err).WithField("dbQuery", dbQuery).Warn("failed in saving transformed Query json Marshal. Hence skipping.")
// 			continue
// 		}
// 		statusCode := H.InValidateSavedQueryCache(&dbQuery)
// 		if statusCode != http.StatusOK {
// 			log.WithField("query_id", dbQuery.ID).Error("Failed in invalidating saved query cache.")
// 		}

// 		count += 1

// 		if runSleep && count%5 == 0 {
// 			time.Sleep(30 * time.Second)
// 		}

// 	}
// 	log.WithField("count", count).Warn("Completed with count - Insights")


// 	savedQueries2 := []model.Queries{}

// 	query2 := db.Table("queries").Where("type = 2 AND JSON_EXTRACT_STRING(query, 'query_group', 0, 'cl') = 'events' AND is_deleted = 0 ")
// 	if !allProjects {
// 		query2 = query2.Where("project_id IN (?)", projectIdsArray)
// 	}
// 	query2 = query2.Limit(2000)
// 	if err := query2.Find(&savedQueries2).Error; err !=nil {
// 		log.WithField("err", err.Error()).Warn("Failed during query fetch of saved queries - events ")
// 		return
// 	}

// 	count2 := 0
// 	for _, dbQuery := range savedQueries2 {
// 		queryClass, errMsg := store.GetStore().GetQueryClassFromQueries(dbQuery)
// 		var internalQuery model.QueryGroup
// 		if errMsg != "" {
// 			log.WithField("dbQuery", dbQuery).Warn("Failed in GetQueryClassFromQueries. Hence skipping - events")
// 			continue
// 		}

// 		if queryClass != model.QueryClassEvents {
// 			log.WithField("queryClass", queryClass).Warn("Failed  No  events")
// 		} else {
// 			err := U.DecodePostgresJsonbToStructType(&dbQuery.Query, &internalQuery)
// 			if err != nil {
// 				log.WithField("dbQuery", dbQuery).Warn("Failed in DecodePostgresJsonbToStructType. Hence skipping - events")
// 				continue
// 			}
// 		}

// 		if internalQuery.GroupAnalysis != "" {
// 			continue
// 		}


// 		for index, q := range internalQuery.Queries {
// 			if (q.Type == model.QueryTypeUniqueUsers ) {
// 				internalQuery.Queries[index].GroupAnalysis = "users"
// 			} else if (q.Type == model.QueryTypeEventsOccurrence) {
// 				internalQuery.Queries[index].GroupAnalysis = "events"
// 			} else {
// 				internalQuery.Queries[index].GroupAnalysis = "events"
// 			}
// 		}
// 		queryInPostgresFormat, err := U.EncodeStructTypeToPostgresJsonb(internalQuery)
// 		if err != nil {
// 			log.WithField("dbQuery", dbQuery).Warn("Failed in EncodeStructTypeToPostgresJsonb. Hence skipping - events")
// 			continue
// 		}
// 		dbQuery.Query = *queryInPostgresFormat

// 		if err := db.Table("queries").Save(&dbQuery).Error; err != nil {
// 			log.WithField("err", err).WithField("dbQuery", dbQuery).Warn("failed in saving transformed Query json Marshal. Hence skipping.")
// 			continue
// 		}
// 		statusCode := H.InValidateSavedQueryCache(&dbQuery)
// 		if statusCode != http.StatusOK {
// 			log.WithField("query_id", dbQuery.ID).Error("Failed in invalidating saved query cache.")
// 		}

// 		count2 += 1

// 		if runSleep && count2%5 == 0 {
// 			time.Sleep(30 * time.Second)
// 		}

// 	}
// 	log.WithField("count", count2).Warn("Completed with count - Events")

// }

/* Not Required.
func migrateNonDerivedCustomMetricToGroupUser(allProjects bool, projectIdsArray []int64, objectType, transformationString string)  {

	customMetrics := make([]model.CustomMetric, 0)

	query := db.Table("custom_metrics").Where("type_of_query = 1  AND (name = 'HubSpot Overall Closed Won Deals'  OR name = 'HubSpot Overall Closed Won Revenue') ");
	if !allProjects {
		query = query.Where(" AND project_id IN (?)", projectIdsArray)
	}
	if err := query.Find(&customMetrics).Error; err !=nil {
		log.WithField("err", err.Error()).Warn("Failed during query fetch of custom metrics")
	}

	for _, customMetric := range customMetrics {
		customMetric.ObjectType = "hubspot_deals"

		var transformation CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
		if err != nil {
			return false, "Error during decode of custom metrics transformations - custom_metrics handler."
		}
		transformation
	}
}
*/