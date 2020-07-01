package config

import "C"
import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"factors/vendor_custom/machinery/v1"
	machineryConfig "factors/vendor_custom/machinery/v1/config"

	D "github.com/gamebtc/devicedetector"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	geoip2 "github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"

	U "factors/util"

	"factors/interfaces/maileriface"
	"factors/services/error_collector"
	serviceEtcd "factors/services/etcd"
	"factors/services/mailer"
	serviceSes "factors/services/ses"
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
	AppName                          string
	Env                              string
	Port                             int
	DBInfo                           DBConf
	RedisHost                        string
	RedisPort                        int
	QueueRedisHost                   string
	QueueRedisPort                   int
	EtcdEndpoints                    []string
	GeolocationFile                  string
	DeviceDetectorPath               string
	APIDomain                        string
	APPDomain                        string
	AWSRegion                        string
	AWSKey                           string
	AWSSecret                        string
	Cookiename                       string
	EmailSender                      string
	ErrorReportingInterval           int
	AdminLoginEmail                  string
	AdminLoginToken                  string
	FacebookAppID                    string
	FacebookAppSecret                string
	LoginTokenMap                    map[string]string
	SkipTrackProjectIds              []uint64
	SDKRequestQueueProjectTokens     []string
	SegmentRequestQueueProjectTokens []string
	MergeUspProjectIds               string
	SkipSessionProjectIds            string // comma seperated project ids, supports "*" for all projects.
}

type Services struct {
	Db                 *gorm.DB
	GeoLocation        *geoip2.Reader
	Etcd               *serviceEtcd.EtcdClient
	Redis              *redis.Pool
	QueueClient        *machinery.Server
	patternServersLock sync.RWMutex
	patternServers     map[string]string
	Mailer             maileriface.Mailer
	ErrorCollector     *error_collector.Collector
	DeviceDetector     *D.DeviceDetector
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

func initGeoLocationService(geoLocationFile string) {
	if geoLocationFile == "" {
		log.WithField("geo_location_file",
			geoLocationFile).Fatal("Invalid geolocation file.")
	}

	if services == nil {
		services = &Services{}
	}

	// Ref: https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
	geolocation, err := geoip2.Open(geoLocationFile)
	if err != nil {
		log.WithError(err).WithField("GeolocationFilePath",
			geoLocationFile).Fatal("Failed to initialize geolocation service.")
	}

	log.Info("Geolocation service intialized.")
	services.GeoLocation = geolocation
}
func initDeviceDetectorPath(deviceDetectorPath string) {
	if deviceDetectorPath == "" {
		log.WithField("dev_detect_path",
			deviceDetectorPath).Fatal("Invalid device detector path.")
	}
	if services == nil {
		services = &Services{}
	}
	deviceDetector, err := D.NewDeviceDetector(deviceDetectorPath)
	if err != nil {
		log.WithError(err).WithField("DeviceDetectorPath",
			deviceDetectorPath).Fatal("Failed to initialize device detector service.")
	}

	log.Info("Device Detector Path service intialized.")
	services.DeviceDetector = deviceDetector
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

	InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval)

	initGeoLocationService(config.GeolocationFile)
	initDeviceDetectorPath(config.DeviceDetectorPath)

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

func InitDBWithMaxIdleAndMaxOpenConn(dbConf DBConf,
	maxOpenConns, maxIdleConns int) error {

	if services == nil {
		services = &Services{}
	}

	db, err := gorm.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
			dbConf.Host,
			dbConf.Port,
			dbConf.User,
			dbConf.Name,
			dbConf.Password,
		))
	// Connection Pooling and Logging.
	db.DB().SetMaxOpenConns(maxOpenConns)
	db.DB().SetMaxIdleConns(maxIdleConns)
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
	configuration.DBInfo = dbConf
	return nil
}

func InitDB(dbConf DBConf) error {
	// default configuration.
	return InitDBWithMaxIdleAndMaxOpenConn(dbConf, 50, 10)
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
		MaxActive: 300,
		MaxIdle:   100,
		// IdleTimeout: 240 * time.Second,
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

func InitQueueClient(redisHost string, redisPort int) error {
	if services == nil {
		services = &Services{}
	}

	if redisHost == "" || redisPort == 0 {
		return fmt.Errorf("invalid redis host %s port %d", redisHost, redisPort)
	}

	// format: redis://[password@]host[port][/db_num]
	// Todo: Add password support for other environments.
	redisConnectionString := fmt.Sprintf("redis://%s:%d", redisHost, redisPort)

	config := &machineryConfig.Config{
		Broker: redisConnectionString,
		// No default queue configured. Queue name is decided conditionaly
		// and given on sendTask (enqueue) as routing_key and
		// on customer worker (dequeue) as queue_name.
		// DefaultQueue: "default_queue"
		Redis: &machineryConfig.RedisConfig{
			MaxActive: 300,
			MaxIdle:   100,
		},
		// Result Backend creates individual keys for each task
		// with the state after processing. Expiring the keys in 2 mins.
		// Retry is not using or affected by this, It is using a
		// seperate internal queue.
		ResultBackend:   redisConnectionString,
		ResultsExpireIn: 2 * 60,
		Debug:           !IsProduction(),
	}

	client, err := machinery.NewServer(config)
	if err != nil {
		return err
	}

	services.QueueClient = client

	return nil
}

func InitLogClient(env, appName, emailSender, awsKey, awsSecret,
	awsRegion string, reportingInterval int) {

	InitMailClient(awsKey, awsSecret, awsRegion)
	initCollectorClient(env, appName, "team@factors.ai", emailSender, reportingInterval)
	initLogging(services.ErrorCollector)
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

func initCollectorClient(env, appName, toMail, fromMail string, reportingInterval int) {
	if services == nil {
		services = &Services{}
	}
	dur := time.Second * time.Duration(reportingInterval)
	services.ErrorCollector = error_collector.New(services.Mailer, dur, env, appName, toMail, fromMail)
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
		log.WithField("PatternServers", GetServices().GetPatternServerAddresses()).Info(
			"Updated List of pattern servers")
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

	InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval)

	initiated = true
	return nil
}

func InitSDKService(config *Configuration) error {
	if initiated {
		return fmt.Errorf("Config already initialized")
	}

	configuration = config

	// DB dependency for SDK project_settings.
	if err := InitDB(config.DBInfo); err != nil {
		log.WithError(err).Error("Failed to initialize db on sdk_service.")
	}

	// Cache dependency for requests not using queue.
	InitRedis(config.RedisHost, config.RedisPort)

	initGeoLocationService(config.GeolocationFile)
	initDeviceDetectorPath(config.DeviceDetectorPath)

	err := InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize queue client on init sdk service.")
	}

	InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval)

	initiated = true
	return nil
}

