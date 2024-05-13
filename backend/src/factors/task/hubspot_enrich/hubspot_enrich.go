package hubspot_enrich

import (
	C "factors/config"
	IntHubspot "factors/integration/hubspot"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type SyncStatus struct {
	Status                  []IntHubspot.Status
	HasFailure              bool
	AnyProcessLimitExceeded bool
	Lock                    sync.Mutex
}

type WorkerStatus struct {
	Status     []IntHubspot.Status
	HasFailure bool
	ProjectId  int64
}

func isStatusProcessLimitExceeded(status []IntHubspot.Status) bool {
	for i := range status {
		if status[i].IsProcessLimitExceeded {
			return true
		}
	}

	return false
}
func Producer(hubspotProjectSettingsMap map[int64]*model.HubspotProjectSettings, projectSettingsChannel chan model.HubspotProjectSettings) {
	for _, settings := range hubspotProjectSettingsMap {
		projectSettingsChannel <- *settings
	}
	close(projectSettingsChannel)
}

func AddSyncStatus(s *SyncStatus, status []IntHubspot.Status, hasFailure bool) {
	processLimitExceeded := false
	if isStatusProcessLimitExceeded(status) {
		processLimitExceeded = true
	}

	for i := range status {
		if !processLimitExceeded {
			s.Status = append(s.Status, status[i])
			continue
		}

		if s.AnyProcessLimitExceeded == false {
			s.AnyProcessLimitExceeded = true
		}

		status[i].Message = "Process limit exceeded"
		s.Status = append([]IntHubspot.Status{status[i]}, s.Status...)
	}

	if hasFailure {
		s.HasFailure = hasFailure
	}
}

func StartEnrichmentByProjectIdWorker(projectSettingsChannel chan model.HubspotProjectSettings, syncStatusChannel chan WorkerStatus,
	numDocRoutines int, projectsMaxCreatedAt map[int64]int64, recordsProcessLimit, pullLimit int) {
	for projectSettings := range projectSettingsChannel {

		recordsMaxCreatedAt := projectsMaxCreatedAt[projectSettings.ProjectId]
		var workerStatus WorkerStatus

		if projectSettings.APIKey != "" || projectSettings.RefreshToken != "" {
			datePropertiesByObjectType, err := IntHubspot.GetHubspotPropertiesByDataType(projectSettings.ProjectId, model.GetHubspotAllowedObjects(projectSettings.ProjectId),
				projectSettings.APIKey, projectSettings.RefreshToken, model.HubspotDataTypeDate)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectSettings.ProjectId}).WithError(err).Error("Failed to get date properties.")
				workerStatus.Status = []IntHubspot.Status{{ProjectId: projectSettings.ProjectId, Message: model.CLIENT_TOKEN_EXPIRED, Status: U.CRM_SYNC_STATUS_FAILURES}}
				workerStatus.HasFailure = true
				workerStatus.ProjectId = projectSettings.ProjectId
				syncStatusChannel <- workerStatus
				continue
			}

			timeZone, portalID, err := model.GetHubspotAccountTimezoneAndPortalID(projectSettings.ProjectId, projectSettings.APIKey, projectSettings.RefreshToken, C.GetHubspotAppID(), C.GetHubspotAppSecret())
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectSettings.ProjectId}).WithError(err).Error("Failed to get timezone for enrichment.")
				workerStatus.Status = []IntHubspot.Status{{ProjectId: projectSettings.ProjectId, Message: err.Error(), Status: U.CRM_SYNC_STATUS_FAILURES}}
				workerStatus.HasFailure = true
				workerStatus.ProjectId = projectSettings.ProjectId
				syncStatusChannel <- workerStatus
				continue
			}
			status, hasFailure := IntHubspot.Sync(projectSettings.ProjectId, numDocRoutines, recordsMaxCreatedAt, datePropertiesByObjectType, timeZone, recordsProcessLimit, pullLimit, portalID)
			workerStatus.Status = status
			workerStatus.HasFailure = hasFailure
			workerStatus.ProjectId = projectSettings.ProjectId
			syncStatusChannel <- workerStatus
		} else {
			status, hasFailure := IntHubspot.Sync(projectSettings.ProjectId, numDocRoutines, recordsMaxCreatedAt, nil, "", recordsProcessLimit, pullLimit, "")
			workerStatus.Status = status
			workerStatus.HasFailure = hasFailure
			workerStatus.ProjectId = projectSettings.ProjectId
			syncStatusChannel <- workerStatus
		}
	}

}

