package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetPropertyOverridesByType(projectID int64, typeConstant int, entity int) (int, []string) {
	logFields := log.Fields{
		"project_id": projectID,
		"type":       typeConstant,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	result := make([]string, 0)
	if projectID == 0 || typeConstant == 0 {
		return http.StatusBadRequest, result
	}

	db := C.GetServices().Db

	propertyDetail := &model.PropertyOverride{
		ProjectID:    projectID,
		OverrideType: typeConstant,
		Entity:       entity,
	}

	var propertyOverrides []model.PropertyOverride

	if err := db.Where(propertyDetail).Find(&propertyOverrides).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectID, "type": typeConstant}).WithError(err).Error(
			"Failed to GetPropertyOverridesFromDB.")
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusOK, result
		}
		return http.StatusInternalServerError, result
	}

	for _, prop := range propertyOverrides {
		result = append(result, prop.PropertyName)
	}

	return http.StatusOK, result
}
