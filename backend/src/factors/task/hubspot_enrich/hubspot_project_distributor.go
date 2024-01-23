package hubspot_enrich

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"net/http"
	"time"

	IntHubspot "factors/integration/hubspot"

	log "github.com/sirupsen/logrus"
)

func getAllCRMSettingsAsMap() (map[int64]model.CRMSetting, error) {
	crmSettings, status := store.GetStore().GetAllCRMSetting()
	if status != http.StatusFound && status != http.StatusNotFound {
		log.Error("Failed to get all crm settings as map")
		return nil, errors.New("failed to get crm settings as map")
	}

	crmSettingsMap := make(map[int64]model.CRMSetting, 0)
	for i := range crmSettings {
		crmSettingsMap[crmSettings[i].ProjectID] = crmSettings[i]
	}

	return crmSettingsMap, nil
}

func getAllProjectsDocumentCount(allHubspotProjects []model.HubspotProjectSettings, projectdocumentCount []model.HubspotDocumentCount) []model.HubspotDocumentCount {
	// Add projects which have zero counts
	emptyProjectsDocumentCount := make([]model.HubspotDocumentCount, 0)
	for i := range allHubspotProjects {
		projectID := allHubspotProjects[i].ProjectId
		projectFound := false
		for j := range projectdocumentCount {
			if projectdocumentCount[j].ProjectID == projectID {
				projectFound = true
				break
			}
		}

		if !projectFound {
			emptyProjectsDocumentCount = append(emptyProjectsDocumentCount, model.HubspotDocumentCount{ProjectID: projectID, Count: 0})
		}
	}

	return append(projectdocumentCount, emptyProjectsDocumentCount...)
}

// RunHubspotProjectDistributer to be used only with light job
func RunHubspotProjectDistributer(configs map[string]interface{}) (map[string]interface{}, bool) {

	countThreshold := configs["light_projects_count_threshold"].(int)
	defaultHealthcheckPingID := configs["health_check_ping_id"].(string)
	hubspotMaxCreatedAt := configs["max_record_created_at"].(int64)
	overrideHealthcheckPingID := configs["override_healthcheck_ping_id"].(string)
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, overrideHealthcheckPingID)

	hubspotEnabledProjectSettings, errCode := store.GetStore().GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Error("No projects enabled hubspot integration.")
		return nil, false
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

	projectIDs := make([]int64, 0)
	for i := range hubspotEnabledProjectSettings {
		projectIDs = append(projectIDs, hubspotEnabledProjectSettings[i].ProjectId)

	}

	projectdocumentCount, status := store.GetStore().GetHubspotDocumentCountForSync(projectIDs, IntHubspot.GetHubspotObjectTypeForSync(), time.Now().Unix())
	if status != http.StatusFound {
		log.Error("Failed to get hubspot document counts for project.")
		return nil, false
	}

	projectdocumentCount = getAllProjectsDocumentCount(hubspotEnabledProjectSettings, projectdocumentCount)

	crmSettingsMap, err := getAllCRMSettingsAsMap()
	if err != nil {
		log.WithError(err).Error("Failed to get all crm settings for hubspot project distributer.")
		return nil, false
	}

	newHeavyProjects := make([]int64, 0)
	newLightProjects := make([]int64, 0)
	for i := range projectdocumentCount {
		projectID := projectdocumentCount[i].ProjectID
		count := projectdocumentCount[i].Count
		// ignore projects which are already under heavy processing. These project will be marked as false once it gets completed under enrich heavy
		if isEnrichHeavyProject(projectID, crmSettingsMap) {
			continue
		}

		if count > countThreshold {
			newHeavyProjects = append(newHeavyProjects, projectID)
		} else {
			newLightProjects = append(newLightProjects, projectID)
		}
	}

	anyFailure := false
	for heavy, projects := range map[bool][]int64{true: newHeavyProjects, false: newLightProjects} {
		for i := range projects {
			if heavy {
				status := store.GetStore().CreateOrUpdateCRMSettingHubspotEnrich(projects[i], true, &hubspotMaxCreatedAt)
				if status != http.StatusAccepted && status != http.StatusCreated {
					log.WithFields(log.Fields{"project_id": projects[i]}).Error("Failed to update crm settings for hubspot project distributer.")
					anyFailure = true
				}
				continue
			}

			status := store.GetStore().CreateOrUpdateCRMSettingHubspotEnrich(projects[i], false, nil)
			if status != http.StatusAccepted && status != http.StatusCreated {
				log.WithFields(log.Fields{"project_id": projects[i]}).Error("Failed to update crm settings for hubspot project distributer.")
				anyFailure = true
			}
		}
	}

	newDestribution := map[string]interface{}{
		"heavy_projects": newHeavyProjects,
		"light_projects": newLightProjects,
	}

	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, newDestribution)
		return nil, false
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, newDestribution)

	return newDestribution, true
}