func StartEnrichment(numProjectRoutines int, projectsMaxCreatedAt map[int64]int64, hubspotProjectSettingsMap map[int64]*model.HubspotProjectSettings,
	numDocRoutines int, recordsProcessLimit int, enrichPullLimit int) SyncStatus {

	syncStatusChannel := make(chan WorkerStatus)
	projectSettingsChannel := make(chan model.HubspotProjectSettings)
	overallSyncStatus := SyncStatus{}

	go Producer(hubspotProjectSettingsMap, projectSettingsChannel)

	for i := 0; i < numProjectRoutines; i++ {
		go StartEnrichmentByProjectIdWorker(projectSettingsChannel, syncStatusChannel, numDocRoutines, projectsMaxCreatedAt, recordsProcessLimit, enrichPullLimit)
	}

	for i := 0; i < len(hubspotProjectSettingsMap); i++ {
		workerStatus := <-syncStatusChannel
		syncStatus := workerStatus.Status
		anyFailure := workerStatus.HasFailure
		AddSyncStatus(&overallSyncStatus, syncStatus, anyFailure)

		// mark enrich_heavy as false irrespective of failure, let the distributer re distribute the project
		if !isStatusProcessLimitExceeded(workerStatus.Status) {
			errCode := store.GetStore().UpdateCRMSetting(workerStatus.ProjectId, model.HubspotEnrichHeavy(false, nil))
			if errCode != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": workerStatus.ProjectId}).Error("Failed to mark hubspot project for crm settings")
			}

			errCode = store.GetStore().UpdateCRMSetting(workerStatus.ProjectId, model.HubspotFirstTimeEnrich(false))
			if errCode != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": workerStatus.ProjectId}).Error("Failed to remove hubspot project from first time enrich")
			}
		}
	}

	close(syncStatusChannel)
	return overallSyncStatus
}

func isEnrichHeavyProject(projectID int64, settings map[int64]model.CRMSetting) bool {
	if _, exists := settings[projectID]; !exists {
		return false
	}

	return settings[projectID].HubspotEnrichHeavy
}

func isFirsTimeProject(projectID int64, settings map[int64]model.CRMSetting) bool {
	return settings[projectID].HubspotFirstTimeEnrich
}

/*
getProjectMaxCreatedAt returns the maximum created_at timestamp for a records in a project to be processed in the current job
Max created at should be limited so as to avoid future partial updates to get processed which can cause inconsistency is user properties state
For light projects maximum created_at will be the job start time, since they process all records in single job

For heavy projects maximum created_at is decided by project distributer and set to project distributer start time, since all records till that time has led it to heavy project.
Heavy job will process all records till created_at in multiple runs and then exit from heavy job.

For first time enrich minimun created_at + 1 will used as limit
*/
func getProjectMaxCreatedAt(projectID int64, jobMaxCreatedAt int64, settings map[int64]model.CRMSetting) int64 {
	if !isEnrichHeavyProject(projectID, settings) && !isFirsTimeProject(projectID, settings) {
		return jobMaxCreatedAt
	}

	if isFirsTimeProject(projectID, settings) {
		projectMinCreatedAt, status := store.GetStore().GetHubspotHubspotDocumentMinCreatedAt(projectID)
		if status != http.StatusFound && status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID}).
				Error("Failed to get min created_at for first time enrich.")
			return 0
		}
		return time.Unix(projectMinCreatedAt, 0).AddDate(0, 0, 1).Unix()
	}

	if settings[projectID].HubspotEnrichHeavyMaxCreatedAt == nil {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Found empty max created at on hubspot enrich heavy.")
		return 0
	}

	return *settings[projectID].HubspotEnrichHeavyMaxCreatedAt

}

