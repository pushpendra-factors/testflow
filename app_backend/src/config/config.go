package config

import (
	json "encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var configFilePath = flag.String("config_filepath", "../config/config.json", "")

const DEVELOPMENT = "development"

type Configuration struct {
	Env    string `json:"env"`
	Port   int    `json:"port"`
	DbHost string `json:"db_host"`
	DbPort int    `json:"db_port"`
}
type Services struct {
	Db *gorm.DB
}

var configuration *Configuration = nil
var services *Services = nil

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
	raw, err := ioutil.ReadFile(configFileAbsPath)
	if err != nil {
		log.WithFields(log.Fields{"file": configFileAbsPath}).Error("Failed to load config")
		return err
	}

	json.Unmarshal(raw, &configuration)
	log.WithFields(log.Fields{"file": configFileAbsPath, "config": &configuration}).Info("Config File Loaded")
	return nil
}

func initServices() error {
	db, err := gorm.Open("sqlite3", "../gorm.db")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed Db Initialization")
		return err
	}
	log.Info("Db Service initialized")

	services = &Services{Db: db}
	return nil
}

func Init() error {
	initFlags()
	initLogging()
	err := initConfigFromFile()
	if err != nil {
		return err
	}

	err = initServices()
	if err != nil {
		return err
	}
	return nil
}

func GetConfig() *Configuration {
	return configuration
}

func GetServices() *Services {
	return services
}
