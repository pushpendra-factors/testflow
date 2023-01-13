package main

import (
	"errors"
	C "factors/config"
	"factors/filestore"
	PC "factors/pattern_client"
	patternserver "factors/pattern_server"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceGCS "factors/services/gcstorage"
	"flag"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/rpc"

	rpcjson "github.com/gorilla/rpc/json"

	"factors/model/model"
	"factors/model/store"
	modelstore "factors/model/store"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	InitialDefaultMetadata = "version1"
	DefaultTTLSeconds      = 10
	Development            = "development"
	Staging                = "staging"
	Production             = "production"
)

// Add Test
func makeProjectModelChunkLookup(projectDatas []model.ProjectModelMetadata) map[int64]patternserver.ModelChunkMapping {
	pD := make(map[int64]patternserver.ModelChunkMapping)
	for _, p := range projectDatas {
		mCM, exists := pD[p.ProjectId]
		if !exists {
			mCM = patternserver.ModelChunkMapping{}
		}
		mD := patternserver.ModelData{Type: p.ModelType, Chunks: strings.Split(p.Chunks, ","), StartTimestamp: p.StartTime, EndTimestamp: p.EndTime}
		mCM[p.ModelId] = mD
		pD[p.ProjectId] = mCM
	}
	return pD
}

type config struct {
	Environment   string
	IP            string
	RPCPort       string
	HTTPPort      string
	EtcdEndpoints []string
	BucketName    string
	DiskBaseDir   string
	NewBucketName string
}

func isValidEnv(env string) bool {
	return env == Development || env == Staging || env == Production
}

func NewConfig(env, ip, rpcPort, httpPort, etcd, diskBaseDir, bucketName, newBucketName string) (*config, error) {
	if !isValidEnv(env) {
		return nil, errors.New("Invalid Environment")
	}
	if ip == "" {
		return nil, errors.New("Invalid IP")
	}
	if rpcPort == "" {
		return nil, errors.New("Invalid RPCPort")
	}

	if httpPort == "" {
		return nil, errors.New("Invalid HTTPPort")
	}

	if diskBaseDir == "" {
		return nil, errors.New("Invalid baseDiskDir")
	}

	if bucketName == "" {
		return nil, errors.New("Invalid BucketName")
	}

	if newBucketName == "" {
		return nil, errors.New("Invalid BucketName")
	}

	etcds := strings.Split(etcd, ",")
	if len(etcds) == 0 {
		return nil, errors.New("Invalid EtcdEndpoints")
	}

	c := config{
		Environment:   env,
		IP:            ip,
		RPCPort:       rpcPort,
		HTTPPort:      httpPort,
		EtcdEndpoints: etcds,
		DiskBaseDir:   diskBaseDir,
		BucketName:    bucketName,
		NewBucketName: newBucketName,
	}

	return &c, nil
}

func (c *config) GetEnvironment() string {
	return c.Environment
}

func (c *config) IsDevelopment() bool {
	return c.Environment == Development
}

func (c *config) GetIP() string {
	return c.IP
}

func (c *config) GetRPCPort() string {
	return c.RPCPort
}

func (c *config) GetHTTPPort() string {
	return c.HTTPPort
}

func (c *config) GetBaseDiskDir() string {
	return c.DiskBaseDir
}

func (c *config) GetEtcdEndpoints() []string {
	return c.EtcdEndpoints
}

func (c *config) GetBucketName() string {
	return c.BucketName
}

func (c *config) GetNewBucketName(new bool) string {
	if new {
		return c.NewBucketName
	}
	return c.BucketName
}

// Process crashes if started with IP and port already registered with etcd.
// TTL on etcd is 10 seconds.
// Monit / Kubernetes will keep trying to restart the process, it should succeed after 10 seconds / till key expires.

