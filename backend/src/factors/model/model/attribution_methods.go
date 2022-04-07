package model

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"sort"
)

// ApplyAttribution This method maps the user to the attribution key based on given attribution methodology.
type AttributionKeyWeight struct {
	Key    string
	Weight float64
}

func ApplyAttributionKPI(attributionType string,
	method string,
	sessions map[string]map[string]UserSessionData,
	kpiData map[string]KPIInfo,
	lookbackDays int, campaignFrom, campaignTo int64,
	attributionKey string) (map[string][]AttributionKeyWeight, error) {

	usersAttribution := make(map[string][]AttributionKeyWeight)
	lookbackPeriod := int64(lookbackDays) * SecsInADay

	for kpiID, kpiInfo := range kpiData {
		// kpiID := kpiInfo.KpiID
		conversionTime := kpiInfo.Timestamp
		userSessions := sessions[kpiID]

		var attributionKeys []AttributionKeyWeight
		switch method {
		case AttributionMethodFirstTouch:
			attributionKeys = getFirstTouchId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLastTouch:
			attributionKeys = getLastTouchId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break
		case AttributionMethodUShaped:
			attributionKeys = getUShaped(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodFirstTouchNonDirect:
			attributionKeys = getFirstTouchNDId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo, attributionKey)
			break

		case AttributionMethodLastTouchNonDirect:
			attributionKeys = getLastTouchNDId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo, attributionKey)
			break

		case AttributionMethodLinear:
			attributionKeys = getLinearTouch(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break
		case AttributionMethodTimeDecay:
			attributionKeys = getTimeDecay(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		default:
			break
		}
		log.WithFields(log.Fields{"KPI_ID": kpiID, "attributionKeys": attributionKeys}).Info(fmt.Sprintf("KPI-Attribution attributionKeys"))
		// In case a successful attribution could not happen, remove converted user.
		if len(attributionKeys) == 0 {
			delete(usersAttribution, kpiID)
			continue
		}
		usersAttribution[kpiID] = attributionKeys
	}
	return usersAttribution, nil
}

func ApplyAttribution(attributionType string, method string, conversionEvent string, usersToBeAttributed []UserEventInfo,
	sessions map[string]map[string]UserSessionData, coalUserIdConversionTimestamp map[string]int64,
	lookbackDays int, campaignFrom, campaignTo int64, attributionKey string) (map[string][]AttributionKeyWeight, map[string]map[string][]AttributionKeyWeight, error) {

	usersAttribution := make(map[string][]AttributionKeyWeight)
	linkedEventUserCampaign := make(map[string]map[string][]AttributionKeyWeight)
	lookbackPeriod := int64(lookbackDays) * SecsInADay

	for _, val := range usersToBeAttributed {
		userId := val.CoalUserID
		eventName := val.EventName
		conversionTime := coalUserIdConversionTimestamp[val.CoalUserID]
		var attributionKeys []AttributionKeyWeight
		userSessions := sessions[userId]

		switch method {
		case AttributionMethodFirstTouch:
			attributionKeys = getFirstTouchId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodLastTouch:
			attributionKeys = getLastTouchId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break
		case AttributionMethodUShaped:
			attributionKeys = getUShaped(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case AttributionMethodFirstTouchNonDirect:
			attributionKeys = getFirstTouchNDId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo, attributionKey)
			break

		case AttributionMethodLastTouchNonDirect:
			attributionKeys = getLastTouchNDId(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo, attributionKey)
			break

		case AttributionMethodLinear:
			attributionKeys = getLinearTouch(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break
		case AttributionMethodTimeDecay:
			attributionKeys = getTimeDecay(attributionType, userSessions, conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		default:
			break
		}
		// In case a successful attribution could not happen, remove converted user.
		if len(attributionKeys) == 0 {
			delete(usersAttribution, userId)
			continue
		}
		if eventName == conversionEvent && val.EventType == EventTypeGoalEvent {
			usersAttribution[userId] = attributionKeys
		} else {
			if _, exist := linkedEventUserCampaign[eventName]; !exist {
				linkedEventUserCampaign[eventName] = make(map[string][]AttributionKeyWeight)
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

func SortInteractionTime(interactions []Interaction, sortingType string) []Interaction {

	// sort the interactions by interaction time ascending order
	if sortingType == SortASC {
		sort.Slice(interactions, func(i, j int) bool {
			return interactions[i].InteractionTime <= interactions[j].InteractionTime
		})
	} else if sortingType == SortDESC {
		sort.Slice(interactions, func(i, j int) bool {
			return interactions[i].InteractionTime >= interactions[j].InteractionTime
		})
	}
	return interactions
}

// returns list of attribution keys and corresponding weights from given attributionKeyTime map
func getLinearTouch(attributionType string, attributionTimerange map[string]UserSessionData,
	conversionTime, lookbackPeriod, from, to int64) []AttributionKeyWeight {

	var keys []AttributionKeyWeight
	interactions := getMergedInteractions(attributionTimerange)

	switch attributionType {
	case AttributionQueryTypeConversionBased:
		for _, interaction := range interactions {
			if isAdTouchWithinLookback(interaction.InteractionTime, conversionTime, lookbackPeriod) {
				keys = append(keys, AttributionKeyWeight{Key: interaction.AttributionKey, Weight: 0})
			}
		}

	case AttributionQueryTypeEngagementBased:
		for _, interaction := range interactions {
			if isAdTouchWithinLookback(interaction.InteractionTime, conversionTime,
				lookbackPeriod) && isAdTouchWithinCampaignOrQueryPeriod(interaction.InteractionTime, from, to) {
				keys = append(keys, AttributionKeyWeight{Key: interaction.AttributionKey, Weight: 0})
			}
		}

	}
	for i := range keys {
		keys[i].Weight = 1 / float64(len(keys))
	}

	return keys
}

// returns list of attribution keys and corresponding weights from given attributionKeyTime map using time decay attribution model.
func getTimeDecay(attributionType string, attributionTimerange map[string]UserSessionData,
	conversionTime, lookbackPeriod, from, to int64) []AttributionKeyWeight {

	var keys []AttributionKeyWeight
	var weight float64
	interactions := getMergedInteractions(attributionTimerange)
	totalWeight := 0.0

	switch attributionType {
	case AttributionQueryTypeConversionBased:
		for _, interaction := range interactions {
			if isAdTouchWithinLookback(interaction.InteractionTime, conversionTime, lookbackPeriod) {
				weight = calculateWeightForTimeDecay(conversionTime, interaction.InteractionTime)
				totalWeight += weight
				keys = append(keys, AttributionKeyWeight{Key: interaction.AttributionKey, Weight: weight})

			}
		}

	case AttributionQueryTypeEngagementBased:
		for _, interaction := range interactions {
			if isAdTouchWithinLookback(interaction.InteractionTime, conversionTime,
				lookbackPeriod) && isAdTouchWithinCampaignOrQueryPeriod(interaction.InteractionTime, from, to) {
				weight = calculateWeightForTimeDecay(conversionTime, interaction.InteractionTime)
				totalWeight += weight
				keys = append(keys, AttributionKeyWeight{Key: interaction.AttributionKey, Weight: weight})
			}
		}
	}
	for i := range keys {
		keys[i].Weight = keys[i].Weight / totalWeight
	}

	return keys
}

// returns the first attributionId and corresponding weight
func getFirstTouchId(attributionType string, attributionTimerange map[string]UserSessionData,
	conversionTime, lookbackPeriod, from, to int64) []AttributionKeyWeight {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = SortInteractionTime(interactions, SortASC)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) {
					return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) {
					return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
				}
			}
		}

	}
	return []AttributionKeyWeight{}
}

// returns the last attributionId and corresponding weight
func getLastTouchId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime,
	lookbackPeriod, from, to int64) []AttributionKeyWeight {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = SortInteractionTime(interactions, SortDESC)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) {
					return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) {
					return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
				}
			}
		}

	}
	return []AttributionKeyWeight{}
}

