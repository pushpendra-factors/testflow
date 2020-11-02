package tests

import (
	C "factors/config"
	"flag"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

var config *C.Configuration

func TestMain(m *testing.M) {
	env := flag.String("env", "development", "")
	port := flag.Int("port", 8100, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	deviceDetectorPath := flag.String("dev_detect_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")

	flag.Parse()

	config = &C.Configuration{
		Env:           *env,
		Port:          *port,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		GeolocationFile:     *geoLocFilePath,
		DeviceDetectorPath:  *deviceDetectorPath,
		APIDomain:           *apiDomain,
		APPDomain:           *appDomain,
	}

	// Setup.
	// Initialize configs and connections.
	if err := C.Init(config); err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitQueueClient(config.RedisHost, config.RedisPort)

	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Fatal("Environment is not Development.")
		os.Exit(1)
	}

	retCode := m.Run()
	os.Exit(retCode)
}