func InitQueueWorker(config *Configuration) error {
	if initiated {
		return fmt.Errorf("Config already initialized")
	}

	configuration = config

	err := InitDB(config.DBInfo)
	if err != nil {
		return err
	}
	InitRedis(config.RedisHost, config.RedisPort)

	initGeoLocationService(config.GeolocationFile)
	initDeviceDetectorPath(config.DeviceDetectorPath)

	// Todo: Use different redis instance for queue for production env.
	err = InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize queue client on init queue worker.")
	}

	InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval)

	initiated = true
	return nil
}

func GetConfig() *Configuration {
	return configuration
}

func GetServices() *Services {
	return services
}

func GetCacheRedisConnection() redis.Conn {
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

func GetFacebookAppId() string {
	return configuration.FacebookAppID
}

func GetFacebookAppSecret() string {
	return configuration.FacebookAppSecret
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

func GetTokensFromStringListAsUint64(stringList string) []uint64 {
	uint64Tokens := make([]uint64, 0, 0)

	if stringList == "" {
		return uint64Tokens
	}

	tokens := strings.Split(stringList, ",")
	for _, token := range tokens {
		uint64Token, err := strconv.ParseUint(strings.TrimSpace(token), 10, 64)
		if err != nil {
			log.WithError(err).
				Error("Failed to parse project_id on string list config.")
			return uint64Tokens
		}

		uint64Tokens = append(uint64Tokens, uint64Token)
	}

	return uint64Tokens
}

func GetTokensFromStringListAsString(stringList string) []string {
	stringTokens := make([]string, 0, 0)

	if stringList == "" {
		return stringTokens
	}

	tokens := strings.Split(stringList, ",")
	for _, token := range tokens {
		stringTokens = append(stringTokens, strings.TrimSpace(token))
	}

	return stringTokens
}

func GetSDKRequestQueueAllowedTokens() []string {
	return configuration.SDKRequestQueueProjectTokens
}

func GetSegmentRequestQueueAllowedTokens() []string {
	return configuration.SegmentRequestQueueProjectTokens
}

/*
GetProjectsFromListWithAllProjectSupport -
If project list string is '*':
  Returns all_projects as true and empty allowed projects and disallowed projects.
else:
  Returns all_projects as false, given projects ids after skipping disallowed
	projects and disallowed projects.
Returns: allProject flag, list of allowed & disallowed and map of allowed & disallowed projects
*/
func GetProjectsFromListWithAllProjectSupport(projectIdsList,
	disallowedProjectIdsList string) (allProjects bool, allowedMap, disallowedMap map[uint64]bool) {
	//allowedProjectIds, skipProjectIds []uint64,
	disallowedProjectIdsList = strings.TrimSpace(disallowedProjectIdsList)
	skipProjectIds := GetTokensFromStringListAsUint64(disallowedProjectIdsList)

	disallowedMap = make(map[uint64]bool)
	for i := range skipProjectIds {
		disallowedMap[skipProjectIds[i]] = true
	}

	projectIdsList = strings.TrimSpace(projectIdsList)
	if projectIdsList == "*" {
		return true, map[uint64]bool{}, disallowedMap
	}

	projectIds := GetTokensFromStringListAsUint64(projectIdsList)

	allowedProjectIds := make([]uint64, 0, len(projectIds))
	for i, cpid := range projectIds {
		//Prioritizing the skip list over project list!
		if _, exists := disallowedMap[cpid]; !exists {
			allowedProjectIds = append(allowedProjectIds, projectIds[i])
		}
	}

	allowedMap = make(map[uint64]bool)
	for i := range allowedProjectIds {
		allowedMap[allowedProjectIds[i]] = true
	}

	return false, allowedMap, disallowedMap
}

func ProjectIdsFromProjectIdBoolMap(mp map[uint64]bool) []uint64 {

	keys := make([]uint64, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	return keys
}

func GetSkipSessionProjects() (allProjects bool, projectIds []uint64) {
	allProjects, projectIDsMap, _ := GetProjectsFromListWithAllProjectSupport(
		configuration.SkipSessionProjectIds, "")
	projectIds = ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	return allProjects, projectIds
}
