package model

type CRMSetting struct {
	ProjectID                      uint64 `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	HubspotEnrichHeavy             bool   `gorm:"default:false" json:"hubspot_enrich_heavy"`
	HubspotEnrichHeavyMaxCreatedAt *int64 `gorm:"default:null" json:"hubspot_enrich_heavy_max_created_at"`
}

type CRMSettingOption func(FieldsToUpdate)

func HubspotEnrichHeavy(isHeavy bool, heavyMaxCreatedAt *int64) CRMSettingOption {
	if isHeavy {
		return func(fields FieldsToUpdate) {
			fields["hubspot_enrich_heavy"] = true
			fields["hubspot_enrich_heavy_max_created_at"] = *heavyMaxCreatedAt
		}
	}

	return func(fields FieldsToUpdate) {
		fields["hubspot_enrich_heavy"] = false
		fields["hubspot_enrich_heavy_max_created_at"] = nil
	}
}
