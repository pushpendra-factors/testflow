package v1

import (
	"errors"
	mid "factors/middleware"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetPricingForPlansAndAddonsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
	}
	var defaultReponse model.PlansAndAddOnsPrices
	defaultReponse = append(defaultReponse, model.SubscriptionProductPrice{
		Name:         "Basic Monthly USD",
		ExternalName: "Basic",
		Type:         "Plan",
		ID:           "basic-monthly-usd",
		Price:        100,
		Period:       1,
	})
	defaultReponse = append(defaultReponse, model.SubscriptionProductPrice{
		Name:   "Additional Account",
		Type:   "AddOn",
		ID:     "additional-accounts",
		Price:  50,
		Period: 1,
	})

	c.JSON(http.StatusOK, defaultReponse)
}

func UpgradePlanHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
	}
	planID := c.Query("updated_plan_id")
	if planID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PLAN"))
	}
	defaultResponse := `{
		"hosted_page": {
			"url": "https://factors-test.chargebee.com/pages/v3/3Wcu6cdCdwA8t1bOsFUOh0cutvxwviakKzF/",
		}
	}`
	c.JSON(http.StatusOK, defaultResponse)
}

func GetSubscriptionDetailsHander(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
	}
	var subscriptionDetails []model.SubscriptionDetail
	subscriptionDetails = append(subscriptionDetails, model.SubscriptionDetail{
		Type:   "Plan",
		ID:     "Basic-USD-monthly",
		Amount: 100,
	})

	defaultResponse := model.Subscription{
		Status:              "active",
		RenewsOn:            time.Now(),
		SubscriptionDetails: subscriptionDetails,
	}
	c.JSON(http.StatusOK, defaultResponse)
}