// RunHubspotEnrich task management compatible function for hubspot enrichment job
func RunHubspotEnrich(configs map[string]interface{}) (map[string]interface{}, bool) {

	projectIDList := configs["project_ids"].(string)
	disabledProjectIDList := configs["disabled_project_ids"].(string)
	numDocRoutines := configs["num_unique_doc_routines"].(int)
	defaultHealthcheckPingID := configs["health_check_ping_id"].(string)
	overrideHealthcheckPingID := configs["override_healthcheck_ping_id"].(string)
	numProjectRoutines := configs["num_project_routines"].(int)
	hubspotMaxCreatedAt := configs["max_record_created_at"].(int64)
	enrichHeavy := configs["enrich_heavy"].(bool)
	firstTimeEnrich := configs["first_time_enrich"].(bool)
	recordsProcessLimit := configs["record_process_limit_per_project"].(int)
	enrichPullLimit := configs["enrich_pull_limit"].(int)
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, overrideHealthcheckPingID)

	hubspotEnabledProjectSettings, errCode := store.GetStore().GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Panic("No projects enabled hubspot integration.")
	}

	featureProjectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_HUBSPOT, false)
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot feature enabled projects.")
		return nil, false
	}

	featureEnabledProjectSettings := []model.HubspotProjectSettings{}
	for i := range hubspotEnabledProjectSettings {
		if util.ContainsInt64InArray(featureProjectIDs, hubspotEnabledProjectSettings[i].ProjectId) {
			featureEnabledProjectSettings = append(featureEnabledProjectSettings, hubspotEnabledProjectSettings[i])
		}
	}
	hubspotEnabledProjectSettings = featureEnabledProjectSettings

	var propertyDetailSyncStatus []IntHubspot.Status
	anyFailure := false
	panicError := true
	jobStatus := make(map[string]interface{})
	defer func() {
		if panicError || anyFailure {
			C.PingHealthcheckForFailure(healthcheckPingID, jobStatus)
		} else {
			C.PingHealthcheckForSuccess(healthcheckPingID, jobStatus)
		}
	}()

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		projectIDList, disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	if len(disabledProjects) > 0 {
		log.WithField("excluded_projects", disabledProjectIDList).Info("Running with exclusion of projects.")
	}

	crmSettingsMap, err := getAllCRMSettingsAsMap()
	if err != nil || len(crmSettingsMap) == 0 { // crmSettingsMap cannot be 0 since project distributer would have created entries
		log.WithError(err).Error("Failed to get all crm settings for hubspot enrich.")
		return nil, false
	}

	projectsMaxCreatedAt := make(map[int64]int64)
	hubspotProjectSettingsMap := make(map[int64]*model.HubspotProjectSettings, 0)
	for i, settings := range hubspotEnabledProjectSettings {
		if exists := disabledProjects[settings.ProjectId]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[settings.ProjectId]; !exists {
				continue
			}
		}

		if (enrichHeavy && !isEnrichHeavyProject(settings.ProjectId, crmSettingsMap)) ||
			(!enrichHeavy && isEnrichHeavyProject(settings.ProjectId, crmSettingsMap)) ||
			(firstTimeEnrich && !isFirsTimeProject(settings.ProjectId, crmSettingsMap)) ||
			(!firstTimeEnrich && isFirsTimeProject(settings.ProjectId, crmSettingsMap)) {
			continue
		}

		projectsMaxCreatedAt[settings.ProjectId] = getProjectMaxCreatedAt(settings.ProjectId, hubspotMaxCreatedAt, crmSettingsMap)

		if C.IsEnabledPropertyDetailByProjectID(settings.ProjectId) {
			log.Info(fmt.Sprintf("Starting sync property details for project %d", settings.ProjectId))

			failure, propertyDetailStatus := IntHubspot.SyncDatetimeAndNumericalProperties(settings.ProjectId, settings.APIKey, settings.RefreshToken)
			propertyDetailSyncStatus = append(propertyDetailSyncStatus, propertyDetailStatus...)
			if failure {
				anyFailure = true
			}

			log.Info(fmt.Sprintf("Synced property details for project %d", settings.ProjectId))
		}

		if C.AllowSyncReferenceFields(settings.ProjectId) {
			log.Info(fmt.Sprintf("Starting sync reference fields for project %d", settings.ProjectId))
			allowedDocTypes := U.GetKeysofStringMapAsArray(*model.GetHubspotAllowedObjects(settings.ProjectId))

			failure := IntHubspot.SyncReferenceField(settings.ProjectId, settings.APIKey, settings.RefreshToken, allowedDocTypes, projectsMaxCreatedAt[settings.ProjectId])
			if failure {
				anyFailure = true
			}

			log.Info(fmt.Sprintf("Synced reference fields for project %d", settings.ProjectId))
		}

		hubspotProjectSettingsMap[settings.ProjectId] = &hubspotEnabledProjectSettings[i]
	}

	// Runs enrichment for list of project_ids synchronously
	syncStatus := StartEnrichment(numProjectRoutines, projectsMaxCreatedAt, hubspotProjectSettingsMap, numDocRoutines, recordsProcessLimit, enrichPullLimit)

	anyFailure = anyFailure || syncStatus.HasFailure

	jobStatus = map[string]interface{}{
		"document_sync":      syncStatus.Status,
		"property_type_sync": propertyDetailSyncStatus,
	}
	panicError = false

	for _, state := range syncStatus.Status {
		if state.IsProcessLimitExceeded {
			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(state.ProjectId, model.HUBSPOT, model.HEAVY_DELAYED)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": state.ProjectId}).Warn("Failed to update integration status")

			}
		}

		if state.Message == model.CLIENT_TOKEN_EXPIRED {
			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(state.ProjectId, model.HUBSPOT, model.CLIENT_TOKEN_EXPIRED)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": state.ProjectId}).Warn("Failed to update integration status")

			}
		}

		if state.Message == model.SUCCESS {
			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(state.ProjectId, model.HUBSPOT, model.SUCCESS)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": state.ProjectId}).Warn("Failed to update integration status")

			}
		}
	}

	return jobStatus, true

}
