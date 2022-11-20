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
				"display_name": strings.Replace(channelAndProperties.Name+"_"+property.Name, "_", " ", -1),
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
	currentChannel, err = GetChannelFromKPIQuery(kpiQuery.DisplayCategory, kpiQuery.Category)
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

func GetChannelFromKPIQuery(displayCategory string, category string) (string, error) {
	var currentChannel string
	var exists bool
	if category == CustomChannelCategory {
		return displayCategory, nil
	}
	if currentChannel, exists = MapOfCategoryToChannel[displayCategory]; !exists {
		return "", errors.New("wrong Display Category given for channels")
	}
	return currentChannel, nil
}

func GetCustomChannelFromKPIQuery() (string, error) {
	return "custom_ads", nil
}

func GetTransformedHeadersForChannels(headers []string, hasAnyGroupByTimestamp bool, hasAnyGroupBy bool) []string {
	if headers[0] == AliasError {
		return headers
	}
	currentHeaders := headers
	size := len(currentHeaders)
	currentHeaders[size-1] = AliasAggr
	if hasAnyGroupBy && hasAnyGroupByTimestamp {
		resultantHeaders := make([]string, 0)
		resultantHeaders = append(resultantHeaders, currentHeaders[size-2])
		resultantHeaders = append(resultantHeaders, currentHeaders[:size-2]...)
		resultantHeaders = append(resultantHeaders, currentHeaders[size-1])
		currentHeaders = resultantHeaders
	}
	return currentHeaders
}

func GetChannelFiltersGrouped(properties []ChannelFilterV1) [][]ChannelFilterV1 {
	groupedProperties := make([][]ChannelFilterV1, 0)
	currentGroupedProperties := make([]ChannelFilterV1, 0)
	for index, p := range properties {
		if index == 0 || p.LogicalOp != "AND" {
			currentGroupedProperties = append(currentGroupedProperties, p)
		} else {
			groupedProperties = append(groupedProperties, currentGroupedProperties)

			currentGroupedProperties = make([]ChannelFilterV1, 0)
			currentGroupedProperties = append(currentGroupedProperties, p)
		}
	}
	if len(currentGroupedProperties) != 0 {
		groupedProperties = append(groupedProperties, currentGroupedProperties)
	}
	return groupedProperties
}
