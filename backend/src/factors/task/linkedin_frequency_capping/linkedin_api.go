package linkedin_frequency_capping

import (
	"bytes"
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func GetExclusionFromDBAndPushToLinkedin(projectID int64, startOfMonthDateInt int64, accessToken string) (string, int) {
	linkedinExclusions, errCode := store.GetStore().GetNonPushedExclusionsForMonth(projectID, startOfMonthDateInt)
	if errCode != http.StatusOK {
		return "Failed to get exclusions for linkedin sync", errCode
	}
	campaignsMap := *(GetCampaignTargetingCriteriaMapInDataStore())
	updatedTargetingCriteriaMap := make(map[string]model.CampaignNameTargetingCriteria)

	for _, exclusion := range linkedinExclusions {
		var campaigns []model.CampaignsIDName
		_ = U.DecodePostgresJsonbToStructType(exclusion.Campaigns, campaigns)
		for _, campaign := range campaigns {
			if _, exists := updatedTargetingCriteriaMap[campaign.CampaignID]; !exists {
				updatedTargetingCriteriaMap[campaign.CampaignID] = campaignsMap[campaign.CampaignID]
			}
			updatedTargetingCriteriaMap[campaign.CampaignID].TargetingCriteria[exclusion.OrgID] = true
		}
	}

	for campaignID, campaignTargetingCriteria := range updatedTargetingCriteriaMap {
		errMsg, errCode := pushTargetingCriteriaToLinkedin(accessToken, campaignID, campaignTargetingCriteria)
		if errMsg != "" {
			return errMsg, errCode
		}
	}
	for _, exclusion := range linkedinExclusions {
		errCode = store.GetStore().UpdateLinkedinPushSyncStatusForOrgAndRule(projectID, startOfMonthDateInt, exclusion.OrgID, exclusion.RuleID)
		if errCode != http.StatusOK {
			return fmt.Sprintf("failed to update linkedin sync status for inclusion in db org_id %s and rule_id %s", exclusion.OrgID, exclusion.RuleID), http.StatusInternalServerError
		}
	}
	return "", http.StatusOK
}

func GetExclusionFromDBAndRemoveFromLinkedin(projectID int64, startOfMonthDateInt int64, endOfMonthDate int64, accessToken string) (string, int) {
	linkedinExclusions, errCode := store.GetStore().GetNonRemovedExclusionsForMonth(projectID, startOfMonthDateInt, endOfMonthDate)
	if errCode != http.StatusOK {
		return "Failed to get exclusions for linkedin sync", errCode
	}
	campaignsMap := *(GetCampaignTargetingCriteriaMapInDataStore())
	updatedTargetingCriteriaMap := make(map[string]model.CampaignNameTargetingCriteria)

	for _, exclusion := range linkedinExclusions {
		var campaigns []model.CampaignsIDName
		_ = U.DecodePostgresJsonbToStructType(exclusion.Campaigns, campaigns)
		for _, campaign := range campaigns {
			if _, exists := updatedTargetingCriteriaMap[campaign.CampaignID]; !exists {
				updatedTargetingCriteriaMap[campaign.CampaignID] = campaignsMap[campaign.CampaignID]
			}
			delete(updatedTargetingCriteriaMap[campaign.CampaignID].TargetingCriteria, exclusion.OrgID)
		}
	}

	for campaignID, campaignTargetingCriteria := range updatedTargetingCriteriaMap {
		errMsg, errCode := pushTargetingCriteriaToLinkedin(accessToken, campaignID, campaignTargetingCriteria)
		if errMsg != "" {
			return errMsg, errCode
		}
	}
	for _, exclusion := range linkedinExclusions {
		errCode = store.GetStore().UpdateLinkedinRemoveSyncStatusForOrgAndRule(projectID, startOfMonthDateInt, endOfMonthDate, exclusion.OrgID, exclusion.RuleID)
		if errCode != http.StatusOK {
			return fmt.Sprintf("failed to update linkedin sync status for exclusion in db org_id %s and rule_id %s", exclusion.OrgID, exclusion.RuleID), http.StatusInternalServerError
		}
	}
	return "", http.StatusOK
}

func pushTargetingCriteriaToLinkedin(accessToken string, campaignID string, campaignTargetingCriteria model.CampaignNameTargetingCriteria) (string, int) {
	targetingCriteriaList := getKeyFromMap(campaignTargetingCriteria.TargetingCriteria)
	payload := map[string]map[string]map[string][]string{
		"exclude": {
			"or": {
				"urn:li:adTargetingFacet:employers": targetingCriteriaList,
			},
		},
	}
	patchPayload := map[string]map[string]map[string]map[string]map[string]map[string][]string{
		"patch": {
			"$set": {
				"targetingCriteria": payload,
			},
		},
	}

	url := fmt.Sprintf("https://api.linkedin.com/rest/adAccounts/%s/adCampaigns/%s", campaignTargetingCriteria.AdAccountID, campaignID)

	patchPayloadByte, err := json.Marshal(patchPayload)
	if err != nil {
		return "failed to build payload for campaign_id " + campaignID, http.StatusInternalServerError
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(patchPayloadByte))
	if err != nil {
		return "failed to create POST request in linkedin targeting criteria update.", http.StatusInternalServerError
	}
	req.Header = http.Header{
		"LinkedIn-Version":          {"202305"},
		"Authorization":             {fmt.Sprintf("Bearer %s", accessToken)},
		"X-Restli-Protocol-Version": {"2.0.0"},
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to fetch linkedin campaign")
		return "failed in request build to update linkedin campaign_id " + campaignID, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "failed in api call to update linkedin campaign_id " + campaignID, resp.StatusCode
	}
	return "", http.StatusOK
}

type LinkedinCampaignResponse struct {
	Paging    map[string]interface{}   `json:"paging"`
	Campaigns []map[string]interface{} `json:"elements"`
}

// This is also very generic.
func getAllActivePausedCampaignsFromLinkedin(projectID int64, adAccounts []string, accessToken string) ([]map[string]interface{}, int) {
	campaigns := make([]map[string]interface{}, 0)
	for _, adAccount := range adAccounts {
		isEndReached := false
		start, count := 0, 1000
		for !isEndReached {
			url := fmt.Sprintf("https://api.linkedin.com/rest/adAccounts/%s/adCampaigns?q=search&search=(status:(values:List(ACTIVE,PAUSED)))&start=%d&count=%d", adAccount, start, count)
			client := http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.WithError(err).Error("failed to build request for linkedin campaign fetch")
				return campaigns, http.StatusInternalServerError
			}
			req.Header = http.Header{
				"LinkedIn-Version":          {"202305"},
				"Authorization":             {fmt.Sprintf("Bearer %s", accessToken)},
				"X-Restli-Protocol-Version": {"2.0.0"},
			}

			resp, err := client.Do(req)
			if err != nil {
				log.WithError(err).Error("failed to fetch linkedin campaign")
				return campaigns, http.StatusInternalServerError
			}
			// log.Fatal(resp.Body)
			if resp.StatusCode != http.StatusOK {
				log.Error("failed to get camapigns from linkedin api")
				return make([]map[string]interface{}, 0), resp.StatusCode
			}
			defer resp.Body.Close()
			var campaignResponse LinkedinCampaignResponse
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&campaignResponse); err != nil {
				return make([]map[string]interface{}, 0), http.StatusInternalServerError
			}

			if len(campaignResponse.Campaigns) == 0 {
				isEndReached = true
				break
			}
			campaigns = append(campaigns, campaignResponse.Campaigns...)
			start += count
		}
	}
	return campaigns, http.StatusOK
}
