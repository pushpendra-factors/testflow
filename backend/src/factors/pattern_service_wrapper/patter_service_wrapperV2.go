package pattern_service_wrapper

import (
	"factors/model/model"
	P "factors/pattern"
	PC "factors/pattern_client"
	U "factors/util"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type PatternServiceWrapperV2 struct {
	projectId         int64
	modelId           uint64
	pMap              map[string]*P.Pattern
	totalEventCount   uint
	userAndEventsInfo *P.UserAndEventsInfo
}

type ExplainV2Goals struct {
	GoalRule          model.FactorsGoalRule `json:"goal"`
	Insights          []*FactorsInsights    `json:"insights"`
	GoalUserCount     float64               `json:"goal_user_count"`
	TotalUsersCount   float64               `json:"total_users_count"`
	OverallPercentage float64               `json:"overall_percentage"`
	OverallMultiplier float64               `json:"overall_multiplier"`
	Type              string                `json:"type"`
	StartTimestamp    int64                 `json:"sts"`
	EndTimestamp      int64                 `json:"ets"`
}

func (pw *PatternServiceWrapperV2) GetUserAndEventsInfo() *P.UserAndEventsInfo {
	userAndEventsInfo, _ := PC.GetUserAndEventsInfoV2("", pw.projectId, pw.modelId)
	pw.userAndEventsInfo = userAndEventsInfo
	return userAndEventsInfo
}

func (pw *PatternServiceWrapperV2) GetPerUserCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	pattern := pw.GetPattern(reqId, eventNames)
	if pattern != nil {
		count, err := pattern.GetPerUserCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *PatternServiceWrapperV2) GetPerOccurrenceCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	pattern := pw.GetPattern(reqId, eventNames)
	if pattern != nil {
		count, err := pattern.GetPerOccurrenceCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}
func (pw *PatternServiceWrapperV2) GetCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints, countType string) (uint, bool) {
	if countType == P.COUNT_TYPE_PER_USER {
		return pw.GetPerUserCount(reqId, eventNames, patternConstraints)
	} else if countType == P.COUNT_TYPE_PER_OCCURRENCE {
		if len(eventNames) == 1 && eventNames[0] == U.SEN_ALL_EVENTS {
			return pw.GetTotalEventCount(reqId), true
		}
		return pw.GetPerOccurrenceCount(reqId, eventNames, patternConstraints)
	}
	log.Errorf(fmt.Sprintf("Unrecognized count type: %s", countType))
	return 0, false
}

func (pw *PatternServiceWrapperV2) GetPattern(reqId string, eventNames []string) *P.Pattern {
	var pattern *P.Pattern = nil
	var found bool
	eventsHash := P.EventArrayToString(eventNames)
	if pattern, found = pw.pMap[eventsHash]; !found {
		// Fetch from server.
		patterns, err := PC.GetPatternsV2(reqId, pw.projectId, pw.modelId, [][]string{eventNames})
		if err == nil && len(patterns) == 1 && P.EventArrayToString(patterns[0].EventNames) == eventsHash {
			pattern = patterns[0]
			// Add it to cache.
			pw.pMap[eventsHash] = pattern
		} else {
			log.WithFields(log.Fields{
				"error": err, "modelId": pw.modelId,
				"projectId": pw.projectId, "eventNames": eventNames,
				"returned Patterns": patterns}).Error(
				"Get Patterns failed.")
		}
	}
	return pattern
}

func (pw *PatternServiceWrapperV2) GetAllPatterns(reqId,
	startEvent, endEvent string) ([]*P.Pattern, error) {
	// Fetch from server.
	log.Infof("inside get all patterns v2 project id :%d : model_id:%d", pw.projectId, pw.modelId)

	patterns, err := PC.GetAllPatternsV2(reqId, pw.projectId, pw.modelId, startEvent, endEvent)
	// Add it to cache.
	for _, p := range patterns {
		pw.pMap[P.EventArrayToString(p.EventNames)] = p
	}
	return patterns, err
}

func (pw *PatternServiceWrapperV2) GetAllContainingPatterns(reqId, event string) ([]*P.Pattern, error) {
	// Fetch from server.
	patterns, err := PC.GetAllContainingPatternsV2(reqId, pw.projectId, pw.modelId, event)
	// Add it to cache.
	for _, p := range patterns {
		pw.pMap[P.EventArrayToString(p.EventNames)] = p
	}
	return patterns, err
}

func (pw *PatternServiceWrapperV2) GetTotalEventCount(reqId string) uint {
	if pw.totalEventCount > 0 {
		// Fetch from cache.
		return pw.totalEventCount
	}
	if tc, err := PC.GetTotalEventCountV2(reqId, pw.projectId, pw.modelId); err != nil {
		return 0
	} else {
		pw.totalEventCount = uint(tc)
		return pw.totalEventCount
	}
}
