package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetLeadSquaredMarker(ProjectID int64, Delta int64, Document string, Tag string) (int, int, bool) {
	db := C.GetServices().Db
	var leadsquaredMarker model.LeadsquaredMarker
	if err := db.Where("project_id = ? AND delta = ? AND document = ? AND tag = ?", ProjectID, Delta, Document, Tag).Find(&leadsquaredMarker).Error; err != nil {
		return 0, 0, false
	}
	return leadsquaredMarker.IndexNumber, leadsquaredMarker.NoOfRetries, leadsquaredMarker.IsDone
}

func (store *MemSQL) CreateLeadSquaredMarker(marker model.LeadsquaredMarker) int {
	db := C.GetServices().Db
	index, noOfRetries, _ := store.GetLeadSquaredMarker(marker.ProjectID, marker.Delta, marker.Document, marker.Tag)
	if index == 0 && noOfRetries == 0 {
		marker.CreatedAt = gorm.NowFunc()
		marker.UpdatedAt = gorm.NowFunc()
		marker.NoOfRetries = 1
		if err := db.Create(&marker).Error; err != nil {
			log.WithError(err).Error("Failed Updating leadsquared marker")
			return http.StatusInternalServerError
		}
	} else {
		updatedFields := map[string]interface{}{
			"index_number":  marker.IndexNumber,
			"no_of_retries": noOfRetries + 1,
			"is_done": 		 marker.IsDone,
			"updated_at":    gorm.NowFunc(),
		}
		dbErr := db.Model(&model.LeadsquaredMarker{}).Where("project_id = ? AND delta = ? AND document = ? AND tag = ?", marker.ProjectID, marker.Delta, marker.Document, marker.Tag).Update(updatedFields).Error

		if dbErr != nil {
			log.WithError(dbErr).Error("updating leadsquared marker failed")
			return http.StatusInternalServerError
		}
	}
	return http.StatusOK
}
