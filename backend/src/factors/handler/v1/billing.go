package v1

import (
	"encoding/json"
	"errors"
	billing "factors/billing/chargebee"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chargebee/chargebee-go/v3/models/event"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetPricingForPlansAndAddonsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
		return
	}
	itemPrices, err := billing.ListPlansAndAddOnsPricesFromChargebee()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.PlansAndAddOnsPrices{})
		return
	}
	var res model.PlansAndAddOnsPrices
	for _, itemPrice := range itemPrices {
		res = append(res, model.SubscriptionProductPrice{
			Type:         string(itemPrice.ItemType),
			Name:         itemPrice.Name,
			ExternalName: itemPrice.ExternalName,
			ID:           itemPrice.Id,
			Price:        formatPrice(itemPrice.Price),
			PeriodUnit:   string(itemPrice.PeriodUnit),
		})
	}
	c.JSON(http.StatusOK, res)
}

func formatPrice(price int64) float64 {
	return float64(price) / 100
}

func GetDifferentialPricingForAddOns(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
		return
	}
	diffPrices, err := billing.ListDifferentialPricingFromChargebee()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.DifferentialPrices{})
		return

	}
	var res model.DifferentialPrices
	for _, diffPrice := range diffPrices {
		res = append(res, model.DifferentialPrice{
			ID:           diffPrice.Id,
			ItemPriceID:  diffPrice.Id,
			ParentItemID: diffPrice.ParentItemId,
			Price:        formatPrice(diffPrice.Price),
		})
	}
	c.JSON(http.StatusOK, res)
}

func UpdateSubscriptionHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
		return
	}
	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithError(http.StatusBadRequest, errors.New("BAD REQUEST"))
		return
	}

	var updateSubscriptionParams model.UpdateSubscriptionParams
	err := c.BindJSON(&updateSubscriptionParams)
	if err != nil {
		log.WithError(err).Error("failed to bind request params")
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID REQUEST"))
		return
	}
	// increase the addons quantity (current + new one)
	hostedPage, _, err := billing.GetUpgradeChargebeeSubscriptionCheckoutURL(projectId, project.BillingSubscriptionID, updateSubscriptionParams)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	resp := make(map[string]string)
	resp["url"] = hostedPage.Url
	c.JSON(http.StatusOK, resp)

}

func GetSubscriptionDetailsHander(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
		return
	}
	var subscription model.Subscription
	var subscriptionDetails []model.SubscriptionDetail

	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.JSON(http.StatusBadRequest, subscription)
		return
	}

	res, err := billing.GetCurrentSubscriptionDetails(projectId, project.BillingSubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, subscription)
		return
	}

	subscription.Status = string(res.Status)
	subscription.RenewsOn = time.Unix(res.NextBillingAt, 0)
	subscription.BillingPeriodUnit = string(res.BillingPeriodUnit)

	for _, item := range res.SubscriptionItems {
		subscriptionDetails = append(subscriptionDetails, model.SubscriptionDetail{
			Type:         string(item.ItemType),
			ID:           item.ItemPriceId,
			Amount:       formatPrice(item.Amount),
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

	if state != "succeeded" {
		// do not sync if purchase was failed
		c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf(C.GetProtocol()+C.GetAPPDomain()+"/pricing&state=%s", state))
	}

	projectIDString := c.Query("project_id")
	projectID, err := strconv.ParseInt(projectIDString, 10, 64)
	if err != nil {
		c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing?error=INVALID PROJECT")
	}

	err = store.GetStore().TriggerSyncChargebeeToFactors(projectID)
	if err != nil {
		c.Redirect(http.StatusPermanentRedirect, C.GetProtocol()+C.GetAPPDomain()+"/pricing?error=SERVER_ERROR")
		return
	}
	c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf(C.GetProtocol()+C.GetAPPDomain()+"/settings/pricing?state=%s", state))
}

func ListAllInvoicesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
		return
	}

	var invoices []model.Invoice
	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.JSON(http.StatusBadRequest, invoices)
		return
	}

	invoicesTemp, err := billing.ListAllInvoicesForSubscription(project.BillingSubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, invoices)
		return
	}

	for _, invoice := range invoicesTemp {
		inv := model.Invoice{
			ID:          invoice.Id,
			BillingDate: time.Unix(invoice.Date, 0),
			Amount:      formatPrice(invoice.Total),
			AmountPaid:  formatPrice(invoice.AmountPaid),
			AmountDue:   formatPrice(invoice.AmountDue),
		}
		var items []string
		for _, item := range invoice.LineItems {
			items = append(items, item.Description)
		}
		inv.Items = items

		invoices = append(invoices, inv)
	}

	c.JSON(http.StatusOK, invoices)

}

func DownloadInvoiceHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID PROJECT ID"))
		return
	}
	invoiceID := c.Query("invoice_id")

	if invoiceID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID INVOICE ID"))
		return
	}

	var invoice model.DownloadInvoice

	iv, err := billing.DownloadInvoiceByInvoiceID(invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, invoice)
		return
	}

	invoice.Url = iv.DownloadUrl
	invoice.ValidTill = time.Unix(iv.ValidTill, 0)

	c.JSON(http.StatusOK, invoice)
}

func BillingSubscriptionChangedWebhookListner(c *gin.Context) {
	var event event.Event
	err := c.BindJSON(&event)

	if err != nil {
		log.Error("failed to read subscription changed event body")
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID JSON"))
		return
	}

	log.Info("billing subscription changed trigger body ", event)

	var content model.Content

	err = json.Unmarshal(event.Content, &content)
	if err != nil {
		log.Error("failed to parse content data")
		c.AbortWithError(http.StatusInternalServerError, errors.New("INTERNAL SERVER ERROR"))
		return
	}

	if content.Subscription.Id == "" {
		log.WithError(err).Error("Subscription ID not found")
		c.AbortWithError(http.StatusBadRequest, errors.New("BAD REQUEST"))
		return
	}

	billingSubscriptionID := content.Subscription.Id

	// get the project ID from subscription id
	projectID, status := store.GetStore().GetProjectIDByBillingSubscriptionID(billingSubscriptionID)
	if status != http.StatusFound {
		log.WithError(err).Error("Failed to get project from subscription id")
		c.AbortWithError(http.StatusInternalServerError, errors.New("INTERNAL SERVER ERROR"))
		return
	}

	// trigger sync
	err = store.GetStore().TriggerSyncChargebeeToFactors(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to sync chargbee data to internal db")
		c.AbortWithError(http.StatusInternalServerError, errors.New("INTERNAL SERVER ERROR"))
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func BillingInvoiceGeneratedWebhookListner(c *gin.Context) {
	var event event.Event
	err := c.BindJSON(&event)

	if err != nil {
		log.Error("failed to read invoice generated event body")
		c.AbortWithError(http.StatusBadRequest, errors.New("INVALID JSON"))
		return
	}

	log.Info("billing invoice generated trigger body ", event)

	var content model.Content

	err = json.Unmarshal(event.Content, &content)
	if err != nil {
		log.Error("failed to parse invoice generated content data")
		c.AbortWithError(http.StatusInternalServerError, errors.New("INTERNAL SERVER ERROR"))
		return
	}

	if content.Invoice.SubscriptionId == "" {
		log.WithError(err).Error("Invoice ID not found")
		c.AbortWithError(http.StatusBadRequest, errors.New("BAD REQUEST"))
		return
	}

	billingSubscriptionID := content.Invoice.SubscriptionId

	// get the project ID from subscription id
	projectID, status := store.GetStore().GetProjectIDByBillingSubscriptionID(billingSubscriptionID)
	if status != http.StatusFound {
		log.WithError(err).Error("Failed to get project from subscription id")
		c.AbortWithError(http.StatusInternalServerError, errors.New("INTERNAL SERVER ERROR"))
		return
	}

	// trigger sync
	err = store.GetStore().TriggerSyncChargebeeToFactors(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to sync chargbee data to internal db")
		c.AbortWithError(http.StatusInternalServerError, errors.New("INTERNAL SERVER ERROR"))
		return
	}
	c.JSON(http.StatusOK, "ok")
}
