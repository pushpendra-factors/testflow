package delta

func GetPriorityKeysMapConversion(projectId int64, version int) map[string]float64 {
	PriorityKeysConversion_ := map[string]float64{
		// add new values here along with boost factor 2 = high, 1.1 = medium
		"$source":          2,
		"$medium":          2,
		"$campaign":        2,
		"$referrer_domain": 2,
		"$landing_page":    2,
		"$country":         2,
		"$device_type":     1.1,
		"$device_model":    1.1,
		"$os":              1.1,
		"$browser":         1.1,
		"$browser_version": 1.1,
		"$city":            1.1,
	}
	PriorityKeysConversionV2_ := map[string]float64{
		// add new values here along with boost factor 2 = high, 1.1 = medium
		"$campaign":         2,
		"$referrer_domain":  2,
		"$landing_page":     2,
		"$country":          2,
		"$device_model":     1.1,
		"$browser_version":  1.1,
		"$city":             1.1,
		"$initial_campaign": 1.5,
		"$initial_channel":  1.5,
		"$initial_page_url": 1.5,
		"$latest_channel":   1.5,
		"$latest_campaign":  1.5,
		"$latest_source":    1.5,
		"$latest_medium":    1.5,
		"$channel":          1.5,
		"$medium":           1.5,
		"$source":           1.5,
		"$device_type":      0.5,
		"$os_version":       0.5,
		"$os":               0.5,
		"$browser":          0.5,
		"$platform":         0.5,
		"$device_brand":     0.5,
		"$continent":        0.5,
		"$postal_code":      0.01,
		"$campaign_id":      0.01,
		"$adgroup_id":       0.01,
		"$creative":         0.5,
		"$content":          0.5,
	}
	if version == 1 {
		return PriorityKeysConversion_
	}
	if version == 2 {
		propertiesFromFile := GetPropertiesFromFile(projectId)
		for property, _ := range propertiesFromFile {
			PriorityKeysConversionV2_[property] = 4
		}
		return PriorityKeysConversionV2_
	}
	return PriorityKeysConversion_
}
func GetPriorityKeysMapDistribution(projectId int64, version int) map[string]float64 {
	PriorityKeysDistribution_ := map[string]float64{}
	PriorityKeysDistributionV2_ := map[string]float64{
		"$initial_campaign": 1.5,
		"$initial_channel":  1.5,
		"$initial_page_url": 1.5,
		"$latest_channel":   1.5,
		"$latest_campaign":  1.5,
		"$latest_source":    1.5,
		"$latest_medium":    1.5,
		"$channel":          1.5,
		"$medium":           1.5,
		"$source":           1.5,
		"$device_type":      0.5,
		"$os_version":       0.5,
		"$os":               0.5,
		"$browser":          0.5,
		"$platform":         0.5,
		"$device_brand":     0.5,
		"$continent":        0.5,
		"$postal_code":      0.01,
		"$campaign_id":      0.01,
		"$adgroup_id":       0.01,
		"$creative":         0.5,
		"$content":          0.5,
	}
	if version == 1 {
		return PriorityKeysDistribution_
	}
	if version == 2 {
		propertiesFromFile := GetPropertiesFromFile(projectId)
		for property, _ := range propertiesFromFile {
			PriorityKeysDistributionV2_[property] = 4
		}
		return PriorityKeysDistributionV2_
	}
	return PriorityKeysDistribution_
}
func GetBlackListedKeys() map[string]bool {
	BlackListedKeys_ := map[string]bool{
		"$day_of_week":                 true,
		"$page_raw_url":                true,
		"$initial_page_domain":         true,
		"$timestamp":                   true,
		"$initial_page_raw_url":        true,
		"$session_latest_page_raw_url": true,
		"$session_latest_page_url":     true,
		"$gclid":                       true,
		"$hubspot_contact_hs_calculated_form_submissions": true,
		"$latest_gclid":               true,
		"$initial_gclid":              true,
		"$joinTime":                   true,
		"$ip":                         true,
		"$latest_referrer":            true,
		"$latest_referrer_url":        true,
		"$initial_referrer":           true,
		"$initial_referrer_url":       true,
		"$referrer":                   true,
		"$referrer_url":               true,
		"$latest_page_url":            true,
		"$latest_page_domain":         true,
		"$latest_page_raw_url":        true,
		"$latest_page_load_time":      true,
		"$latest_page_spent_time":     true,
		"$latest_page_scroll_percent": true,
		"$email":                      true,
		"$os_with_version":            true,
		"$hour_of_first_event":        true,
		"$day_of_first_event":         true,
		"$fbclid":                     true,
	}
	return BlackListedKeys_
}
