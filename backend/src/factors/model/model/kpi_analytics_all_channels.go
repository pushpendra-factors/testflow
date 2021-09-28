package model

import (
	U "factors/util"
)

const (
	AllChannelsDisplayCategory = "all_channels_metrics"
)

func GetKPIConfigsForAllChannels() map[string]interface{} {
	allChannelProperties := []map[string]string{
		{"name": Id, "display_name": Id, "data_type": U.PropertyTypeCategorical, "entity": CAFilterCampaign},
		{"name": Name, "display_name": Name, "data_type": U.PropertyTypeCategorical, "entity": CAFilterCampaign},
		{"name": Id, "display_name": Id, "data_type": U.PropertyTypeCategorical, "entity": CAFilterAdGroup},
		{"name": Name, "display_name": Name, "data_type": U.PropertyTypeCategorical, "entity": CAFilterAdGroup},
	}
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": AllChannelsDisplayCategory,
		"properties":       allChannelProperties,
	}
	config["metrics"] = GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}
