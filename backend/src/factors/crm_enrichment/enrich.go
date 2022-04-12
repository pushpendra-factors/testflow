package crm_enrichment

import (
	"errors"
	"factors/model/model"
	"factors/model/store"
	"net/http"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

type CRMSourceConfig struct {
	projectID       uint64
	sourceAlias     string
	objectTypeAlias map[int]string
	userTypes       map[int]bool
	groupTypes      map[int]bool
	activityTypes   map[int]bool
}

const (
	TableNameCRMusers         = "crm_users"
	TableNameCRMGroups        = "crm_groups"
	TableNameCRMRelationships = "crm_relationships"
	TableNameCRMActivities    = "crm_activities"
)

var enrichOrder = []string{
	TableNameCRMusers,
	TableNameCRMActivities,
}

func (c *CRMSourceConfig) GetCRMObjectTypeAlias(objectType int) (string, error) {
	if c.objectTypeAlias[objectType] == "" {
		return "", errors.New("invalid object type")
	}

	return c.objectTypeAlias[objectType], nil
}

func getCRMRecordByTypeForSync(projectID uint64, source model.CRMSource, tableName string) (interface{}, int) {
	switch tableName {
	case TableNameCRMusers:
		return store.GetStore().GetCRMUsersInOrderForSync(projectID, source)
	case TableNameCRMActivities:
		return store.GetStore().GetCRMActivityInOrderForSync(projectID, source)
	}

	return nil, http.StatusBadRequest
}

func enrichByType(project *model.Project, config *CRMSourceConfig, recordsInt interface{}) (map[string]bool, error) {

	switch records := recordsInt.(type) {
	case []model.CRMUser:
		return enrichAllCRMUser(project, config, records), nil
	case []model.CRMActivity:
		return enrichAllCRMActivity(project, config, records), nil
	default:
		return nil, errors.New("invalid records type")
	}

}

type EnrichStatus struct {
	TableName  string
	ObjectType string
	Status     string
}

func NewCRMEnrichmentConfig(sourceAlias string, sourceObjectTypeAndAlias map[int]string,
	userTypes, groupTypes, activityTypes map[int]bool) (*CRMSourceConfig, error) {

	// for validating sdk request source
	if !model.IsValidUserSource(sourceAlias) {
		return nil, errors.New("invalid source on request source mapping")
	}

	_, err := model.GetCRMSourceByAliasName(sourceAlias)
	if err != nil {
		return nil, errors.New("invalid source")
	}

	if !model.IsAllowedCRMSourceForOverwrites(sourceAlias) {
		return nil, errors.New("source not allowed for properties overwrite")
	}

	// crm sources should not overwrite the properties update timestamp used by other source
	if !model.BlacklistUserPropertiesUpdateTimestampBySource[sourceAlias] {
		return nil, errors.New("properties update timestamp overwrite is allowed")
	}

	// get sdk user mapped request source
	source := model.UserSourceMap[sourceAlias]
	if source == 0 {
		return nil, errors.New("invalid source on request source mapping")
	}

	// check for prefix support, if not adde the prefix are escaped with '_'
	if !U.AllowedCRMPropertyPrefix[U.NAME_PREFIX+sourceAlias+"_"] {
		return nil, errors.New("source support not added in allowed property prefix")
	}

	// for user properties updated make sure object level property is defined for overwrite decision
	for userType := range userTypes {
		typeAlias := sourceObjectTypeAndAlias[userType]
		property := model.GetSourceUserPropertyOverwritePropertySuffix(sourceAlias, typeAlias)
		if property == "" {
			return nil, errors.New("missing overwrite user property key")
		}
	}

	sourceConfig := &CRMSourceConfig{
		sourceAlias:     sourceAlias,
		objectTypeAlias: sourceObjectTypeAndAlias,
		userTypes:       userTypes,
		groupTypes:      groupTypes,
		activityTypes:   activityTypes,
	}

	return sourceConfig, nil
}

func Enrich(projectID uint64, sourceConfig *CRMSourceConfig) []EnrichStatus {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "crm_source_config": sourceConfig})

	if sourceConfig == nil {
		logCtx.Error("Missing source config.")
		return nil
	}

	if projectID == 0 {
		logCtx.Error("Missing project_id.")
		return nil
	}

	sourceConfig.projectID = projectID
	source, err := model.GetCRMSourceByAliasName(sourceConfig.sourceAlias)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get source alias.")
		return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES}}
	}

	if !model.AllowedCRMBySource(source) {
		logCtx.WithError(err).Error("Invalid source.")
		return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES}}
	}

	project, status := store.GetStore().GetProject(sourceConfig.projectID)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			logCtx.WithError(err).Error("Invalid project_id.")
		} else {
			logCtx.WithError(err).Error("Failed to get project.")
		}

		return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES}}
	}

	projectStatus := make([]EnrichStatus, 0)
	for _, tableName := range enrichOrder {

		records, status := getCRMRecordByTypeForSync(sourceConfig.projectID, source, tableName)
		if status != http.StatusFound {
			if status == http.StatusNotFound {
				continue
			}
			logCtx.WithFields(log.Fields{"table_name": tableName}).Error("Failed to get crm records.")
			continue
		}

		typeFailure, err := enrichByType(project, sourceConfig, records)
		if err != nil {
			logCtx.WithFields(log.Fields{"table_name": tableName}).WithError(err).Error("Failed to begin processing records.")
			projectStatus = append(projectStatus, EnrichStatus{TableName: tableName, Status: U.CRM_SYNC_STATUS_FAILURES})
			continue
		}

		for objectType := range typeFailure {
			crmEnrichStatus := EnrichStatus{
				TableName:  tableName,
				ObjectType: objectType,
			}
			if typeFailure[objectType] == true {
				crmEnrichStatus.Status = U.CRM_SYNC_STATUS_FAILURES
			} else {
				crmEnrichStatus.Status = U.CRM_SYNC_STATUS_SUCCESS
			}
			projectStatus = append(projectStatus, crmEnrichStatus)
		}
	}
	return projectStatus
}
