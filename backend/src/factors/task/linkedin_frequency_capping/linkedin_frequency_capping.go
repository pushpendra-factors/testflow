package linkedin_frequency_capping

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	"net/http"
	"strconv"
	"strings"
	"time"
)

func PerformLinkedinExclusionsForProject(linkedinSetting model.LinkedinProjectSettings, dryRun, pushToLinkedin, removeFromLinkedinCustom bool) (string, int) {
	defer ResetDatastore()
	projectIDstr, accessToken, adAccounts := linkedinSetting.ProjectId, linkedinSetting.IntLinkedinAccessToken, strings.Split(linkedinSetting.IntLinkedinAdAccount, ",")

	projectID, _ := strconv.ParseInt(projectIDstr, 10, 64)
	currMonthStart, currDate := GetStartAndTodayDateForCurrMonth()
	prevMonthStart, prevMonthEnd := GetStartAndEndDateForPreviousMonth()
	isTodayFirstOfMonth := currDate == currMonthStart

	errMsg, errCode := setExistingExclusionsInDataStore(projectID, adAccounts, accessToken, currMonthStart, currDate)
	if errMsg != "" {
		return errMsg, errCode
	}
	if pushToLinkedin && (isTodayFirstOfMonth || removeFromLinkedinCustom) {
		errMsg, errCode := GetExclusionFromDBAndRemoveFromLinkedin(projectID, prevMonthStart, prevMonthEnd, accessToken)
		if errMsg != "" {
			return errMsg, errCode
		}
	}

	errMsg, errCode = ApplyRulesAndGetRuleMatchedDataset(projectID, currMonthStart, currDate)
	if errMsg != "" {
		return errMsg, errCode
	}

	errMsg, errCode = BuildAndPushExclusionToDBFromMatchedRules(projectID, currDate)
	if errMsg != "" {
		return errMsg, errCode
	}
	if pushToLinkedin {
		errMsg, errCode = GetExclusionFromDBAndPushToLinkedin(projectID, currMonthStart, accessToken)
		if errMsg != "" {
			return errMsg, errCode
		}
	}
	return "", http.StatusOK
}

// Comments
func setExistingExclusionsInDataStore(projectID int64, adAccounts []string, accessToken string, currMonthStart, currDate int64) (string, int) {
	// Variable/method name change.
	existingExclusionsMap := GetExistingExclusionsInDataStore()
	// TODO think.
	campaigns, errCode := getAllActivePausedCampaignsFromLinkedin(projectID, adAccounts, accessToken)
	if errCode != http.StatusOK {
		return "Failed to find campaign from linkedin", errCode
	}
	// think of a name
	campaignsMap := buildMapOfCampaignToTargetingCriteria(campaigns)
	SetCampaignTargetingCriteriaMapInDataStore(campaignsMap)
	// Can remove
	SetCampaignIDsInDataStore(campaignsMap)

	exclusions, errCode := store.GetStore().GetAllLinkedinCappingExclusionsForTimerange(projectID, currMonthStart, currDate)
	if errCode != http.StatusFound {
		return "Failed to find exclusion from db for current month", errCode
	}
	(*existingExclusionsMap) = buildExistingExclusionsMapFromCampaignAndExclusionTable(campaignsMap, exclusions)

	return "", http.StatusOK
}

func ApplyRulesAndGetRuleMatchedDataset(projectID int64, startOfMonthDateInt int64, currentDateInt int64) (string, int) {

	for _, objectType := range model.LINKEDIN_FREQUENCY_CAPPING_OBJECTS {
		errMsg, errCode := ApplyRuleOnLinkedinCappingDataSetForObjectType(projectID, startOfMonthDateInt, currentDateInt, objectType)
		if errMsg != "" || errCode != http.StatusOK {
			return errMsg, errCode
		}
	}

	return "", http.StatusOK
}

