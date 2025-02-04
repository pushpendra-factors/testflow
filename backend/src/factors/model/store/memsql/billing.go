package memsql

import (
	"errors"
	billing "factors/billing/chargebee"
	"factors/model/model"
	"net/http"
	"strings"
	"time"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// TBD : Trigger this method from webhook or callback ?

func (store *MemSQL) TriggerSyncChargebeeToFactors(projectID int64) error { // Chargebee Recommends using webhooks for this instead of redirect url
	// get the project billing subscription id
	// get the latest subscription details

	// update the plan-price-id to project_plan_mapping table
	// update billing last synced at in project_plan_mappings table
	// update billing last synced at in projects table

	logCtx := log.Fields{"project_id": projectID}
	project, errCode := store.GetProject(projectID)
	if errCode != http.StatusFound {
		log.WithFields(logCtx).Error("Failed to get project")
		return errors.New("Failed to get project ")
	}

	subscriptionID := project.BillingSubscriptionID

	if subscriptionID == "" {
		log.WithFields(logCtx).Error("Subscription doesn't exist for this project ")
		return errors.New("Subscription doesn't exist for this project ")
	}

	latestSubscription, err := billing.GetCurrentSubscriptionDetails(projectID, subscriptionID)

	if err != nil {
		log.WithFields(logCtx).Error("Failed to get subscription details from chargebee ")
		return errors.New("Failed to get subscription details from chargebee ")
	}

	var planMapping model.ProjectPlanMapping
	var addOns model.BillingAddons

	for _, subscriptionItem := range latestSubscription.SubscriptionItems {
		if subscriptionItem.ItemType == "plan" {
			planName := strings.Split(subscriptionItem.ItemPriceId, "-")
			if len(planName) > 0 {
				updatedPlanID, err := model.GetPlanIDFromPlanName(strings.ToUpper(planName[0]))
				if err != nil {
					log.WithFields(logCtx).WithError(err).Error("Failed to update new plan id")
					return errors.New("Failed to update new plan id")
				}
				planMapping.PlanID = int64(updatedPlanID)
			}
			planMapping.BillingPlanID = subscriptionItem.ItemPriceId
			planMapping.BillingLastSyncedAt = time.Now()
		} else if subscriptionItem.ItemType == "addon" {
			log.Info("$$$$ updating billing addons")
			addOn := model.BillingAddOn{
				ItemPriceID: subscriptionItem.ItemPriceId,
				Quantity:    int(subscriptionItem.Quantity),
			}
			addOns = append(addOns, addOn)
		}
	}
	if len(addOns) != 0 {
		addOnsJson, err := U.EncodeStructTypeToPostgresJsonb(addOns)
		if err != nil {
			log.WithFields(logCtx).Error("Failed to encode addons to Json")
			return errors.New("Failed to encode addons to Json")
		}
		planMapping.BillingAddons = addOnsJson
	}

	status := store.UpdateProjectPlanMapping(projectID, &planMapping)
	if status != http.StatusOK {
		log.WithFields(logCtx).Error("Failed to update project plan mapping")
		return errors.New("Failed to update project plan mapping")
	}

	var tempProject model.Project
	tempProject.BillingSubscriptionID = latestSubscription.Id
	tempProject.BillingLastSyncedAt = time.Now()

	status = store.UpdateProject(projectID, &tempProject)
	if status != 0 {
		log.WithFields(logCtx).Error("Failed to update project plan mapping")
		return errors.New("Failed to update project plan mapping")
	}

	// Call Action on Plan Change
	features, _, err := store.GetPlanDetailsAndAddonsForProject(projectID)
	if err != nil {
		log.WithFields(logCtx).WithError(err)
	} else if err = store.OnFeatureEnableOrDisableHook(projectID, features); err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to update configs on plan update")
	}

	return nil
}
