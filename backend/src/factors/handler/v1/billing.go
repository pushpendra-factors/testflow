package v1

import (
	"errors"
	billing "factors/billing/chargebee"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
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
		Name:         "Basic Yearly USD",
		ExternalName: "Basic",
		Type:         "Plan",
		ID:           "basic-yearly-usd",
		Price:        1200,
		Period:       1,
	})
	defaultReponse = append(defaultReponse, model.SubscriptionProductPrice{
		Name:         "Free montly USD",
		ExternalName: "Free",
		Type:         "Plan",
		ID:           "basic-yearly-usd",
		Price:        0,
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

func UpdateSubscriptionHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
	}
	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithError(http.StatusBadRequest, errors.New("BAD REQUEST"))
	}

	var updateSubscriptionParams model.UpdateSubscriptionParams
	err := c.BindJSON(&updateSubscriptionParams)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID REQUEST"))
		return
	}
	hostedPage, _, err := billing.GetUpgradeChargebeeSubscriptionCheckoutURL(project.BillingSubscriptionID, updateSubscriptionParams)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	url := buildCheckoutUrl(projectId, hostedPage.Url)

	// redirect
	c.Redirect(http.StatusPermanentRedirect, url)
}

func GetSubscriptionDetailsHander(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
	}
	var subscription model.Subscription
	var subscriptionDetails []model.SubscriptionDetail

	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.JSON(http.StatusBadRequest, subscription)
	}

	res, err := billing.GetCurrentSubscriptionDetails(project.BillingSubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, subscription)
	}

	subscription.Status = string(res.Status)
	subscription.RenewsOn = time.Unix(res.NextBillingAt, 0)

	for _, item := range res.SubscriptionItems {
		subscriptionDetails = append(subscriptionDetails, model.SubscriptionDetail{
			Type:   string(item.ItemType),
			ID:     item.ItemPriceId,
			Amount: item.Amount,
		})
	}

	subscription.SubscriptionDetails = subscriptionDetails

	c.JSON(http.StatusOK, subscription)
}

func buildCheckoutUrl(projectID int64, checkouturl string) string {
	callBackUrl := C.GetProtocol() + C.GetAPIDomain() + "/billing/upgarde/callback" + "&state=" + fmt.Sprint(projectID)
	return fmt.Sprintf("%s&redirect_url=%s", checkouturl, callBackUrl) // change this to const later
}

func BillingUpgradeCallbackHandler(c *gin.Context) {
	state := c.Query("state")
	projectID, _ := strconv.ParseInt(state, 10, 64)
	err := store.GetStore().TriggerSyncChargebeeToFactors(projectID)
	if err != nil {
		c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing?error=SERVER_ERROR")
		return
	subscriptionDetails = append(subscriptionDetails, model.SubscriptionDetail{
		Type:         "Plan",
		ID:           "Basic-USD-monthly",
		Amount:       100,
		ExternalName: "Basic",
	})

	defaultResponse := model.Subscription{
		Status:              "active",
		RenewsOn:            time.Now(),
		BillingPeriod:       1,
		BillingPeriodUnit:   "month",
		SubscriptionDetails: subscriptionDetails,
	}
	c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing")
}
