package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesPBAMForeignConstraints(pbam model.ProjectBillingAccountMapping) int {
	_, projectErrCode := store.GetProject(pbam.ProjectID)
	baExists := store.existsBillingAccountByID(pbam.BillingAccountID)
	if projectErrCode != http.StatusFound || !baExists {
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func (store *MemSQL) createProjectBillingAccountMapping(projectID uint64, billingAccID string) (*model.ProjectBillingAccountMapping, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "billing_account_id": billingAccID})

	if projectID == 0 || billingAccID == "" {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db

	pBAM := &model.ProjectBillingAccountMapping{ProjectID: projectID, BillingAccountID: billingAccID}
	if errCode := store.satisfiesPBAMForeignConstraints(*pBAM); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}
	if err := db.Create(&pBAM).Error; err != nil {
		logCtx.WithError(err).Error("Creating model.ProjectBillingAccountMapping failed")
		return nil, http.StatusInternalServerError
	}
	return pBAM, http.StatusCreated
}

func (store *MemSQL) GetProjectBillingAccountMappings(billingAccountID string) ([]model.ProjectBillingAccountMapping, int) {
	if billingAccountID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	pBAMS := make([]model.ProjectBillingAccountMapping, 0, 0)
	if err := db.Where("billing_account_id = ?", billingAccountID).Find(&pBAMS).Error; err != nil {
		log.WithField("biling_account_id", billingAccountID).WithError(err).Error("Getting ProjectBillingAccountMappings failed")
		return nil, http.StatusInternalServerError
	}

	if len(pBAMS) == 0 {
		return nil, http.StatusNotFound
	}

	return pBAMS, http.StatusFound
}

func (store *MemSQL) GetProjectBillingAccountMapping(projectID uint64) (*model.ProjectBillingAccountMapping, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	pBAM := model.ProjectBillingAccountMapping{}
	if err := db.Where("project_id = ? ", projectID).Limit(1).Find(&pBAM).Error; err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Finding model.ProjectBillingAccountMapping failed using PropjectId")

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &pBAM, http.StatusFound
}
