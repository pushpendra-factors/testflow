package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"sort"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) DisableFiveTranMapping(ProjectID uint64, Integration string, ConnectorId string) error {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	updatedFields := map[string]interface{}{
		"status": false,
	}
	dbErr := db.Model(&model.FivetranMappings{}).Where("project_id = ? AND integration = ? AND connector_id = ?", ProjectID, Integration, ConnectorId).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating fivetran mappings failed")
		return dbErr
	}
	return nil
}

func (store *MemSQL) EnableFiveTranMapping(ProjectID uint64, Integration string, ConnectorId string, Accounts string) error {
	// check if not any for a project-id/source is status = true
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	allActiveMapping, err := store.GetAllActiveFiveTranMapping(ProjectID, Integration)
	if err != nil {
		logCtx.WithError(err).Error("getting active fivetran mappings failed")
		return err
	}
	if len(allActiveMapping) > 0 {
		return errors.New("Already active integration exists")
	}
	updatedFields := map[string]interface{}{
		"status":   true,
		"accounts": Accounts,
	}
	dbErr := db.Model(&model.FivetranMappings{}).Where("project_id = ? AND integration = ? AND connector_id = ?", ProjectID, Integration, ConnectorId).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating fivetran mappings failed")
		return dbErr
	}
	return nil
}

func (store *MemSQL) UpdateFiveTranMappingAccount(ProjectID uint64, Integration string, ConnectorId string, Accounts string) error {
	// check if not any for a project-id/source is status = true
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	updatedFields := map[string]interface{}{
		"accounts": Accounts,
	}
	dbErr := db.Model(&model.FivetranMappings{}).Where("project_id = ? AND integration = ? AND connector_id = ?", ProjectID, Integration, ConnectorId).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating fivetran mappings failed")
		return dbErr
	}
	return nil
}

func (store *MemSQL) GetFiveTranMapping(ProjectID uint64, Integration string) (string, error) {
	db := C.GetServices().Db
	var records []model.FivetranMappings
	if err := db.Where("project_id = ?", ProjectID).Where("integration = ?", Integration).Find(&records).Error; err != nil {
		log.Error(err)
		return "", err
	}
	if len(records) > 0 {
		for _, record := range records {
			if record.Status == true {
				return record.ConnectorID, nil
			}
		}
		return "", errors.New("Active mapping not found")
	} else {
		return "", errors.New("Mapping not found")
	}
}

func (store *MemSQL) GetActiveFiveTranMapping(ProjectID uint64, Integration string) (model.FivetranMappings, error) {
	db := C.GetServices().Db
	var records model.FivetranMappings
	if err := db.Where("project_id = ?", ProjectID).Where("integration = ?", Integration).Where("status = ?", true).First(&records).Error; err != nil {
		log.Error(err)
		return records, err
	}
	return records, nil
}

func (store *MemSQL) GetLatestFiveTranMapping(ProjectID uint64, Integration string) (string, error) {
	db := C.GetServices().Db
	var records []model.FivetranMappings
	if err := db.Where("project_id = ?", ProjectID).Where("integration = ?", Integration).Find(&records).Error; err != nil {
		log.Error(err)
		return "", err
	}
	if len(records) > 0 {
		sort.Slice(records, func(i, j int) bool {
			return (*(records[i].UpdatedAt)).After(*(records[j].UpdatedAt))
		})
		return records[0].ConnectorID, nil
	} else {
		return "", errors.New("Mapping not found")
	}
}

func (store *MemSQL) GetAllActiveFiveTranMapping(ProjectID uint64, Integration string) ([]string, error) {
	db := C.GetServices().Db
	var records []model.FivetranMappings
	if err := db.Where("project_id = ?", ProjectID).Where("integration = ?", Integration).Where("status = ?", true).Find(&records).Error; err != nil {
		log.Error(err)
		return nil, err
	}
	mappings := make([]string, 0)
	if len(records) > 0 {
		for _, record := range records {
			if record.Status == true {
				mappings = append(mappings, record.ConnectorID)
			}
		}
		return mappings, nil
	} else {
		return mappings, nil
	}
}

func (store *MemSQL) GetAllActiveFiveTranMappingByIntegration(Integration string) ([]model.FivetranMappings, error) {
	db := C.GetServices().Db
	var records []model.FivetranMappings
	if err := db.Where("integration = ?", Integration).Where("status = ?", true).Find(&records).Error; err != nil {
		log.Error(err)
		return nil, err
	}
	return records, nil
}

func (store *MemSQL) PostFiveTranMapping(ProjectID uint64, Integration string, ConnectorId string, SchemaId string, Accounts string) error {
	// Check if there is an active mapping already
	db := C.GetServices().Db
	allActiveMapping, err := store.GetAllActiveFiveTranMapping(ProjectID, Integration)

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	if err != nil {
		logCtx.WithError(err).Error("getting active fivetran mappings failed")
		return err
	}
	if len(allActiveMapping) > 0 {
		return errors.New("Already active integration exists")
	}
	transTime := gorm.NowFunc()

	mapping := model.FivetranMappings{
		ID:          U.GetUUID(),
		ProjectID:   ProjectID,
		Integration: Integration,
		ConnectorID: ConnectorId,
		SchemaID:    SchemaId,
		Accounts:    Accounts,
		CreatedAt:   &transTime,
		UpdatedAt:   &transTime,
	}

	if err := db.Create(&mapping).Error; err != nil {

		logCtx.WithError(err).Error("Insert into fivetran mapping table failed")
		return err
	}
	return nil
}
