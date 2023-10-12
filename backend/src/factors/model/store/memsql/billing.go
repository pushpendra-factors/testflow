package memsql

import (
	"errors"
	"factors/chargebee"
	"factors/model/model"
	"net/http"
	"time"

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

	latestSubscription, err := chargebee.GetCurrentSubscriptionDetails(subscriptionID)

	if err != nil {
		log.WithFields(logCtx).Error("Failed to get subscription details from chargebee ")
		return errors.New("Failed to get subscription details from chargebee ")
	}

	var planMapping model.ProjectPlanMapping

	for _, subscriptionItem := range latestSubscription.SubscriptionItems {
		if subscriptionItem.ItemType == "plan" {
			planMapping.BillingPlanID = subscriptionItem.ItemPriceId
			planMapping.BillingLastSyncedAt = time.Now()
		}
	}

	status := store.UpdateProjectPlanMapping(projectID, &planMapping)
	if status != http.StatusOK {
		log.WithFields(logCtx).Error("Failed to update project plan mapping")
		return errors.New("Failed to update project plan mapping")
	}

	var tempProject model.Project
	tempProject.BillingLastSyncedAt = time.Now()

	status = store.UpdateProject(projectID, &tempProject)
	if status != 0 {
		log.WithFields(logCtx).Error("Failed to update project plan mapping")
		return errors.New("Failed to update project plan mapping")
	}

	return nil
}
