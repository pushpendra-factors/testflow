package config

import (
	json "encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var configFilePath = flag.String("config_filepath", "../config/config.json", "")
var initiated bool = false

const DEVELOPMENT = "development"

type Configuration struct {
	Env        string `json:"env"`
	Port       int    `json:"port"`
	DbHost     string `json:"db_host"`
	DbPort     int    `json:"db_port"`
	DbUser     string `json:"db_user"`
	DbName     string `json:"db_name"`
	DbPassword string `json:"db_password"`
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
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		configuration.DbHost,
		configuration.DbPort,
		configuration.DbUser,
		configuration.DbName,
		configuration.DbPassword))
	// Connection Pooling.
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed Db Initialization")
		return err
	}
	log.Info("Db Service initialized")

	services = &Services{Db: db}
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

func GetServices() *Services {
	return services
}
