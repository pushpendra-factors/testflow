package model

import (
	"sort"
)

type pair struct {
	key   string
	value int64
}

// This method maps the user to the attribution key based on given attribution methodology.
func ApplyAttribution(attributionType string, method string, conversionEvent string, usersToBeAttributed []UserEventInfo,
	userInitialSession map[string]map[string]UserSessionTimestamp,
	coalUserIdConversionTimestamp map[string]int64,
	lookbackDays int, campaignFrom, campaignTo int64) (map[string][]string, map[string]map[string][]string, error) {

	usersAttribution := make(map[string][]string)
	linkedEventUserCampaign := make(map[string]map[string][]string)
	lookbackPeriod := int64(lookbackDays) * SecsInADay
	for _, val := range usersToBeAttributed {
		userId := val.CoalUserID
		eventName := val.EventName
		conversionTime := coalUserIdConversionTimestamp[val.CoalUserID]
		attributionKeys := []string{PropertyValueNone}
		switch method {
		case AttributionMethodFirstTouch:
			attributionKeys = getFirstTouchId(attributionType, userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLastTouch:
			attributionKeys = getLastTouchId(attributionType, userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodFirstTouchNonDirect:
			attributionKeys = getFirstTouchNDId(attributionType, userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLastTouchNonDirect:
			attributionKeys = getLastTouchNDId(attributionType, userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLinear:
			attributionKeys = getLinearTouch(attributionType, userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		default:
			break
		}
		// In case a successful attribution could not happen, remove converted user.
		if len(attributionKeys) == 0 {
			delete(usersAttribution, userId)
		}
		if eventName == conversionEvent {
			usersAttribution[userId] = attributionKeys
		} else {
			if _, exist := linkedEventUserCampaign[eventName]; !exist {
				linkedEventUserCampaign[eventName] = make(map[string][]string)
			}
			linkedEventUserCampaign[eventName][userId] = attributionKeys
		}
	}
	return usersAttribution, linkedEventUserCampaign, nil
}

// returns list of attribution keys from given attributionKeyTime map
func getLinearTouch(attributionType string, attributionTimerange map[string]UserSessionTimestamp, conversionTime, lookbackPeriod, from, to int64) []string {

	var keys []string

	switch attributionType {
	case AttributionQueryTypeConversionBased:
		for aId, rangeTime := range attributionTimerange {
			if isAdTouchWithinLookback(rangeTime.MinTimestamp, conversionTime, lookbackPeriod) {
				keys = append(keys, aId)
			}
		}
	case AttributionQueryTypeEngagementBased:
		for aId, rangeTime := range attributionTimerange {
			if isAdTouchWithinLookback(rangeTime.MinTimestamp, conversionTime,
				lookbackPeriod) && isAdTouchWithinCampaignOrQueryPeriod(rangeTime.MinTimestamp, from, to) {
				keys = append(keys, aId)
			}
		}
	}

	return keys
}

// returns the first attributionId
func getFirstTouchId(attributionType string, attributionTimerange map[string]UserSessionTimestamp, conversionTime, lookbackPeriod, from, to int64) []string {
	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MinTimestamp for FirstTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MinTimestamp})
	}
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value < attributionIds[j].value
		})

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(attributionIds); i++ {
				if isAdTouchWithinLookback(attributionIds[i].value, conversionTime, lookbackPeriod) {
					return []string{attributionIds[i].key}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(attributionIds); i++ {
				if isAdTouchWithinLookback(attributionIds[i].value, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(attributionIds[i].value, from, to) {
					return []string{attributionIds[i].key}
				}
			}
		}

	}
	return []string{}
}

// returns the last attributionId
func getLastTouchId(attributionType string, attributionTimerange map[string]UserSessionTimestamp, conversionTime, lookbackPeriod, from, to int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MaxTimestamp for LastTouch.
		attributionIds = append(attributionIds, pair{aId, rangeT.MaxTimestamp})
	}
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value > attributionIds[j].value
		})

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(attributionIds); i++ {
				if isAdTouchWithinLookback(attributionIds[i].value, conversionTime,
					lookbackPeriod) {
					return []string{attributionIds[i].key}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(attributionIds); i++ {
				if isAdTouchWithinLookback(attributionIds[i].value, conversionTime,
					lookbackPeriod) && isAdTouchWithinCampaignOrQueryPeriod(attributionIds[i].value, from, to) {
					return []string{attributionIds[i].key}
				}
			}
		}

	}
	return []string{}
}

// returns the first non $none attributionId
func getFirstTouchNDId(attributionType string, attributionTimerange map[string]UserSessionTimestamp, conversionTime, lookbackPeriod, from, to int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MinTimestamp for FirstTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MinTimestamp})
	}
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value < attributionIds[j].value
		})

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for _, atPair := range attributionIds {
				if isAdTouchWithinLookback(atPair.value, conversionTime,
					lookbackPeriod) && atPair.key != PropertyValueNone {
					return []string{atPair.key}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for _, atPair := range attributionIds {
				if isAdTouchWithinLookback(atPair.value, conversionTime,
					lookbackPeriod) && atPair.key != PropertyValueNone &&
					isAdTouchWithinCampaignOrQueryPeriod(atPair.value, from, to) {
					return []string{atPair.key}
				}
			}
		}

	}
	return []string{}
}

// returns the last non $none attributionId
func getLastTouchNDId(attributionType string, attributionTimerange map[string]UserSessionTimestamp, conversionTime, lookbackPeriod, from, to int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MaxTimestamp for LastTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MaxTimestamp})
	}
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value > attributionIds[j].value
		})

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for _, atPair := range attributionIds {
				if isAdTouchWithinLookback(atPair.value, conversionTime,
					lookbackPeriod) && atPair.key != PropertyValueNone {
					return []string{atPair.key}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for _, atPair := range attributionIds {
				if isAdTouchWithinLookback(atPair.value, conversionTime,
					lookbackPeriod) && atPair.key != PropertyValueNone &&
					isAdTouchWithinCampaignOrQueryPeriod(atPair.value, from, to) {
					return []string{atPair.key}
				}
			}
		}
	}
	return []string{}
}

// isAdTouchWithinCampaignOrQueryPeriod checks if attribution time is within query (campaign) period.
func isAdTouchWithinCampaignOrQueryPeriod(attributionTime int64, from, to int64) bool {
	return attributionTime >= from && attributionTime <= to
}

// isAdTouchWithinLookback checks if attribution time is within lookback period.
// i.e. the conversion should be after
func isAdTouchWithinLookback(campaignTouchPoint int64,
	conversionTime int64, lookbackPeriod int64) bool {

	if conversionTime >= campaignTouchPoint {
		if (conversionTime - campaignTouchPoint) <= lookbackPeriod {
			return true
		}
	}
	return false
}
