package model

import (
	"sort"
)

// ApplyAttribution This method maps the user to the attribution key based on given attribution methodology.
func ApplyAttribution(attributionType string, method string, conversionEvent string, usersToBeAttributed []UserEventInfo,
	sessions map[string]map[string]UserSessionData, coalUserIdConversionTimestamp map[string]int64,
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
			attributionKeys = getFirstTouchId(attributionType, sessions[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLastTouch:
			attributionKeys = getLastTouchId(attributionType, sessions[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodFirstTouchNonDirect:
			attributionKeys = getFirstTouchNDId(attributionType, sessions[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLastTouchNonDirect:
			attributionKeys = getLastTouchNDId(attributionType, sessions[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLinear:
			attributionKeys = getLinearTouch(attributionType, sessions[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		default:
			break
		}
		// In case a successful attribution could not happen, remove converted user.
		if len(attributionKeys) == 0 {
			delete(usersAttribution, userId)
		}
		if eventName == conversionEvent && val.EventType == EventTypeGoalEvent {
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

type Interaction struct {
	AttributionKey  string
	InteractionTime int64
}

func getMergedInteractions(attributionTimerange map[string]UserSessionData) []Interaction {

	var interactions []Interaction
	for key, value := range attributionTimerange {
		for _, timestamp := range value.TimeStamps {
			interactions = append(interactions, Interaction{AttributionKey: key, InteractionTime: timestamp})
		}
	}
	return interactions
}

func sortInteractionTime(interactions []Interaction, sortingType string) []Interaction {

	// sort the interactions by interaction time ascending order
	if sortingType == SortASC {
		sort.Slice(interactions, func(i, j int) bool {
			return interactions[i].InteractionTime < interactions[j].InteractionTime
		})
	} else if sortingType == SortDESC {
		sort.Slice(interactions, func(i, j int) bool {
			return interactions[i].InteractionTime > interactions[j].InteractionTime
		})
	}
	return interactions
}

// returns list of attribution keys from given attributionKeyTime map
func getLinearTouch(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime, lookbackPeriod, from, to int64) []string {

	var keys []string
	interactions := getMergedInteractions(attributionTimerange)

	switch attributionType {
	case AttributionQueryTypeConversionBased:
		for _, interaction := range interactions {
			if isAdTouchWithinLookback(interaction.InteractionTime, conversionTime, lookbackPeriod) {
				keys = append(keys, interaction.AttributionKey)
			}
		}
	case AttributionQueryTypeEngagementBased:
		for _, interaction := range interactions {
			if isAdTouchWithinLookback(interaction.InteractionTime, conversionTime,
				lookbackPeriod) && isAdTouchWithinCampaignOrQueryPeriod(interaction.InteractionTime, from, to) {
				keys = append(keys, interaction.AttributionKey)
			}
		}
	}

	return keys
}

// returns the first attributionId
func getFirstTouchId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime, lookbackPeriod, from, to int64) []string {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = sortInteractionTime(interactions, SortASC)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) {
					return []string{interactions[i].AttributionKey}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) {
					return []string{interactions[i].AttributionKey}
				}
			}
		}

	}
	return []string{}
}

// returns the last attributionId
func getLastTouchId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime, lookbackPeriod, from, to int64) []string {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = sortInteractionTime(interactions, SortDESC)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime,
					lookbackPeriod) {
					return []string{interactions[i].AttributionKey}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime,
					lookbackPeriod) && isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) {
					return []string{interactions[i].AttributionKey}
				}
			}
		}

	}
	return []string{}
}

// returns the first non $none attributionId
func getFirstTouchNDId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime, lookbackPeriod, from, to int64) []string {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = sortInteractionTime(interactions, SortASC)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime,
					lookbackPeriod) && interactions[i].AttributionKey != PropertyValueNone {
					return []string{interactions[i].AttributionKey}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) && interactions[i].AttributionKey != PropertyValueNone {
					return []string{interactions[i].AttributionKey}
				}
			}
		}

	}
	return []string{}
}

// returns the last non $none attributionId
func getLastTouchNDId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime, lookbackPeriod, from, to int64) []string {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = sortInteractionTime(interactions, SortDESC)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime,
					lookbackPeriod) && interactions[i].AttributionKey != PropertyValueNone {
					return []string{interactions[i].AttributionKey}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) && interactions[i].AttributionKey != PropertyValueNone {
					return []string{interactions[i].AttributionKey}
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
