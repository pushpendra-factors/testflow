package main

import (
	C "factors/config"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	IntHubspot "factors/integration/hubspot"
	M "factors/model"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	// init DB, etcd
	config := &C.Configuration{
		AppName: "hubspot_enrich_job",
		Env:     *env,
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

	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Fatal("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)

	hubspotEnabledProjectSettings, errCode := M.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Fatal("No projects enabled hubspot integration.")
	}

	statusList := make([]IntHubspot.Status, 0, 0)
	for _, settings := range hubspotEnabledProjectSettings {
		status := IntHubspot.Sync(settings.ProjectId)
		statusList = append(statusList, status...)
	}

	err = util.NotifyThroughSNS("hubspot_enrich", *env, statusList)
	if err != nil {
		log.WithError(err).Fatal("Failed to notify through SNS on hubspot sync.")
	}
}
