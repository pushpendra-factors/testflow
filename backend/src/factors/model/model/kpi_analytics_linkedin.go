package model

import "strings"

const (
	LinkedinDisplayCategory = "linkedin_metrics"
)

var KpiLinkedinConfig = map[string]interface{}{
	"category":         ChannelCategory,
	"display_category": LinkedinDisplayCategory,
}

// Similar method is found in KPI_analytics_channel_common.
func TransformLinkedinChannelsPropertiesConfigToKpiPropertiesConfig(channelsWithProperties []ChannelObjectAndProperties) []map[string]string {
	var resultantPropertiesConfig []map[string]string
	var tempPropertyConfig map[string]string

	for _, channelAndProperties := range channelsWithProperties {
		displayNameForObjectType, _ := ObjectToDisplayCategoryForLinkedin[channelAndProperties.Name]
		for _, property := range channelAndProperties.Properties {
			tempPropertyConfig = map[string]string{
				"name":         channelAndProperties.Name + "_" + property.Name,
				"display_name": strings.Replace(displayNameForObjectType+"_"+property.Name, "_", " ", -1),
				"data_type":    property.Type,
				"object_type":  channelAndProperties.Name,
				"entity":       EventEntity,
			}
			resultantPropertiesConfig = append(resultantPropertiesConfig, tempPropertyConfig)
		}
	}
	return resultantPropertiesConfig
}


func GetKPIMetricsForLinkedin() []map[string]string {
	return GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
}
