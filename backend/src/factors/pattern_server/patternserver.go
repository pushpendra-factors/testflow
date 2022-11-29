package patternserver

import (
	"factors/filestore"
	"factors/pattern"
	client "factors/pattern_client"
	store "factors/pattern_server/store"
	serviceEtcd "factors/services/etcd"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"sync"

	modelstore "factors/model/store"

	"github.com/gin-gonic/gin"
	E "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

const (
	IdSeparator           = ":"
	PatternServerNotFound = -1
)

type ModelData struct {
	Type           string
	Chunks         []string
	StartTimestamp int64
	EndTimestamp   int64
}

// model_id -> ModelData
type ModelChunkMapping map[uint64]ModelData

type state struct {
	patternServerNodes []serviceEtcd.KV
	myNum              int

	projectDataVersion string

	// project_id -> ModelChunkMapping
	projectsModelChunkData map[int64]ModelChunkMapping

	// project_id -> true
	projectsToServe map[int64]bool

	// project_id:model_id -> true
	projectModelsToServe map[string]bool

	// project_id:model_id:chunk_id -> true
	projectModelChunksToServe map[string]bool
}

func (s *state) getPatternServerNodes() []serviceEtcd.KV {
	return s.patternServerNodes
}

func (s *state) getMyNum() int {
	return s.myNum
}

func (s *state) getProjectDataVersion() string {
	return s.projectDataVersion
}

func (s *state) getProjectModelChunkData() map[int64]ModelChunkMapping {
	return s.projectsModelChunkData
}

func (s *state) getProjectModelsToServe() map[string]bool {
	return s.projectModelsToServe
}

func (s *state) getProjectsToServe() map[int64]bool {
	return s.projectsToServe
}

func (s *state) getProjectModelChunksToServe() map[string]bool {
	return s.projectModelChunksToServe
}

type PatternServer struct {
	ip          string
	rpcPort     string
	httpPort    string
	etcdLeaseID clientv3.LeaseID
	etcdClient  *serviceEtcd.EtcdClient

	stateLock sync.RWMutex
	state     *state

	store *store.PatternStore
}

func New(ip, rpcPort, httpPort string, etcdClient *serviceEtcd.EtcdClient, diskFileManager, cloudFileManger, newCloudFileManager filestore.FileManager, projectIdsV2 []int64, chunkCacheSize, eventInfoCacheSize int) (*PatternServer, error) {

	store, err := store.New(chunkCacheSize, eventInfoCacheSize, diskFileManager, cloudFileManger, newCloudFileManager, projectIdsV2)
	if err != nil {
		return &PatternServer{}, E.Wrap(err, "Failed To Create Pattern Store")
	}

	state := &state{
		myNum:                  -1,
		projectsModelChunkData: make(map[int64]ModelChunkMapping),
	}

	ps := &PatternServer{
		ip:         ip,
		rpcPort:    rpcPort,
		httpPort:   httpPort,
		state:      state,
		etcdClient: etcdClient,
		store:      store,
	}

	return ps, nil
}

func (ps *PatternServer) SetState(myNum int, psNodes []serviceEtcd.KV, projectDataVersion string, projectsModelChunkData map[int64]ModelChunkMapping) {

	log.Debugln("Computing New State")

	newState := state{
		myNum:                  myNum,
		patternServerNodes:     psNodes,
		projectDataVersion:     projectDataVersion,
		projectsModelChunkData: projectsModelChunkData,
	}

	projectsToServe := computeProjectsToServe(newState.projectsModelChunkData, uint64(newState.myNum), uint64(len(newState.patternServerNodes)))
	modelsToServe := computeModelsToServe(newState.projectsModelChunkData, uint64(newState.myNum), uint64(len(newState.patternServerNodes)))
	chunksToServe := computeChunksToServe(newState.projectsModelChunkData, uint64(newState.myNum), uint64(len(newState.patternServerNodes)))

	ps.stateLock.Lock()
	defer ps.stateLock.Unlock()
	newState.projectModelsToServe = modelsToServe
	newState.projectModelChunksToServe = chunksToServe
	newState.projectsToServe = projectsToServe

	log.WithFields(log.Fields{
		"MyNum":                            newState.myNum,
		"PatternServersRegisteredWithEtcd": newState.patternServerNodes,
		"ProjectDataVersion":               newState.projectDataVersion,
		"ProjectModelChunkData":            newState.projectsModelChunkData,
		"ProjectsToServe":                  newState.projectsToServe,
		"ProjectModelsToServe":             newState.projectModelsToServe,
		"ProjectModelChunksToServe":        newState.projectModelChunksToServe,
	}).Infoln("Setting New State")

	ps.state = &newState
}

func (ps *PatternServer) GetState() *state {
	ps.stateLock.RLock()
	defer ps.stateLock.RUnlock()

	return ps.state
}

func computeProjectsToServe(projectDatas map[int64]ModelChunkMapping, myNum, noOfPatternServers uint64) map[int64]bool {
	projectsToServe := make(map[int64]bool)
	for projectId := range projectDatas {
		if serveProject(projectId, myNum, noOfPatternServers) {
			projectsToServe[projectId] = true
		}
	}
	return projectsToServe
}

func computeModelsToServe(projectDatas map[int64]ModelChunkMapping, myNum, noOfPatternServers uint64) map[string]bool {
	modelsToServe := make(map[string]bool)
	for projectId, pD := range projectDatas {
		for modelId := range pD {
			if serveProjectModel(projectId, modelId, myNum, noOfPatternServers) {
				modelsToServe[GetModelKey(projectId, modelId)] = true
			}
		}
	}
	return modelsToServe
}

func computeChunksToServe(projectDatas map[int64]ModelChunkMapping, myNum, noOfPatternServers uint64) map[string]bool {
	chunksToServe := make(map[string]bool)
	for projectId, pd := range projectDatas {
		for modelId, modelData := range pd {
			for _, chunk := range modelData.Chunks {
				if serveProjectModelChunk(projectId, modelId, chunk, myNum, noOfPatternServers) {
					chunksToServe[GetChunkKey(projectId, modelId, chunk)] = true
				}
			}
		}
	}
	return chunksToServe
}

func generateHashNum(str string) uint64 {
	f := fnv.New64a()
	_, err := f.Write([]byte(str)) // write never returns error
	if err != nil {
		log.WithError(err).WithField("str", str).Errorln("Failed to generate hash num for string")
	}
	return f.Sum64()
}

func serveProject(projectId int64, myNum, noOfPatternServers uint64) bool {
	str := fmt.Sprintf("%v", projectId)
	hashN := generateHashNum(str)
	return hashN%noOfPatternServers == myNum%noOfPatternServers
}

func serveProjectModel(projectId int64, modelId, myNum, noOfPatternServers uint64) bool {
	str := GetModelKey(projectId, modelId)
	hashN := generateHashNum(str)
	return hashN%noOfPatternServers == myNum%noOfPatternServers
}

func serveProjectModelChunk(projectId int64, modelId uint64, chunkId string, myNum, noOfPatternServers uint64) bool {
	str := GetChunkKey(projectId, modelId, chunkId)
	hashN := generateHashNum(str)
	return hashN%noOfPatternServers == myNum%noOfPatternServers
}

func (ps *PatternServer) GetProjectModelChunks(projectId int64, modelId uint64) ([]string, bool) {
	// call db to get this: janani
	modelData, _, _ := modelstore.GetStore().GetProjectModelMetadata(projectId)
	chunkIds := make([]string, 0, 0)
	for _, data := range modelData {
		if data.ModelId == modelId {
			chunkIds = strings.Split(data.Chunks, ",")
		}
	}
	return chunkIds, true
}

func (ps *PatternServer) GetProjectModelLatestInterval(projectId int64) (client.ModelInfo, error) {
	// call db to get this: janani
	modelMetadata, _, msg := modelstore.GetStore().GetProjectModelMetadata(projectId)
	var err error
	if msg != "" {
		return client.ModelInfo{}, E.Wrap(err, fmt.Sprintf("ProjectID %d, Missing ProjectModelIntervals, %s", projectId, msg))
	}
	return client.ModelInfo{ModelId: modelMetadata[0].ModelId,
		ModelType:      modelMetadata[0].ModelType,
		StartTimestamp: modelMetadata[0].StartTime,
		EndTimestamp:   modelMetadata[0].EndTime}, nil
}

func (ps *PatternServer) GetProjectDataVersion() string {
	return ps.GetState().getProjectDataVersion()
}

func (ps *PatternServer) WatchPatternServers() clientv3.WatchChan {
	return ps.etcdClient.Watch(serviceEtcd.PatternServerPrefix, clientv3.WithPrefix())
}

func (ps *PatternServer) WatchProjectsFile() clientv3.WatchChan {
	return ps.etcdClient.Watch(serviceEtcd.ProjectVersionKey)
}

func (ps *PatternServer) KeepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return ps.etcdClient.KeepAlive(ps.etcdLeaseID)
}

