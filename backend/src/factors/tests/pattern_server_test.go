package tests

import (
	patternserver "factors/pattern_server"
	serviceEtcd "factors/services/etcd"
	U "factors/util"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectModels(t *testing.T) {
	ps := patternserver.PatternServer{}
	projectData := make(map[int64]patternserver.ModelChunkMapping)
	projectId1 := U.RandomInt64()
	modelId1 := U.RandomUint64()
	chunkId1 := U.RandomString(5)

	modelId2 := U.RandomUint64()
	chunkId2 := U.RandomString(5)
	chunkId3 := U.RandomString(5)

	project1ModelChunkMapping := patternserver.ModelChunkMapping{
		modelId1: patternserver.ModelData{
			StartTimestamp: time.Now().Unix(),
			EndTimestamp:   time.Now().Unix(),
			Chunks:         []string{chunkId1},
		},
		modelId2: patternserver.ModelData{
			StartTimestamp: time.Now().Unix(),
			EndTimestamp:   time.Now().Unix(),
			Chunks:         []string{chunkId2, chunkId3},
		},
	}

	projectData[projectId1] = project1ModelChunkMapping
	// mock this
	ps.SetState(0, []serviceEtcd.KV{
		{
			Key:   "localhost:2379",
			Value: "localhost:2379",
		},
	}, "", projectData)

	models := []uint64{modelId1, modelId2}
	sort.Slice(models, func(i, j int) bool { return models[i] < models[j] })
}

func TestGetChunkKey(t *testing.T) {
	projectId := U.RandomInt64()
	modelId := U.RandomUint64()
	chunkId := U.RandomString(5)
	key := patternserver.GetChunkKey(projectId, modelId, chunkId)
	expKey := fmt.Sprintf("%d%s%d%s%s", projectId, patternserver.IdSeparator, modelId, patternserver.IdSeparator, chunkId)
	assert.Equal(t, expKey, key)
}

func TestGetModelKey(t *testing.T) {
	projectId := U.RandomInt64()
	modelId := U.RandomUint64()
	key := patternserver.GetModelKey(projectId, modelId)
	expKey := fmt.Sprintf("%d%s%d", projectId, patternserver.IdSeparator, modelId)
	assert.Equal(t, expKey, key)
}

func TestCalculateMyNum(t *testing.T) {

	kv := []serviceEtcd.KV{serviceEtcd.KV{
		Key:   "/prefix/127.0.0.1:8080",
		Value: "127.0.0.1:8080",
	},
		serviceEtcd.KV{
			Key:   "/prefix/127.0.0.1:8081",
			Value: "127.0.0.1:8081",
		},
		serviceEtcd.KV{
			Key:   "/prefix/127.0.0.1:8083",
			Value: "127.0.0.1:8083",
		},
	}

	ip := "127.0.0.1"
	port := "8083"
	num := patternserver.CalculateMyNum(ip, port, kv)
	assert.Equal(t, 2, num)
}
