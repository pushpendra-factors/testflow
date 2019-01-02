package main

import (
	"bufio"
	encjson "encoding/json"
	"errors"
	"factors/filestore"
	patternserver "factors/patternserver"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceS3 "factors/services/s3"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/rpc"

	rpcjson "github.com/gorilla/rpc/json"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultTTLSeconds = 10
	Development       = "development"
)

type projectData struct {
	ID        uint64    `json:"pid"`
	ModelID   uint64    `json:"mid"`
	StartDate time.Time `json:"sd"`
	EndDate   time.Time `json:"ed"`
	Chunks    []string  `json:"cs"`
}

func parseProjectsDataFile(reader io.Reader) []projectData {

	scanner := bufio.NewScanner(reader)
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	projectDatas := make([]projectData, 0, 0)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var p projectData
		if err := encjson.Unmarshal([]byte(line), &p); err != nil {
			log.WithFields(log.Fields{"lineNum": lineNum, "err": err}).Error("Failed to unmarshal project")
			continue
		}
		projectDatas = append(projectDatas, p)
	}
	err := scanner.Err()
	if err != nil {
		log.WithError(err).Errorln("Scanner error")
	}

	return projectDatas
}

// Add Test
func makeProjectModelChunkLookup(projectDatas []projectData) map[uint64]patternserver.ModelChunkMapping {
	pD := make(map[uint64]patternserver.ModelChunkMapping)
	for _, p := range projectDatas {
		mCM, exists := pD[p.ID]
		if !exists {
			mCM = patternserver.ModelChunkMapping{}
		}
		mD := patternserver.ModelData{Chunks: p.Chunks, StartDate: p.StartDate, EndDate: p.EndDate}
		mCM[p.ModelID] = mD
		pD[p.ID] = mCM
	}
	return pD
}

type config struct {
	Environment    string
	IP             string
	Port           string
	EtcdEndpoints  []string
	S3BucketName   string
	S3BucketRegion string
	DiskBaseDir    string
}

func NewConfig(env, ip, port, etcd, diskBaseDir, s3Bucket, s3BucketRegion string) (*config, error) {
	if env != Development {
		return nil, errors.New("Invalid Environment")
	}
	if ip == "" {
		return nil, errors.New("Invalid IP")
	}
	if port == "" {
		return nil, errors.New("Invalid Port")
	}

	if diskBaseDir == "" {
		return nil, errors.New("Invalid baseDiskDir")
	}

	if s3Bucket == "" {
		return nil, errors.New("Invalid S3BucketName")
	}

	if s3BucketRegion == "" {
		return nil, errors.New("Invalid S3BucketRegion")
	}

	etcds := strings.Split(etcd, ",")
	if len(etcds) == 0 {
		return nil, errors.New("Invalid EtcdEndpoints")
	}

	c := config{
		Environment:    env,
		IP:             ip,
		Port:           port,
		EtcdEndpoints:  etcds,
		DiskBaseDir:    diskBaseDir,
		S3BucketName:   s3Bucket,
		S3BucketRegion: s3BucketRegion,
	}

	return &c, nil
}

func (c *config) GetEnvironment() string {
	return c.Environment
}

func (c *config) GetIP() string {
	return c.IP
}

func (c *config) GetPort() string {
	return c.Port
}

func (c *config) GetBaseDiskDir() string {
	return c.DiskBaseDir
}

func (c *config) GetEtcdEndpoints() []string {
	return c.EtcdEndpoints
}

func (c *config) GetS3BucketName() string {
	return c.S3BucketName
}

func (c *config) GetS3BucketRegion() string {
	return c.S3BucketRegion
}

// Process crashes if started with IP and port already registered with etcd.
// TTL on etcd is 10 seconds.
// Monit / Kubernetes will keep trying to restart the process, it should succeed after 10 seconds / till key expires.

