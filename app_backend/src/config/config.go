package config

import (
	"bufio"
	json "encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	P "pattern"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var configFilePath = flag.String("config_filepath", "../config/config.json", "")
var initiated bool = false

const DEVELOPMENT = "development"

type Configuration struct {
	Env          string `json:"env"`
	Port         int    `json:"port"`
	DbHost       string `json:"db_host"`
	DbPort       int    `json:"db_port"`
	DbUser       string `json:"db_user"`
	DbName       string `json:"db_name"`
	DbPassword   string `json:"db_password"`
	PatternsFile string `json:"patterns_file"`
}
type Services struct {
	Db             *gorm.DB
	PatternService *P.PatternService
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
	// Connection Pooling and Logging.
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(true)

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed Db Initialization")
		return err
	}
	log.Info("Db Service initialized")

	patternsFileAbsPath, _ := filepath.Abs(configuration.PatternsFile)
	file, err := os.Open(patternsFileAbsPath)
	if err != nil {
		log.WithFields(log.Fields{"file": patternsFileAbsPath}).Error("Failed to load patterns")
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	patterns := []*P.Pattern{}
	for scanner.Scan() {
		line := scanner.Text()
		var pattern P.Pattern
		if err := json.Unmarshal([]byte(line), &pattern); err != nil {
			log.WithFields(log.Fields{"file": patternsFileAbsPath, "line": line}).Error("Failed to unmarshal pattern.")
			return err
		}
		patterns = append(patterns, &pattern)
	}
	patternService, err := P.NewPatternService(patterns)
	if err != nil {
		log.WithFields(log.Fields{"file": patternsFileAbsPath}).Error("Failed to load patterns")
	}

	services = &Services{Db: db, PatternService: patternService}
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
