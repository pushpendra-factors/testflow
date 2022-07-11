package hubspot_enrich

import (
	C "factors/config"
	IntHubspot "factors/integration/hubspot"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type SyncStatus struct {
	Status                  []IntHubspot.Status
	HasFailure              bool
	AnyProcessLimitExceeded bool
	Lock                    sync.Mutex
}

func isStatusProcessLimitExceeded(status []IntHubspot.Status) bool {
	for i := range status {
		if status[i].IsProcessLimitExceeded {
			return true
		}
	}

	return false
}

func (s *SyncStatus) AddSyncStatus(status []IntHubspot.Status, hasFailure bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
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

func syncWorker(projectID int64, wg *sync.WaitGroup, numDocRoutines int, syncStatus *SyncStatus, recordsMaxCreatedAt int64, hubspotProjectSettings *model.HubspotProjectSettings, recordsProcessLimit int) {
	defer wg.Done()
	datePropertiesByObjectType, err := IntHubspot.GetHubspotPropertiesByDataType(projectID, model.GetHubspotAllowedObjects(projectID), hubspotProjectSettings.APIKey, model.HubspotDataTypeDate)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed to get date properties.")

		syncStatus.AddSyncStatus([]IntHubspot.Status{{ProjectId: projectID, Message: err.Error(), Status: U.CRM_SYNC_STATUS_FAILURES}}, true)
		return
	}

	timeZone, err := IntHubspot.GetHubspotAccountTimezone(projectID, hubspotProjectSettings.APIKey)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed to get timezone for enrichment.")
		syncStatus.AddSyncStatus([]IntHubspot.Status{{ProjectId: projectID, Message: err.Error(), Status: U.CRM_SYNC_STATUS_FAILURES}}, true)
		return
	}

	status, hasFailure := IntHubspot.Sync(projectID, numDocRoutines, recordsMaxCreatedAt, datePropertiesByObjectType, timeZone, recordsProcessLimit)
	syncStatus.AddSyncStatus(status, hasFailure)

	// mark enrich_heavy as false irrespective of failure, let the distributer re distribute the project
	if !isStatusProcessLimitExceeded(status) {
		errCode := store.GetStore().UpdateCRMSetting(projectID, model.HubspotEnrichHeavy(false, nil))
		if errCode != http.StatusAccepted {
			log.WithFields(log.Fields{"project_id": projectID}).Error("Failed to mark hubspot project for crm settings")
		}
	}

}

func isEnrichHeavyProject(projectID int64, settings map[int64]model.CRMSetting) bool {
	if _, exists := settings[projectID]; !exists {
		return false
	}

	return settings[projectID].HubspotEnrichHeavy
}

/*
getProjectMaxCreatedAt returns the maximum created_at timestamp for a records in a project to be processed in the current job
Max created at should be limited so as to avoid future partial updates to get processed which can cause inconsistency is user properties state
For light projects maximum created_at will be the job start time, since they process all records in single job

For heavy projects maximum created_at is decided by project distributer and set to project distributer start time, since all records till that time has led it to heavy project.
Heavy job will process all records till created_at in multiple runs and then exit from heavy job.
*/
func getProjectMaxCreatedAt(projectID int64, jobMaxCreatedAt int64, settings map[int64]model.CRMSetting) int64 {
	if !isEnrichHeavyProject(projectID, settings) {
		return jobMaxCreatedAt
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
	recordsProcessLimit := configs["record_process_limit_per_project"].(int)
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, overrideHealthcheckPingID)

	hubspotEnabledProjectSettings, errCode := store.GetStore().GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Panic("No projects enabled hubspot integration.")
	}

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

	projectIDs := make([]int64, 0, 0)
	projectsMaxCreatedAt := make(map[int64]int64)
	hubspotProjectSettingsMap := make(map[int64]*model.HubspotProjectSettings, 0)
	for _, settings := range hubspotEnabledProjectSettings {
		if exists := disabledProjects[settings.ProjectId]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[settings.ProjectId]; !exists {
				continue
			}
		}

		if (enrichHeavy && !isEnrichHeavyProject(settings.ProjectId, crmSettingsMap)) ||
			(!enrichHeavy && isEnrichHeavyProject(settings.ProjectId, crmSettingsMap)) {
			continue
		}

		projectsMaxCreatedAt[settings.ProjectId] = getProjectMaxCreatedAt(settings.ProjectId, hubspotMaxCreatedAt, crmSettingsMap)

		if C.IsEnabledPropertyDetailByProjectID(settings.ProjectId) {
			log.Info(fmt.Sprintf("Starting sync property details for project %d", settings.ProjectId))

			failure, propertyDetailStatus := IntHubspot.SyncDatetimeAndNumericalProperties(settings.ProjectId, settings.APIKey)
			propertyDetailSyncStatus = append(propertyDetailSyncStatus, propertyDetailStatus...)
			if failure {
				anyFailure = true
			}

			log.Info(fmt.Sprintf("Synced property details for project %d", settings.ProjectId))
		}

		projectIDs = append(projectIDs, settings.ProjectId)
		hubspotProjectSettingsMap[settings.ProjectId] = &settings
	}

	// Runs enrichment for list of project_ids as batch using go routines.
	batches := U.GetInt64ListAsBatch(projectIDs, numProjectRoutines)
	log.WithFields(log.Fields{"project_batches": batches}).Info("Running for batches.")
	syncStatus := SyncStatus{}
	for bi := range batches {
		batch := batches[bi]

		var wg sync.WaitGroup
		for pi := range batch {
			wg.Add(1)
			go syncWorker(batch[pi], &wg, numDocRoutines, &syncStatus, projectsMaxCreatedAt[batch[pi]], hubspotProjectSettingsMap[batch[pi]], recordsProcessLimit)
		}
		wg.Wait()
	}
	anyFailure = anyFailure || syncStatus.HasFailure

	jobStatus = map[string]interface{}{
		"document_sync":      syncStatus.Status,
		"property_type_sync": propertyDetailSyncStatus,
	}
	panicError = false

	// For enrichment heavy, if any project limit is exceeded then return false to re-run the task on a new pod
	if enrichHeavy && syncStatus.AnyProcessLimitExceeded == true {
		return jobStatus, false
	}

	return jobStatus, true
}
