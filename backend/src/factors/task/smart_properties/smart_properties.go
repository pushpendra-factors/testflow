package smart_properties

import (
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"reflect"
	"strings"

	U "factors/util"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm/dialects/postgres"
)

func EnrichSmartPropertiesForChangedRulesForProject(projectID uint64) int {
	log.Warn("Smart properties enrichment for changed rule started for project ", projectID)
	var recordsEvaluated = 0
	var recordsUpdated = 0

	smartPropertiesRules, errCode := store.GetStore().GetAllChangedSmartPropertiesRulesForProject(projectID)
	if errCode != http.StatusFound {
		log.Error("failed to get smart properties rules for project ", projectID)
		return http.StatusInternalServerError
	}
	if len(smartPropertiesRules) == 0 {
		log.Warn("No updated rules found for project ", projectID)
		return http.StatusOK
	}
	log.Warn("No of changed smart properties rules: ", len(smartPropertiesRules))
	adwordsCampaigns, adwordsAdGroups := store.GetStore().GetLatestMetaForAdwordsForGivenDays(projectID, 30)
	facebookCampaigns, facebookAdGroups := store.GetStore().GetLatestMetaForFacebookForGivenDays(projectID, 30)
	linkedinCampaigns, linkedinAdGroups := store.GetStore().GetLatestMetaForLinkedinForGivenDays(projectID, 30)

	for _, smartPropertiesRule := range smartPropertiesRules {
		switch checkState(smartPropertiesRule) {
		case model.CREATED:
			noOfRecordsUpdated, noOfRecordsEvaluated, errCode := createSmartPropertiesFromRuleObject(smartPropertiesRule, adwordsCampaigns, adwordsAdGroups, facebookCampaigns, facebookAdGroups, linkedinCampaigns, linkedinAdGroups)
			recordsEvaluated += noOfRecordsEvaluated
			recordsUpdated += noOfRecordsUpdated
			if errCode != http.StatusCreated {
				continue
			}
		case model.DELETED:
			noOfRecordsUpdated, noOfRecordsEvaluated, errCode := deleteSmartPropertiesFromRule(smartPropertiesRule)
			recordsEvaluated += noOfRecordsEvaluated
			recordsUpdated += noOfRecordsUpdated
			if errCode != http.StatusAccepted {
				continue
			}
		case model.UPDATED:
			noOfRecordsUpdated, noOfRecordsEvaluated, errCode := updateSmartPropertiesFromRule(smartPropertiesRule, adwordsCampaigns, adwordsAdGroups, facebookCampaigns, facebookAdGroups, linkedinCampaigns, linkedinAdGroups)
			recordsEvaluated += noOfRecordsEvaluated
			recordsUpdated += noOfRecordsUpdated
			if errCode != http.StatusAccepted {
				continue
			}
		default:
			log.Error("Invalid status")
		}

		smartPropertiesRule.Picked = true
		errMsg, errCode := store.GetStore().UpdateSmartPropertiesRules(smartPropertiesRule.ProjectID, smartPropertiesRule.ID, smartPropertiesRule)
		if errCode != http.StatusAccepted {
			log.WithFields(log.Fields{"errMsg": errMsg, "smart_properties_rule": smartPropertiesRule}).Error("failed to update smart properties rule.")
			continue
		}
	}
	log.Warn("No. of records evaluated: ", recordsEvaluated)
	log.Warn("No. of records updated: ", recordsUpdated)
	log.Warn("Smart properties enrichment for changed rules ended for project ", projectID)
	return http.StatusOK
}

