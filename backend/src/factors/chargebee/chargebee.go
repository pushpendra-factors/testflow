package chargebee

import (
	C "factors/config"
	"factors/model/model"
	"github.com/chargebee/chargebee-go/v3"
	customerAction "github.com/chargebee/chargebee-go/v3/actions/customer"
	hostedPageAction "github.com/chargebee/chargebee-go/v3/actions/hostedpage"
	itemAction "github.com/chargebee/chargebee-go/v3/actions/item"
	itemPriceAction "github.com/chargebee/chargebee-go/v3/actions/itemprice"
	subscriptionAction "github.com/chargebee/chargebee-go/v3/actions/subscription"

	"github.com/chargebee/chargebee-go/v3/models/customer"
	"github.com/chargebee/chargebee-go/v3/models/hostedpage"
	"github.com/chargebee/chargebee-go/v3/models/item"
	"github.com/chargebee/chargebee-go/v3/models/itemprice"
	"github.com/chargebee/chargebee-go/v3/models/subscription"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func CreateChargebeeCustomer(agent model.Agent) (customer.Customer, int, error) {
	logCtx := log.Fields{"uuid": agent.UUID}
	// TODO : set api key in secrets
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())
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
func CreateChargebeeSubscriptionForCustomer(customerID string, planPriceID string) (subscription.Subscription, int, error) {
	logCtx := log.Fields{"customer_id": customerID}
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())
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

func GetUpgradeChargebeeSubscriptionCheckoutURL(subscriptionID string, params model.UpdateSubscriptionParams) (hostedpage.HostedPage, int, error) {
	logCtx := log.Fields{"subscription_ID": subscriptionID}
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())

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
	res, err := hostedPageAction.CheckoutExistingForItems(&hostedpage.CheckoutExistingForItemsRequestParams{
		Subscription: &hostedpage.CheckoutExistingForItemsSubscriptionParams{
			Id: subscriptionID,
		},
		SubscriptionItems: subscriptionItems,
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
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())
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
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())
	res, err := itemPriceAction.List(&itemprice.ListRequestParams{
		Limit: chargebee.Int32(10),
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

func GetCurrentSubscriptionDetails(subscriptionID string) (subscription.Subscription, error) {
	logCtx := log.Fields{"subscription_ID": subscriptionID}
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())
	res, err := subscriptionAction.Retrieve("__test__8asukSOXe0W3SU").Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get subscription details")
		return subscription.Subscription{}, err
	} else {
		return *res.Subscription, nil
	}
}

func GetItemDetailsFromItemPriceID(itemPriceID string) (itemprice.ItemPrice, error) {
	logCtx := log.Fields{"subscription_ID": itemPriceID}
	chargebee.Configure(C.GetChargebeeApiKey(), C.GetChargebeeSiteName())
	res, err := itemPriceAction.Retrieve("basic-USD-monthly").Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get subscription details")
		return itemprice.ItemPrice{}, err
	} else {
		return *res.ItemPrice, nil
	}

}

func SyncChargebeePostPurchaseAction(projectID int64) { // Chargebee Recommends using webhooks for this instead of redirect url
	// get the project billing subscription id
	// get the latest subscription details

	// update the plan-price-id to project_plan_mapping table
	// update billing last synced at in project_plan_mappings table
	// update billing last synced at in projects table
}
