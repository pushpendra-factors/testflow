package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	chargebee "factors/chargebee"

	"github.com/chargebee/chargebee-go/v3/actions/subscription"
	"github.com/gin-gonic/gin"
)

func GetPlansAndAddonsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	items, err := chargebee.ListPlansAndAddOnsFromChargebee()
	if err != nil {
		return nil, http.StatusInternalServerError, "Error", ErrorMessages[err.Error()], true
	}
	return items, http.StatusFound, "", "", false
}

func GetPricingForPlansAndAddonsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	itemPrices, err := chargebee.ListPlansAndAddOnsPricesFromChargebee()
	if err != nil {
		return nil, http.StatusInternalServerError, "Error", ErrorMessages[err.Error()], true
	}
	return itemPrices, http.StatusFound, "", "", false

}

func UpgradePlanHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	planID := c.Query("updated_plan_id")
	if planID == "" {
		return nil, http.StatusBadRequest, "Error", ErrorMessages["Plan not found"], true
	}
	project, status := store.GetStore().GetProject(projectId)
	if status != http.StatusFound {
		return nil, http.StatusInternalServerError, "Error", ErrorMessages["Failed to get project"], true
	}
	hostedPage, _, err := chargebee.GetUpgradeChargebeeSubscriptionCheckoutURL(project.BillingSubscriptionID, planID)
	if err != nil {
		return nil, http.StatusInternalServerError, "Failed to upgrade plan", ErrorMessages["Failed to upgrade plan"], true
	}
	// redirect from backend ?
	// c.Redirect(http.StatusTemporaryRedirect, hostedPage.Url)
	return hostedPage, http.StatusAccepted, "", "", false
}

func GetSubscriptionDetailsHander(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	project, status := store.GetStore().GetProject(projectId)
	if status != http.StatusFound {
		return nil, http.StatusInternalServerError, "Error", ErrorMessages["Failed to get project"], true
	}
	subscription, err := chargebee.GetCurrentSubscriptionDetails(project.BillingSubscriptionID)
	if err != nil {
		return nil, http.StatusInternalServerError, "Failed to get subscription details", ErrorMessages["Failed to get subscription details"], true
	}
	return subscription, http.StatusFound, "", "", false
}
