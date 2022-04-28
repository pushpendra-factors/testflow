package postgres

import (
	C "factors/config"
	"factors/model/model"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const weeklyInsights string = "weekly_insights"

func (pg *Postgres) PostFeedback(ProjectID uint64, agentUUID string, Feature string, Property *postgres.Jsonb, VoteType int) (int, string) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()

	var feedback model.Feedback

	feedback = model.Feedback{
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
func (pg *Postgres) GetRecordsFromFeedback(projectID uint64, agentUUID string) ([]model.Feedback, error) {
	db := C.GetServices().Db
	var records []model.Feedback
	if err := db.Where("project_id = ?", projectID).Where("created_by = (?) AND feature = (?)", agentUUID, weeklyInsights).Find(&records).Error; err != nil {
		log.Error(err)
		return nil, err
	}
	return records, nil
}
