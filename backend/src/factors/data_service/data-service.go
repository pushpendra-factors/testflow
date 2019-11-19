package main

import (
	C "factors/config"
	"flag"
	"strconv"

	H "factors/handler"
	mid "factors/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", "development", "")
	port := flag.Int("port", 8089, "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	flag.Parse()

	config := &C.Configuration{
		Env:  *env,
		Port: *port,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
	}

	err := C.InitDataService(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}

	if !C.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(mid.RequestIdGenerator())
	r.Use(mid.Logger())
	r.Use(mid.Recovery())

	// Initialize routes.
	H.InitDataServiceRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
