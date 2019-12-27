package config

import (
	"factors/interfaces/maileriface"
	"factors/services/error_collector"
	serviceEtcd "factors/services/etcd"
	"factors/services/mailer"
	serviceSes "factors/services/ses"
	U "factors/util"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	geoip2 "github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

var initiated bool = false

const DEVELOPMENT = "development"
const STAGING = "staging"
const PRODUCTION = "production"

// Warning: Any changes to the cookie name has to be
// in sync with other services which uses the cookie.
const FactorsSessionCookieName = "factors-sid"

type DBConf struct {
	Host     string
	Port     int
	User     string
	Name     string
	Password string
}

type Configuration struct {
	Env                    string
	Port                   int
	DBInfo                 DBConf
	RedisHost              string
	RedisPort              int
	EtcdEndpoints          []string
	GeolocationFile        string
	APIDomain              string
	APPDomain              string
	AWSRegion              string
	AWSKey                 string
	AWSSecret              string
	Cookiename             string
	EmailSender            string
	ErrorReportingInterval int
	AdminLoginEmail        string
	AdminLoginToken        string
	LoginTokenMap          map[string]string
	SkipTrackProjectIds    []uint64
}

type Services struct {
	Db                 *gorm.DB
	GeoLocation        *geoip2.Reader
	Etcd               *serviceEtcd.EtcdClient
	Redis              *redis.Pool
	patternServersLock sync.RWMutex
	patternServers     map[string]string
	Mailer             maileriface.Mailer
	ErrorCollector     *error_collector.Collector
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

var configuration *Configuration
var services *Services = nil

func initLogging(collector *error_collector.Collector) {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	if IsDevelopment() {
		log.SetLevel(log.DebugLevel)
	}

	log.SetReportCaller(true)

	if collector != nil {
		hook := &U.Hook{C: services.ErrorCollector}
		log.AddHook(hook)
	}
}

func initServices(config *Configuration) error {
	services = &Services{patternServers: make(map[string]string)}

	err := InitDB(config.DBInfo)
	if err != nil {
		return err
	}

	InitRedis(config.RedisHost, config.RedisPort)

	err = InitEtcd(config.EtcdEndpoints)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize etcd")
	}

	InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)

	initCollectorClient(config.Env, "team@factors.ai", config.EmailSender)

	// Ref: https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
	geolocation, err := geoip2.Open(config.GeolocationFile)
	if err != nil {
		log.WithError(err).WithField("GeolocationFilePath",
			config.GeolocationFile).Fatal("Failed to initialize geolocation service")
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

	initCookieInfo(configuration.Env)

	return nil
}

func initCookieInfo(env string) {
	// Warning: Any changes to the cookie name has to be
	// in sync with other services which uses the cookie.

	cookieName := fmt.Sprintf("%s%s", FactorsSessionCookieName, "d")

	if env == STAGING {
		cookieName = fmt.Sprintf("%s%s", FactorsSessionCookieName, "s")
	} else if env == PRODUCTION {
		cookieName = FactorsSessionCookieName
	}

	configuration.Cookiename = cookieName
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
	if IsDevelopment() {
		db.LogMode(true)
	} else {
		db.LogMode(false)
	}

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed Db Initialization")
		return err
	}
	log.Info("Db Service initialized")
	services.Db = db
	configuration.DBInfo = DBInfo
	return nil
}

func InitRedis(host string, port int) {
	if host == "" || port == 0 {
		log.WithField("host", host).WithField("port", port).Fatal(
			"Invalid redis host or port.")
	}

	if services == nil {
		services = &Services{}
	}

	conn := fmt.Sprintf("%s:%d", host, port)
	redisPool := &redis.Pool{
		MaxActive:   100,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conn)
			if err != nil {
				// do not panic. connection dial would be called
				// on pool refill too.
				log.WithError(err).Error("Redis connection dial error.")
				return nil, err
			}

			return c, err
		},

		// Tests connection before idle connection being reused.
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
			if err != nil {
				log.WithError(err).Error("Redis connection test on borrow error.")
			}

			return err
		},
	}

	log.Info("Redis Service initialized.")
	configuration.RedisHost = host
	configuration.RedisPort = port
	services.Redis = redisPool
}

