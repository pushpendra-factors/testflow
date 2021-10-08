package delta

func GetPriorityKeysMapConversion() map[string]float64 {
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
	return PriorityKeysConversion_
}
func GetPriorityKeysMapDistribution() map[string]float64 {
	PriorityKeysDistribution_ := map[string]float64{}
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
	}
	return BlackListedKeys_
}
