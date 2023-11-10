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
	log "github.com/sirupsen/logrus"
	"net/http"
	UR "net/url"
	"strconv"
	"strings"
	"time"
)

func GetPricingForPlansAndAddonsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
	}
	itemPrices, err := billing.ListPlansAndAddOnsPricesFromChargebee()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.PlansAndAddOnsPrices{})
	}
	var res model.PlansAndAddOnsPrices
	for _, itemPrice := range itemPrices {
		res = append(res, model.SubscriptionProductPrice{
			Type:         string(itemPrice.ItemType),
			Name:         itemPrice.Name,
			ExternalName: itemPrice.ExternalName,
			ID:           itemPrice.Id,
			Price:        itemPrice.Price,
			PeriodUnit:   string(itemPrice.PeriodUnit),
		})
	}
	c.JSON(http.StatusOK, res)
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
		log.WithError(err).Error("failed to bind request params")
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID REQUEST"))
		return
	}
	hostedPage, _, err := billing.GetUpgradeChargebeeSubscriptionCheckoutURL(project.BillingSubscriptionID, updateSubscriptionParams)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	url := buildCheckoutUrl(projectId, hostedPage.Url)

	// redirect
	c.JSON(http.StatusOK, UR.QueryEscape(url))
	// c.Redirect(http.StatusPermanentRedirect, url)
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
	subscription.BillingPeriodUnit = string(res.BillingPeriodUnit)

	for _, item := range res.SubscriptionItems {
		subscriptionDetails = append(subscriptionDetails, model.SubscriptionDetail{
			Type:         string(item.ItemType),
			ID:           item.ItemPriceId,
			Amount:       item.Amount,
			ExternalName: getExternalNameFromPlanID(item.ItemPriceId),
		})
	}

	subscription.SubscriptionDetails = subscriptionDetails

	c.JSON(http.StatusOK, subscription)
}
func getExternalNameFromPlanID(planID string) string {
	arr := strings.Split(planID, "-")
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}
func buildCheckoutUrl(projectID int64, checkouturl string) string {
	callBackUrl := C.GetProtocol() + C.GetAPIDomain() + "/billing/upgarde/callback" + "&state=" + fmt.Sprint(projectID)
	return fmt.Sprintf("%s?redirect_url=%s", checkouturl, callBackUrl) // change this to const later
}

func BillingUpgradeCallbackHandler(c *gin.Context) {
	log.Info("redirect happened billing")
	state := c.Query("state") // TODO :change this to session cookie id
	projectID, err := strconv.ParseInt(state, 10, 64)
	if err != nil {
		c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing?error=BAD_REQUEST")
	}
	err = store.GetStore().TriggerSyncChargebeeToFactors(projectID)
	if err != nil {
		c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing?error=SERVER_ERROR")
		return
	}
	c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing")
}
