package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetLinkedinCappingExclusionsForRule(projectID int64, ruleID string) ([]model.LinkedinExclusion, int) {
	linkedinExclusions := make([]model.LinkedinExclusion, 0)
	linkedinExclusions = append(linkedinExclusions, model.SampleExclusion)
	return linkedinExclusions, http.StatusOK
}
func (store *MemSQL) GetAllLinkedinCappingExclusionsForTimerange(projectID int64, startTimestamp int64, endTimestamp int64) ([]model.LinkedinExclusion, int) {
	linkedinExclusions := make([]model.LinkedinExclusion, 0)
	linkedinExclusions = append(linkedinExclusions, model.SampleExclusion)
	return linkedinExclusions, http.StatusOK
}
