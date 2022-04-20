package postgres

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"reflect"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const error_duplicateFactorsGoalName string = "pq: duplicate key value violates unique constraint \"name_projectid_unique_idx\""

func isDuplicateName(err error) bool {
	return err.Error() == error_duplicateFactorsGoalName
}

// GetAllFactorsGoals - get all the goals for a project
func (pg *Postgres) GetAllFactorsGoals(ProjectID uint64) ([]model.FactorsGoal, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var goals []model.FactorsGoal
	if err := db.Limit(1000).Where("project_id = ?", ProjectID).Find(&goals).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusFound
		}
		logCtx.WithError(err).Error("Get All FactorsGoals failed")
		return nil, http.StatusInternalServerError
	}
	return goals, http.StatusFound
}

// GetAllActiveFactorsGoals - get all the goals for a project
func (pg *Postgres) GetAllActiveFactorsGoals(ProjectID uint64) ([]model.FactorsGoal, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var goals []model.FactorsGoal
	if err := db.Limit(1000).Where("project_id = ?", ProjectID).Where("is_active = ?", true).Find(&goals).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusFound
		}
		logCtx.WithError(err).Error("Get All FactorsGoals failed")
		return nil, http.StatusInternalServerError
	}
	return goals, http.StatusFound
}

// CreateFactorsGoal - create a new goal
func (pg *Postgres) CreateFactorsGoal(ProjectID uint64, Name string, Rule model.FactorsGoalRule, agentUUID string) (int64, int, string) {
	if pg.isActiveFactorsGoalsLimitExceeded(ProjectID) {
		return 0, http.StatusBadRequest, "FactorsGoals count exceeded"
	}
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	insertType := U.UserCreated
	if agentUUID == "" {
		insertType = U.AutoTracked
	}
	transTime := gorm.NowFunc()
	if status, errMsg := pg.isRuleValid(Rule, ProjectID); status == false {
		return 0, http.StatusBadRequest, errMsg
	}
	ruleJSON, _ := json.Marshal(Rule)
	ruleJsonb := postgres.Jsonb{ruleJSON}
	if isDulplicateFactorsGoalRule(ProjectID, Rule) {
		return 0, http.StatusConflict, "Rule already exist"
	}
	var goal model.FactorsGoal
	if insertType == "UC" {
		goal = model.FactorsGoal{
			ProjectID:     ProjectID,
			Name:          Name,
			Rule:          ruleJsonb,
			Type:          insertType,
			CreatedBy:     &agentUUID,
			LastTrackedAt: new(time.Time),
			IsActive:      true,
			CreatedAt:     &transTime,
			UpdatedAt:     &transTime,
		}
	} else {
		goal = model.FactorsGoal{
			ProjectID:     ProjectID,
			Name:          Name,
			Rule:          ruleJsonb,
			Type:          insertType,
			LastTrackedAt: new(time.Time),
			IsActive:      true,
			CreatedAt:     &transTime,
			UpdatedAt:     &transTime,
		}
	}
	if err := db.Create(&goal).Error; err != nil {
		if isDuplicateName(err) {
			logCtx.WithError(err).Error("Duplicate name")
			return 0, http.StatusConflict, "duplicate name"
		}
		logCtx.WithError(err).Error("Insert into goals table failed")
		return 0, http.StatusInternalServerError, ""
	}
	return int64(goal.ID), http.StatusCreated, ""
}

func isDulplicateFactorsGoalRule(ProjectID uint64, Rule model.FactorsGoalRule) bool {
	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	ruleJSON, err := json.Marshal(Rule)
	if err != nil {
		logCtx.WithError(err).Error("FactorsGoal rule marshall failed")
		return true
	}
	ruleJsonb := postgres.Jsonb{ruleJSON}

	var goal model.FactorsGoal
	if err := db.Where("rule = ?", ruleJsonb).Where("project_id = ?", ProjectID).Where("is_active =?", true).Take(&goal).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
	}
	if goal.IsActive == true {
		return true
	}
	return false
}