func EnrichSmartPropertiesForCurrentDayForProject(projectID uint64) int {
	log.Warn("Smart properties enrichment for current day's data started for project ", projectID)
	smartPropertiesRules, errCode := store.GetStore().GetSmartPropertiesRules(projectID)
	if errCode != http.StatusFound {
		log.Error("No rules found for project ", projectID)
		return http.StatusOK
	}
	recordsUpdated := 0
	recordsEvaluated := 0
	adwordsCampaigns, adwordsAdGroups := store.GetStore().GetLatestMetaForAdwordsForGivenDays(projectID, 1)
	facebookCampaigns, facebookAdGroups := store.GetStore().GetLatestMetaForFacebookForGivenDays(projectID, 1)
	linkedinCampaigns, linkedinAdGroups := store.GetStore().GetLatestMetaForLinkedinForGivenDays(projectID, 1)

	changedAdwordsCampaigns, changedAdwordsAdGroups := getChangedChannelDocuments(projectID, adwordsCampaigns, adwordsAdGroups, "adwords")
	changedFacebookCampaigns, changedFacebookAdGroups := getChangedChannelDocuments(projectID, facebookCampaigns, facebookAdGroups, "facebook")
	changedLinkedinCampaigns, changedLinkedinAdGroups := getChangedChannelDocuments(projectID, linkedinCampaigns, linkedinAdGroups, "linkedin")

	for _, smartPropertiesRule := range smartPropertiesRules {
		noOfRecordsUpdated, noOfRecordsEvaluated, errCode := createSmartPropertiesFromRuleObject(smartPropertiesRule, changedAdwordsCampaigns, changedAdwordsAdGroups, changedFacebookCampaigns, changedFacebookAdGroups, changedLinkedinCampaigns, changedLinkedinAdGroups)
		recordsUpdated += noOfRecordsUpdated
		recordsEvaluated += noOfRecordsEvaluated
		if errCode != http.StatusCreated {
			log.WithField("smart_properties_rule", smartPropertiesRule).Error("Failed to create smart properties for rule")
			continue
		}
	}

	log.Warn("No. of records evaluated: ", recordsEvaluated)
	log.Warn("No. of records updated: ", recordsUpdated)
	log.Warn("Smart properties enrichment for current day's data ended for project ", projectID)
	return http.StatusOK
}
func checkState(smartPropertiesRule model.SmartPropertiesRules) string {
	if smartPropertiesRule.CreatedAt == smartPropertiesRule.UpdatedAt {
		return model.CREATED
	}
	if (smartPropertiesRule.CreatedAt != smartPropertiesRule.UpdatedAt) && smartPropertiesRule.IsDeleted {
		return model.DELETED
	} else {
		return model.UPDATED
	}
}

