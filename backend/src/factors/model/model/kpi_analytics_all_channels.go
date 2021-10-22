package model

import (
	U "factors/util"
)

const (
	AllChannelsDisplayCategory = "all_channels_metrics"
)

func GetKPIConfigsForAllChannels() map[string]interface{} {
	allChannelProperties := []map[string]string{
		{"name": Id, "display_name": Id, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterCampaign},
		{"name": Name, "display_name": Name, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterCampaign},
		{"name": Id, "display_name": Id, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterAdGroup},
		{"name": Name, "display_name": Name, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterAdGroup},
	}
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": AllChannelsDisplayCategory,
		"properties":       allChannelProperties,
	}
	config["metrics"] = GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}
