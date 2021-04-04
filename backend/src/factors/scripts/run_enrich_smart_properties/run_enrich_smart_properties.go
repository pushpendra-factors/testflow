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
	projectIDs := flag.String("project_ids", "", "Projects for which the smart properties are to be populated")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		AppName: "enrich_smart_properties_job",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to enrich smart properties. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	projectIDMap := util.GetIntBoolMapFromStringList(projectIDs)
	if len(projectIDMap) > 0 {
		for projectID, _ := range projectIDMap {
			errCode := SP.EnrichSmartPropertiesForChangedRulesForProject(projectID)
			if errCode != http.StatusOK {
				log.Error("smart properties enrichment for rule changes failed for project ", projectID)
			}
			errCode = SP.EnrichSmartPropertiesForCurrentDayForProject(projectID)
			if errCode != http.StatusOK {
				log.Error("smart properties enrichment for current day's data failed for project ", projectID)
			}
		}
	} else {
		projectIDs, errCode := store.GetStore().GetProjectIDsHavingSmartPropertiesRules()
		if errCode != http.StatusFound {
			log.Warn("Failed to get any projects with smart properties rules")
		}
		for _, projectID := range projectIDs {
			errCode := SP.EnrichSmartPropertiesForChangedRulesForProject(projectID)
			if errCode != http.StatusOK {
				log.Error("smart properties enrichment for rule changes failed for project ", projectID)
			}
			errCode = SP.EnrichSmartPropertiesForCurrentDayForProject(projectID)
			if errCode != http.StatusOK {
				log.Error("smart properties enrichment for current day's data failed for project ", projectID)
			}
		}
	}

	log.Warn("End of enrich smart property job")
}
