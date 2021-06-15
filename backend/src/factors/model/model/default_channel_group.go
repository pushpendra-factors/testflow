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
	direct        = "Direct"
	paidSearch    = "Paid Search"
	paidSocial    = "Paid Social"
	organicSearch = "Organic Search"
	organicSocial = "Organic Social"
	email         = "Email"
	affiliate     = "Affiliate"
	referral      = "Referral"
)

var DefaultChannelPropertyRules = []ChannelPropertyRule{
	{
		Channel: direct,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_GCLID,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_FBCLID,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_CAMPAIGN,
				Condition: COMPARE_EQUAL,
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
				Condition: COMPARE_NOT_EQUAL,
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
				Condition: COMPARE_EQUAL,
				Value:     "google",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "bing",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "adwords",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "youtube",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "adwords",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "display",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
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
				Condition: COMPARE_CONTAINS,
				Value:     "google.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "bing.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "duckduckgo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "yahoo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "yandex.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "baidu.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_CAMPAIGN,
				Condition: COMPARE_NOT_EQUAL,
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
				Condition: COMPARE_NOT_EQUAL,
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
				Condition: COMPARE_EQUAL,
				Value:     "facebook",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "fb",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "linkedin",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "twitter",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "quora",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "pinterest",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "snapchat",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
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
				Condition: COMPARE_EQUAL,
				Value:     "paidsocial",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "paidsocial",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: paidSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "cpm",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "facebook.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "linkedin.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "quora.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "pinterest.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "twitter.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "snapchat.",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: organicSocial,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_FBCLID,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_NOT_EQUAL,
				Value:     "paid",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_NOT_EQUAL,
				Value:     "cpc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_NOT_EQUAL,
				Value:     "ppc",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_NOT_EQUAL,
				Value:     "cpm",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "facebook.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "linkedin.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "quora.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "pinterest.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "twitter.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "snapchat.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "youtube.",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: organicSearch,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_GCLID,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_FBCLID,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_CAMPAIGN,
				Condition: COMPARE_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "google.",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "bing.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "duckduckgo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "yahoo.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
				Value:     "yandex.",
				LogicalOp: LOGICAL_OP_OR,
			},
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_CONTAINS,
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
				Condition: COMPARE_EQUAL,
				Value:     "email",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "email",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: affiliate,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.EP_SOURCE,
				Condition: COMPARE_EQUAL,
				Value:     "affiliate",
				LogicalOp: LOGICAL_OP_AND,
			},
			{
				Property:  U.EP_MEDIUM,
				Condition: COMPARE_EQUAL,
				Value:     "affiliate",
				LogicalOp: LOGICAL_OP_OR,
			},
		},
	},
	{
		Channel: referral,
		Conditions: []ChannelPropertyFilter{
			{
				Property:  U.SP_INITIAL_REFERRER_DOMAIN,
				Condition: COMPARE_NOT_EQUAL,
				Value:     "$none",
				LogicalOp: LOGICAL_OP_AND,
			},
		},
	},
}

func EvaluateChannelPropertyRules(channelGroupRules []ChannelPropertyRule, sessionPropertiesMap U.PropertiesMap) string {
	for _, rule := range channelGroupRules {
		var checkCondition bool
		for index, filter := range rule.Conditions {
			if index == 0 {
				checkCondition = checkFilter(sessionPropertiesMap, filter)
			} else {
				if filter.LogicalOp == LOGICAL_OP_OR {
					checkCondition = checkCondition || checkFilter(sessionPropertiesMap, filter)
				} else {
					checkCondition = checkCondition && checkFilter(sessionPropertiesMap, filter)
				}
			}
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
	case COMPARE_EQUAL:
		return compareEqual(isExists, lowerCasePropertyValue, lowerCaseFilterValue)
	case COMPARE_NOT_EQUAL:
		return !compareEqual(isExists, lowerCasePropertyValue, lowerCaseFilterValue)
	case COMPARE_CONTAINS:
		return strings.Contains(lowerCasePropertyValue, lowerCaseFilterValue)
	case COMPARE_NOT_CONTAINS:
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
