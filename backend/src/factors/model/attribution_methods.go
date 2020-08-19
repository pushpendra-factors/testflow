package model

import "sort"

type pair struct {
	key   string
	value int64
}

// This method maps the user to the attribution key based on given attribution methodology.
func applyAttribution(method string, conversionEvent string, usersToBeAttributed []UserEventInfo,
	userInitialSession map[string]map[string]RangeTimestamp) (map[string][]string, map[string]map[string][]string, error) {

	usersAttribution := make(map[string][]string)
	linkedEventUserCampaign := make(map[string]map[string][]string)

	for _, val := range usersToBeAttributed {
		userId := val.coalUserId
		eventName := val.eventName

		attributionKeys := []string{PropertyValueNone}
		switch method {
		case ATTRIBUTION_METHOD_FIRST_TOUCH:
			attributionKeys = getFirstTouchId(userInitialSession[userId])
			break

		case ATTRIBUTION_METHOD_LAST_TOUCH:
			attributionKeys = getLastTouchId(userInitialSession[userId])
			break

		case ATTRIBUTION_METHOD_FIRST_TOUCH_NON_DIRECT:
			attributionKeys = getFirstTouchNDId(userInitialSession[userId])
			break

		case ATTRIBUTION_METHOD_LAST_TOUCH_NON_DIRECT:
			attributionKeys = getLastTouchNDId(userInitialSession[userId])
			break

		case ATTRIBUTION_METHOD_LINEAR:
			attributionKeys = getLinearTouch(userInitialSession[userId])
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

// returns the last non $none attributionId
func getLastTouchNDId(attributionTimerange map[string]RangeTimestamp) []string {

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
			if pair.key != PropertyValueNone {
				key = pair.key
				// break on first non $none
				break
			}
		}
	}
	return []string{key}
}

// returns list of attribution keys from given attributionKeyTime map
func getLinearTouch(attributionTimerange map[string]RangeTimestamp) []string {

	var keys []string
	for aId, _ := range attributionTimerange {
		keys = append(keys, aId)
	}
	return keys
}

// returns the last attributionId
func getLastTouchId(attributionTimerange map[string]RangeTimestamp) []string {

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
		return []string{attributionIds[0].key}
	}
	return []string{key}
}

// returns the first attributionId
func getFirstTouchId(attributionTimerange map[string]RangeTimestamp) []string {
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
		return []string{attributionIds[0].key}
	}
	return []string{key}
}

// returns the first non $none attributionId
func getFirstTouchNDId(attributionTimerange map[string]RangeTimestamp) []string {

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
			if pair.key != PropertyValueNone {
				key = pair.key
				// break on first non $none
				break
			}
		}
	}
	return []string{key}
}