func ApplyRuleOnLinkedinCappingDataSetForObjectType(projectID int64, startOfMonthDateInt int64, currentDateInt int64, objectType string) (string, int) {
	linkedinCappingRules, errCode := store.GetStore().GetAllActiveLinkedinCappingRulesByObjectType(projectID, objectType)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		return "Failed to find rules", errCode
	}
	if len(linkedinCappingRules) == 0 {
		return "", http.StatusOK
	}
	decodedRules, err := decodeRules(linkedinCappingRules)
	if err != nil {
		return err.Error() + " - rule decode", http.StatusInternalServerError
	}
	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound {
		return "failed to get group", errCode
	}

	linkedinCappingDataSet, errCode := store.GetStore().GetDataSetForFrequencyCappingForMonthForObjectType(
		projectID, startOfMonthDateInt, objectType)
	if errCode != http.StatusOK {
		return "fail to get data set for account level rules", errCode
	}

	for _, linkedinCappingData := range linkedinCappingDataSet {
		groupData, err := GetDomainToGroupRelatedDataInDataStore(projectID, group.ID, linkedinCappingData.CompanyDomain)
		if err != nil {
			return err.Error() + " - get domain data", http.StatusInternalServerError
		}
		/* to think through, not required now
		if its excluded
		may be all 3 levels or 2 levels
		for all levels.
		campaigns = getListOFCampaignsWhichArenotExlcluded
		getLatestExclusionRecord/AuditLOgFromDb(org, AcccountLevel)
		createa new record with all the above metadata and add the current Campaigns
		Save it
		*/

		for _, ruleWithDecodedValue := range decodedRules {
			campaignsReq := getCampaignsRequiredForRuleEvaluationForDomain(linkedinCappingData.OrgID, ruleWithDecodedValue)
			if len(campaignsReq) == 0 {
				continue
			}
			isMatched, ruleMatchedData, err := store.GetStore().ApplyRuleOnLinkedinCappingData(projectID, linkedinCappingData,
				ruleWithDecodedValue, groupData)
			if err != nil {
				return err.Error() + " - rule execution", http.StatusInternalServerError
			}
			if isMatched {
				ruleMatchedData.Campaigns = campaignsReq
				AppendRuleMatchedDataSetInDataStore(ruleMatchedData)
				UpdateExsitingExclusionsMapWithNewValues(linkedinCappingData.OrgID, campaignsReq)
				break
			}
		}
	}
	return "", http.StatusOK
}

// Adwords - CampaignGroup. CampaignGroup - []

func getCampaignsRequiredForRuleEvaluationForDomain(orgID string, ruleWithdecodedValue model.LinkedinCappingRuleWithDecodedValues) []model.CampaignsIDName {
	campaignsIDName := make([]model.CampaignsIDName, 0)
	existingExclusionsMap := GetExistingExclusionsInDataStore()
	// unclear name.
	// will change.
	requiredCampaignsMap, err := getListOfMatchedCampaignsMapBasedOnCriteria(ruleWithdecodedValue, *GetCampaignTargetingCriteriaMapInDataStore()) // mathced campaign/ or similar somethikng
	if err != nil {
		return campaignsIDName
	}
	for campaignID, targetingCriteria := range requiredCampaignsMap {
		if _, exists := (*existingExclusionsMap)[campaignID][orgID]; !exists {
			campaignsIDName = append(campaignsIDName, model.CampaignsIDName{CampaignID: campaignID, CampaignName: targetingCriteria.CampaignName})
		}
	}
	return campaignsIDName
}

func BuildAndPushExclusionToDBFromMatchedRules(projectID int64, currentDate int64) (string, int) {
	exclusionsToAdd := make([]model.LinkedinExclusion, 0)
	ruleMatchedDataSet := GetRuleMatchedDataSetInDataStore()
	for _, data := range *ruleMatchedDataSet {
		exclusions, err := buildExclusionsWithMatchedRules(projectID, currentDate, data.Rule, data)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		exclusionsToAdd = append(exclusionsToAdd, exclusions...)
	}
	for _, exclusion := range exclusionsToAdd {
		errCode := store.GetStore().CreateLinkedinExclusion(projectID, exclusion)
		if errCode != http.StatusCreated {
			return "failed to create exclusion entry in db", errCode
		}
	}
	return "", http.StatusOK
}

func buildExclusionsWithMatchedRules(projectID int64, timestamp int64, rule model.LinkedinCappingRule, ruleMatchedData model.RuleMatchedDataSet) ([]model.LinkedinExclusion, error) {

	exclusionsToAdd := make([]model.LinkedinExclusion, 0)
	encodedRule, err := U.EncodeStructTypeToPostgresJsonb(rule)
	if err != nil {
		return make([]model.LinkedinExclusion, 0), err
	}
	_, _, currDay := time.Now().Date()
	daysRemainingInMonth := 30 - currDay
	data := ruleMatchedData.CappingData

	encodedProps, err := U.EncodeStructTypeToPostgresJsonb(ruleMatchedData.PropertiesMatched)
	if err != nil {
		return make([]model.LinkedinExclusion, 0), err
	}
	encodedCampaigns, err := U.EncodeStructTypeToPostgresJsonb(ruleMatchedData.Campaigns)
	if err != nil {
		return make([]model.LinkedinExclusion, 0), err
	}
	exclusion := model.LinkedinExclusion{
		ProjectID:          projectID,
		OrgID:              data.OrgID,
		Timestamp:          timestamp,
		CompanyName:        data.CompanyName,
		Campaigns:          encodedCampaigns,
		RuleID:             ruleMatchedData.Rule.ID,
		RuleObjectType:     ruleMatchedData.Rule.ObjectType,
		RuleSnapshot:       encodedRule,
		PropertiesSnapshot: encodedProps,
		ImpressionsSaved:   int64(daysRemainingInMonth * int(data.Impressions) / currDay),
		ClicksSaved:        int64(daysRemainingInMonth * int(data.Clicks) / currDay),
	}

	exclusionsToAdd = append(exclusionsToAdd, exclusion)
	return exclusionsToAdd, nil
}