func (ps *PatternServer) SetLeaseId(leaseId clientv3.LeaseID) {
	ps.etcdLeaseID = leaseId
}

func (ps *PatternServer) GetLeaseId() clientv3.LeaseID {
	return ps.etcdLeaseID
}

func (ps *PatternServer) GetMyNum() int {
	return ps.GetState().getMyNum()
}

func (ps *PatternServer) GetPatternServerNodes() []serviceEtcd.KV {
	return ps.GetState().getPatternServerNodes()
}

func (ps *PatternServer) GetNoOfPatternNodes() int {
	return len(ps.GetState().getPatternServerNodes())
}

func (ps *PatternServer) GetEtcdClient() *serviceEtcd.EtcdClient {
	return ps.etcdClient
}

func (ps *PatternServer) GetIp() string {
	return ps.ip
}

func (ps *PatternServer) GetRPCPort() string {
	return ps.rpcPort
}

func (ps *PatternServer) GetHTTPPort() string {
	return ps.httpPort
}

func (ps *PatternServer) GetProjectModelsToServe() map[string]bool {
	return ps.GetState().getProjectModelsToServe()
}

func (ps *PatternServer) GetProjectModelChunksToServe() map[string]bool {
	return ps.GetState().getProjectModelChunksToServe()
}

func (ps *PatternServer) GetProjectModelChunkData() map[int64]ModelChunkMapping {
	return ps.GetState().getProjectModelChunkData()
}