// ./pattern-app --env=development --ip=127.0.0.1 --port=8100 --etcd=localhost:2379 --disk_dir=/tmp/factors --s3=/tmp/factors-dev --s3_region=us-east-1
func main() {

	env := flag.String("env", "development", "")
	ip := flag.String("ip", "127.0.0.1", "")
	port := flag.String("port", "8100", "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")
	diskBaseDir := flag.String("disk_dir", "/tmp/factors", "")
	s3Bucket := flag.String("s3", "/tmp/factors-dev", "")
	s3BucketRegion := flag.String("s3_region", "us-east-1", "")
	flag.Parse()

	config, err := NewConfig(*env, *ip, *port, *etcd, *diskBaseDir, *s3Bucket, *s3BucketRegion)
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"IP":            config.GetIP(),
		"Port":          config.GetPort(),
		"Env":           config.GetEnvironment(),
		"EtcdEndpoints": config.GetEtcdEndpoints(),
		"DiskBaseDir":   config.GetBaseDiskDir(),
		"S3Bucket":      config.GetS3BucketName(),
		"S3Region":      config.GetS3BucketRegion(),
	}).Infoln("Initialising with config")

	if config.GetEnvironment() == Development {
		log.SetLevel(log.DebugLevel)
	}

	logCtx := log.WithFields(log.Fields{
		"IP":   config.GetIP(),
		"Port": config.GetPort(),
	})

	logCtx.WithFields(log.Fields{
		"EtcdEndpoints": config.GetEtcdEndpoints(),
	}).Infoln("Pattern Server")

	etcdClient, err := serviceEtcd.New(config.GetEtcdEndpoints())
	if err != nil {
		logCtx.WithError(err).Errorln("Falied to initialize etcd client")
		panic(err)
	}

	if isRegistered, err := etcdClient.IsRegistered(config.GetIP(), config.GetPort()); err != nil {
		logCtx.WithError(err).Errorln("Falied to check registered status with etcd")
		panic(err)
	} else if isRegistered {
		str := fmt.Sprintf("Pattern Server Already registered with IP:Port [ %s:%s ]", config.GetIP(), config.GetPort())
		logCtx.Errorln(str)
		panic(str)
	}
	var cloudManger filestore.FileManager
	if config.GetEnvironment() == "development" {
		cloudManger = serviceDisk.New(config.GetS3BucketName())
	} else {
		cloudManger = serviceS3.New(config.GetS3BucketName(), config.GetS3BucketRegion())
	}

	diskManager := serviceDisk.New(config.GetBaseDiskDir())

	ps, err := patternserver.New(config.GetIP(), config.GetPort(), etcdClient, diskManager, cloudManger)
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

	err = ps.GetEtcdClient().RegisterPatternServer(config.GetIP(), config.GetPort(), lease.ID)
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

	myNum := patternserver.CalculateMyNum(config.GetIP(), config.GetPort(), patternServersRegisteredWithEtcd)
	if myNum == patternserver.PatternServerNotFound {
		panic("calculate my num failed: pattern server not registered")
	}

	version, err := etcdClient.GetProjectVersion()
	if err != nil {
		logCtx.WithError(err).Errorln("Failed to fetch projects metadata version from etcd")
		panic(err)
	}

	projectDataMap, err := getProjectsDataFromVersion(version, cloudManger)

	ps.SetState(myNum, patternServersRegisteredWithEtcd, version, projectDataMap)

	go watchAndHandleEtcdEvents(ps, cloudManger)

	go keepEtcdLeaseAlive(ps)

	addr := ps.GetIp() + ":" + ps.GetPort()
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

	n := patternserver.CalculateMyNum(ps.GetIp(), ps.GetPort(), psNodes)
	if n == patternserver.PatternServerNotFound {
		log.WithError(err).Errorln("pattern server not registered")
		return errors.New("pattern server not registered")
	}

	// no of pattern servers have changed
	// hence n and psNodes will only change
	ps.SetState(n, psNodes, ps.GetProjectDataVersion(), ps.GetProjectModelChunkData())

	return nil
}

