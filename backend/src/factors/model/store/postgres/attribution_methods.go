package postgres

import (
	"factors/model/model"
	"sort"
)

type pair struct {
	key   string
	value int64
}

// This method maps the user to the attribution key based on given attribution methodology.
func (pg *Postgres) ApplyAttribution(method string, conversionEvent string, usersToBeAttributed []model.UserEventInfo,
	userInitialSession map[string]map[string]model.RangeTimestamp,
	coalUserIdConversionTimestamp map[string]int64,
	lookbackDays int, campaignFrom, campaignTo int64) (map[string][]string, map[string]map[string][]string, error) {

	usersAttribution := make(map[string][]string)
	linkedEventUserCampaign := make(map[string]map[string][]string)
	lookbackPeriod := int64(lookbackDays) * SecsInADay
	for _, val := range usersToBeAttributed {
		userId := val.CoalUserID
		eventName := val.EventName
		conversionTime := coalUserIdConversionTimestamp[val.CoalUserID]
		attributionKeys := []string{model.PropertyValueNone}
		switch method {
		case model.AttributionMethodFirstTouch:
			attributionKeys = getFirstTouchId(userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case model.AttributionMethodLastTouch:
			attributionKeys = getLastTouchId(userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case model.AttributionMethodFirstTouchNonDirect:
			attributionKeys = getFirstTouchNDId(userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case model.AttributionMethodLastTouchNonDirect:
			attributionKeys = getLastTouchNDId(userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		case model.AttributionMethodLinear:
			attributionKeys = getLinearTouch(userInitialSession[userId], conversionTime,
				lookbackPeriod, campaignFrom, campaignTo)
			break

		default:
			break
		}
		// in case a successful attribution could not happen, remove converted user
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
func getLinearTouch(attributionTimerange map[string]model.RangeTimestamp, conversionTime int64,
	lookbackPeriod int64, from, to int64) []string {

	var keys []string
	for aId, rangeTime := range attributionTimerange {
		if isConversionWithinLookback(rangeTime.MinTimestamp, conversionTime,
			lookbackPeriod) && isConversionWithinQueryPeriod(rangeTime.MinTimestamp, from, to) {
			keys = append(keys, aId)
		}
	}
	return keys
}

// returns the first attributionId
func getFirstTouchId(attributionTimerange map[string]model.RangeTimestamp, conversionTime int64,
	lookbackPeriod int64, from, to int64) []string {
	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MinTimestamp for FirstTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MinTimestamp})
	}
	key := model.PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value < attributionIds[j].value
		})
		for i := 0; i < len(attributionIds); i++ {
			if isConversionWithinLookback(attributionIds[i].value, conversionTime,
				lookbackPeriod) {
				// exit condition: Attribution not in query range
				if !isConversionWithinQueryPeriod(attributionIds[i].value, from, to) {
					return []string{}
				}
				return []string{attributionIds[i].key}
			}
		}
	}
	return []string{key}
}

// returns the last attributionId
func getLastTouchId(attributionTimerange map[string]model.RangeTimestamp, conversionTime int64,
	lookbackPeriod int64, from, to int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MaxTimestamp for LastTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MaxTimestamp})
	}
	key := model.PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value > attributionIds[j].value
		})
		for i := 0; i < len(attributionIds); i++ {
			if isConversionWithinLookback(attributionIds[i].value, conversionTime,
				lookbackPeriod) {
				// exit condition: Attribution not in query range
				if !isConversionWithinQueryPeriod(attributionIds[i].value, from, to) {
					return []string{}
				}
				return []string{attributionIds[i].key}
			}
		}
	}
	return []string{key}
}

// returns the first non $none attributionId
func getFirstTouchNDId(attributionTimerange map[string]model.RangeTimestamp, conversionTime int64,
	lookbackPeriod int64, from, to int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MinTimestamp for FirstTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MinTimestamp})
	}
	key := model.PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value < attributionIds[j].value
		})
		key = attributionIds[0].key
		for _, atPair := range attributionIds {
			if isConversionWithinLookback(atPair.value, conversionTime,
				lookbackPeriod) {
				if atPair.key != model.PropertyValueNone {
					// exit condition: Attribution not in query range
					if !isConversionWithinQueryPeriod(atPair.value, from, to) {
						return []string{}
					}
					key = atPair.key
					// break on first non $none
					break
				}
			}
		}
	}
	return []string{key}
}

// returns the last non $none attributionId
func getLastTouchNDId(attributionTimerange map[string]model.RangeTimestamp, conversionTime int64,
	lookbackPeriod int64, from, to int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MaxTimestamp for LastTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MaxTimestamp})
	}
	key := model.PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value > attributionIds[j].value
		})
		key = attributionIds[0].key
		for _, atPair := range attributionIds {
			if isConversionWithinLookback(atPair.value, conversionTime,
				lookbackPeriod) {
				if atPair.key != model.PropertyValueNone {
					// exit condition: Attribution not in query range
					if !isConversionWithinQueryPeriod(atPair.value, from, to) {
						return []string{}
					}
					key = atPair.key
					// break on first non $none
					break
				}
			}
		}
	}
	return []string{key}
}

// isConversionWithinQueryPeriod checks if attribution time is within query (campaign) period
func isConversionWithinQueryPeriod(attributionTime int64, from, to int64) bool {
	return attributionTime >= from && attributionTime <= to
}

// isConversionWithinLookback checks if attribution time is within lookback period
func isConversionWithinLookback(attributionTime int64,
	conversionTime int64, lookbackPeriod int64) bool {

	if conversionTime >= attributionTime {
		if (conversionTime - attributionTime) <= lookbackPeriod {
			return true
		}
	}
	return false
}
