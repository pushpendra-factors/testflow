package crm_enrichment

import (
	"errors"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"sync"
	"time"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

type CRMSourceConfig struct {
	projectID       int64
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
	TableNameCRMProperties    = "crm_properties"
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

func getCRMRecordByTypeForSync(projectID int64, source U.CRMSource, tableName string, startTimestamp, endTimestamp int64) (interface{}, int) {
	log.WithFields(log.Fields{"project_id": projectID, "start_time": startTimestamp, "end_time": endTimestamp, "table_name": tableName}).
		Info("Getting records for following range.")

	switch tableName {
	case TableNameCRMusers:
		return store.GetStore().GetCRMUsersInOrderForSync(projectID, source, startTimestamp, endTimestamp)
	case TableNameCRMActivities:
		return store.GetStore().GetCRMActivityInOrderForSync(projectID, source, startTimestamp, endTimestamp)
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
	TableName  string `json:"table_name"`
	ObjectType string `json:"type"`
	Status     string `json:"status"`
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

	if !model.IsCRMSource(sourceAlias) {
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

	// prefix check for crm properties for overwriting
	if !U.IsCRMPropertyKeyBySource(sourceAlias, model.GetCRMEnrichPropertyKeyByType(sourceAlias, " ", " ")) {
		return nil, errors.New("missing prefix check for source")
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

func getMinimumTimestampForSync(projectID int64, source U.CRMSource, minTimestampForSync int64) (int64, int) {
	if minTimestampForSync > 0 {
		return minTimestampForSync, http.StatusOK
	}

	userMinTimestamp, status := store.GetStore().GetCRMUsersMinimumTimestampForSync(projectID, source)
	if status != http.StatusFound && status != http.StatusNotFound {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Failed to get crm user min timstamp for enrichmnt.")
		return 0, http.StatusInternalServerError
	}

	if userMinTimestamp == 0 {
		userMinTimestamp = time.Now().Unix()
	}

	activityMinTimestamp, status := store.GetStore().GetCRMActivityMinimumTimestampForSync(projectID, source)
	if status != http.StatusFound && status != http.StatusNotFound {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Failed to get crm activity min timstamp for enrichmnt.")
		return 0, http.StatusInternalServerError
	}

	if activityMinTimestamp == 0 {
		activityMinTimestamp = time.Now().Unix()
	}

	return U.Min(userMinTimestamp, activityMinTimestamp), http.StatusOK
}

func GetBatchRecordInOrder(recordsInt interface{}, batchSize int) ([]map[string]interface{}, error) {
	switch records := recordsInt.(type) {
	case []model.CRMUser:
		return model.GetCRMUserBatchedOrderedRecordsByID(records, batchSize), nil
	case []model.CRMActivity:
		return model.GetCRMActivityBatchedOrderedRecordsByID(records, batchSize), nil
	default:
		return nil, errors.New("invalid records type")
	}
}

type enrichWorkerStatus struct {
	TypeStatus map[string]bool
	Lock       sync.Mutex
}

func enrichAllWorker(project *model.Project, wg *sync.WaitGroup, enrichStatus *enrichWorkerStatus,
	config *CRMSourceConfig, recordsInt interface{}) {
	defer wg.Done()
	defer log.WithFields(log.Fields{"project_id": project.ID}).Info("Completed processing records.")

	log.WithFields(log.Fields{"project_id": project.ID}).Info("Beginning processing records.")
	typeFailure, err := enrichByType(project, config, recordsInt)
	if err != nil {
		log.WithFields(log.Fields{"project_id": project.ID}).WithError(err).Error("Failed to begin process records.")
		typeFailure = map[string]bool{
			"Failed to begin process records": true,
		}
	}

	enrichStatus.Lock.Lock()
	defer enrichStatus.Lock.Unlock()
	for typ, failure := range typeFailure {
		if !failure && enrichStatus.TypeStatus[typ] != true {
			enrichStatus.TypeStatus[typ] = false
		} else {
			enrichStatus.TypeStatus[typ] = true
		}
	}
}

func getAllCRMEventNames(projectID int64, config *CRMSourceConfig) ([]string, error) {
	source, err := model.GetCRMSourceByAliasName(config.sourceAlias)
	if err != nil {
		return nil, err
	}

	crmUsersTypeAndAction, errCode := store.GetStore().GetCRMUsersTypeAndAction(projectID, source)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		return nil, errors.New("Failed to get crm users type and action.")
	}

	eventNames := make([]string, 0)

	if errCode == http.StatusFound {
		for _, crmUserTypeAndAction := range crmUsersTypeAndAction {
			userTypeAlias, err := config.GetCRMObjectTypeAlias(crmUserTypeAndAction.Type)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"project_id": projectID,
					"source":     config.sourceAlias,
					"type":       crmUserTypeAndAction.Type,
					"action":     crmUserTypeAndAction.Action,
				}).Error("Failed to get crm users type and action.")
				continue
			}

			eventNames = append(eventNames, GetCRMEventNameByAction(config.sourceAlias, userTypeAlias, crmUserTypeAndAction.Action))
		}
	} else {
		log.WithFields(log.Fields{"project_id": projectID, "source": config.sourceAlias}).Info("No crm users found. Skipped adding.")
	}

	crmActivityNames, errCode := store.GetStore().GetCRMActivityNames(projectID, source)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		return nil, errors.New("Failed to get crm activity name.")
	}

	if errCode == http.StatusFound {
		for _, name := range crmActivityNames {
			activityName := getActivityEventName(config.sourceAlias, name)
			eventNames = append(eventNames, activityName)
		}
	} else {
		log.WithFields(log.Fields{"project_id": projectID, "source": config.sourceAlias}).Info("No crm activities found. Skipped adding.")
	}

	return eventNames, nil
}

func CreateOrGetCRMEventNames(projectID int64, config *CRMSourceConfig) int {
	eventNames, err := getAllCRMEventNames(projectID, config)
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Failed to get crm event names from getAllCRMEventNames.")
		return http.StatusInternalServerError
	}

	//TODO: move to config based creation
	for _, eventName := range eventNames {
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			log.WithField("project_id", projectID).Error("Failed to create event names for crm.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func Enrich(projectID int64, sourceConfig *CRMSourceConfig, batchSize int, minTimestampForSync int64) []EnrichStatus {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "crm_source_config": sourceConfig})

	if sourceConfig == nil {
		logCtx.Error("Missing source config.")
		return nil
	}

	if projectID == 0 {
		logCtx.Error("Missing project_id.")
		return nil
	}

	status := CreateOrGetCRMEventNames(projectID, sourceConfig)
	if status != http.StatusOK {
		return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES}}
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
	minTimestamp, status := getMinimumTimestampForSync(projectID, source, minTimestampForSync)
	if status != http.StatusOK {
		logCtx.Error("Failed to get min timstamp for enrichmnt.")
		return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES}}
	}
	logCtx.WithFields(log.Fields{"min_timestamp": minTimestamp}).Info("Using minimum timestamp.")
	orderedTimeSeries := model.GetCRMTimeSeriesByStartTimestamp(projectID, minTimestamp, U.CRM_SOURCE_NAME_MARKETO)

	typeFailures := make(map[string]map[string]bool)
	for i := range orderedTimeSeries {
		for _, tableName := range enrichOrder {

			records, status := getCRMRecordByTypeForSync(sourceConfig.projectID, source, tableName, orderedTimeSeries[i][0], orderedTimeSeries[i][1])
			if status != http.StatusFound {
				if status == http.StatusNotFound {
					continue
				}
				logCtx.WithFields(log.Fields{"table_name": tableName}).Error("Failed to get crm records.")
				return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES, TableName: tableName}}
			}

			batches, err := GetBatchRecordInOrder(records, batchSize)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get batched records.")
				return []EnrichStatus{{Status: U.CRM_SYNC_STATUS_FAILURES, TableName: tableName}}
			}

			workerIndex := 0
			var enrichStatus enrichWorkerStatus
			enrichStatus.TypeStatus = make(map[string]bool)
			for i := range batches {
				batch := batches[i]
				var wg sync.WaitGroup
				for docID := range batch {
					logCtx.WithFields(log.Fields{"worker": workerIndex, "doc_id": docID, "table_name": tableName}).Info("Processing Batch by doc_id")
					wg.Add(1)
					go enrichAllWorker(project, &wg, &enrichStatus, sourceConfig, batch[docID])
					workerIndex++
				}
				wg.Wait()
			}

			for objectType, failure := range enrichStatus.TypeStatus {
				if _, exist := typeFailures[tableName]; !exist {
					typeFailures[tableName] = make(map[string]bool)
				}

				if failure == true && typeFailures[tableName][objectType] != true {
					typeFailures[tableName][objectType] = true
				} else {
					typeFailures[tableName][objectType] = false
				}
			}
		}

	}

	for tableName, typeFailure := range typeFailures {
		crmEnrichStatus := EnrichStatus{
			TableName: tableName,
		}
		for objectType, failure := range typeFailure {
			typeStatus := crmEnrichStatus
			typeStatus.ObjectType = objectType
			if failure {
				typeStatus.Status = U.CRM_SYNC_STATUS_FAILURES
			} else {
				typeStatus.Status = U.CRM_SYNC_STATUS_SUCCESS
			}
			projectStatus = append(projectStatus, typeStatus)
		}
	}

	return projectStatus
}
