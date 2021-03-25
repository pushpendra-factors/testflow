package postgres

import (
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	U "factors/util"
	"time"
	"strconv"
)

func (pg *Postgres) GetEventUserCountsOfAllProjects(lastNDays int) (map[string][]*model.ProjectAnalytics, error) {
	currentDate := time.Now().UTC()
	result := make(map[string][]*model.ProjectAnalytics, 0)
		for i := 0; i < lastNDays; i++ {
			dateKey := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
			if result[dateKey] == nil {
				result[dateKey] = make([]*model.ProjectAnalytics, 0)
			}
			totalUniqueUsersKey, err := model.UserCountAnalyticsCacheKey(dateKey)
			if err != nil {
				return nil, err
			}
			users, err := cacheRedis.ZrangeWithScoresPersistent(true, totalUniqueUsersKey)
			if err != nil {
				return nil, err
			}
			totalUniqueEventsKey, err := model.UniqueEventNamesAnalyticsCacheKey(dateKey)
			if err != nil {
				return nil, err
			}
			uniqueEvents, err := cacheRedis.ZrangeWithScoresPersistent(true, totalUniqueEventsKey)
			if err != nil {
				return nil, err
			}
			totalEventsKey, err := model.EventsCountAnalyticsCacheKey(dateKey)
			if err != nil {
				return nil, err
			}
			totalEvents, err := cacheRedis.ZrangeWithScoresPersistent(true, totalEventsKey)
			if err != nil {
				return nil, err
			}
			for projId, count := range users {
				uniqueUsers, _ := strconv.Atoi(count)
				totalEvents, _ := strconv.Atoi(totalEvents[projId])
				uniqueEvents, _ := strconv.Atoi(uniqueEvents[projId])
				projIdInt, _ := strconv.Atoi(projId)
				result[dateKey] = append(result[dateKey], &model.ProjectAnalytics{
					ProjectID:         uint64(projIdInt),
					TotalEvents:       uint64(totalEvents),
					TotalUniqueEvents: uint64(uniqueEvents),
					TotalUniqueUsers:  uint64(uniqueUsers),
				})

			}
			
		}
	
	return result, nil
}