package model

import (
	C "factors/config"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
)

type ProjectBillingAccountMapping struct {
	ProjectID        uint64 `gorm:"primary_key:true"`
	BillingAccountID uint64 `gorm:"primary_key:true"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func createProjectBillingAccountMapping(projectID, billingAccID uint64) (*ProjectBillingAccountMapping, int) {
	if projectID == 0 || billingAccID == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db

	pBAM := &ProjectBillingAccountMapping{ProjectID: projectID, BillingAccountID: billingAccID}
	if err := db.Create(&pBAM).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	return pBAM, http.StatusCreated
}

func GetProjectBillingAccountMappings(billingAccountID uint64) ([]ProjectBillingAccountMapping, int) {
	if billingAccountID == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	pBAMS := make([]ProjectBillingAccountMapping, 0, 0)
	if err := db.Where("billing_account_id = ?", billingAccountID).Find(&pBAMS).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(pBAMS) == 0 {
		return nil, http.StatusNotFound
	}

	return pBAMS, http.StatusFound
}

func GetProjectBillingAccountMapping(projectID uint64) (*ProjectBillingAccountMapping, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	pBAM := ProjectBillingAccountMapping{}
	if err := db.Where("project_id = ? ", projectID).Limit(1).Find(&pBAM).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &pBAM, http.StatusFound
}