func handlerVersionFileUpdates(ps *patternserver.PatternServer, newVersion string, cloudManger filestore.FileManager) error {
	newProjectModelChunkData, err := getProjectsDataFromVersion(newVersion, cloudManger)
	if err != nil {
		log.WithError(err).Errorln("failed to get updated projectModelChunkData")
		return err
	}
	// only version file has changed
	// hence only version and projectModelChunkData will change
	ps.SetState(ps.GetMyNum(), ps.GetPatternServerNodes(), newVersion, newProjectModelChunkData)
	return nil
}

func watchAndHandleEtcdEvents(ps *patternserver.PatternServer, cloudManger filestore.FileManager) {
	logCtx := log.WithFields(log.Fields{
		"IP":   ps.GetIp(),
		"Port": ps.GetPort(),
	})
	psUpdateChan := ps.WatchPatternServers()
	versionFileUpdateChan := ps.WatchProjectsFile()
	for {
		select {
		case psUpdateEvent := <-psUpdateChan:
			if len(psUpdateEvent.Events) > 0 {
				for _, event := range psUpdateEvent.Events {
					logCtx.WithFields(log.Fields{
						"Type":  event.Type,
						"Key":   string(event.Kv.Key),
						"Value": string(event.Kv.Value),
					}).Infoln("Event Received on PatternServerUpdateChannel")
				}
				err := handlePatternServerNodeUpdates(ps)
				if err != nil {
					logCtx.WithError(err).Errorln("failed to Update list of pattern servers")
				}
			}

		case versionFileUpdateEvent := <-versionFileUpdateChan:
			for _, event := range versionFileUpdateEvent.Events {
				logCtx.WithFields(log.Fields{
					"Type":  event.Type,
					"Key":   string(event.Kv.Key),
					"Value": string(event.Kv.Value),
				}).Infoln("Event Received on versionFileUpdateChan")
				newVersion := string(event.Kv.Value)
				logCtx.WithField("NewVersion", newVersion).Infoln("Update ProjectFileVersion")
				err := handlerVersionFileUpdates(ps, newVersion, cloudManger)
				if err != nil {
					logCtx.WithError(err).Errorln("failed to Update ProjectFileVersion")
				}
			}
		}

	}
}

func getProjectsDataFromVersion(version string, cloudManger filestore.FileManager) (map[uint64]patternserver.ModelChunkMapping, error) {

	// always read from cloud
	// do not copy on disk ?
	projectDataFilePath, fName := cloudManger.GetProjectsDataFilePathAndName(version)

	log.WithFields(log.Fields{
		"Version": version,
		"path":    projectDataFilePath,
	}).Info("ProjectsDataFile")

	projectDataFile, err := cloudManger.Get(projectDataFilePath, fName)
	if err != nil {
		log.WithError(err).Errorln("Failed to open projects data file")
		return make(map[uint64]patternserver.ModelChunkMapping), err
	}
	defer projectDataFile.Close()

	projectDataMap := makeProjectModelChunkLookup(parseProjectsDataFile(projectDataFile))
	return projectDataMap, nil
}

func keepEtcdLeaseAlive(ps *patternserver.PatternServer) {
	log.Println("Starting to watch on keepAliveChannel")
	keepAliveChannel, err := ps.KeepAlive()
	if err != nil {
		log.WithError(err).Errorln("Error ETCD Keep Alive")
	}
	for {
		ls := <-keepAliveChannel
		log.WithFields(log.Fields{
			"TTL":     ls.TTL,
			"LeaseId": ps.GetLeaseId(),
		}).Debugln("Success ETCD Keep Alive")
	}
}

func initRpcServer(ps *patternserver.PatternServer) *mux.Router {
	s := rpc.NewServer()
	s.RegisterCodec(rpcjson.NewCodec(), "application/json")
	s.RegisterCodec(rpcjson.NewCodec(), "application/json;charset=UTF-8")
	s.RegisterService(ps, patternserver.RPCServiceName)
	r := mux.NewRouter()
	r.Handle(patternserver.RPCEndpoint, s)
	return r
}
