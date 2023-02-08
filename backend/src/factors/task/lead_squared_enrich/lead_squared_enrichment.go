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
	recordProcessLimit := config["record_process_limit"].(int)

	sourceObjectTypeAndAlias := model.GetLeadSquaredTypeToAliasMap(model.LeadSquaredDocumentTypeAlias)

	userTypes := map[int]bool{
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_LEAD]: true,
	}

	activityTypes := map[int]bool{
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_SALES_ACTIVITY]:                                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SENT]:                                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_INFO]:                                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_HAD_A_CALL]:                                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_CALLED_TO_COLLECT_REFERRAL]:                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_BOUNCED]:                                     true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_LINK_CLICKED]:                                true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED]:             true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_MARKED_SPAM]:                                 true,
		model.LeadSquaredDocumentTypeAlias[model.LEASQUARED_EMAIL_NEGATIVE_RESPONSE]:                            true,
		model.LeadSquaredDocumentTypeAlias[model.LEASQUARED_EMAIL_NEUTRAL_RESPONSE]:                             true,
		model.LeadSquaredDocumentTypeAlias[model.LEASQUARED_EMAIL_POSITIVE_RESPONSE]:                            true,
		model.LeadSquaredDocumentTypeAlias[model.LEASQUARED_EMAIL_OPENED]:                                       true,
		model.LeadSquaredDocumentTypeAlias[model.LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL]:                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEASQUARED_EMAIL_RESUBSCRIBED]:                                 true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS]:                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION]:   true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION]:                  true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY]:                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST]:                          true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION]:                  true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION]: true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER]:                  true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION]:                true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY]:                      true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST]:                        true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED]:                    true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_UNSUBSCRIBED]:                                true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED]:                true,
		model.LeadSquaredDocumentTypeAlias[model.LEADSQUARED_EMAIL_RECEIVED]:                                    true,
	}

	sourceConfig, err := enrichment.NewCRMEnrichmentConfig(U.CRM_SOURCE_NAME_LEADSQUARED, sourceObjectTypeAndAlias, userTypes, nil, activityTypes, recordProcessLimit)
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