//to do: If we intend to parallelise this at per project level, It might be better to evaluate in chunks. Can check later.
func getChangedChannelDocuments(projectID uint64, campaigns []model.ChannelDocumentsWithFields, adGroups []model.ChannelDocumentsWithFields,
	source string) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	smartPropertiesCampaigns, errCode := store.GetStore().GetSmartPropertiesByProjectIDAndSourceAndObjectType(projectID, source, 1)
	if errCode != http.StatusFound {
		return make([]model.ChannelDocumentsWithFields, 0, 0), make([]model.ChannelDocumentsWithFields, 0, 0)
	}
	channelDocumentsCampaignsMap := make(map[string]model.ChannelDocumentsWithFields)
	for _, smartProperty := range smartPropertiesCampaigns {
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

	smartPropertiesAdGroups, errCode := store.GetStore().GetSmartPropertiesByProjectIDAndSourceAndObjectType(projectID, source, 2)
	if errCode != http.StatusFound {
		return make([]model.ChannelDocumentsWithFields, 0, 0), make([]model.ChannelDocumentsWithFields, 0, 0)
	}
	channelDocumentsAdGroupsMap := make(map[string]model.ChannelDocumentsWithFields)
	for _, smartProperty := range smartPropertiesAdGroups {
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

func createSmartPropertiesFromRuleObject(smartPropertiesRule model.SmartPropertiesRules, adwordsCampaigns []model.ChannelDocumentsWithFields,
	adwordsAdGroups []model.ChannelDocumentsWithFields, facebookCampaigns []model.ChannelDocumentsWithFields,
	facebookAdGroups []model.ChannelDocumentsWithFields, linkedinCampaigns []model.ChannelDocumentsWithFields,
	linkedinAdGroups []model.ChannelDocumentsWithFields) (int, int, int) {

	recordsUpdated := 0
	recordsEvaluated := 0
	var rules []model.Rule
	err := U.DecodePostgresJsonbToStructType(smartPropertiesRule.Rules, &rules)
	if err != nil {
		log.Error("failed to decode rules")
		return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
	}
	noOfRecordsUpdated, noOfRecordsEvaluated, errCode := createSmartPropertiesFromRuleObjectForSource(smartPropertiesRule, adwordsCampaigns, adwordsAdGroups, rules, "adwords")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
	}

	noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertiesFromRuleObjectForSource(smartPropertiesRule, facebookCampaigns, facebookAdGroups, rules, "facebook")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
	}

	noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertiesFromRuleObjectForSource(smartPropertiesRule, linkedinCampaigns, linkedinAdGroups, rules, "linkedin")
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode != http.StatusCreated {
		return recordsUpdated, recordsEvaluated, errCode
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
func createSmartPropertiesFromRuleObjectForSource(smartPropertiesRule model.SmartPropertiesRules,
	campaigns []model.ChannelDocumentsWithFields, adGroups []model.ChannelDocumentsWithFields, rules []model.Rule, source string) (int, int, int) {
	logCtx := log.WithFields(log.Fields{"smart_properties_rule": smartPropertiesRule})
	recordsEvaluated := 0
	recordsUpdated := 0
	switch smartPropertiesRule.Type {
	case 1:
		for _, campaign := range campaigns {
			recordsEvaluated += 1
			for _, rule := range rules {
				if rule.Source != source && rule.Source != "all" {
					continue
				}
				isRuleApplicable := checkIfRuleApplicable(campaign, rule.Filters)
				if isRuleApplicable {
					errCode := store.GetStore().CreateSmartPropertiesFromChannelDocumentAndRule(&smartPropertiesRule, rule, campaign, source)
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
					errCode := store.GetStore().CreateSmartPropertiesFromChannelDocumentAndRule(&smartPropertiesRule, rule, adGroup, source)
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

func deleteSmartPropertiesFromRule(smartPropertiesRule model.SmartPropertiesRules) (int, int, int) {
	noOfUpdatedRecords, noOfEvaluatedRecords, errCode := store.GetStore().DeleteSmartPropertiesByRuleID(smartPropertiesRule.ProjectID, smartPropertiesRule.ID)
	if errCode != http.StatusAccepted {
		return noOfUpdatedRecords, noOfEvaluatedRecords, errCode
	}
	return noOfUpdatedRecords, noOfEvaluatedRecords, http.StatusAccepted
}

func updateSmartPropertiesFromRule(smartPropertiesRule model.SmartPropertiesRules, adwordsCampaigns []model.ChannelDocumentsWithFields,
	adwordsAdGroups []model.ChannelDocumentsWithFields, facebookCampaigns []model.ChannelDocumentsWithFields,
	facebookAdGroups []model.ChannelDocumentsWithFields, linkedinCampaigns []model.ChannelDocumentsWithFields,
	linkedinAdGroups []model.ChannelDocumentsWithFields) (int, int, int) {
	recordsEvaluated := 0
	recordsUpdated := 0
	noOfRecordsUpdated, noOfRecordsEvaluated, errCode := deleteSmartPropertiesFromRule(smartPropertiesRule)
	recordsEvaluated += noOfRecordsEvaluated
	recordsUpdated += noOfRecordsUpdated
	if errCode == http.StatusAccepted {
		noOfRecordsUpdated, noOfRecordsEvaluated, errCode = createSmartPropertiesFromRuleObject(smartPropertiesRule, adwordsCampaigns,
			adwordsAdGroups, facebookCampaigns, facebookAdGroups, linkedinCampaigns, linkedinAdGroups)
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
