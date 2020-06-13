package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/util"

	"factors/task/session"
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

	// projectIds: supports * (asterisk) for all projects.
	projectIds := flag.String("project_ids", "", "Allowed projects to create sessions offline.")
	numRoutines := flag.Int("num_routines", 1, "Number of routines to use.")
	maxLookbackDays := flag.Int64("max_lookback_days", 0, "Max lookback days to look for session existence.")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#AddSession"
	defer util.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "add_session",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:              *redisHost,
		RedisPort:              *redisPort,
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
	}

	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}

	// Cache dependency for requests not using queue.
	C.InitRedis(config.RedisHost, config.RedisPort)

	C.InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval)

	allowedProjectIds, errCode := session.GetAddSessionAllowedProjects(*projectIds)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get add session allowed project ids.")
		os.Exit(0)
	}

	var maxLookbackTimestamp int64
	if *maxLookbackDays > 0 {
		maxLookbackTimestamp = util.UnixTimeBeforeDuration(time.Hour * 24 * time.Duration(*maxLookbackDays))
	}

	status, _ := session.AddSession(allowedProjectIds, maxLookbackTimestamp, *numRoutines)
	if err := util.NotifyThroughSNS(taskID, *env, status); err != nil {
		log.Fatalf("Failed to notify status %+v", status)
	}

	log.WithField("no_of_projects", len(allowedProjectIds)).Info("Successfully added sessions.")
}
