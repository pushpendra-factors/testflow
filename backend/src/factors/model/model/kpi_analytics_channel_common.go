package model

import "errors"

// Common/Util methods - to both adwords, facebook and all channels.
func tranformChannelConfigStructToKPISpecificConfig(channelConfig map[string]map[string]PropertiesAndRelated) []map[string]string {
	var resultantPropertiesConfig []map[string]string
	var tempPropertyConfig map[string]string

	// transforming properties.
	for objectType, mapOfPropertyToData := range channelConfig {
		for property, data := range mapOfPropertyToData {
			tempPropertyConfig = map[string]string{
				"name":         property,
				"display_name": property,
				"data_type":    data.TypeOfProperty,
				"object_type":  objectType,
				"entity":       EventEntity,
			}
			resultantPropertiesConfig = append(resultantPropertiesConfig, tempPropertyConfig)
		}
	}
	return resultantPropertiesConfig
}

func TransformKPIQueryToChannelsV1Query(kpiQuery KPIQuery) (ChannelQueryV1, error) {
	var currentChannel string
	var exists bool
	channelSelectMetrics := make([]string, 0)
	channelSelectMetrics = append(channelSelectMetrics, kpiQuery.Metrics...)
	channelQueryV1 := ChannelQueryV1{
		SelectMetrics:    channelSelectMetrics,
		GroupByTimestamp: kpiQuery.GroupByTimestamp,
		Timezone:         kpiQuery.Timezone,
		From:             kpiQuery.From,
		To:               kpiQuery.To,
	}
	channelQueryV1.GroupBy = transformGroupByKPIToChannelsV1(kpiQuery.GroupBy)
	channelQueryV1.Filters = transformFiltersKPIToChannelsV1(kpiQuery.Filters)
	if currentChannel, exists = MapOfCategoryToChannel[kpiQuery.DisplayCategory]; !exists {
		return ChannelQueryV1{}, errors.New("wrong Display Category given for channels")
	}
	channelQueryV1.Channel = currentChannel
	return channelQueryV1, nil
}

func transformGroupByKPIToChannelsV1(kpiGroupBys []KPIGroupBy) []ChannelGroupBy {
	var resultChannelGroupBys []ChannelGroupBy
	var tempChannelGroupBy ChannelGroupBy
	for _, kpiGroupBy := range kpiGroupBys {
		tempChannelGroupBy = ChannelGroupBy{
			Object:   kpiGroupBy.ObjectType,
			Property: kpiGroupBy.PropertyName,
		}
		resultChannelGroupBys = append(resultChannelGroupBys, tempChannelGroupBy)
	}
	return resultChannelGroupBys
}

func transformFiltersKPIToChannelsV1(kpiFilters []KPIFilter) []ChannelFilterV1 {
	var resultChannelFilters []ChannelFilterV1
	var tempChannelFilter ChannelFilterV1
	for _, kpiFilter := range kpiFilters {
		tempChannelFilter = ChannelFilterV1{
			Object:    kpiFilter.ObjectType,
			Property:  kpiFilter.PropertyName,
			Condition: kpiFilter.Condition,
			Value:     kpiFilter.Value,
			LogicalOp: kpiFilter.LogicalOp,
		}
		resultChannelFilters = append(resultChannelFilters, tempChannelFilter)
	}
	return resultChannelFilters
}
