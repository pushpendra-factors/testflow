package config

import (
	json "encoding/json"
	P "factors/pattern"
	PS "factors/patternserver/store"
	serviceEtcd "factors/services/etcd"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
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
	Env                 string               `json:"env"`
	Port                int                  `json:"port"`
	DBInfo              DBConf               `json:"db"`
	ProjectModelMapping map[uint64]string    `json:"project_model_mapping"`
	EtcdEndpoints       []string             `json:"etcd_endpoints"`
	GeolocationFile     string               `json:"geolocation_file"`
	SubdomainLogin      SubdomainLoginConfig `json:"subdomain_login"`
}

type Services struct {
	Db             *gorm.DB
	GeoLocation    *geoip2.Reader
	PatternService *P.PatternService
	Etcd           *serviceEtcd.EtcdClient

	patternServersLock sync.RWMutex
	patternServers     map[string]string
}

func (service *Services) GetPatternServerAddresses() []string {
	service.patternServersLock.RLock()
	defer service.patternServersLock.RUnlock()

	ps := make([]string, 0, 0)
	for _, addr := range service.patternServers {
		ps = append(ps, addr)
	}
	return ps
}

func (service *Services) addPatternServer(key, addr string) {
	log.Infof("Add Pattern Server Key:%s, addr: %s", key, addr)
	service.patternServersLock.Lock()
	defer service.patternServersLock.Unlock()

	service.patternServers[key] = addr
}

func (service *Services) removePatternServer(key string) {
	log.Infof("Remove Pattern Server Key: %s", key)
	service.patternServersLock.Lock()
	defer service.patternServersLock.Unlock()

	delete(services.patternServers, key)
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

	if IsDevelopment() {
		log.SetLevel(log.DebugLevel)
	}

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
		logCtx.WithError(err).Error("Failed to load config")
		return err
	}

	if err := json.Unmarshal(raw, &configuration); err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal json")
		return err
	}
	logCtx.WithFields(log.Fields{"config": configuration}).Info("Config File Loaded")
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
	projectToUserAndEventsInfoMap := make(map[uint64]*P.UserAndEventsInfo)
	for projectId, modelId := range configuration.ProjectModelMapping {

		eventInfoFilePath := fmt.Sprintf("/tmp/factors-dev/projects/%v/models/%v/event_info_%v.txt", projectId, modelId, modelId)
		eventInfofile, err := os.Open(eventInfoFilePath)
		if err != nil {
			log.WithFields(log.Fields{"file": eventInfoFilePath}).Error("Failed to load eventInfoFile")
			return err
		}
		defer eventInfofile.Close()
		userAndEventsInfo, err := PS.CreatePatternEventInfoFromScanner(PS.CreateScannerFromFile(eventInfofile))
		if err != nil {
			log.WithError(err).WithField("file", eventInfoFilePath).Error("Failed to create eventInfo from File")
			return err
		}

		patternsFilePath := fmt.Sprintf("/tmp/factors-dev/projects/%v/models/%v/patterns_%v.txt", projectId, modelId, modelId)
		patternsfile, err := os.Open(patternsFilePath)
		if err != nil {
			log.WithError(err).WithField("file", patternsFilePath).Error("Failed to load patternsFile")
			return err
		}
		defer patternsfile.Close()

		patterns, err := PS.CreatePatternsFromScanner(PS.CreateScannerFromFile(patternsfile))
		if err != nil {
			log.WithError(err).WithField("file", patternsFilePath).Error("Failed to create patterns from File")
			return err
		}

		patternsMap[projectId] = patterns
		projectToUserAndEventsInfoMap[projectId] = &userAndEventsInfo
		log.Info(fmt.Sprintf("Loaded %d patterns for project %d", len(patterns), projectId))
	}

	patternService, err := P.NewPatternService(patternsMap, projectToUserAndEventsInfoMap)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize pattern service")
	}

	// Ref: https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
	geolocation, err := geoip2.Open(configuration.GeolocationFile)
	if err != nil {
		log.WithError(err).WithField("GeolocationFilePath", configuration.GeolocationFile).Fatal("Failed to initialize geolocation service")
	}
	log.Info("Geolocation service intialized")

	etcdClient, err := serviceEtcd.New(configuration.EtcdEndpoints)
	if err != nil {
		log.WithError(err).Errorln("Falied to initialize etcd client")
		return err
	}
	log.Infof("ETCD Service Initialized with endpoints: %v", configuration.EtcdEndpoints)

	services = &Services{Db: db, PatternService: patternService, Etcd: etcdClient, patternServers: make(map[string]string), GeoLocation: geolocation}

	regPatternServers, err := etcdClient.DiscoverPatternServers()
	if err != nil && err != serviceEtcd.NotFound {
		log.WithError(err).Errorln("Falied to initialize discover pattern servers")
		return err
	}

	for _, ps := range regPatternServers {
		services.addPatternServer(ps.Key, ps.Value)
	}

	go func() {
		psUpdateChannel := etcdClient.Watch(serviceEtcd.PatternServerPrefix, clientv3.WithPrefix())
		watchPatternServers(psUpdateChannel)
	}()

	return nil
}

func watchPatternServers(psUpdateChannel clientv3.WatchChan) {
	log.Infoln("Starting to watch on psUpdateChannel")
	for {
		msg := <-psUpdateChannel
		for _, event := range msg.Events {
			log.WithFields(log.Fields{
				"Type":  event.Type,
				"Key":   string(event.Kv.Key),
				"Value": string(event.Kv.Value),
			}).Infoln("Event Received on PatternServerUpdateChannel")

			if event.Type == mvccpb.PUT {
				GetServices().addPatternServer(string(event.Kv.Key), string(event.Kv.Value))
			} else if event.Type == mvccpb.DELETE {
				GetServices().removePatternServer(string(event.Kv.Key))
			}
		}
		log.WithField("PatternServers", GetServices().GetPatternServerAddresses()).Info("Updated List of pattern servers")
	}
}

func Init() error {
	if initiated {
		return fmt.Errorf("Config already initialized")
	}
	initFlags()
	err := initConfigFromFile()
	if err != nil {
		return err
	}
	initLogging()

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
