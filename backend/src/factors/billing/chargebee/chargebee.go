package billing_plan_id

import (
	"factors/model/model"

	"github.com/chargebee/chargebee-go/v3"
	customerAction "github.com/chargebee/chargebee-go/v3/actions/customer"
	differentialPriceAction "github.com/chargebee/chargebee-go/v3/actions/differentialprice"
	hostedPageAction "github.com/chargebee/chargebee-go/v3/actions/hostedpage"
	invoiceAction "github.com/chargebee/chargebee-go/v3/actions/invoice"
	itemAction "github.com/chargebee/chargebee-go/v3/actions/item"
	itemPriceAction "github.com/chargebee/chargebee-go/v3/actions/itemprice"
	subscriptionAction "github.com/chargebee/chargebee-go/v3/actions/subscription"
	"github.com/chargebee/chargebee-go/v3/models/differentialprice"

	"net/http"

	C "factors/config"
	"fmt"
	"github.com/chargebee/chargebee-go/v3/filter"
	"github.com/chargebee/chargebee-go/v3/models/customer"
	"github.com/chargebee/chargebee-go/v3/models/download"
	"github.com/chargebee/chargebee-go/v3/models/hostedpage"
	"github.com/chargebee/chargebee-go/v3/models/invoice"
	"github.com/chargebee/chargebee-go/v3/models/item"
	"github.com/chargebee/chargebee-go/v3/models/itemprice"
	"github.com/chargebee/chargebee-go/v3/models/subscription"
	log "github.com/sirupsen/logrus"
)

func CreateChargebeeCustomer(agent model.Agent) (customer.Customer, int, error) {
	logCtx := log.Fields{"uuid": agent.UUID}
	// TODO : set api key in secrets
	res, err := customerAction.Create(&customer.CreateRequestParams{
		FirstName: agent.FirstName,
		LastName:  agent.LastName,
		Email:     agent.Email,
		Phone:     agent.Phone,
	}).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to create customer on chargebee")
		return customer.Customer{}, http.StatusInternalServerError, err
	}
	return *res.Customer, http.StatusCreated, nil
}

// only used to create free subscription which doesn't require a card
func CreateChargebeeSubscriptionForCustomer(projectID int64, customerID string, planPriceID string) (subscription.Subscription, int, error) {
	logCtx := log.Fields{"customer_id": customerID,
		"project_id": projectID,
	}

	res, err := subscriptionAction.CreateWithItems(customerID, &subscription.CreateWithItemsRequestParams{
		SubscriptionItems: []*subscription.CreateWithItemsSubscriptionItemParams{
			{
				ItemPriceId: planPriceID,
			},
		},
	}).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to create subscription on chargebee")
		return subscription.Subscription{}, http.StatusInternalServerError, err
	} else {
		return *res.Subscription, http.StatusCreated, nil
	}
}

func GetUpgradeChargebeeSubscriptionCheckoutURL(projectID int64, subscriptionID string, params model.UpdateSubscriptionParams) (hostedpage.HostedPage, int, error) {
	logCtx := log.Fields{"subscription_ID": subscriptionID}

	// manual logic
	var subscriptionItems []*hostedpage.CheckoutExistingForItemsSubscriptionItemParams
	if params.UpdatedPlanID != "" {
		subscriptionItems = append(subscriptionItems, &hostedpage.CheckoutExistingForItemsSubscriptionItemParams{
			ItemPriceId: params.UpdatedPlanID,
		})
	}

	for _, addOn := range params.Addons {
		subscriptionItems = append(subscriptionItems, &hostedpage.CheckoutExistingForItemsSubscriptionItemParams{
			ItemPriceId: addOn.AddOnID,
			Quantity:    &addOn.Quantity,
		})
	}
	replaceItems := true
	res, err := hostedPageAction.CheckoutExistingForItems(&hostedpage.CheckoutExistingForItemsRequestParams{
		Subscription: &hostedpage.CheckoutExistingForItemsSubscriptionParams{
			Id: subscriptionID,
		},
		SubscriptionItems: subscriptionItems,
		RedirectUrl:       GetRedirectUrl(projectID),
		ReplaceItemsList:  &replaceItems,
	}).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get checkout url for upgrade subscription on chargebee")
		return hostedpage.HostedPage{}, http.StatusInternalServerError, err
	}
	return *res.HostedPage, http.StatusAccepted, nil
}

