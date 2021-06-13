package main

import (
	C "factors/config"
	"factors/model/store"
	SP "factors/task/smart_properties"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	dryRunSmartProperties := flag.Bool("dry_run_smart_properties", false, "Dry run mode for smart properties job")
	projectIDs := flag.String("project_ids", "", "Projects for which the smart properties are to be populated")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "enrich_smart_properties_job"
	defaultHealthcheckPingID := C.HealthCheckSmartPropertiesPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:      *primaryDatastore,
		DryRunSmartProperties: *dryRunSmartProperties,
	}
	C.InitConf(config)
	C.InitSmartPropertiesMode(config.DryRunSmartProperties)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to enrich smart properties. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	syncStatusFailures := make([]SP.Status, 0)
	syncStatusSuccesses := make([]SP.Status, 0)

	projectIDMap := util.GetIntBoolMapFromStringList(projectIDs)
	if len(projectIDMap) > 0 {
		for projectID, _ := range projectIDMap {
			errCode := SP.EnrichSmartPropertyForChangedRulesForProject(projectID)
			if errCode != http.StatusOK {
				errMsg := "smart properties enrichment for rule changes failed for project "
				log.Error(errMsg, projectID)
				syncStatusFailure := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    errMsg,
					Type:      SP.Rule_change,
				}
				syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
			} else {
				syncStatusSuccess := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    "",
					Type:      SP.Rule_change,
				}
				syncStatusSuccesses = append(syncStatusSuccesses, syncStatusSuccess)
			}
			errCode = SP.EnrichSmartPropertyForCurrentDayForProject(projectID)
			if errCode != http.StatusOK {
				errMsg := "smart properties enrichment for current day's data failed for project "
				log.Error(errMsg, projectID)
				syncStatusFailure := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    errMsg,
					Type:      SP.Current_day,
				}
				syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
			} else {
				syncStatusSuccess := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    "",
					Type:      SP.Current_day,
				}
				syncStatusSuccesses = append(syncStatusSuccesses, syncStatusSuccess)
			}
		}
	} else {
		projectIDs, errCode := store.GetStore().GetProjectIDsHavingSmartPropertyRules()
		if errCode != http.StatusFound {
			log.Warn("Failed to get any projects with smart properties rules")
		}
		for _, projectID := range projectIDs {
			errCode := SP.EnrichSmartPropertyForChangedRulesForProject(projectID)
			if errCode != http.StatusOK {
				errMsg := "smart properties enrichment for rule changes failed for project "
				log.Error(errMsg, projectID)
				syncStatusFailure := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    errMsg,
					Type:      SP.Rule_change,
				}
				syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
			} else {
				syncStatusSuccess := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    "",
					Type:      SP.Rule_change,
				}
				syncStatusSuccesses = append(syncStatusSuccesses, syncStatusSuccess)
			}
			errCode = SP.EnrichSmartPropertyForCurrentDayForProject(projectID)
			if errCode != http.StatusOK {
				errMsg := "smart properties enrichment for current day's data failed for project "
				log.Error(errMsg, projectID)
				syncStatusFailure := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    errMsg,
					Type:      SP.Current_day,
				}
				syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
			} else {
				syncStatusSuccess := SP.Status{
					ProjectID: projectID,
					ErrCode:   errCode,
					ErrMsg:    "",
					Type:      SP.Current_day,
				}
				syncStatusSuccesses = append(syncStatusSuccesses, syncStatusSuccess)
			}
		}
	}

	log.Warn("End of enrich smart property job")
	syncStatus := map[string]interface{}{
		"Success": syncStatusSuccesses,
		"Failure": syncStatusFailures,
	}
	if !*dryRunSmartProperties {
		if len(syncStatusFailures) > 0 {
			C.PingHealthcheckForFailure(healthcheckPingID, syncStatus)
			return
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, syncStatus)
	}
}
