package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetLinkedinCappingConfig(projectID int64) ([]model.LinkedinCappingConfig, int) {
	linkedinCappingConfig := make([]model.LinkedinCappingConfig, 0)
	linkedinCappingConfig = append(linkedinCappingConfig, model.SampleCampaignGroupConfig...)
	linkedinCappingConfig = append(linkedinCappingConfig, model.SampleCampaignConfig...)
	return linkedinCappingConfig, http.StatusOK
}
func (store *MemSQL) CreateLinkedinCappingRule(projectID int64, linkedinCappingRule *model.LinkedinCappingRule) int {
	return http.StatusCreated
}

func (store *MemSQL) GetAllLinkedinCappingRules(projectID int64) ([]model.LinkedinCappingRule, int) {
	linkedinCappingRules := make([]model.LinkedinCappingRule, 0)
	linkedinCappingRules = append(linkedinCappingRules, model.SampleCappingRule)
	return linkedinCappingRules, http.StatusFound
}

func (store *MemSQL) GetLinkedinCappingRule(projectID int64, ruleID string) (model.LinkedinCappingRule, int) {
	return model.SampleCappingRule, http.StatusFound
}
func (store *MemSQL) UpdateLinkedinCappingRule(projectID int64, ruleID string) int {
	return http.StatusAccepted
}

func (store *MemSQL) DeleteLinkedinCappingRule(projectID int64, ruleID string) int {
	return http.StatusAccepted
}
