package config

import (
	"bufio"
	json "encoding/json"
	P "factors/pattern"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
)

var configFilePath = flag.String("config_filepath", "../config/config.json", "")
var initiated bool = false

const DEVELOPMENT = "development"

type DBConf struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type SubdomainLoginConfig struct {
	Enabled        bool   `json:"enabled"`
	ConfigFilepath string `json:"config_filepath"`
}

type Configuration struct {
	Env             string               `json:"env"`
	Port            int                  `json:"port"`
	DBInfo          DBConf               `json:"db"`
	PatternFiles    map[uint64]string    `json:"pattern_files"`
	GeolocationFile string               `json:"geolocation_file"`
	SubdomainLogin  SubdomainLoginConfig `json:"subdomain_login"`
}

type Services struct {
	Db             *gorm.DB
	GeoLocation    *geoip2.Reader
	PatternService *P.PatternService
}

type SubdomainLoginCache struct {
	Map map[string][]uint64 `json:"token_projects"`
}

var configuration *Configuration = nil
var services *Services = nil
var subdomainLoginCache *SubdomainLoginCache = nil

func initFlags() {
	flag.Parse()
}

func initLogging() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	// log.SetLevel(log.WarnLevel)
}

func initConfigFromFile() error {

	configFileAbsPath, _ := filepath.Abs(*configFilePath)

	logCtx := log.WithFields(log.Fields{
		"file": configFileAbsPath,
	})

	raw, err := ioutil.ReadFile(configFileAbsPath)
	if err != nil {
		logCtx.WithError(err).Fatal("Failed to load config")
	}

	if err := json.Unmarshal(raw, &configuration); err != nil {
		logCtx.WithError(err).Fatal("Failed to unmarshal json")
	}
	logCtx.WithFields(log.Fields{"config": &configuration}).Info("Config File Loaded")
	return nil
}

func initSubdomainLoginCache() {
	subdomainLoginConfig := GetConfig().SubdomainLogin
	if !subdomainLoginConfig.Enabled {
		return
	}

	raw, err := ioutil.ReadFile(subdomainLoginConfig.ConfigFilepath)
	if err != nil {
		log.WithFields(log.Fields{"config": subdomainLoginConfig,
			"err": err}).Fatal("Failed reading subdomain login config file.")
	}

	// Loading cache.
	if err := json.Unmarshal(raw, &subdomainLoginCache); err != nil {
		log.WithFields(log.Fields{"config": subdomainLoginConfig,
			"err": err}).Fatal("Failed to unmarshal subdomain login config file.")
	}

	log.WithFields(log.Fields{"cache": &subdomainLoginCache}).Info("Initialized subdomain login cache.")
}

func initServices() error {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		configuration.DBInfo.Host,
		configuration.DBInfo.Port,
		configuration.DBInfo.User,
		configuration.DBInfo.Name,
		configuration.DBInfo.Password))
	// Connection Pooling and Logging.
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(true)

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed Db Initialization")
		return err
	}
	log.Info("Db Service initialized")

	patternsMap := make(map[uint64][]*P.Pattern)
	projectEventInfoMap := make(map[uint64]*P.EventInfoMap)
	for projectId, patternsFile := range configuration.PatternFiles {
		patterns := []*P.Pattern{}
		var eventInfoMap P.EventInfoMap

		patternsFileAbsPath, _ := filepath.Abs(patternsFile)
		file, err := os.Open(patternsFileAbsPath)
		if err != nil {
			log.WithFields(log.Fields{"file": patternsFileAbsPath}).Error("Failed to load patterns")
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		// Adjust scanner buffer capacity to 10MB per line.
		const maxCapacity = 10 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if lineNum == 1 {
				// First line is all the event and event properties information
				// seen in the data.
				if err := json.Unmarshal([]byte(line), &eventInfoMap); err != nil {
					log.WithFields(log.Fields{
						"file": patternsFileAbsPath, "lineNum": lineNum, "err": err}).Error(
						"Failed to unmarshal events info.")
					return err
				}
			} else {
				var pattern P.Pattern
				if err := json.Unmarshal([]byte(line), &pattern); err != nil {
					log.WithFields(log.Fields{
						"file": patternsFileAbsPath, "lineNum": lineNum, "err": err}).Error(
						"Failed to unmarshal pattern.")
					return err
				}
				patterns = append(patterns, &pattern)
			}
		}
		err = scanner.Err()
		if err != nil {
			log.WithFields(log.Fields{"err": err, "file": patternsFileAbsPath}).Error("Scanner error")
			return err
		}
		patternsMap[projectId] = patterns
		projectEventInfoMap[projectId] = &eventInfoMap
		log.Info(fmt.Sprintf("Loaded %d patterns for project %d", len(patterns), projectId))
	}

	patternService, err := P.NewPatternService(patternsMap, projectEventInfoMap)
	if err != nil {
		log.Fatal("Failed to initialize pattern service")
	}

	// Ref: https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
	geolocation, err := geoip2.Open(configuration.GeolocationFile)
	if err != nil {
		log.Fatal("Failed to initialize geolocation service. Falied opening geolocation db file")
	}
	log.Info("Geolocation service intialized")

	services = &Services{Db: db, PatternService: patternService, GeoLocation: geolocation}
	return nil
}

func Init() error {
	if initiated {
		return fmt.Errorf("Config already initialized")
	}
	initFlags()
	initLogging()
	err := initConfigFromFile()
	if err != nil {
		return err
	}

	initSubdomainLoginCache()

	err = initServices()
	if err != nil {
		return err
	}

	initiated = true
	return nil
}

func GetConfig() *Configuration {
	return configuration
}

func GetLoginTokenCache() *SubdomainLoginCache {
	return subdomainLoginCache
}

func GetServices() *Services {
	return services
}

func IsDevelopment() bool {
	return (strings.Compare(configuration.Env, DEVELOPMENT) == 0)
}

func IsTokenLoginEnabled() bool {
	return GetConfig().SubdomainLogin.Enabled
}
