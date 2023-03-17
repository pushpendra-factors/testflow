package task

import (
	ch "factors/cache/redis"
	M "factors/model/model"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func ComputeResultAndCache(project_id int64, model_id uint64, qr M.ExplainV2Query) (string, error) {

	var result string
	expiry := M.QueryCacheMutableResultMonth
	result, err := Compute_result_string(project_id, model_id, qr)
	if err != nil {
		log.Errorf("Unable to compute result string")
	}

	err = SetResultCache(project_id, model_id, float64(expiry), result)
	if err != nil {
		log.Errorf("Unable to set result in cache")
		return "", err
	}

	return result, nil

}

func createCacheKey(projectId int64, model_id uint64) (*ch.Key, error) {
	modelIdString := strconv.FormatUint(model_id, 10)
	cache_key, err := ch.NewKey(projectId, "expv2", modelIdString)
	if err != nil {
		log.Errorf("Unable to create explain v2 redis key : %d , %s", projectId, modelIdString)
		return nil, err
	}
	return cache_key, nil
}

func SetResultCache(projectId int64, modelId uint64, expiry float64, result string) error {

	cacheKey, err := createCacheKey(projectId, modelId)
	if err != nil {
		return err
	}
	err = ch.SetPersistent(cacheKey, result, expiry)
	if err != nil {
		log.Errorf("Unable to set key/value in cache  ")
		return err
	}

	return nil

}

func GetResultCache(projectId int64, modelId uint64) (string, error) {

	cacheKey, err := createCacheKey(projectId, modelId)
	if err != nil {
		return "", err
	}
	result, exist, err := ch.GetIfExistsPersistent(cacheKey)
	if err != nil {

		log.Errorf("Unable to create explain v2 redis key : %d , %d", projectId, modelId)
		return "", err
	}
	if !exist {
		return "", nil
	}

	return result, nil
}

func RemoveCachedKey(projectId int64, modelId uint64) (bool, error) {

	cacheKey, err := createCacheKey(projectId, modelId)
	if err != nil {
		return false, err
	}

	err = ch.DelPersistent(cacheKey)
	if err != nil {
		return false, err
	}
	return true, nil
}