// lists only plans without billing frequencies available
func ListPlansAndAddOnsFromChargebee() ([]item.Item, error) {
	items := []item.Item{}
	res, err := itemAction.List(&item.ListRequestParams{
		Limit: chargebee.Int32(10),
	}).ListRequest()
	if err != nil {
		log.WithError(err).Error("Failed to fetch plans and addons")
		return items, nil
	} else {
		for idx := 0; idx < len(res.List); idx++ {
			items = append(items, *res.List[idx].Item)
		}
	}
	return items, nil
}

// list plans and their prices according to billing frequencies with plan-price-id
func ListPlansAndAddOnsPricesFromChargebee() ([]itemprice.ItemPrice, error) {
	itemPrices := []itemprice.ItemPrice{}
	res, err := itemPriceAction.List(&itemprice.ListRequestParams{
		Limit: chargebee.Int32(50),
	}).ListRequest()
	if err != nil {
		log.WithError(err).Error("Failed to fetch items prices")
		return itemPrices, nil
	} else {
		for idx := 0; idx < len(res.List); idx++ {
			itemPrices = append(itemPrices, *res.List[idx].ItemPrice)
		}
	}
	return itemPrices, nil
}

func ListDifferentialPricingFromChargebee() ([]differentialprice.DifferentialPrice, error) {
	differentialPrices := []differentialprice.DifferentialPrice{}
	res, err := differentialPriceAction.List(&differentialprice.ListRequestParams{
		Limit: chargebee.Int32(50),
	}).ListRequest()
	if err != nil {
		log.WithError(err).Error("Failed to fetch differential items prices")
		return differentialPrices, nil
	} else {
		for idx := 0; idx < len(res.List); idx++ {
			differentialPrices = append(differentialPrices, *res.List[idx].DifferentialPrice)
		}
	}
	return differentialPrices, nil
}

func GetCurrentSubscriptionDetails(projectID int64, subscriptionID string) (subscription.Subscription, error) {
	logCtx := log.Fields{"subscription_ID": subscriptionID,
		"project_id": projectID,
	}

	res, err := subscriptionAction.Retrieve(subscriptionID).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get subscription details")
		return subscription.Subscription{}, err
	} else {
		return *res.Subscription, nil
	}
}

func GetItemDetailsFromItemPriceID(itemPriceID string) (itemprice.ItemPrice, error) {
	logCtx := log.Fields{"item_price_id": itemPriceID}

	res, err := itemPriceAction.Retrieve(itemPriceID).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get subscription details")
		return itemprice.ItemPrice{}, err
	} else {
		return *res.ItemPrice, nil
	}

}

func ListAllInvoicesForSubscription(subscriptionID string) ([]invoice.Invoice, error) {
	logCtx := log.Fields{"subscription_ID": subscriptionID}

	var invoices []invoice.Invoice
	res, err := invoiceAction.List(&invoice.ListRequestParams{
		Limit: chargebee.Int32(100),
		SubscriptionId: &filter.StringFilter{
			Is: subscriptionID,
		},
		// SortBy: date,
	}).ListRequest()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get invoices from chargebee")
		return invoices, err
	} else {
		for idx := 0; idx < len(res.List); idx++ {
			invoices = append(invoices, *res.List[idx].Invoice)
		}
	}
	return invoices, nil
}

func DownloadInvoiceByInvoiceID(invoiceID string) (download.Download, error) {
	logCtx := log.Fields{"invoice_id": invoiceID}
	res, err := invoiceAction.Pdf(invoiceID, nil).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to download invoice from chargebee")
	}
	return *res.Download, nil
}

func GetRedirectUrl(projectID int64) string {
	callBackUrl := C.GetProtocol() + C.GetAPIDomain() + fmt.Sprintf("/billing/upgrade/callback?project_id=%d", projectID)
	return callBackUrl
}
