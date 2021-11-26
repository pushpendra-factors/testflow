package model

import (
	"errors"
	"strings"
)

// Common/Util methods - to both adwords, facebook and all channels.
func TransformChannelsPropertiesConfigToKpiPropertiesConfig(channelsWithProperties []ChannelObjectAndProperties) []map[string]string {
	var resultantPropertiesConfig []map[string]string
	var tempPropertyConfig map[string]string

	for _, channelAndProperties := range channelsWithProperties {
		for _, property := range channelAndProperties.Properties {
			tempPropertyConfig = map[string]string{
				"name":         channelAndProperties.Name + "_" + property.Name,
				"display_name": channelAndProperties.Name + "_" + property.Name,
				"data_type":    property.Type,
				"object_type":  channelAndProperties.Name,
				"entity":       EventEntity,
			}
			resultantPropertiesConfig = append(resultantPropertiesConfig, tempPropertyConfig)
		}
	}
	return resultantPropertiesConfig
}

func TransformKPIQueryToChannelsV1Query(kpiQuery KPIQuery) (ChannelQueryV1, error) {
	var currentChannel string
	var err error
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
	currentChannel, err = GetChannelFromKPIQuery(kpiQuery.DisplayCategory)
	if err != nil {
		return ChannelQueryV1{}, err
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
			Property: strings.TrimPrefix(kpiGroupBy.PropertyName, kpiGroupBy.ObjectType+"_"),
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
			Property:  strings.TrimPrefix(kpiFilter.PropertyName, kpiFilter.ObjectType+"_"),
			Condition: kpiFilter.Condition,
			Value:     kpiFilter.Value,
			LogicalOp: kpiFilter.LogicalOp,
		}
		resultChannelFilters = append(resultChannelFilters, tempChannelFilter)
	}
	return resultChannelFilters
}

func GetChannelFromKPIQuery(displayCategory string) (string, error) {
	var currentChannel string
	var exists bool
	if currentChannel, exists = MapOfCategoryToChannel[displayCategory]; !exists {
		return "", errors.New("wrong Display Category given for channels")
	}
	return currentChannel, nil
}