func InitMailClient(key, secret, region string) {
	if services == nil {
		services = &Services{}
	}
	if IsDevelopment() {
		services.Mailer = mailer.New()
		return
	}
	services.Mailer = serviceSes.New(key, secret, region)
}

func InitSenderEmail(senderEmail string) {
	if services == nil {
		services = &Services{}
	}
	configuration.EmailSender = senderEmail
}

func initCollectorClient(env, toMail, fromMail string) {
	if services == nil {
		services = &Services{}
	}
	dur := time.Second * time.Duration(configuration.ErrorReportingInterval)
	services.ErrorCollector = error_collector.New(services.Mailer, dur, env, toMail, fromMail)
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

	err := initServices(config)
	if err != nil {
		return err
	}

	initLogging(services.ErrorCollector)

	initiated = true
	return nil
}

func InitDataService(config *Configuration) error {
	if initiated {
		return fmt.Errorf("Config already initialized")
	}

	configuration = config
	err := InitDB(config.DBInfo)
	if err != nil {
		return err
	}

	InitRedis(config.RedisHost, config.RedisPort)

	// init error collector, error mailer, and log hook.
	InitMailClient(config.AWSKey, config.AWSSecret, config.AWSRegion)
	initCollectorClient(config.Env, "team@factors.ai", config.EmailSender) // inits error_collector.
	initLogging(services.ErrorCollector)

	initiated = true
	return nil
}

func GetConfig() *Configuration {
	return configuration
}

func GetServices() *Services {
	return services
}

func GetRedisConn() redis.Conn {
	return services.Redis.Get()
}

func IsDevelopment() bool {
	return (strings.Compare(configuration.Env, DEVELOPMENT) == 0)
}

func IsStaging() bool {
	return (strings.Compare(configuration.Env, STAGING) == 0)
}

func IsProduction() bool {
	return (strings.Compare(configuration.Env, PRODUCTION) == 0)
}

func GetAPPDomain() string {
	return configuration.APPDomain
}

func GetAPIDomain() string {
	return configuration.APIDomain
}

func UseSecureCookie() bool {
	return !IsDevelopment()
}

func UseHTTPOnlyCookie() bool {
	return !IsDevelopment()
}

func GetProtocol() string {
	if IsDevelopment() {
		return "http://"
	}
	return "https://"
}

func GetFactorsSenderEmail() string {
	return configuration.EmailSender
}

func GetCookieDomian() string {
	domain := GetAPIDomain()
	port := ":" + strconv.Itoa(configuration.Port)
	if strings.Contains(domain, port) {
		return strings.Split(domain, port)[0]
	}
	return domain
}

func GetFactorsCookieName() string {
	return configuration.Cookiename
}

func GetSkipTrackProjectIds() []uint64 {
	return configuration.SkipTrackProjectIds
}

// ParseConfigStringToMap - Parses config string
// "k1:v1,k2:v2"-> map[string]string{k1: v1, k2: v2}
func ParseConfigStringToMap(configStr string) map[string]string {
	configMap := make(map[string]string, 0)

	if configStr == "" {
		return configMap
	}

	commaSplit := strings.Split(configStr, ",")
	if len(commaSplit) == 0 {
		return configMap
	}

	for _, cs := range commaSplit {
		kv := strings.Split(cs, ":")
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			configMap[kv[0]] = kv[1]
		}
	}

	return configMap
}

func GetProjectIdsFromStringList(projectsList string) []uint64 {
	projectIds := make([]uint64, 0, 0)

	if projectsList == "" {
		return projectIds
	}

	tokens := strings.Split(projectsList, ",")
	for _, token := range tokens {
		projectId, err := strconv.ParseUint(strings.TrimSpace(token), 10, 64)
		if err != nil {
			return projectIds
		}

		projectIds = append(projectIds, projectId)
	}
	return projectIds
}
