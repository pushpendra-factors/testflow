package model

import "sort"

type pair struct {
	key   string
	value int64
}

// This method maps the user to the attribution key based on given attribution methodology.
func ApplyAttribution(method string, conversionEvent string, usersToBeAttributed []UserEventInfo,
	userInitialSession map[string]map[string]RangeTimestamp,
	coalUserIdConversionTimestamp map[string]int64,
	lookbackDays int) (map[string][]string, map[string]map[string][]string, error) {

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
			attributionKeys = getFirstTouchId(userInitialSession[userId], conversionTime, lookbackPeriod)
			break

		case AttributionMethodLastTouch:
			attributionKeys = getLastTouchId(userInitialSession[userId], conversionTime, lookbackPeriod)
			break

		case AttributionMethodFirstTouchNonDirect:
			attributionKeys = getFirstTouchNDId(userInitialSession[userId], conversionTime, lookbackPeriod)
			break

		case AttributionMethodLastTouchNonDirect:
			attributionKeys = getLastTouchNDId(userInitialSession[userId], conversionTime, lookbackPeriod)
			break

		case AttributionMethodLinear:
			attributionKeys = getLinearTouch(userInitialSession[userId], conversionTime, lookbackPeriod)
			break

		default:
			break
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
func getLinearTouch(attributionTimerange map[string]RangeTimestamp, conversionTime int64,
	lookbackPeriod int64) []string {

	var keys []string
	for aId, rangeTime := range attributionTimerange {
		if isConversionWithinLookback(rangeTime.MinTimestamp, conversionTime,
			lookbackPeriod) {
			keys = append(keys, aId)
		}
	}
	return keys
}

// returns the first attributionId
func getFirstTouchId(attributionTimerange map[string]RangeTimestamp, conversionTime int64,
	lookbackPeriod int64) []string {
	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MinTimestamp for FirstTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MinTimestamp})
	}
	key := PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value < attributionIds[j].value
		})
		for i := 0; i < len(attributionIds); i++ {
			if isConversionWithinLookback(attributionIds[i].value, conversionTime,
				lookbackPeriod) {
				return []string{attributionIds[i].key}
			}
		}
	}
	return []string{key}
}

// returns the last attributionId
func getLastTouchId(attributionTimerange map[string]RangeTimestamp, conversionTime int64,
	lookbackPeriod int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MaxTimestamp for LastTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MaxTimestamp})
	}
	key := PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value > attributionIds[j].value
		})
		for i := 0; i < len(attributionIds); i++ {
			if isConversionWithinLookback(attributionIds[i].value, conversionTime,
				lookbackPeriod) {
				return []string{attributionIds[i].key}
			}
		}
	}
	return []string{key}
}

// returns the first non $none attributionId
func getFirstTouchNDId(attributionTimerange map[string]RangeTimestamp, conversionTime int64,
	lookbackPeriod int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MinTimestamp for FirstTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MinTimestamp})
	}
	key := PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value < attributionIds[j].value
		})
		key = attributionIds[0].key
		for _, pair := range attributionIds {
			if isConversionWithinLookback(pair.value, conversionTime,
				lookbackPeriod) {
				if pair.key != PropertyValueNone {
					key = pair.key
					// break on first non $none
					break
				}
			}
		}
	}
	return []string{key}
}

// returns the last non $none attributionId
func getLastTouchNDId(attributionTimerange map[string]RangeTimestamp, conversionTime int64,
	lookbackPeriod int64) []string {

	var attributionIds []pair
	for aId, rangeT := range attributionTimerange {
		// MaxTimestamp for LastTouch
		attributionIds = append(attributionIds, pair{aId, rangeT.MaxTimestamp})
	}
	key := PropertyValueNone
	if len(attributionIds) > 0 {
		sort.Slice(attributionIds, func(i, j int) bool {
			return attributionIds[i].value > attributionIds[j].value
		})
		key = attributionIds[0].key
		for _, pair := range attributionIds {
			if isConversionWithinLookback(pair.value, conversionTime,
				lookbackPeriod) {
				if pair.key != PropertyValueNone {
					key = pair.key
					// break on first non $none
					break
				}
			}
		}
	}
	return []string{key}
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
