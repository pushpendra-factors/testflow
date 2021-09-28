package model

import (
	U "factors/util"
)

const (
	LinkedinDisplayCategory = "linkedin_metrics"
)

func GetKPIConfigsForLinkedin() map[string]interface{} {
	linkedInProperties := []map[string]string{
		{"name": Id, "display_name": Id, "data_type": U.PropertyTypeCategorical, "entity": CAFilterCampaign},
		{"name": Name, "display_name": Name, "data_type": U.PropertyTypeCategorical, "entity": CAFilterCampaign},
		{"name": Id, "display_name": Id, "data_type": U.PropertyTypeCategorical, "entity": CAFilterAdGroup},
		{"name": Name, "display_name": Name, "data_type": U.PropertyTypeCategorical, "entity": CAFilterAdGroup},
	}
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": FacebookDisplayCategory,
		"properties":       linkedInProperties,
	}
	config["metrics"] = GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}
