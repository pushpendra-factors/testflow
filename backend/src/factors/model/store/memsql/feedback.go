package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const weeklyInsights string = "weekly_insights"

func (store *MemSQL) PostFeedback(ProjectID int64, agentUUID string, Feature string, Property *postgres.Jsonb, VoteType int) (int, string) {
	logFields := log.Fields{
		"project_id": ProjectID,
		"agent_uuid": agentUUID,
		"feature":    Feature,
		"property":   Property,
		"vote_type":  VoteType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)
	transTime := gorm.NowFunc()

	var feedback model.Feedback

	feedback = model.Feedback{
		ID:        U.GetUUID(),
		ProjectID: ProjectID,
		Feature:   Feature,
		Property:  Property,
		VoteType:  VoteType,
		CreatedBy: agentUUID,
		CreatedAt: &transTime,
		UpdatedAt: &transTime,
	}

	if err := db.Create(&feedback).Error; err != nil {

		logCtx.WithError(err).Error("Insert into feedback table failed")
		return http.StatusInternalServerError, ""
	}
	return http.StatusCreated, ""
}
func (store *MemSQL) GetRecordsFromFeedback(projectID int64, agentUUID string) ([]model.Feedback, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var records []model.Feedback
	if err := db.Where("project_id = ?", projectID).Where("created_by = (?) AND feature = (?)", agentUUID, weeklyInsights).Find(&records).Error; err != nil {
		log.WithError(err).Error("Failure in GetRecordsFromFeedback")
		return nil, err
	}
	return records, nil
}
