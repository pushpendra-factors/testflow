package main

import (
	C "factors/config"
	"flag"
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"encoding/json"
	"factors/model/store"
	"strconv"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultHealthcheckPingID := C.HealthcheckCurrencyUploadPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)

	defaultAppName := "currency_conversion"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	todayDateinYYYYMMFormat := time.Unix(time.Now().Unix(), 0).UTC().Format("200601")
	todayDateinYYYYMMDDFormat := time.Unix(time.Now().Unix(), 0).UTC().Format("2006-01-01")
	exchangeRateUrl := "https://openexchangerates.org/api/historical/"+todayDateinYYYYMMDDFormat+".json?app_id=b61633badf274600a9b20b398e6c1768"
	fmt.Println(exchangeRateUrl)
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, exchangeRateUrl).
		WithHeader("Content-Type", "application/json")

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Failed to build request.")
	}

	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("Failed to make GET call for currency conversion.")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Failed to make GET call for currency conversion.")
	}

	var responseDetails map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseDetails)
	if err != nil {
		log.Errorf("Unable to decode response from GET request: %v", resp.Body)
	}
	log.Info(responseDetails["rates"])

	currency := responseDetails["rates"].(map[string]interface{})
	currencyOffsetWithINR := make(map[string]float64)
	offset := currency["INR"].(float64)
	for key, value := range currency {
		currencyOffsetWithINR[key] = value.(float64)/offset
	}
	dateAsInt64, _ := strconv.ParseInt(todayDateinYYYYMMFormat, 10, 64)
	allSuccess := true
	finalStatus := make(map[string]interface{})
	for key, value := range currencyOffsetWithINR {
		err := store.GetStore().CreateCurrencyDetails(key, dateAsInt64, value)
		if(err != nil){
			finalStatus[key+"err"] = err.Error()
			allSuccess = false
		} else {
			finalStatus[key] = value
		}
	}
	if allSuccess == false {
		C.PingHealthcheckForFailure(healthcheckPingID, finalStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, finalStatus)
	}
}