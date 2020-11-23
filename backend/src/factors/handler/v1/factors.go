package v1

import (
	mid "factors/middleware"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Factors Constants
var CAMPAIGNTYPE string = "campaign"
var ATTRIBUTETYPE string = "attribute"
var JOURNEYTYPE string = "journey"

// Factors object
type Factors struct {
	Insights              []FactorsInsights `json:"insights"`
	InsightsUserCount     float64           `json:"insights_user_count"`
	TotalUsersCount       float64           `json:"total_users_count"`
	OverallPercentageText string            `json:"overall_percentage_text"`
	OverallPercentage     float64           `json:"overall_percentage"`
}

// FactorsAttributeTuple object
type FactorsAttributeTuple struct {
	FactorsAttributeKey   string `json:"factors_attribute_key"`
	FactorsAttributeValue string `json:"factors_attribute_value"`
}

// FactorsInsights object
type FactorsInsights struct {
	FactorsInsightsAttribute      []FactorsAttributeTuple `json:"factors_insights_attribute"`
	FactorsInsightsKey            string                  `json:"factors_insights_key"`
	FactorsInsightsText           string                  `json:"factors_insights_text"`
	FactorsInsightsMultiplier     float64                 `json:"factors_insights_multiplier"`
	FactorsInsightsPercentage     float64                 `json:"factors_insights_percentage"`
	FactorsUsersCount             float64                 `json:"factors_users_count"`
	FactorsMultiplierIncreaseFlag bool                    `json:"factors_multiplier_increase_flag"`
	FactorsInsightsType           string                  `json:"factors_insights_type"`
	FactorsSubInsights            []*FactorsInsights      `json:"factors_sub_insights"`
}

// GetAllFactorsHandler - Factors handler
func GetAllFactorsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	factorsJourneyWithJourney := FactorsInsights{
		FactorsInsightsKey:            "www.acme.com/colloboration",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     25,
		FactorsInsightsText:           "of which users coming via .../product shows 2x lower goal completion",
		FactorsUsersCount:             75,
		FactorsMultiplierIncreaseFlag: false,
		FactorsInsightsType:           JOURNEYTYPE,
	}
	factorsJourneyWithAttributes := FactorsInsights{
		FactorsInsightsMultiplier:     2.5,
		FactorsInsightsPercentage:     75,
		FactorsInsightsText:           "where $city= chennai shows 2.5x higher goal completion",
		FactorsUsersCount:             175,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           JOURNEYTYPE,
		FactorsInsightsAttribute: []FactorsAttributeTuple{
			FactorsAttributeTuple{
				FactorsAttributeKey:   "$city",
				FactorsAttributeValue: "chennai",
			},
		},
	}
	factorsJourney := FactorsInsights{
		FactorsInsightsKey:            "www.acme.com/product",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     50,
		FactorsInsightsText:           "of which users coming via .../product shows 2x higher goal completion",
		FactorsUsersCount:             250,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           JOURNEYTYPE,
		FactorsSubInsights: []*FactorsInsights{
			&factorsJourneyWithAttributes,
			&factorsJourneyWithJourney,
		},
	}

	factorsCampaignLevel2 := FactorsInsights{
		FactorsInsightsKey:            "FreshDeskCampaign2",
		FactorsInsightsText:           "and then FreshDeskBrand2 shows 2x lower conversion",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     25,
		FactorsUsersCount:             50,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           CAMPAIGNTYPE,
	}
	factorsCampaignLevel1 := FactorsInsights{
		FactorsInsightsKey:            "FreshDeskCampaign1",
		FactorsInsightsText:           "and then FreshDeskBrand1 shows 2x higher conversion",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     25,
		FactorsUsersCount:             175,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           CAMPAIGNTYPE,
		FactorsSubInsights: []*FactorsInsights{
			&factorsCampaignLevel2,
		},
	}
	factorsCampaign := FactorsInsights{
		FactorsInsightsKey:            "FreshDeskCampaign",
		FactorsInsightsText:           "of which visitors from the campaign FreshDeskBrand shows 2x higher conversion",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     25,
		FactorsUsersCount:             250,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           CAMPAIGNTYPE,
		FactorsSubInsights: []*FactorsInsights{
			&factorsCampaignLevel1,
		},
	}

	factorsAttributeLevel1 := FactorsInsights{
		FactorsInsightsAttribute: []FactorsAttributeTuple{
			FactorsAttributeTuple{
				FactorsAttributeKey:   "$osversion",
				FactorsAttributeValue: "1.0",
			},
		},
		FactorsInsightsText:           "of which visitors with $osversion=1.0 shows 2x higher goal conversion",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     50,
		FactorsUsersCount:             250,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           ATTRIBUTETYPE,
	}

	factorsAttribute := FactorsInsights{
		FactorsInsightsAttribute: []FactorsAttributeTuple{
			FactorsAttributeTuple{
				FactorsAttributeKey:   "$deviceType",
				FactorsAttributeValue: "ios",
			},
		},
		FactorsInsightsText:           "of which visitors with $deviceType=ios shows 2x higher goal conversion",
		FactorsInsightsMultiplier:     2,
		FactorsInsightsPercentage:     50,
		FactorsUsersCount:             500,
		FactorsMultiplierIncreaseFlag: true,
		FactorsInsightsType:           ATTRIBUTETYPE,
		FactorsSubInsights: []*FactorsInsights{
			&factorsAttributeLevel1,
		},
	}
	factors := Factors{
		TotalUsersCount:       1000,
		OverallPercentage:     50,
		InsightsUserCount:     500,
		OverallPercentageText: "50% of all users have completed this goal",
		Insights: []FactorsInsights{
			factorsJourney, factorsCampaign, factorsAttribute,
		},
	}
	c.JSON(http.StatusOK, factors)
}
