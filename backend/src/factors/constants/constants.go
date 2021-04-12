package constants

import (
	"factors/model/model"
)

var SmartPropertyReservedNames = make(map[string]bool)

func SetSmartPropertiesReservedNames() {
	adwordsReservedNames := GetAdwordsReservedNamesForSmartProperties()
	facebookReservedNames := GetFacebookReservedNamesForSmartProperties()
	linkedinReservedNames := GetLinkedinReservedNamesForSmartProperties()
	for key, value := range adwordsReservedNames {
		SmartPropertyReservedNames[key] = value
	}
	for key, value := range facebookReservedNames {
		SmartPropertyReservedNames[key] = value
	}
	for key, value := range linkedinReservedNames {
		SmartPropertyReservedNames[key] = value
	}

}

func GetAdwordsReservedNamesForSmartProperties() map[string]bool {
	adwordsReservedNames := make(map[string]bool)
	for key, value := range model.AdwordsExtToInternal {
		adwordsReservedNames[key] = true
		adwordsReservedNames[value] = true
	}
	for key, value := range model.AdwordsInternalPropertiesToJobsInternal {
		adwordsReservedNames[key] = true
		adwordsReservedNames[value] = true
	}
	for key, value := range model.AdwordsInternalPropertiesToReportsInternal {
		adwordsReservedNames[key] = true
		adwordsReservedNames[value] = true
	}
	return adwordsReservedNames
}

func GetFacebookReservedNamesForSmartProperties() map[string]bool {
	facebookReservedNames := make(map[string]bool)
	for key, value := range model.FacebookExternalRepresentationToInternalRepresentation {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range model.FacebookInternalRepresentationToExternalRepresentation {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range model.ObjectToValueInFacebookJobsMapping {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range model.ObjectAndKeyInFacebookToPropertyMapping {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	for key, value := range model.FacebookObjectMapForSmartProperty {
		facebookReservedNames[key] = true
		facebookReservedNames[value] = true
	}
	return facebookReservedNames
}
func GetLinkedinReservedNamesForSmartProperties() map[string]bool {
	linkedinReservedNames := make(map[string]bool)
	for key, value := range model.LinkedinExternalRepresentationToInternalRepresentation {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range model.LinkedinInternalRepresentationToExternalRepresentation {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range model.ObjectToValueInLinkedinJobsMapping {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range model.ObjectAndKeyInLinkedinToPropertyMapping {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range model.ObjectAndKeyInLinkedinToPropertyMapping {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	for key, value := range model.LinkedinObjectMapForSmartProperty {
		linkedinReservedNames[key] = true
		linkedinReservedNames[value] = true
	}
	return linkedinReservedNames
}
