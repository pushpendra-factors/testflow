package patternserver

import (
	"errors"
	"factors/filestore"
	"factors/pattern"
	client "factors/pattern_client"
	store "factors/pattern_server/store"
	serviceEtcd "factors/services/etcd"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

const (
	IdSeparator           = ":"
	ChunkCacheSize        = 5
	EventInfoCacheSize    = 10
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
	projectsModelChunkData map[uint64]ModelChunkMapping

	// project_id -> true
	projectsToServe map[uint64]bool

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

func (s *state) getProjectModelChunkData() map[uint64]ModelChunkMapping {
	return s.projectsModelChunkData
}

func (s *state) getProjectModelsToServe() map[string]bool {
	return s.projectModelsToServe
}

func (s *state) getProjectsToServe() map[uint64]bool {
	return s.projectsToServe
}

func (s *state) getProjectModelChunksToServe() map[string]bool {
	return s.projectModelChunksToServe
}

type PatternServer struct {
	ip          string
	port        string
	etcdLeaseID clientv3.LeaseID
	etcdClient  *serviceEtcd.EtcdClient

	stateLock sync.RWMutex
	state     *state

	store *store.PatternStore
}

func New(ip, port string, etcdClient *serviceEtcd.EtcdClient, diskFileManager, cloudFileManger filestore.FileManager) (*PatternServer, error) {

	store, err := store.New(ChunkCacheSize, EventInfoCacheSize, diskFileManager, cloudFileManger)
	if err != nil {
		return &PatternServer{}, err
	}

	state := &state{
		myNum:                  -1,
		projectsModelChunkData: make(map[uint64]ModelChunkMapping),
	}

	ps := &PatternServer{
		ip:         ip,
		port:       port,
		state:      state,
		etcdClient: etcdClient,
		store:      store,
	}

	return ps, nil
}

func (ps *PatternServer) SetState(myNum int, psNodes []serviceEtcd.KV, projectDataVersion string, projectsModelChunkData map[uint64]ModelChunkMapping) {

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

func computeProjectsToServe(projectDatas map[uint64]ModelChunkMapping, myNum, noOfPatternServers uint64) map[uint64]bool {
	projectsToServe := make(map[uint64]bool)
	for projectId := range projectDatas {
		if serveProject(projectId, myNum, noOfPatternServers) {
			projectsToServe[projectId] = true
		}
	}
	return projectsToServe
}

func computeModelsToServe(projectDatas map[uint64]ModelChunkMapping, myNum, noOfPatternServers uint64) map[string]bool {
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

func computeChunksToServe(projectDatas map[uint64]ModelChunkMapping, myNum, noOfPatternServers uint64) map[string]bool {
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

func serveProject(projectId, myNum, noOfPatternServers uint64) bool {
	str := fmt.Sprintf("%v", projectId)
	hashN := generateHashNum(str)
	return hashN%noOfPatternServers == myNum%noOfPatternServers
}

func serveProjectModel(projectId, modelId, myNum, noOfPatternServers uint64) bool {
	str := GetModelKey(projectId, modelId)
	hashN := generateHashNum(str)
	return hashN%noOfPatternServers == myNum%noOfPatternServers
}

func serveProjectModelChunk(projectId, modelId uint64, chunkId string, myNum, noOfPatternServers uint64) bool {
	str := GetChunkKey(projectId, modelId, chunkId)
	hashN := generateHashNum(str)
	return hashN%noOfPatternServers == myNum%noOfPatternServers
}

func (ps *PatternServer) GetProjectModels(projectId uint64) (ModelChunkMapping, bool) {
	projectModels := ps.GetState().getProjectModelChunkData()
	projectData, exists := projectModels[projectId]
	return projectData, exists
}

func (ps *PatternServer) GetProjectModel(projectId, modelId uint64) (ModelData, bool) {
	projectModels, found := ps.GetProjectModels(projectId)
	if !found {
		return ModelData{}, found
	}
	modelData, found := projectModels[modelId]
	return modelData, found
}

func (ps *PatternServer) GetProjectModelChunks(projectId, modelId uint64) ([]string, bool) {
	modelData, found := ps.GetProjectModel(projectId, modelId)
	if !found {
		return []string{}, found
	}
	chunkIds := make([]string, 0, 0)
	for _, cId := range modelData.Chunks {
		chunkIds = append(chunkIds, cId)
	}
	return chunkIds, true
}

func (ps *PatternServer) GetProjectModelIntervals(projectId uint64) ([]client.ModelInfo, error) {

	modelInfos := make([]client.ModelInfo, 0, 0)

	modelChunkData, exists := ps.GetProjectModels(projectId)
	if !exists {
		err := errors.New("MissingModelChunkData")
		return modelInfos, err
	}

	for mid, modelData := range modelChunkData {
		mi := client.ModelInfo{
			ModelId:        mid,
			ModelType:      modelData.Type,
			StartTimestamp: modelData.StartTimestamp,
			EndTimestamp:   modelData.EndTimestamp,
		}
		modelInfos = append(modelInfos, mi)
	}

	// latest interval is first
	sort.Slice(modelInfos, func(i, j int) bool {
		return modelInfos[i].StartTimestamp > modelInfos[j].StartTimestamp
	})

	return modelInfos, nil
}

func (ps *PatternServer) GetProjectModelLatestInterval(projectId uint64) (client.ModelInfo, error) {
	modelInfos, err := ps.GetProjectModelIntervals(projectId)
	if err != nil {
		return client.ModelInfo{}, err
	}

	return modelInfos[0], nil
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

func (ps *PatternServer) GetPort() string {
	return ps.port
}

func (ps *PatternServer) GetProjectModelsToServe() map[string]bool {
	return ps.GetState().getProjectModelsToServe()
}

func (ps *PatternServer) GetProjectModelChunksToServe() map[string]bool {
	return ps.GetState().getProjectModelChunksToServe()
}

func (ps *PatternServer) GetProjectModelChunkData() map[uint64]ModelChunkMapping {
	return ps.GetState().getProjectModelChunkData()
}

func (ps *PatternServer) IsProjectServable(projectId uint64) bool {
	val, exists := ps.GetState().getProjectsToServe()[projectId]
	return val && exists
}

func (ps *PatternServer) IsProjectModelServable(projectId, modelId uint64) bool {
	val, exists := ps.GetState().getProjectModelsToServe()[GetModelKey(projectId, modelId)]
	return val && exists
}

func (ps *PatternServer) IsProjectModelChunkServable(projectId, modelId uint64, chunkId string) bool {
	val, exists := ps.GetState().getProjectModelChunksToServe()[GetChunkKey(projectId, modelId, chunkId)]
	return val && exists
}

func (ps *PatternServer) GetModelEventInfo(projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {
	return ps.store.GetModelEventInfo(projectId, modelId)
}

func GetModelKey(projectId, modelId uint64) string {
	return fmt.Sprintf("%d%s%d", projectId, IdSeparator, modelId)
}

func GetChunkKey(projectId, modelId uint64, chunkId string) string {
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
