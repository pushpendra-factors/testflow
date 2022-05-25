package model

type CRMSetting struct {
	ProjectID          uint64 `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	HubspotEnrichHeavy bool   `gorm:"default:false" json:"hubspot_enrich_heavy"`
}

type CRMSettingOption func(FieldsToUpdate)

func HubspotEnrichHeavy(value bool) CRMSettingOption {
	return func(fields FieldsToUpdate) {
		fields["hubspot_enrich_heavy"] = value
	}
}