func (pg *Postgres) isEventObjectValid(event string, eventFilters []model.KeyValueTuple, ProjectID uint64) (bool, string) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	eventData, err := pg.GetEventName(event, ProjectID)
	if err != http.StatusFound {
		logCtx.Error("Get Event details failed")
		return false, "event doesnt exist"
	}
	existingFactorsTrackedEvent, dbErr := pg.GetFactorsTrackedEvent(eventData.ID, ProjectID)
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("Events not in tracked list")
		return false, "event not tracked"
	}
	if existingFactorsTrackedEvent.IsActive != true {
		logCtx.WithError(dbErr).Error("event not active")
		return false, "tracked event not active"
	}
	eventProperties := make([]string, 0)
	for _, filter := range eventFilters {
		eventProperties = append(eventProperties, filter.Key)
	}
	res, msg := pg.isEventPropertiesValid(ProjectID, event, eventProperties)
	if res == false {
		logCtx.WithError(dbErr).Error(msg)
		return false, msg
	}
	return true, ""
}
func (pg *Postgres) isRuleValid(Rule model.FactorsGoalRule, ProjectID uint64) (bool, string) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	res, msg := pg.isEventObjectValid(Rule.EndEvent, Rule.Rule.EndEnEventFitler, ProjectID)
	if res == false {
		return res, msg
	}
	if Rule.StartEvent != "" {
		res, msg := pg.isEventObjectValid(Rule.StartEvent, Rule.Rule.StartEnEventFitler, ProjectID)
		if res == false {
			return res, msg
		}
	}
	userProperties := make([]string, 0)
	for _, filter := range Rule.Rule.GlobalFilters {
		userProperties = append(userProperties, filter.Key)
	}
	res, msg = pg.isUserPropertiesValid(ProjectID, userProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	userProperties = make([]string, 0)
	for _, filter := range Rule.Rule.StartEnUserFitler {
		userProperties = append(userProperties, filter.Key)
	}
	res, msg = pg.isUserPropertiesValid(ProjectID, userProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	userProperties = make([]string, 0)
	for _, filter := range Rule.Rule.EndEnUserFitler {
		userProperties = append(userProperties, filter.Key)
	}
	res, msg = pg.isUserPropertiesValid(ProjectID, userProperties)
	if res == false {
		logCtx.Error(msg)
		return false, msg
	}
	return true, ""
}

func (pg *Postgres) isUserPropertiesValid(ProjectID uint64, UserProperties []string) (bool, string) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	allUserPropertiesByCategory, err := pg.GetUserPropertiesByProject(ProjectID, 10000, 30)
	if err != nil {
		logCtx.WithError(err).Error("Get user Properties from cache failed")
		return false, "user proeprties missing"
	}
	userPropertiesMap := make(map[string]bool)
	for _, properties := range allUserPropertiesByCategory {
		for _, property := range properties {
			userPropertiesMap[property] = true
		}
	}
	for _, userProperty := range UserProperties {
		if userPropertiesMap[userProperty] == false {
			logCtx.Error("User Property not associated with project")
			return false, "user property not associated to this project"
		}
		exitingFactorsTrackedUserProperty, dbErr := pg.GetFactorsTrackedUserProperty(userProperty, ProjectID)
		if dbErr != nil {
			logCtx.Error("User Property not tracked")
			return false, "user property not tracked"
		}
		if exitingFactorsTrackedUserProperty.IsActive != true {
			logCtx.Error("User Property not active")
			return false, "user property not active"
		}
	}
	return true, ""
}

func (pg *Postgres) isEventPropertiesValid(ProjectID uint64, EventName string, EventProperties []string) (bool, string) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	allEventPropertiesByCategory, err := pg.GetPropertiesByEvent(ProjectID, EventName, 10000, 30)
	if err != nil {
		logCtx.WithError(err).Error("Get event Properties from cache failed")
		return false, "event proeprties missing"
	}
	eventPropertiesMap := make(map[string]bool)
	for _, properties := range allEventPropertiesByCategory {
		for _, property := range properties {
			eventPropertiesMap[property] = true
		}
	}
	for _, eventProperty := range EventProperties {
		if eventPropertiesMap[eventProperty] == false {
			logCtx.Error("event Property not associated with project")
			return false, "event property not associated to this project"
		}
	}
	return true, ""
}