// returns the first and the last attributionId and corresponding weights
func getUShaped(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime,
	lookbackPeriod, from, to int64) []AttributionKeyWeight {
	firstTouch := getFirstTouchId(attributionType, attributionTimerange, conversionTime, lookbackPeriod, from, to)
	lastTouch := getLastTouchId(attributionType, attributionTimerange, conversionTime, lookbackPeriod, from, to)
	keys := append(firstTouch, lastTouch...)
	if len(keys) > 0 {
		keys[0].Weight = 0.5
		keys[1].Weight = 0.5
	}

	return keys
}

// returns the first non $none attributionId and corresponding weight
func getFirstTouchNDId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime,
	lookbackPeriod, from, to int64, attributionKey string) []AttributionKeyWeight {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = SortInteractionTime(interactions, SortASC)
	directSessionExists := false
	noneKey := GetNoneKeyForAttributionType(attributionKey)

	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) {
					if interactions[i].AttributionKey != noneKey {
						return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
					}
					directSessionExists = true
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) {
					if interactions[i].AttributionKey != noneKey {
						return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
					}
					directSessionExists = true
				}
			}
		}

	}
	// return $none key only if Direct session was seen
	if directSessionExists {
		return []AttributionKeyWeight{{Key: noneKey, Weight: 1}}
	} else {
		return []AttributionKeyWeight{}
	}
}

// returns the last non $none attributionId and corresponding weight
func getLastTouchNDId(attributionType string, attributionTimerange map[string]UserSessionData, conversionTime,
	lookbackPeriod, from, to int64, attributionKey string) []AttributionKeyWeight {

	interactions := getMergedInteractions(attributionTimerange)
	interactions = SortInteractionTime(interactions, SortDESC)
	directSessionExists := false
	noneKey := GetNoneKeyForAttributionType(attributionKey)
	if len(interactions) > 0 {

		switch attributionType {
		case AttributionQueryTypeConversionBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) {
					if interactions[i].AttributionKey != noneKey {
						return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
					}
					directSessionExists = true
				}
			}
		case AttributionQueryTypeEngagementBased:
			for i := 0; i < len(interactions); i++ {
				if isAdTouchWithinLookback(interactions[i].InteractionTime, conversionTime, lookbackPeriod) &&
					isAdTouchWithinCampaignOrQueryPeriod(interactions[i].InteractionTime, from, to) {
					if interactions[i].AttributionKey != noneKey {
						return []AttributionKeyWeight{{Key: interactions[i].AttributionKey, Weight: 1}}
					}
					directSessionExists = true
				}
			}
		}
	}
	// return $none key only if Direct session was seen
	if directSessionExists {
		return []AttributionKeyWeight{{Key: noneKey, Weight: 1}}
	} else {
		return []AttributionKeyWeight{}
	}
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

// calculateWeightForTimeDecay returns weight based on conversion time and interaction time using following formula:
// y = pow(2,-x/7), where x is number of days interaction happened prior to the conversion.
// In the formula,'7' is half-life. If touchpoint 'x1' is 7 days before a different touchpoint 'x2' and 'x2' receives credit 'y',
// then 'x1' will receive credit 'y/2'. It is ensured by the formula.
func calculateWeightForTimeDecay(conversionTime int64, interactionTime int64) float64 {
	halfLife := 7.0
	days := (conversionTime - interactionTime) / SecsInADay
	weight := math.Pow(2, float64(-days)/halfLife)
	return weight
}
