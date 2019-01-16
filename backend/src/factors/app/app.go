package main

import (
	C "factors/config"
	H "factors/handler"
	mid "factors/middleware"
	"flag"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ./app --env=development --port=8080 --etcd=localhost:2379 --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --geo_loc_path=/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb --subdomain_enabled=true --subdomain_conf_path=/usr/local/var/factors/config/subdomain_login_config.json
func main() {

	env := flag.String("env", "development", "")
	port := flag.Int("port", 8080, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	subDomainLogicEnabled := flag.Bool("subdomain_enabled", true, "")
	subDomainLogicFilePath := flag.String("subdomain_conf_path", "/usr/local/var/factors/config/subdomain_login_config.json", "")

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
		SubdomainLogin: C.SubdomainLoginConfig{
			Enabled:        *subDomainLogicEnabled,
			ConfigFilepath: *subDomainLogicFilePath,
		},
	}

	// Initialize configs and connections.
	err := C.Init(config)
	if err != nil {
		log.Fatal("Failed to initialize.")
		return
	}

	r := gin.Default()
	// Group based middlewares should be registered on corresponding init methods.
	// Root middleware for cors.
	r.Use(mid.CustomCors())

	// Initialize routes.
	H.InitAppRoutes(r)
	H.InitSDKRoutes(r)
	H.InitIntRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
