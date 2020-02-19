package main

import (
	C "factors/config"
	H "factors/handler"

	"flag"
	"strconv"

	mid "factors/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")
	port := flag.Int("port", 8090, "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	flag.Parse()

	config := &C.Configuration{
		Env:                    *env,
		Port:                   *port,
		RedisHost:              *redisHost,
		RedisPort:              *redisPort,
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
	}

	err := C.InitSDKService(config)
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

	H.InitSDKRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
