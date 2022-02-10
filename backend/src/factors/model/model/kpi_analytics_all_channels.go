package model

import (
	U "factors/util"
)

const (
	AllChannelsDisplayCategory = "all_channels_metrics"
)

func GetKPIConfigsForAllChannels() map[string]interface{} {
	allChannelProperties := []map[string]string{
		{"name": CAFilterCampaign + "_" + Id, "display_name": CAFilterCampaign + "_" + Id, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterCampaign},
		{"name": CAFilterCampaign + "_" + Name, "display_name": CAFilterCampaign + "_" + Name, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterCampaign},
		{"name": CAFilterAdGroup + "_" + Id, "display_name": CAFilterAdGroup + "_" + Id, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterAdGroup},
		{"name": CAFilterAdGroup + "_" + Name, "display_name": CAFilterAdGroup + "_" + Name, "data_type": U.PropertyTypeCategorical, "object_type": CAFilterAdGroup},
		{"name": "channel_name", "display_name": "channel_name", "data_type": U.PropertyTypeCategorical, "object_type": "channel"},
	}
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": AllChannelsDisplayCategory,
		"properties":       allChannelProperties,
	}
	config["metrics"] = GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}
