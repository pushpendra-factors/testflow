package main

import (
	C "factors/config"
	DD "factors/default_data"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

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

	projectIDsFlag := flag.String("project_ids", "", "List of project_id to run for.")
	disabledProjectIDsFlag := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")

	flag.Parse()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	}

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDsFlag, *disabledProjectIDsFlag)

	log.Info("Starting to initialize database.")
	appName := "add_default_custom_metrics_for_segment_kpi"

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
		PrimaryDatastore:    *primaryDatastore,
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

	createCustomKPIs(allProjects, projectIdsArray)

}

func createCustomKPIs(allProjects bool, projectIdsArray []int64) {
	defaultHealthcheckPingID := C.HealthCheckPreBuiltCustomKPIPingID
	allowedProjectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_CUSTOM_METRICS, false)
	if err != nil {
		errString := "Failed in fetching projects with this feature flag enabled - events_cube_aggregation_deploy"
		log.WithField("err", err).Warn(errString)
		C.PingHealthcheckForFailure(defaultHealthcheckPingID, errString)
		os.Exit(1)
	}

	allowedProjectIdsMap := make(map[int64]bool)
	for _, projectID := range allowedProjectIDs {
		allowedProjectIdsMap[projectID] = true
	}

	// Re evaluate - once projects are run what happens in the following state.
	if allProjects {
		for projectID := range allowedProjectIdsMap {
			widgetGroupsPresent, err := CheckWidgetGroupsPresent(projectID)
			if err != "" {
				log.WithField("err", err).WithField("projectID", projectID).Warn("Errored during CheckWidgetGroupsPresent")
			}
			if !widgetGroupsPresent {
				_, errCode := store.GetStore().CreateWidgetGroups(projectID)
				if errCode != http.StatusCreated {
					log.WithField("err_code", errCode).Error("CreateProject Failed, Create widget groups failed.")
					continue
				}
			}

			present := store.GetStore().IsHubspotIntegrationAvailable(projectID)
			present2 := store.GetStore().IsSalesforceIntegrationAvailable(projectID)

			if present && present2 {
				log.WithField("projectID", projectID).Warn("Not Processing for this")
				continue
			}

			if present {
				buildMissingCustomKPI(projectID, DD.HubspotIntegrationName)
				buildWidgetGroup(projectID, DD.HubspotIntegrationName)
			}

			if present2 {
				buildMissingCustomKPI(projectID, DD.SalesforceIntegrationName)
				buildWidgetGroup(projectID, DD.SalesforceIntegrationName)
			}
		}
	} else {
		for _, projectID := range projectIdsArray {
			if _, exists := allowedProjectIdsMap[projectID]; exists {
				widgetGroupsPresent, errMsg := CheckWidgetGroupsPresent(projectID)
				if errMsg != "" {
					log.WithField("project", projectID).Warn(errMsg)
				}
				if !widgetGroupsPresent {
					_, errCode := store.GetStore().CreateWidgetGroups(projectID)
					if errCode != http.StatusCreated {
						log.WithField("err_code", errCode).Error("CreateProject Failed, Create widget groups failed.")
						continue
					}
				}

				present := store.GetStore().IsHubspotIntegrationAvailable(projectID)
				present2 := store.GetStore().IsSalesforceIntegrationAvailable(projectID)

				if present && present2 {
					log.WithField("projectID", projectID).Warn("Not Processing for this")
					continue
				}

				if present {
					buildMissingCustomKPI(projectID, DD.HubspotIntegrationName)
					buildWidgetGroup(projectID, DD.HubspotIntegrationName)
				}

				if present2 {
					buildMissingCustomKPI(projectID, DD.SalesforceIntegrationName)
					buildWidgetGroup(projectID, DD.SalesforceIntegrationName)
				}
			}
		}
	}
}

func buildMissingCustomKPI(projectID int64, integrationString string) {
	factory := DD.GetDefaultDataCustomKPIFactory(integrationString)
	statusCode2 := factory.Build(projectID)
	if statusCode2 != http.StatusOK {
		_, errMsg := fmt.Printf("Failed during prebuilt custom KPI creation for: %v, %v", projectID, integrationString)
		log.WithField("projectID", projectID).WithField("integration", integrationString).Warn(errMsg)
		// return "", http.StatusInternalServerError, "", "", false
	}
}

func CheckWidgetGroupsPresent(projectID int64) (bool, string) {
	widgetGroups, err, statusCode := store.GetStore().GetWidgetGroupAndWidgetsForConfig(projectID)
	if statusCode == http.StatusNotFound || len(widgetGroups) == 0 {
		return false, ""
	}
	if statusCode != http.StatusFound {
		return false, err
	}
	return true, ""
}

func buildWidgetGroup(projectID int64, integration string) {
	areWidgetsAdded, errMsg, statusCode4 := store.GetStore().AreWidgetsAddedToWidgetGroup(projectID)
	if statusCode4 != http.StatusFound {
		msg := fmt.Sprintf("Failed during fetch of AreWidgetsAddedToWidgetGroup: %s %v", errMsg, projectID)
		C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, msg)
		return
	}

	if areWidgetsAdded {
		return
	}

	if integration == DD.SalesforceIntegrationName {
		_, errMsg, statusCode := store.GetStore().AddWidgetsToWidgetGroup(projectID, memsql.MarketingEngagementWidgetGroup, model.SALESFORCE)
		if statusCode != http.StatusCreated {
			errMsg = errMsg + fmt.Sprintf(" %d", projectID)
			C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
			return
		}
		if statusCode == http.StatusCreated {
			_, errMsg, statusCode = store.GetStore().AddWidgetsToWidgetGroup(projectID, memsql.SalesOppWidgetGroup, model.SALESFORCE)
			if statusCode != http.StatusCreated {
				errMsg = errMsg + fmt.Sprintf(" %d", projectID)
				C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
				return
			}
		}
	}
	if integration == DD.HubspotIntegrationName {
		_, errMsg, statusCode := store.GetStore().AddWidgetsToWidgetGroup(projectID, memsql.MarketingEngagementWidgetGroup, model.HUBSPOT)
		if statusCode != http.StatusCreated {
			errMsg = errMsg + fmt.Sprintf(" %d", projectID)
			C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
		}
		if statusCode == http.StatusCreated {
			_, errMsg, statusCode = store.GetStore().AddWidgetsToWidgetGroup(projectID, memsql.SalesOppWidgetGroup, model.HUBSPOT)
			if statusCode != http.StatusCreated {
				errMsg = errMsg + fmt.Sprintf(" %d", projectID)
				C.PingHealthcheckForFailure(C.HealthCheckPreBuiltCustomKPIPingID, errMsg)
			}
		}
	}

}