// DeactivateFactorsGoal - Mark the existing goal as inactive
func (pg *Postgres) DeactivateFactorsGoal(ID int64, ProjectID uint64) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()
	existingFactorsGoal, dbErr := pg.GetFactorsGoalByID(ID, ProjectID)
	if dbErr == nil {
		if existingFactorsGoal.IsActive == true {
			updatedFields := map[string]interface{}{
				"is_active":  false,
				"updated_at": transTime,
			}
			return updateFactorsGoal(existingFactorsGoal.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Deactivate FactorsGoal failed")
		return 0, http.StatusFound

	}
	logCtx.WithError(dbErr).Error("FactorsGoal not found")
	return 0, http.StatusNotFound
}

// ActivateFactorsGoal - activating the already deactivated goal
func (pg *Postgres) ActivateFactorsGoal(Name string, ProjectID uint64) (int64, int) {
	if pg.isActiveFactorsGoalsLimitExceeded(ProjectID) {
		return 0, http.StatusBadRequest
	}
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()
	existingFactorsGoal, dbErr := pg.GetFactorsGoal(Name, ProjectID)
	if dbErr == nil {
		if existingFactorsGoal.IsActive == false {
			updatedFields := map[string]interface{}{
				"is_active":  true,
				"updated_at": transTime,
			}
			return updateFactorsGoal(existingFactorsGoal.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Activate FactorsGoal failed")
		return 0, http.StatusFound

	}
	logCtx.WithError(dbErr).Error("FactorsGoal not found")
	return 0, http.StatusNotFound
}

// UpdateFactorsGoal - Edit the existing goal's name/rule
func (pg *Postgres) UpdateFactorsGoal(ID int64, Name string, Rule model.FactorsGoalRule, ProjectID uint64) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()
	existingFactorsGoal, dbErr := pg.GetFactorsGoalByID(ID, ProjectID)
	if dbErr == nil {
		if existingFactorsGoal.IsActive == true {

			updatedFields := map[string]interface{}{
				"updated_at": transTime,
			}
			if Name != "" {
				updatedFields["name"] = Name
			}
			if !reflect.DeepEqual(Rule, model.FactorsGoalRule{}) {
				ruleJSON, err := json.Marshal(Rule)
				if err != nil {
					logCtx.WithError(err).Error("FactorsGoal rule marshall failed")
					return 0, http.StatusInternalServerError
				}
				ruleJsonb := postgres.Jsonb{ruleJSON}
				updatedFields["rule"] = ruleJsonb
			}
			return updateFactorsGoal(existingFactorsGoal.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Update FactorsGoal failed")
		return 0, http.StatusFound

	}
	logCtx.WithError(dbErr).Error("FactorsGoal not found")
	return 0, http.StatusNotFound
}

// GetFactorsGoal - Get pariticular goal's details by name
func (pg *Postgres) GetFactorsGoal(Name string, ProjectID uint64) (*model.FactorsGoal, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var goal model.FactorsGoal
	if err := db.Where("name = ?", Name).Where("project_id = ?", ProjectID).Take(&goal).Error; err != nil {
		logCtx.WithFields(log.Fields{"ProjectID": ProjectID}).WithError(err).Error(
			"Getting goal failed on GetFactorsGoal")
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &goal, nil
}

// GetFactorsGoalByID  - Get goal details by ID
func (pg *Postgres) GetFactorsGoalByID(ID int64, ProjectID uint64) (*model.FactorsGoal, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var goal model.FactorsGoal
	if err := db.Where("id = ?", ID).Where("project_id = ?", ProjectID).Take(&goal).Error; err != nil {
		logCtx.WithFields(log.Fields{"ProjectID": ProjectID}).WithError(err).Error(
			"Getting goal failed on GetFactorsGoal")
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &goal, nil
}

func updateFactorsGoal(FactorsGoalID uint64, ProjectID uint64, updatedFields map[string]interface{}) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	dbErr := db.Model(&model.FactorsGoal{}).Where("project_id = ? AND id = ?", ProjectID, FactorsGoalID).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating goal failed")
		return 0, http.StatusInternalServerError
	}
	return int64(FactorsGoalID), http.StatusOK
}

// GetAllFactorsGoalsWithNamePattern - get all the goals for a project matching a specific pattern in the name
func (pg *Postgres) GetAllFactorsGoalsWithNamePattern(ProjectID uint64, NamePattern string) ([]model.FactorsGoal, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	percentagePrefix := "%"
	NamePattern = fmt.Sprintf("%s%s%s", percentagePrefix, NamePattern, percentagePrefix)
	var goals []model.FactorsGoal
	if err := db.Limit(1000).Where("project_id = ?", ProjectID).Where("name LIKE ?", NamePattern).Find(&goals).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusFound
		}
		logCtx.WithError(err).Error("Get All FactorsGoals failed")
		return nil, http.StatusInternalServerError
	}
	return goals, http.StatusFound
}

func (pg *Postgres) isActiveFactorsGoalsLimitExceeded(ProjectID uint64) bool {
	goals, errCode := pg.GetAllActiveFactorsGoals(ProjectID)
	if errCode != http.StatusFound {
		return true
	}
	if len(goals) >= C.GetFactorsGoalsLimit() {
		return true
	}
	return false
}
