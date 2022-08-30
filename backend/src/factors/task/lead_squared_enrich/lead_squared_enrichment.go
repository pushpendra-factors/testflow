package lead_squared_enrich

import (
	enrichment "factors/crm_enrichment"
	"factors/model/model"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

type EnrichStatus struct {
	PropertyEnrich []enrichment.EnrichStatus `json:"enrich_status"`
	Enrich         []enrichment.EnrichStatus `json:"enrich"`
}

func RunLeadSquaredEnrich(projectID int64, config map[string]interface{}) (map[string]interface{}, bool) {
	numDocRoutines := config["document_routines"].(int)
	minSyncTimestamp := config["min_sync_timestamp"].(int64)

	sourceObjectTypeAndAlias := model.GetLeadSquaredTypeToAliasMap(model.LeadSquaredDocumentTypeAlias)

	userTypes := map[int]bool{
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_LEAD]: true,
	}

	activityTypes := map[int]bool{
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_SALES_ACTIVITY]: true,
	}

	sourceConfig, err := enrichment.NewCRMEnrichmentConfig(U.CRM_SOURCE_NAME_LEADSQUARED, sourceObjectTypeAndAlias, userTypes, nil, activityTypes)
	if err != nil {
		log.WithError(err).Error("Failed to create new crm enrichment config for leadsquared.")
		return nil, false
	}

	commonEnrichStatus := U.CRM_SYNC_STATUS_SUCCESS
	propertyEnrichStatus := enrichment.SyncProperties(projectID, sourceConfig)
	for i := range propertyEnrichStatus {
		if propertyEnrichStatus[i].Status == U.CRM_SYNC_STATUS_FAILURES {
			commonEnrichStatus = U.CRM_SYNC_STATUS_FAILURES
		}
	}

	enrichStatus := enrichment.Enrich(projectID, sourceConfig, numDocRoutines, minSyncTimestamp)
	for i := range enrichStatus {
		if enrichStatus[i].Status == U.CRM_SYNC_STATUS_FAILURES {
			commonEnrichStatus = U.CRM_SYNC_STATUS_FAILURES
		}
	}
	projectEnrichStatus := map[string]interface{}{
		commonEnrichStatus: EnrichStatus{
			PropertyEnrich: propertyEnrichStatus,
			Enrich:         enrichStatus,
		},
	}

	return projectEnrichStatus, true
}