func (ps *PatternServer) IsProjectServable(projectId int64) bool {
	val, exists := ps.GetState().getProjectsToServe()[projectId]
	return val && exists
}

func (ps *PatternServer) IsProjectModelServable(projectId int64, modelId uint64) bool {
	val, exists := ps.GetState().getProjectModelsToServe()[GetModelKey(projectId, modelId)]
	return val && exists
}

func (ps *PatternServer) IsProjectModelChunkServable(projectId int64, modelId uint64, chunkId string) bool {
	val, exists := ps.GetState().getProjectModelChunksToServe()[GetChunkKey(projectId, modelId, chunkId)]
	return val && exists
}

func (ps *PatternServer) GetModelEventInfo(projectId int64, modelId uint64) (pattern.UserAndEventsInfo, error) {
	return ps.store.GetModelEventInfo(projectId, modelId)
}

func GetModelKey(projectId int64, modelId uint64) string {
	return fmt.Sprintf("%d%s%d", projectId, IdSeparator, modelId)
}

func GetChunkKey(projectId int64, modelId uint64, chunkId string) string {
	return fmt.Sprintf("%d%s%d%s%s", projectId, IdSeparator, modelId, IdSeparator, chunkId)
}

func CalculateMyNum(ip, port string, patternServers []serviceEtcd.KV) int {
	myAddr := ip + ":" + port
	for i, kv := range patternServers {
		if kv.Value == myAddr {
			return i
		}
	}
	return PatternServerNotFound
}

func (ps *PatternServer) DebugState(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"num":                           ps.GetMyNum(),
			"ip":                            ps.GetIp(),
			"etcd_lease_id":                 ps.GetLeaseId(),
			"project_data_version":          ps.GetProjectDataVersion(),
			"projects_to_serve":             ps.GetState().getProjectsToServe(),
			"project_models_to_serve":       ps.GetState().getProjectModelsToServe(),
			"project_model_chunks_to_serve": ps.GetState().getProjectModelChunksToServe(),
			"peer_pattern_servers":          ps.GetPatternServerNodes(),
			"projects_model_chunk_data":     ps.GetState().getProjectModelChunkData(),
		},
	})
}
