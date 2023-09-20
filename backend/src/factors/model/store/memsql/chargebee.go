package memsql

import (
	// C "factors/config"
	"factors/model/model"
	"github.com/chargebee/chargebee-go/v3"
	customerAction "github.com/chargebee/chargebee-go/v3/actions/customer"
	hostedPageAction "github.com/chargebee/chargebee-go/v3/actions/hostedpage"
	subscriptionAction "github.com/chargebee/chargebee-go/v3/actions/subscription"
	"github.com/chargebee/chargebee-go/v3/models/customer"
	"github.com/chargebee/chargebee-go/v3/models/hostedpage"
	"github.com/chargebee/chargebee-go/v3/models/subscription"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (store *MemSQL) CreateChargebeeCustomer(agent model.Agent) (customer.Customer, int, error) {
	logCtx := log.Fields{"uuid": agent.UUID}
	// TODO : set api key in secrets and pull from config
	chargebee.Configure("{site_api_key}", "{site}")
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
func (store *MemSQL) CreateChargebeeSubscriptionForCustomer(customerID string, planPriceID string, billingCycles int32) (subscription.Subscription, int, error) {
	logCtx := log.Fields{"customer_id": customerID}
	chargebee.Configure("{site_api_key}", "{site}")
	res, err := subscriptionAction.CreateWithItems(customerID, &subscription.CreateWithItemsRequestParams{
		SubscriptionItems: []*subscription.CreateWithItemsSubscriptionItemParams{
			{
				ItemPriceId:   planPriceID,
				BillingCycles: &billingCycles,
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

func (store *MemSQL) GetUpgradeChargebeeSubscriptionCheckoutURL(subscriptionID string, planPriceID string) (hostedpage.HostedPage, int, error) {
	logCtx := log.Fields{"subscription_ID": subscriptionID}
	chargebee.Configure("{site_api_key}", "{site}")
	res, err := hostedPageAction.CheckoutExistingForItems(&hostedpage.CheckoutExistingForItemsRequestParams{
		Subscription: &hostedpage.CheckoutExistingForItemsSubscriptionParams{
			Id: "__test__KyVnGWS4EgP3HA",
		},
		SubscriptionItems: []*hostedpage.CheckoutExistingForItemsSubscriptionItemParams{
			{
				ItemPriceId: planPriceID,
			},
		},
	}).Request()
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to get checkout url for upgrade subscription on chargebee")
		return hostedpage.HostedPage{}, http.StatusInternalServerError, err
	}
	return *res.HostedPage, http.StatusAccepted, nil
}
