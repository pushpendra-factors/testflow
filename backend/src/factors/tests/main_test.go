package tests

import (
	C "factors/config"
	"flag"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	env := flag.String("env", "development", "")
	port := flag.Int("port", 8100, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")

	flag.Parse()

	config := &C.Configuration{
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
		GeolocationFile: *geoLocFilePath,
	}

	// Setup.
	// Initialize configs and connections.
	if err := C.Init(config); err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}
	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Fatal("Environment is not Development.")
		os.Exit(1)
	}
	retCode := m.Run()
	os.Exit(retCode)
}
