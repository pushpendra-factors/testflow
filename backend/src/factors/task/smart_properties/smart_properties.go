package smart_properties

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"reflect"
	"strings"

	U "factors/util"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Status struct {
	ProjectID int64  `json:"project_id"`
	ErrCode   int    `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
	Type      string `json:"type"`
}

const (
	Rule_change = "rule_change"
	Current_day = "current_day"
)

func EnrichSmartPropertyForChangedRulesForProject(projectID int64) int {
	log.Warn("Smart properties enrichment for changed rule started for project ", projectID)
	var recordsEvaluated = 0
	var recordsUpdated = 0

	smartPropertyRules, errCode := store.GetStore().GetAllChangedSmartPropertyRulesForProject(projectID)
	if errCode != http.StatusFound {
		log.Error("failed to get smart properties rules for project ", projectID)
		return http.StatusInternalServerError
	}
	if len(smartPropertyRules) == 0 {
		log.Warn("No updated rules found for project ", projectID)
		return http.StatusOK
	}
	// Need to fetch only required projects that have integration.
	log.Warn("No of changed smart properties rules: ", len(smartPropertyRules))
	sources, _ := store.GetStore().GetCustomAdsSourcesByProject(projectID)
	adwordsCampaigns, adwordsAdGroups := store.GetStore().GetLatestMetaForAdwordsForGivenDays(projectID, 30)
	facebookCampaigns, facebookAdGroups := store.GetStore().GetLatestMetaForFacebookForGivenDays(projectID, 30)
	linkedinCampaigns, linkedinAdGroups := store.GetStore().GetLatestMetaForLinkedinForGivenDays(projectID, 30)
	bingadsCampaigns, bingadsAdGroups := store.GetStore().GetLatestMetaForBingAdsForGivenDays(projectID, 30)
	customadsCampaigns, customadsAdGroups := make([][]model.ChannelDocumentsWithFields, 0), make([][]model.ChannelDocumentsWithFields, 0)
	for _, source := range sources {
		customadsCampaign, customadsAdGroup := store.GetStore().GetLatestMetaForCustomAdsForGivenDays(projectID, source, 90)
		customadsCampaigns = append(customadsCampaigns, customadsCampaign)
		customadsAdGroups = append(customadsAdGroups, customadsAdGroup)
	}

	for _, smartPropertyRule := range smartPropertyRules {
		switch checkState(smartPropertyRule) {
		case model.CREATED:
			noOfRecordsUpdated, noOfRecordsEvaluated, errCode := createSmartPropertyFromRuleObject(smartPropertyRule, adwordsCampaigns, adwordsAdGroups, facebookCampaigns, facebookAdGroups, linkedinCampaigns, linkedinAdGroups, bingadsCampaigns, bingadsAdGroups, customadsCampaigns, customadsAdGroups, sources)
			recordsEvaluated += noOfRecordsEvaluated
			recordsUpdated += noOfRecordsUpdated
			if errCode != http.StatusCreated {
				continue
			}
		case model.DELETED:
			noOfRecordsUpdated, noOfRecordsEvaluated, errCode := deleteSmartPropertyFromRule(smartPropertyRule)
			recordsEvaluated += noOfRecordsEvaluated
			recordsUpdated += noOfRecordsUpdated
			if errCode != http.StatusAccepted {
				continue
			}
		case model.UPDATED:
			noOfRecordsUpdated, noOfRecordsEvaluated, errCode := updateSmartPropertyFromRule(smartPropertyRule, adwordsCampaigns, adwordsAdGroups, facebookCampaigns, facebookAdGroups, linkedinCampaigns, linkedinAdGroups, bingadsCampaigns, bingadsAdGroups, customadsCampaigns, customadsAdGroups, sources)
			recordsEvaluated += noOfRecordsEvaluated
			recordsUpdated += noOfRecordsUpdated
			if errCode != http.StatusAccepted {
				continue
			}
		default:
			log.Error("Invalid status")
		}

		if !C.IsDryRunSmartProperties() {
			smartPropertyRule.EvaluationStatus = model.EvaluationStatusMap["picked"]
			smartPropertyRule.TypeAlias = model.SmartPropertyRulesTypeToTypeAlias[smartPropertyRule.Type]
			_, errMsg, errCode := store.GetStore().UpdateSmartPropertyRules(smartPropertyRule.ProjectID, smartPropertyRule.ID, smartPropertyRule)
			if errCode != http.StatusAccepted {
				log.WithFields(log.Fields{"errMsg": errMsg, "smart_properties_rule": smartPropertyRule}).Error("failed to update smart properties rule.")
				continue
			}
		}
	}
	log.Warn("No. of records evaluated: ", recordsEvaluated)
	log.Warn("No. of records updated: ", recordsUpdated)
	log.Warn("Smart properties enrichment for changed rules ended for project ", projectID)
	return http.StatusOK
}

func EnrichSmartPropertyForCurrentDayForProject(projectID int64) int {
	log.Warn("Smart properties enrichment for current day's data started for project ", projectID)
	smartPropertyRules, errCode := store.GetStore().GetSmartPropertyRules(projectID)
	if errCode != http.StatusFound {
		log.Error("No rules found for project ", projectID)
		return http.StatusOK
	}
	recordsUpdated := 0
	recordsEvaluated := 0
	sources, _ := store.GetStore().GetCustomAdsSourcesByProject(projectID)
	adwordsCampaigns, adwordsAdGroups := store.GetStore().GetLatestMetaForAdwordsForGivenDays(projectID, 1)
	facebookCampaigns, facebookAdGroups := store.GetStore().GetLatestMetaForFacebookForGivenDays(projectID, 1)
	linkedinCampaigns, linkedinAdGroups := store.GetStore().GetLatestMetaForLinkedinForGivenDays(projectID, 1)
	bingadsCampaigns, bingadsAdGroups := store.GetStore().GetLatestMetaForBingAdsForGivenDays(projectID, 3)
	customadsCampaigns, customadsAdGroups := make([][]model.ChannelDocumentsWithFields, 0), make([][]model.ChannelDocumentsWithFields, 0)
	for _, source := range sources {
		customadsCampaign, customadsAdGroup := store.GetStore().GetLatestMetaForCustomAdsForGivenDays(projectID, source, 1)
		customadsCampaigns = append(customadsCampaigns, customadsCampaign)
		customadsAdGroups = append(customadsAdGroups, customadsAdGroup)
	}

	changedAdwordsCampaigns, changedAdwordsAdGroups := getUpdatedAndNonExistingInSmartPropertiesChannelDocuments(projectID, adwordsCampaigns, adwordsAdGroups, "adwords")
	changedFacebookCampaigns, changedFacebookAdGroups := getUpdatedAndNonExistingInSmartPropertiesChannelDocuments(projectID, facebookCampaigns, facebookAdGroups, "facebook")
	changedLinkedinCampaigns, changedLinkedinAdGroups := getUpdatedAndNonExistingInSmartPropertiesChannelDocuments(projectID, linkedinCampaigns, linkedinAdGroups, "linkedin")
	changedBingadsCampaigns, changedBingadsAdGroups := getUpdatedAndNonExistingInSmartPropertiesChannelDocuments(projectID, bingadsCampaigns, bingadsAdGroups, "bingads")
	changedCustomadsCampaigns, changedCustomadsAdGroups := make([][]model.ChannelDocumentsWithFields, 0), make([][]model.ChannelDocumentsWithFields, 0)
	for i, source := range sources {
		changedCustomadsCampaign, changedCustomadsAdGroup := getUpdatedAndNonExistingInSmartPropertiesChannelDocuments(projectID, customadsCampaigns[i], customadsAdGroups[i], source)
		changedCustomadsCampaigns = append(changedCustomadsCampaigns, changedCustomadsCampaign)
		changedCustomadsAdGroups = append(changedCustomadsAdGroups, changedCustomadsAdGroup)
	}
	for _, smartPropertyRule := range smartPropertyRules {
		noOfRecordsUpdated, noOfRecordsEvaluated, errCode := createSmartPropertyFromRuleObject(smartPropertyRule, changedAdwordsCampaigns, changedAdwordsAdGroups, changedFacebookCampaigns, changedFacebookAdGroups, changedLinkedinCampaigns, changedLinkedinAdGroups, changedBingadsCampaigns, changedBingadsAdGroups, changedCustomadsCampaigns, changedCustomadsAdGroups, sources)
		recordsUpdated += noOfRecordsUpdated
		recordsEvaluated += noOfRecordsEvaluated
		if errCode != http.StatusCreated {
			log.WithField("smart_properties_rule", smartPropertyRule).Error("Failed to create smart properties for rule")
			continue
		}
	}

	log.Warn("No. of records evaluated: ", recordsEvaluated)
	log.Warn("No. of records updated: ", recordsUpdated)
	log.Warn("Smart properties enrichment for current day's data ended for project ", projectID)
	return http.StatusOK
}
func checkState(smartPropertyRule model.SmartPropertyRules) string {
	if smartPropertyRule.CreatedAt == smartPropertyRule.UpdatedAt {
		return model.CREATED
	}
	if (smartPropertyRule.CreatedAt != smartPropertyRule.UpdatedAt) && smartPropertyRule.IsDeleted {
		return model.DELETED
	} else {
		return model.UPDATED
	}
}

//to do: If we intend to parallelise this at per project level, It might be better to evaluate in chunks. Can check later.
func getUpdatedAndNonExistingInSmartPropertiesChannelDocuments(projectID int64, campaigns []model.ChannelDocumentsWithFields,
	adGroups []model.ChannelDocumentsWithFields, source string) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	smartPropertyCampaigns, errCode := store.GetStore().GetSmartPropertyByProjectIDAndSourceAndObjectType(projectID, source, 1)
	if errCode != http.StatusFound {
		return make([]model.ChannelDocumentsWithFields, 0, 0), make([]model.ChannelDocumentsWithFields, 0, 0)
	}
	channelDocumentsCampaignsMap := make(map[string]model.ChannelDocumentsWithFields)
	for _, smartProperty := range smartPropertyCampaigns {
		var objectProperties model.ChannelDocumentsWithFields
		err := U.DecodePostgresJsonbToStructType(smartProperty.ObjectProperty, &objectProperties)
		if err != nil {
			log.WithField("smart_property", smartProperty).Error("Failed to decode object properties")
			continue
		}
		channelDocumentsCampaignsMap[objectProperties.CampaignID] = objectProperties
	}
	changedCampaignsChannelDocuments := make([]model.ChannelDocumentsWithFields, 0, 0)
	for _, campaign := range campaigns {
		existingCampaign, isPresent := channelDocumentsCampaignsMap[campaign.CampaignID]
		if !isPresent {
			changedCampaignsChannelDocuments = append(changedCampaignsChannelDocuments, campaign)
			continue
		}
		if !reflect.DeepEqual(campaign, existingCampaign) {
			errCode := store.GetStore().DeleteSmartPropertyByProjectIDAndSourceAndObjectID(projectID, source, campaign.CampaignID)
			if errCode != http.StatusAccepted {
				log.Error("Failed to delete existing smart property for object_id ", campaign.CampaignID)
				continue
			}
			changedCampaignsChannelDocuments = append(changedCampaignsChannelDocuments, campaign)
		}
	}

	smartPropertyAdGroups, errCode := store.GetStore().GetSmartPropertyByProjectIDAndSourceAndObjectType(projectID, source, 2)
	if errCode != http.StatusFound {
		return make([]model.ChannelDocumentsWithFields, 0, 0), make([]model.ChannelDocumentsWithFields, 0, 0)
	}
	channelDocumentsAdGroupsMap := make(map[string]model.ChannelDocumentsWithFields)
	for _, smartProperty := range smartPropertyAdGroups {
		var objectProperties model.ChannelDocumentsWithFields
		err := U.DecodePostgresJsonbToStructType(smartProperty.ObjectProperty, &objectProperties)
		if err != nil {
			log.WithField("smart_property", smartProperty).Error("Failed to decode object properties")
			continue
		}
		channelDocumentsAdGroupsMap[objectProperties.AdGroupID] = objectProperties
	}

	changedAdGroupsChannelDocuments := make([]model.ChannelDocumentsWithFields, 0, 0)
	for _, adGroup := range adGroups {
		existingAdGroup, isPresent := channelDocumentsAdGroupsMap[adGroup.AdGroupID]
		if !isPresent {
			changedAdGroupsChannelDocuments = append(changedAdGroupsChannelDocuments, adGroup)
			continue
		}
		if !reflect.DeepEqual(adGroup, existingAdGroup) {
			errCode := store.GetStore().DeleteSmartPropertyByProjectIDAndSourceAndObjectID(projectID, source, adGroup.AdGroupID)
			if errCode != http.StatusAccepted {
				log.Error("Failed to delete existing smart property for object_id ", adGroup.AdGroupID)
				continue
			}
			changedAdGroupsChannelDocuments = append(changedAdGroupsChannelDocuments, adGroup)
		}
	}

	return changedCampaignsChannelDocuments, changedAdGroupsChannelDocuments
}

func createSmartPropertyFromRuleObject(smartPropertyRule model.SmartPropertyRules, adwordsCampaigns []model.ChannelDocumentsWithFields,
	adwordsAdGroups []model.ChannelDocumentsWithFields, facebookCampaigns []model.ChannelDocumentsWithFields,
	facebookAdGroups []model.ChannelDocumentsWithFields, linkedinCampaigns []model.ChannelDocumentsWithFields,
	linkedinAdGroups []model.ChannelDocumentsWithFields, bingadsCampaigns []model.ChannelDocumentsWithFields,
	bingadsAdGroups []model.ChannelDocumentsWithFields, customadsCampaigns [][]model.ChannelDocumentsWithFields,
	customadsAdGroups [][]model.ChannelDocumentsWithFields, customSources []string) (int, int, int) {

	recordsUpdated := 0
	recordsEvaluated := 0
	var rules []model.Rule
	err := U.DecodePostgresJsonbToStructType(smartPropertyRule.Rules, &rules)
	if err != nil {
		log.Error("failed to decode rules")
		return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
	}
	noOfRecordsUpdated, noOfRecordsEvaluated, errCode := createSmartPropertyFromRuleObjectForSource(smartPropertyRule, adwordsCampaigns, adwordsAdGroups, rules, "adwords")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
	}

	noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertyFromRuleObjectForSource(smartPropertyRule, facebookCampaigns, facebookAdGroups, rules, "facebook")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
	}

	noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertyFromRuleObjectForSource(smartPropertyRule, linkedinCampaigns, linkedinAdGroups, rules, "linkedin")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
	}

	noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertyFromRuleObjectForSource(smartPropertyRule, bingadsCampaigns, bingadsAdGroups, rules, "bingads")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
	}

	for i, _ := range customadsCampaigns {
		noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertyFromRuleObjectForSource(smartPropertyRule, customadsCampaigns[i], customadsAdGroups[i], rules, customSources[i])
		recordsEvaluated += noOfRecordsEvaluated
		recordsUpdated += noOfRecordsUpdated
		if errCode != http.StatusCreated {
			return recordsUpdated, recordsEvaluated, errCode
		}
	}

	return recordsUpdated, recordsEvaluated, http.StatusCreated
}
func checkIfPropertyChanged(objectAndPropertyJson *postgres.Jsonb, channelDocument model.ChannelDocumentsWithFields) bool {
	objectAndProperty := model.ChannelDocumentsWithFields{}
	err := U.DecodePostgresJsonbToStructType(objectAndPropertyJson, &objectAndProperty)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(objectAndProperty, channelDocument)
}
func createSmartPropertyFromRuleObjectForSource(smartPropertyRule model.SmartPropertyRules,
	campaigns []model.ChannelDocumentsWithFields, adGroups []model.ChannelDocumentsWithFields, rules []model.Rule, source string) (int, int, int) {
	logCtx := log.WithFields(log.Fields{"smart_properties_rule": smartPropertyRule})
	recordsEvaluated := 0
	recordsUpdated := 0
	switch smartPropertyRule.Type {
	case 1:
		for _, campaign := range campaigns {
			recordsEvaluated += 1
			for _, rule := range rules {
				if rule.Source != source && rule.Source != "all" {
					continue
				}
				isRuleApplicable := checkIfRuleApplicable(campaign, rule.Filters)
				if isRuleApplicable {
					errCode := store.GetStore().BuildAndCreateSmartPropertyFromChannelDocumentAndRule(&smartPropertyRule, rule, campaign, source)
					if errCode != http.StatusCreated {
						logCtx.WithFields(log.Fields{"rule": rule, "channel_document": campaign}).Error("Failed to create smart properties")
						return recordsUpdated, recordsEvaluated, errCode
					} else {
						recordsUpdated += 1
						break
					}
				}
			}
		}
	case 2:
		for _, adGroup := range adGroups {
			recordsEvaluated += 1
			for _, rule := range rules {
				if rule.Source != source && rule.Source != "all" {
					continue
				}
				isRuleApplicable := checkIfRuleApplicable(adGroup, rule.Filters)
				if isRuleApplicable {
					errCode := store.GetStore().BuildAndCreateSmartPropertyFromChannelDocumentAndRule(&smartPropertyRule, rule, adGroup, source)
					if errCode != http.StatusCreated {
						logCtx.WithFields(log.Fields{"rule": rule, "channel_document": adGroup}).Error("Failed to create smart properties")
						return recordsUpdated, recordsEvaluated, errCode
					} else {
						recordsUpdated += 1
						break
					}
				}
			}
		}
	}

	return recordsUpdated, recordsEvaluated, http.StatusCreated
}

func checkIfRuleApplicable(document model.ChannelDocumentsWithFields, filters []model.ChannelFilterV1) bool {
	flag := false
	for _, filter := range filters {
		var value string
		if filter.Object == "campaign" {
			value = document.CampaignName
		} else if filter.Object == "ad_group" {
			value = document.AdGroupName
		} else {
			return false
		}
		filterApplicable := CheckFilter(value, filter.Condition, filter.Value)
		if !filterApplicable && filter.LogicalOp == "AND" {
			return false
		}
		if filterApplicable && filter.LogicalOp == "OR" {
			return true
		}
		if filterApplicable && filter.LogicalOp == "AND" {
			flag = true
		}
	}
	return flag
}

func CheckFilter(value string, condition string, filterValue string) bool {
	switch condition {
	case model.EqualsOpStr:
		return value == filterValue
	case model.NotEqualOpStr:
		return value != filterValue
	case model.ContainsOpStr:
		return strings.Contains(value, filterValue)
	case model.NotContainsOpStr:
		return !strings.Contains(value, filterValue)
	}
	return false
}

func deleteSmartPropertyFromRule(smartPropertyRule model.SmartPropertyRules) (int, int, int) {
	noOfUpdatedRecords, noOfEvaluatedRecords, errCode := store.GetStore().DeleteSmartPropertyByRuleID(smartPropertyRule.ProjectID, smartPropertyRule.ID)
	if errCode != http.StatusAccepted {
		return noOfUpdatedRecords, noOfEvaluatedRecords, errCode
	}
	return noOfUpdatedRecords, noOfEvaluatedRecords, http.StatusAccepted
}

func updateSmartPropertyFromRule(smartPropertyRule model.SmartPropertyRules, adwordsCampaigns []model.ChannelDocumentsWithFields,
	adwordsAdGroups []model.ChannelDocumentsWithFields, facebookCampaigns []model.ChannelDocumentsWithFields,
	facebookAdGroups []model.ChannelDocumentsWithFields, linkedinCampaigns []model.ChannelDocumentsWithFields,
	linkedinAdGroups []model.ChannelDocumentsWithFields, bingadsCampaigns []model.ChannelDocumentsWithFields,
	bingadsAdGroups []model.ChannelDocumentsWithFields, customadsCampaigns [][]model.ChannelDocumentsWithFields,
	customadsAdGroups [][]model.ChannelDocumentsWithFields, customSources []string) (int, int, int) {
	recordsEvaluated := 0
	recordsUpdated := 0
	noOfRecordsUpdated, noOfRecordsEvaluated, errCode := deleteSmartPropertyFromRule(smartPropertyRule)
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode == http.StatusAccepted {
		noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertyFromRuleObject(smartPropertyRule, adwordsCampaigns,
			adwordsAdGroups, facebookCampaigns, facebookAdGroups, linkedinCampaigns, linkedinAdGroups, bingadsCampaigns, bingadsAdGroups, customadsCampaigns, customadsAdGroups, customSources)
		recordsEvaluated += noOfRecordsEvaluated
		recordsUpdated += noOfRecordsUpdated
		if errCode != http.StatusCreated {
			return recordsUpdated, recordsEvaluated, errCode
		}
	} else {
		return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
	}
	return recordsUpdated, recordsEvaluated, http.StatusAccepted
}
