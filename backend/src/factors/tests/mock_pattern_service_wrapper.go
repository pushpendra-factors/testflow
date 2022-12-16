package tests

import (
	"factors/pattern"
	P "factors/pattern"
	PS "factors/pattern_server"
)

type MockPatternServiceWrapper struct {
	pMap              map[string]*P.Pattern
	patterns          []*P.Pattern
	userAndEventsInfo *P.UserAndEventsInfo
}

func (pw *MockPatternServiceWrapper) GetUserAndEventsInfo() *P.UserAndEventsInfo {
	return pw.userAndEventsInfo
}

func (pw *MockPatternServiceWrapper) GetPerUserCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	if p, ok := pw.pMap[P.EventArrayToString(eventNames)]; ok {
		count, err := p.GetPerUserCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *MockPatternServiceWrapper) GetPerOccurrenceCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints) (uint, bool) {
	if p, ok := pw.pMap[P.EventArrayToString(eventNames)]; ok {
		count, err := p.GetPerOccurrenceCount(patternConstraints)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *MockPatternServiceWrapper) GetCount(reqId string, eventNames []string,
	patternConstraints []P.EventConstraints, countType string) (uint, bool) {
	if p, ok := pw.pMap[P.EventArrayToString(eventNames)]; ok {
		count, err := p.GetCount(patternConstraints, countType)
		if err == nil {
			return count, true
		}
	}
	return 0, false
}

func (pw *MockPatternServiceWrapper) GetPattern(reqId string, eventNames []string) *P.Pattern {
	p, _ := pw.pMap[P.EventArrayToString(eventNames)]
	return p
}

func (pw *MockPatternServiceWrapper) GetAllPatterns(reqId,
	startEvent, endEvent string) ([]*P.Pattern, error) {
	filterPatterns := PS.GetFilter(startEvent, endEvent, "")
	patternsToReturn := make([]*pattern.Pattern, 0)
	for _, p := range pw.patterns {
		if filterPatterns(p.EventNames, startEvent, endEvent, "") {
			patternsToReturn = append(patternsToReturn, p)
		}
	}
	return patternsToReturn, nil
}

func (pw *MockPatternServiceWrapper) GetAllContainingPatterns(reqId,
	event string) ([]*P.Pattern, error) {
	filterPatterns := PS.GetFilter("", "", event)
	patternsToReturn := make([]*pattern.Pattern, 0, 0)
	for _, p := range pw.patterns {
		if filterPatterns(p.EventNames, "", "", event) {
			patternsToReturn = append(patternsToReturn, p)
		}
	}
	return patternsToReturn, nil
}

func (pw *MockPatternServiceWrapper) GetTotalEventCount(reqId string) uint {
	count := uint(0)
	for _, p := range pw.patterns {
		if len(p.EventNames) == 1 {
			count += p.PerOccurrenceCount
		}
	}
	return count
}

func NewMockPatternServiceWrapper(
	patterns []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo) *MockPatternServiceWrapper {
	patternWrapper := MockPatternServiceWrapper{
		patterns: patterns,
	}
	pMap := make(map[string]*P.Pattern)
	for _, p := range patterns {
		pMap[p.String()] = p
	}
	patternWrapper.pMap = pMap
	patternWrapper.userAndEventsInfo = userAndEventsInfo
	return &patternWrapper
}
