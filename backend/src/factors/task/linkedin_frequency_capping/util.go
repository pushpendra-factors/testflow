package linkedin_frequency_capping

import (
	"factors/model/model"
	U "factors/util"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func removeExistingExclusionsFromDataset(linkedinCappingDataSet []model.LinkedinCappingDataSet, exclusionsMap map[string]map[string]bool) []model.LinkedinCappingDataSet {
	sanitizedDataset := make([]model.LinkedinCappingDataSet, 0)
	for _, data := range linkedinCappingDataSet {
		if _, exists := exclusionsMap[data.CampaignID][data.OrgID]; !exists {
			sanitizedDataset = append(sanitizedDataset, data)
		}
	}
	return sanitizedDataset
}
func decodeRules(rules []model.LinkedinCappingRule) ([]model.LinkedinCappingRuleWithDecodedValues, error) {
	decodedRules := make([]model.LinkedinCappingRuleWithDecodedValues, 0)
	for _, rule := range rules {
		objectIDs := make([]string, 0)
		advRules := make([]model.AdvancedRuleFilters, 0)
		if rule.ObjectIDs != nil {
			err := U.DecodePostgresJsonbToStructType(rule.ObjectIDs, &objectIDs)
			if err != nil {
				return decodedRules, err
			}
		}
		if rule.AdvancedRules != nil {
			err := U.DecodePostgresJsonbToStructType(rule.AdvancedRules, &advRules)
			if err != nil {
				return decodedRules, err
			}
		}
		decodedRules = append(decodedRules, model.LinkedinCappingRuleWithDecodedValues{Rule: rule, ObjectIDs: objectIDs, AdvancedRules: advRules})
	}
	return decodedRules, nil
}

func groupRulesByObjectType(rules []model.LinkedinCappingRuleWithDecodedValues) map[string][]model.LinkedinCappingRuleWithDecodedValues {
	groupedRules := make(map[string][]model.LinkedinCappingRuleWithDecodedValues)
	for _, ruleStruct := range rules {
		rule := ruleStruct.Rule
		if _, exists := groupedRules[rule.ObjectType]; !exists {
			groupedRules[rule.ObjectType] = make([]model.LinkedinCappingRuleWithDecodedValues, 0)
		}
		groupedRules[rule.ObjectType] = append(groupedRules[rule.ObjectType], ruleStruct)
	}
	return groupedRules
}

func getKeyFromMap(valueMap map[string]bool) []string {
	keys := make([]string, 0)
	for key := range valueMap {
		keys = append(keys, key)
	}
	return keys
}

func buildExistingExclusionsMapFromCampaignAndExclusionTable(campaigns map[string]model.CampaignNameTargetingCriteria, exclusions []model.LinkedinExclusion) map[string]map[string]bool {
	exclusionsMap := make(map[string]map[string]bool)
	for campaignID, targetingCriteria := range campaigns {
		if _, exists := exclusionsMap[campaignID]; !exists {
			exclusionsMap[campaignID] = make(map[string]bool)
		}
		for orgID := range targetingCriteria.TargetingCriteria {
			exclusionsMap[campaignID][orgID] = true
		}
	}
	for _, exclusion := range exclusions {
		var campaigns []model.CampaignsIDName
		_ = U.DecodePostgresJsonbToStructType(exclusion.Campaigns, campaigns)
		for _, campaign := range campaigns {
			if _, exists := exclusionsMap[campaign.CampaignID]; !exists {
				exclusionsMap[campaign.CampaignID] = make(map[string]bool)
			}
			exclusionsMap[campaign.CampaignID][exclusion.OrgID] = true
		}
	}
	return exclusionsMap
}

func getListOfMatchedCampaignsMapBasedOnCriteria(ruleWithDecodedValue model.LinkedinCappingRuleWithDecodedValues, campaignsMap map[string]model.CampaignNameTargetingCriteria) (map[string]model.CampaignNameTargetingCriteria, error) {
	reqCampaignsMap := make(map[string]model.CampaignNameTargetingCriteria)
	rule := ruleWithDecodedValue.Rule

	if rule.ObjectType == model.LINKEDIN_ACCOUNT {
		return campaignsMap, nil
	} else if rule.ObjectType == model.LINKEDIN_CAMPAIGN_GROUP {
		objectIDs := ruleWithDecodedValue.ObjectIDs
		for campaignID, campaignCriteria := range campaignsMap {
			if checkIfIDExistsInArray(campaignCriteria.CampaignGroupID, objectIDs) {
				reqCampaignsMap[campaignID] = campaignCriteria
			}
		}
	} else {
		objectIDs := ruleWithDecodedValue.ObjectIDs
		for campaignID, campaignCriteria := range campaignsMap {
			if checkIfIDExistsInArray(campaignID, objectIDs) {
				reqCampaignsMap[campaignID] = campaignCriteria
			}
		}
	}
	return reqCampaignsMap, nil
}
func checkIfIDExistsInArray(objectID string, objectIDArr []string) bool {
	for _, id := range objectIDArr {
		if id == objectID {
			return true
		}
	}
	return false
}

func buildMapOfCampaignToTargetingCriteria(campaigns []map[string]interface{}) map[string]model.CampaignNameTargetingCriteria {
	mapToReturn := make(map[string]model.CampaignNameTargetingCriteria)
	for _, campaign := range campaigns {
		excludedCompanies := make(map[string]bool, 0)
		campaignIDInt, campaignName, adAccountID, campaignGroupID := int64(campaign["id"].(float64)), campaign["name"].(string), strings.Split(campaign["account"].(string), ":")[3], strings.Split(campaign["campaignGroup"].(string), ":")[3]
		campaignID := fmt.Sprintf("%d", campaignIDInt)
		if targetingCriteria, ok := campaign["targetingCriteria"].(map[string]interface{}); ok {
			if excludeList, ok := targetingCriteria["exclude"].(map[string]interface{}); ok {
				if excludeListType, ok := excludeList["or"].(map[string]interface{}); ok {
					if orgList, ok := excludeListType["urn:li:adTargetingFacet:employers"].([]interface{}); ok {
						for _, v := range orgList {
							excludedCompanies[v.(string)] = true
						}
					}
				}
			}
		}
		mapToReturn[campaignID] = model.CampaignNameTargetingCriteria{
			CampaignName:      campaignName,
			AdAccountID:       adAccountID,
			CampaignGroupID:   campaignGroupID,
			TargetingCriteria: excludedCompanies,
		}
	}

	return mapToReturn
}

func GetStartAndTodayDateForCurrMonth() (int64, int64) {
	currYear, currMon, _ := time.Now().Date()
	startOfMonth, _ := strconv.ParseInt(time.Date(currYear, currMon, 1, 0, 0, 0, 0, time.Now().Location()).Format("20060102"), 10, 64)
	currDate, _ := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
	return startOfMonth, currDate
}

func GetStartAndEndDateForPreviousMonth() (int64, int64) {
	currYear, currMon, _ := time.Now().Date()
	startOfCurrMonth := time.Date(currYear, currMon, 1, 0, 0, 0, 0, time.Now().Location())
	startOfPrevMonth, _ := strconv.ParseInt(startOfCurrMonth.AddDate(0, -1, 0).Format("20060102"), 10, 64)
	endOfPrevMonth, _ := strconv.ParseInt(startOfCurrMonth.AddDate(0, 0, -1).Format("20060102"), 10, 64)
	return startOfPrevMonth, endOfPrevMonth
}
