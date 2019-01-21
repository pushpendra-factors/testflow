package config

import (
	json "encoding/json"
	serviceEtcd "factors/services/etcd"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	geoip2 "github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

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
	EtcdEndpoints   []string             `json:"etcd_endpoints"`
	GeolocationFile string               `json:"geolocation_file"`
	SubdomainLogin  SubdomainLoginConfig `json:"subdomain_login"`
}

type Services struct {
	Db                 *gorm.DB
	GeoLocation        *geoip2.Reader
	Etcd               *serviceEtcd.EtcdClient
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

var configuration *Configuration
var services *Services = nil
var subdomainLoginCache *SubdomainLoginCache = nil

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

func initSubdomainLoginCache(config *Configuration) {
	subdomainLoginConfig := config.SubdomainLogin
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

func initServices(config *Configuration) error {
	services = &Services{patternServers: make(map[string]string)}

	err := InitDB(config.DBInfo)
	if err != nil {
		return err
	}

	err = InitEtcd(config.EtcdEndpoints)
	if err != nil {
		return err
	}

	// Ref: https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
	geolocation, err := geoip2.Open(config.GeolocationFile)
	if err != nil {
		log.WithError(err).WithField("GeolocationFilePath", config.GeolocationFile).Fatal("Failed to initialize geolocation service")
	}
	log.Info("Geolocation service intialized")
	services.GeoLocation = geolocation

	regPatternServers, err := GetServices().Etcd.DiscoverPatternServers()
	if err != nil && err != serviceEtcd.NotFound {
		log.WithError(err).Errorln("Falied to initialize discover pattern servers")
		return err
	}

	for _, ps := range regPatternServers {
		services.addPatternServer(ps.Key, ps.Value)
	}

	go func() {
		psUpdateChannel := GetServices().Etcd.Watch(serviceEtcd.PatternServerPrefix, clientv3.WithPrefix())
		watchPatternServers(psUpdateChannel)
	}()

	return nil
}

func InitConf(env string) {
	configuration = &Configuration{
		Env: env,
	}
}

func InitEtcd(EtcdEndpoints []string) error {
	etcdClient, err := serviceEtcd.New(EtcdEndpoints)
	if err != nil {
		log.WithError(err).Errorln("Falied to initialize etcd client")
		return err
	}
	log.Infof("ETCD Service Initialized with endpoints: %v", EtcdEndpoints)
	services.Etcd = etcdClient
	configuration.EtcdEndpoints = EtcdEndpoints
	return nil
}

func InitDB(DBInfo DBConf) error {
	if services == nil {
		services = &Services{}
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		DBInfo.Host,
		DBInfo.Port,
		DBInfo.User,
		DBInfo.Name,
		DBInfo.Password))
	// Connection Pooling and Logging.
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(true)

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed Db Initialization")
		return err
	}
	log.Info("Db Service initialized")
	services.Db = db
	configuration.DBInfo = DBInfo
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

func Init(config *Configuration) error {
	if initiated {
		return fmt.Errorf("Config already initialized")
	}

	configuration = config

	initLogging()

	initSubdomainLoginCache(config)

	err := initServices(config)
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