// ./pattern-app --env=development --ip=127.0.0.1 --ps_rpc_port=8100 --ps_http_port=8101 --etcd=localhost:2379 --disk_dir=/usr/local/var/factors/local_disk --bucket_name=/usr/local/var/factors/cloud_storage --chunk_cache_size=5 --event_info_cache_size=10 --aws_region=us-east-1 --aws_key=dummy --aws_secret=dummy --email_sender=support@factors.ai --error_reporting_interval=300
func main() {

	env := flag.String("env", Development, "")
	ip := flag.String("ip", "127.0.0.1", "")
	rpc_port := flag.String("ps_rpc_port", "8100", "")
	http_port := flag.String("ps_http_port", "8101", "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	diskBaseDir := flag.String("disk_dir", "/usr/local/var/factors/local_disk", "")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	bucketNameV2 := flag.String("bucket_name_v2", "/usr/local/var/factors/cloud_storage_models", "")
	useBucketV2 := flag.Bool("use_bucket_v2", false, "Whether to use new bucketing system or not")
	projectIdV2 := flag.String("project_ids_v2", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	chunkCacheSize := flag.Int("chunk_cache_size", 5, "")
	eventInfoCacheSize := flag.Int("event_info_cache_size", 10, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	flag.Parse()

	config, err := NewConfig(*env, *ip, *rpc_port, *http_port, *etcd, *diskBaseDir, *bucketName, *bucketNameV2)
	if err != nil {
		panic(err)
	}

	appName := "pattern_server"
	dbConfig := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Certificate: *memSQLCertificate,
			Password:    *memSQLPass,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(dbConfig)

	C.InitSentryLogging(*sentryDSN, appName)
	defer C.SafeFlushAllCollectors()
	err = C.InitDB(*dbConfig)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()
	// TODO(Ankit):
	// This needs to be handled with graceful shutdown
	// defer ErrorCollector.Flush()

	log.SetReportCaller(true)
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	log.WithFields(log.Fields{
		"IP":            config.GetIP(),
		"Port":          config.GetRPCPort(),
		"Env":           config.GetEnvironment(),
		"EtcdEndpoints": config.GetEtcdEndpoints(),
		"DiskBaseDir":   config.GetBaseDiskDir(),
		"BucketName":    config.GetBucketName(),
	}).Infoln("Initialising with config")

	if config.IsDevelopment() {
		log.SetLevel(log.DebugLevel)
	}

	logCtx := log.WithFields(log.Fields{
		"IP":   config.GetIP(),
		"Port": config.GetRPCPort(),
	})

	logCtx.WithFields(log.Fields{
		"EtcdEndpoints": config.GetEtcdEndpoints(),
	}).Infoln("Pattern Server")

	etcdClient, err := serviceEtcd.New(config.GetEtcdEndpoints())
	if err != nil {
		logCtx.WithError(err).Errorln("Falied to initialize etcd client")
		panic(err)
	}

	if isRegistered, err := etcdClient.IsRegistered(config.GetIP(), config.GetRPCPort()); err != nil {
		logCtx.WithError(err).Errorln("Falied to check registered status with etcd")
		panic(err)
	} else if isRegistered {
		str := fmt.Sprintf("Pattern Server Already registered with IP:Port [ %s:%s ]", config.GetIP(), config.GetRPCPort())
		logCtx.Errorln(str)
		panic(str)
	}

	var cloudManager filestore.FileManager
	var newCloudManager filestore.FileManager
	if config.IsDevelopment() {
		cloudManager = serviceDisk.New(config.GetBucketName())
		newCloudManager = serviceDisk.New(config.GetNewBucketName(*useBucketV2))
	} else {
		cloudManager, err = serviceGCS.New(config.GetBucketName())
		if err != nil {
			logCtx.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
		newCloudManager, err = serviceGCS.New(config.GetNewBucketName(*useBucketV2))
		if err != nil {
			logCtx.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}

	diskManager := serviceDisk.New(config.GetBaseDiskDir())

	projectIdsToRunV2 := make(map[int64]bool, 0)
	var allProjects bool
	allProjects, projectIdsToRunV2, _ = C.GetProjectsFromListWithAllProjectSupport(*projectIdV2, "")
	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("Failed to get all projects and project_ids set to '*'.")
		}
		for _, projectID := range projectIDs {
			projectIdsToRunV2[projectID] = true
		}
	}

	projectIdsArrayV2 := make([]int64, 0)
	for projectId, _ := range projectIdsToRunV2 {
		projectIdsArrayV2 = append(projectIdsArrayV2, projectId)
	}

	ps, err := patternserver.New(config.GetIP(), config.GetRPCPort(), config.GetHTTPPort(), etcdClient, diskManager, cloudManager, newCloudManager, projectIdsArrayV2, *chunkCacheSize, *eventInfoCacheSize)
	if err != nil {
		logCtx.WithError(err).Errorln("Failed to init New PatternServer")
		panic(err)
	}

	lease, err := etcdClient.GrantLease(DefaultTTLSeconds)
	if err != nil {
		logCtx.WithError(err).Errorln("Failed to register lease with etcd")
		panic(err)
	}

	logCtx.WithFields(log.Fields{
		"TTL":     lease.TTL,
		"LeaseID": lease.ID,
	}).Infoln("Success: Registered lease with etcd")

	ps.SetLeaseId(lease.ID)

	err = ps.GetEtcdClient().RegisterPatternServer(config.GetIP(), config.GetRPCPort(), lease.ID)
	if err != nil {
		logCtx.WithError(err).Errorln("Failed to register with etcd")
		panic(err)
	}
	logCtx.Infoln("Success: Registered with etcd")

	patternServersRegisteredWithEtcd, err := ps.GetEtcdClient().DiscoverPatternServers()
	if err != nil {
		logCtx.WithError(err).Errorln("Failed to disconver pattern servers from etcd")
		panic(err)
	}

	logCtx.WithFields(log.Fields{
		"Pattern-servers": patternServersRegisteredWithEtcd,
	}).Infoln("List of pattern servers running")

	myNum := patternserver.CalculateMyNum(config.GetIP(), config.GetRPCPort(), patternServersRegisteredWithEtcd)
	if myNum == patternserver.PatternServerNotFound {
		panic("calculate my num failed: pattern server not registered")
	}

	version, err := etcdClient.GetProjectVersion()
	if err != nil {
		logCtx.WithError(err).Errorln("Failed to fetch projects metadata version from etcd")
		err = etcdClient.SetProjectVersion(InitialDefaultMetadata)
		if err != nil {
			logCtx.WithError(err).Errorln("Failed to set projects metadata version from etcd")
			panic(err)
		}
		version = InitialDefaultMetadata
	}

	projectDataMap, err := getProjectsDataFromVersion(version, cloudManager)

	ps.SetState(myNum, patternServersRegisteredWithEtcd, version, projectDataMap)

	go watchAndHandleEtcdEvents(ps, cloudManager)

	go keepEtcdLeaseAlive(ps)

	go runHttpStatus(ps.GetIp(), ps.GetHTTPPort(), ps, config.IsDevelopment())

	addr := ps.GetIp() + ":" + ps.GetRPCPort()
	logCtx.Printf("Starting rpc pattern server at %s", addr)
	r := initRpcServer(ps)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}

func handlePatternServerNodeUpdates(ps *patternserver.PatternServer) error {
	psNodes, err := ps.GetEtcdClient().DiscoverPatternServers()
	if err != nil {
		log.WithError(err).Errorln("failed to disover pattern servers")
		return err
	}

	n := patternserver.CalculateMyNum(ps.GetIp(), ps.GetRPCPort(), psNodes)
	if n == patternserver.PatternServerNotFound {
		err := errors.New("pattern server not registered")
		log.WithError(err).Errorln("handlePatternServerNodeUpdates pattern server not registered")
		return err
	}

	// no of pattern servers have changed
	// hence n and psNodes will only change
	ps.SetState(n, psNodes, ps.GetProjectDataVersion(), ps.GetProjectModelChunkData())

	return nil
}

func handlerVersionFileUpdates(ps *patternserver.PatternServer, newVersion string, cloudManger filestore.FileManager) error {
	newProjectModelChunkData, err := getProjectsDataFromVersion(newVersion, cloudManger)
	if err != nil {
		log.WithError(err).Errorln("handlerVersionFileUpdates failed to get updated projectModelChunkData")
		return err
	}
	// only version file has changed
	// hence only version and projectModelChunkData will change
	ps.SetState(ps.GetMyNum(), ps.GetPatternServerNodes(), newVersion, newProjectModelChunkData)
	return nil
}

func isListOfServersEqual(curNodes, newNodes []serviceEtcd.KV) bool {
	if len(curNodes) != len(newNodes) {
		return false
	}
	for i := 0; i < len(newNodes); i++ {
		if newNodes[i] != curNodes[i] {
			return false
		}
	}
	return true
}

func handleRecomputation(ps *patternserver.PatternServer) error {
	psNodes, err := ps.GetEtcdClient().DiscoverPatternServers()
	if err != nil {
		log.WithError(err).Errorln("failed to disover pattern servers")
		return err
	}

	n := patternserver.CalculateMyNum(ps.GetIp(), ps.GetRPCPort(), psNodes)
	if n == patternserver.PatternServerNotFound {
		err := errors.New("pattern server not registered")
		log.WithError(err).Errorln("handleRecomputation pattern server not registered")
		return err
	}

	oldNum := ps.GetMyNum()
	oldNodes := ps.GetPatternServerNodes()
	if n == oldNum && isListOfServersEqual(oldNodes, psNodes) {
		log.Debugf("OldNum: %d, NewNum: %d Same Returning", oldNum, n)
		return nil
	}

	log.Errorf("State Error, Rebalancing OldNum: %d, NewNum: %d changed, Recomputing, OldNodes: %+v, NewNodes: %+v", oldNum, n, oldNodes, psNodes)

	// my num has changed, update
	ps.SetState(n, psNodes, ps.GetProjectDataVersion(), ps.GetProjectModelChunkData())

	return nil
}

func watchAndHandleEtcdEvents(ps *patternserver.PatternServer, cloudManger filestore.FileManager) {
	logCtx := log.WithFields(log.Fields{
		"IP":      ps.GetIp(),
		"RPCPort": ps.GetRPCPort(),
	})
	psUpdateChan := ps.WatchPatternServers()
	versionFileUpdateChan := ps.WatchProjectsFile()

	ticker := time.NewTicker(5 * time.Minute)

	for {
		select {

		case t := <-ticker.C:
			logCtx.Debugf("Recomputing State at: %v", t)
			err := handleRecomputation(ps)
			if err != nil {
				logCtx.WithError(err).Errorln("recomputation failed to Update list of pattern servers")
				panic(err)
			}
		case psUpdateEvent := <-psUpdateChan:

			if err := psUpdateEvent.Err(); err != nil {
				logCtx.WithError(err).Errorln("psUpdateEvent Channel Received err from etcd")
				panic(err)
			}

			if len(psUpdateEvent.Events) > 0 {
				for _, event := range psUpdateEvent.Events {
					logCtx.WithFields(log.Fields{
						"UnitType": event.Type,
						"Key":      string(event.Kv.Key),
						"Value":    string(event.Kv.Value),
					}).Infoln("Event Received on PatternServerUpdateChannel")
				}
				err := handlePatternServerNodeUpdates(ps)
				if err != nil {
					logCtx.WithError(err).Errorln("failed to Update list of pattern servers")
					panic(err)
				}
			}

		case versionFileUpdateEvent := <-versionFileUpdateChan:

			if err := versionFileUpdateEvent.Err(); err != nil {
				logCtx.WithError(err).Errorln("versionFileUpdateEvent Channel Received err from etcd")
				panic(err)
			}

			for _, event := range versionFileUpdateEvent.Events {
				logCtx.WithFields(log.Fields{
					"UnitType": event.Type,
					"Key":      string(event.Kv.Key),
					"Value":    string(event.Kv.Value),
				}).Infoln("Event Received on versionFileUpdateChan")
				newVersion := string(event.Kv.Value)
				logCtx.WithField("NewVersion", newVersion).Infoln("Update ProjectFileVersion")
				err := handlerVersionFileUpdates(ps, newVersion, cloudManger)
				if err != nil {
					logCtx.WithError(err).Errorln("failed to Update ProjectFileVersion")
					panic(err)
				}
			}
		}

	}
}

func getProjectsDataFromVersion(version string, cloudManger filestore.FileManager) (map[int64]patternserver.ModelChunkMapping, error) {

	modelMetadata, _, msg := modelstore.GetStore().GetAllProjectModelMetadata()

	if msg != "" {
		return nil, errors.New(msg)
	}

	return makeProjectModelChunkLookup(modelMetadata), nil
}

func keepEtcdLeaseAlive(ps *patternserver.PatternServer) {
	log.Println("Starting to watch on keepAliveChannel")
	keepAliveChannel, err := ps.KeepAlive()
	if err != nil {
		log.WithError(err).Errorln("Error ETCD Keep Alive")
	}
	for {
		<-keepAliveChannel
	}
}

func runHttpStatus(ip, port string, ps *patternserver.PatternServer, isDev bool) {
	r := initHttpStatusServer(isDev, ps)
	addr := ip + ":" + port
	r.Run(addr)
}

func initRpcServer(ps *patternserver.PatternServer) *mux.Router {
	s := rpc.NewServer()
	s.RegisterCodec(rpcjson.NewCodec(), "application/json")
	s.RegisterCodec(rpcjson.NewCodec(), "application/json;charset=UTF-8")
	s.RegisterService(ps, PC.RPCServiceName)
	s.RegisterBeforeFunc(func(i *rpc.RequestInfo) {
		reqId := ""
		if len(i.Request.Header["X-Req-Id"]) > 0 {
			reqId = i.Request.Header["X-Req-Id"][0]
		}

		method := i.Method
		startedAt := time.Now().UnixNano()

		i.Request.Header["Started-At"] = []string{fmt.Sprintf("%v", startedAt)}

		log.WithFields(log.Fields{
			"reqId":  reqId,
			"method": method,
		}).Info("Seen Request")
	})
	s.RegisterAfterFunc(func(i *rpc.RequestInfo) {
		reqId := ""
		if len(i.Request.Header["X-Req-Id"]) > 0 {
			reqId = i.Request.Header["X-Req-Id"][0]
		}

		method := i.Method
		err := i.Error
		statusCode := i.StatusCode

		startedAt := time.Now().UnixNano()
		if len(i.Request.Header["Started-At"]) > 0 {
			startedAt, _ = strconv.ParseInt(i.Request.Header["Started-At"][0], 10, 64)
		}

		endedAt := time.Now().UnixNano()

		latency := endedAt - startedAt

		logCtx := log.WithFields(log.Fields{
			"reqId":       reqId,
			"method":      method,
			"latency(ms)": int(math.Ceil(float64(latency) / 1000000.0)),
			"statusCode":  statusCode,
		})

		if err != nil {
			logCtx.WithError(err).Error("Error Processing Request")
		} else {
			logCtx.Info("Processed Request")
		}
	})
	r := mux.NewRouter()
	r.Handle(PC.RPCEndpoint, s)
	return r
}

func initHttpStatusServer(isDev bool, ps *patternserver.PatternServer) *gin.Engine {

	if !isDev {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/state", ps.DebugState)
	r.GET("/status", func(c *gin.Context) {
		resp := map[string]string{
			"status": "success",
		}
		c.JSON(http.StatusOK, resp)
		return
	})
	return r
}
