package model

import (
	U "factors/util"
	"fmt"
	"strings"
)

type ChannelPropertyFilter struct {
	Property  string `json:"property"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
	LogicalOp string `json:"logical_operator"`
}

type ChannelPropertyRule struct {
	Channel    string                  `json:"channel"`
	Conditions []ChannelPropertyFilter `json:"conditions"`
}

const (
	direct         = "Direct"
	paidSearch     = "Paid Search"
	paidSocial     = "Paid Social"
	organicSearch  = "Organic Search"
	organicSocial  = "Organic Social"
	email          = "Email"
	affiliate      = "Affiliate"
	otherCampaigns = "Other Campaigns"
	referral       = "Referral"
)

var DefaultChannelPropertyRules = []ChannelPropertyRule{
	{
		Channel: direct,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_GCLID,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_FBCLID,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_CAMPAIGN,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: paidSearch,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_GCLID,
				Condition: NotEqualOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: paidSearch,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "google",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "bing",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "adwords",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "youtube",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "adwords",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "display",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "cpm",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: paidSearch,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "google.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "bing.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "duckduckgo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "yahoo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "yandex.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "baidu.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_CAMPAIGN,
				Condition: NotEqualOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: paidSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_FBCLID,
				Condition: NotEqualOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: paidSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "facebook",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "fb",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "linkedin",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "twitter",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "quora",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "pinterest",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "snapchat",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "instagram",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "cpm",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: paidSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "paidsocial",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: paidSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "paidsocial",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: paidSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "cpm",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "facebook.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "linkedin.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "quora.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "pinterest.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "twitter.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "snapchat.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "instagram.",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: organicSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_FBCLID,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: NotEqualOpStr,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: NotEqualOpStr,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: NotEqualOpStr,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: NotEqualOpStr,
				Value:     "cpm",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "facebook.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "linkedin.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "quora.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "pinterest.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "twitter.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "snapchat.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "youtube.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "instagram.",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: organicSearch,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_GCLID,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_FBCLID,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_CAMPAIGN,
				Condition: EqualsOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "google.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "bing.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "duckduckgo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "yahoo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "yandex.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: ContainsOpStr,
				Value:     "baidu.",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: email,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "email",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: email,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "email",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: affiliate,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: EqualsOpStr,
				Value:     "affiliate",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: affiliate,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_MEDIUM,
				Condition: EqualsOpStr,
				Value:     "affiliate",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: otherCampaigns,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_CAMPAIGN,
				Condition: NotEqualOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
	{
		Channel: referral,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: NotEqualOpStr,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
}

// condition : (medium=paid OR medium=cpc ) AND (referral domain contains either of ("facebook.","linkedin.")
// in code : rules = [{ property: medium, L_OP: AND, OP: contains, value: paid}, {property: medium, L_OP: OR, OP: contains, value: cpc}, {property: ref_domain, L_OP: AND, OP: contains, value: 'facebook.'}, {property: ref_domain, L_OP: OR, OP: contains, value: 'linkedin.'}]

// solution for now:
// group rule based off of property and then run filter checks on top of them

func groupConditionsBasedOnProperty(conditions []ChannelPropertyFilter) map[string][]ChannelPropertyFilter {
	groupedConditions := make(map[string][]ChannelPropertyFilter, 0)
	for _, filter := range conditions {
		if groupedConditions[filter.Property] == nil {
			groupedConditions[filter.Property] = make([]ChannelPropertyFilter, 0)
		}
		groupedConditions[filter.Property] = append(groupedConditions[filter.Property], filter)
	}
	return groupedConditions
}
func EvaluateChannelPropertyRules(channelGroupRules []ChannelPropertyRule, sessionPropertiesMap U.PropertiesMap, projectID uint64) string {
	for _, rule := range channelGroupRules {
		groupedConditions := groupConditionsBasedOnProperty(rule.Conditions)
		checkCondition := true
		for _, conditions := range groupedConditions {
			var checkConditionForProperty bool
			for index, filter := range conditions {
				if index == 0 {
					checkConditionForProperty = checkFilter(sessionPropertiesMap, filter)
				} else {
					if filter.LogicalOp == LOGICAL_OP_OR {
						checkConditionForProperty = checkConditionForProperty || checkFilter(sessionPropertiesMap, filter)
					} else {
						checkConditionForProperty = checkConditionForProperty && checkFilter(sessionPropertiesMap, filter)
					}
				}
			}
			checkCondition = checkCondition && checkConditionForProperty
		}
		if checkCondition {
			return rule.Channel
		}
	}
	return "Others"
}

func checkFilter(sessionPropertesMap U.PropertiesMap, filter ChannelPropertyFilter) bool {
	propertyValueInterface, isExists := sessionPropertesMap[filter.Property]
	propertyValue := fmt.Sprintf("%v", propertyValueInterface)

	lowerCaseFilterValue := strings.ToLower(filter.Value)
	lowerCasePropertyValue := strings.ToLower(propertyValue)

	switch filter.Condition {
	case EqualsOpStr:
		return compareEqual(isExists, lowerCasePropertyValue, lowerCaseFilterValue)
	case NotEqualOpStr:
		return !compareEqual(isExists, lowerCasePropertyValue, lowerCaseFilterValue)
	case ContainsOpStr:
		return strings.Contains(lowerCasePropertyValue, lowerCaseFilterValue)
	case NotContainsOpStr:
		return !strings.Contains(lowerCasePropertyValue, lowerCaseFilterValue)
	}
	return false
}
func compareEqual(isExists bool, propertyValue string, filterValue string) bool {
	if filterValue == "$none" {
		return !isExists || propertyValue == filterValue || propertyValue == ""
	} else {
		return propertyValue == filterValue
	}
}
