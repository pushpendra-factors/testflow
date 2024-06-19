package linkedin_frequency_capping

import (
	"factors/model/model"
	"factors/model/store"
)

var exclusionAlreadyPresent = make(map[string]map[string]bool)
var domainToGroupRelatedDataMap = make(map[string]model.GroupRelatedData)
var nonMatchedDataSet = make([]model.LinkedinCappingDataSet, 0)
var ruleMatchedDataSet = make([]model.RuleMatchedDataSet, 0)
var campaignTargetingCriteriaMap = make(map[string]model.CampaignNameTargetingCriteria)
var allCampaigns = make([]string, 0)
var campaignGroupToCampaignIDsMap = make(map[string][]string, 0)

func GetExistingExclusionsInDataStore() *map[string]map[string]bool {
	return &exclusionAlreadyPresent
}

func UpdateExsitingExclusionsMapWithNewValues(orgID string, campaigns []model.CampaignsIDName) {
	for _, campaign := range campaigns {
		if _, exists := exclusionAlreadyPresent[campaign.CampaignID]; !exists {
			exclusionAlreadyPresent[campaign.CampaignID] = make(map[string]bool)
		}
		exclusionAlreadyPresent[campaign.CampaignID][orgID] = true
	}
}

func GetNonMatchedDataSetInDataStore() *[]model.LinkedinCappingDataSet {
	return &nonMatchedDataSet
}
func GetDomainToGroupRelatedDataMapInDataStore() *map[string]model.GroupRelatedData {
	return &domainToGroupRelatedDataMap
}

func SetDomainToGroupRelatedDataMapInDataStore(updatedData *map[string]model.GroupRelatedData) {
	domainToGroupRelatedDataMap = *updatedData
}

func SetNonMatchedDataSetInDataStore(dataset []model.LinkedinCappingDataSet) {
	nonMatchedDataSet = dataset
}

func GetRuleMatchedDataSetInDataStore() *[]model.RuleMatchedDataSet {
	return &ruleMatchedDataSet
}

func AppendRuleMatchedDataSetInDataStore(dataset model.RuleMatchedDataSet) {
	ruleMatchedDataSet = append(ruleMatchedDataSet, dataset)
}

func GetCampaignTargetingCriteriaMapInDataStore() *map[string]model.CampaignNameTargetingCriteria {
	return &campaignTargetingCriteriaMap
}

func SetCampaignTargetingCriteriaMapInDataStore(targetingCriteriaMap map[string]model.CampaignNameTargetingCriteria) {
	campaignTargetingCriteriaMap = targetingCriteriaMap
}

func SetCampaignIDsInDataStore(targetingCriteriaMap map[string]model.CampaignNameTargetingCriteria) {
	for key, value := range targetingCriteriaMap {
		allCampaigns = append(allCampaigns, key)

		if _, exists := campaignGroupToCampaignIDsMap[value.CampaignGroupID]; !exists {
			campaignGroupToCampaignIDsMap[value.CampaignGroupID] = make([]string, 0)
		}
		campaignGroupToCampaignIDsMap[value.CampaignGroupID] = append(campaignGroupToCampaignIDsMap[value.CampaignGroupID], key)
	}
}

func SetGroupRelatedDataInDataStore(projectID int64, groupID int, linkedinCappingDataSet []model.LinkedinCappingDataSet) error {
	for _, data := range linkedinCappingDataSet {
		groupRelatedData, err := store.GetStore().GetGroupRelatedData(projectID, groupID, data.CompanyDomain, domainToGroupRelatedDataMap)
		if err != nil {
			return err
		}
		domainToGroupRelatedDataMap[data.CompanyDomain] = groupRelatedData
	}
	return nil
}
func GetDomainToGroupRelatedDataInDataStore(projectID int64, groupID int, domain string) (model.GroupRelatedData, error) {
	if _, exists := domainToGroupRelatedDataMap[domain]; exists {
		return domainToGroupRelatedDataMap[domain], nil
	}
	groupRelatedData, err := store.GetStore().GetGroupRelatedData(projectID, groupID, domain, domainToGroupRelatedDataMap)
	if err != nil {
		return model.GroupRelatedData{}, err
	}
	domainToGroupRelatedDataMap[domain] = groupRelatedData

	return domainToGroupRelatedDataMap[domain], nil
}

func ResetDatastore() {
	exclusionAlreadyPresent = make(map[string]map[string]bool)
	domainToGroupRelatedDataMap = make(map[string]model.GroupRelatedData)
	nonMatchedDataSet = make([]model.LinkedinCappingDataSet, 0)
	ruleMatchedDataSet = make([]model.RuleMatchedDataSet, 0)
	campaignTargetingCriteriaMap = make(map[string]model.CampaignNameTargetingCriteria)
}
