package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type SmartProperties struct {
	ProjectID      uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ObjectType     int             `gorm:"primary_key:true;auto_increment:false" json:"object_type"`
	ObjectID       string          `gorm:"primary_key:true;auto_increment:false" json:"object_id"`
	ObjectProperty *postgres.Jsonb `json:"object_property"`
	Properties     *postgres.Jsonb `json:"properties"`
	RulesRef       *postgres.Jsonb `json:"rules_ref"`
	Source         string          `gorm:"primary_key:true;auto_increment:false" json:"source"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

var SmartPropertyReservedNames = make(map[string]bool)

func SetSmartPropertiesReservedNames() {
	adwordsReservedNames := GetAdwordsReservedNamesForSmartProperties()
	facebookReservedNames := GetFacebookReservedNamesForSmartProperties()
	linkedinReservedNames := GetLinkedinReservedNamesForSmartProperties()
	bingadsReservedNames := GetBingAdsReservedNamesForSmartProperties()
	for key, value := range adwordsReservedNames {
		SmartPropertyReservedNames[key] = value
	}
	for key, value := range facebookReservedNames {
		SmartPropertyReservedNames[key] = value
	}
	for key, value := range linkedinReservedNames {
		SmartPropertyReservedNames[key] = value
	}
	for key, value := range bingadsReservedNames {
		SmartPropertyReservedNames[key] = value
	}

}

func GetAdwordsReservedNamesForSmartProperties() map[string]bool {
	adwordsReservedNames := make(map[string]bool)
	for key, value := range AdwordsExtToInternal {
		adwordsReservedNames[key] = true
		adwordsReservedNames[value] = true
	}
	for key, value := range AdwordsInternalPropertiesToJobsInternal {
		adwordsReservedNames[key] = true
		adwordsReservedNames[value] = true
	}
	for key, value := range AdwordsInternalPropertiesToReportsInternal {
		adwordsReservedNames[key] = true
		adwordsReservedNames[value] = true
	}
	return adwordsReservedNames
}

func GetFacebookReservedNamesForSmartProperties() map[string]bool {
	facebookReservedNames := make(map[string]bool)
	for key, value := range FacebookExternalRepresentationToInternalRepresentation {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range FacebookInternalRepresentationToExternalRepresentation {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range ObjectToValueInFacebookJobsMapping {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range ObjectAndKeyInFacebookToPropertyMapping {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range FacebookObjectMapForSmartProperty {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	return facebookReservedNames
}
func GetLinkedinReservedNamesForSmartProperties() map[string]bool {
	linkedinReservedNames := make(map[string]bool)
	for key, value := range LinkedinExternalRepresentationToInternalRepresentation {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range LinkedinInternalRepresentationToExternalRepresentation {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range ObjectToValueInLinkedinJobsMapping {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range ObjectAndKeyInLinkedinToPropertyMapping {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range ObjectAndKeyInLinkedinToPropertyMapping {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range LinkedinObjectMapForSmartProperty {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	return linkedinReservedNames
}
func GetBingAdsReservedNamesForSmartProperties() map[string]bool {
	bingadsReservedNames := make(map[string]bool)
	for key, value := range BingAdsInternalRepresentationToExternalRepresentation {
		bingadsReservedNames[key] = true
		bingadsReservedNames[value] = true
	}
	for key, value := range BingAdsInternalRepresentationToExternalRepresentationForReports {
		bingadsReservedNames[key] = true
		bingadsReservedNames[value] = true
	}
	return bingadsReservedNames
}
