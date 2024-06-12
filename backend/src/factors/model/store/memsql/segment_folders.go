package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Checking If Folder by name or id Already Exists or NOT
func (store *MemSQL) isFolderExists(projectID int64, name string, folderType string, id string ) bool {


	var tmpCount int64
	
	logFields := log.Fields{
		"project_id": projectID,
		"name":    name,
		"id": id,
	}
	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return false
	}

	db := C.GetServices().Db
	if(id != ""){
		db.Model(&model.SegmentFolder{}).Where("project_id = ? and id = ? and folder_type = ?", projectID, id, folderType).Count(&tmpCount)
	}else if(name != ""){
		db.Model(&model.SegmentFolder{}).Where("project_id = ? and name = ? and folder_type = ?", projectID, name, folderType).Count(&tmpCount)
	}
	if tmpCount > 0{
		return true
	}
	log.WithFields(logFields).Error("Segment Folder Not Found")
	return false
}


func (store *MemSQL) CreateSegmentFolder(projectID int64, name string, folderType string) int {
	
	
	if name == ""{
		return http.StatusBadRequest
	}
	
	logFields := log.Fields{
		"project_id": projectID,
		"name":    name,
	}

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return http.StatusBadRequest
	}

	if store.isFolderExists(projectID, name, folderType, ""){
		return http.StatusConflict
	}

	folder := model.SegmentFolder{
		Name: name,
		ProjectId: projectID,
		FolderType: folderType,
		Id: U.GetUUID(),
	}
	db := C.GetServices().Db
	// Creating New Segment Folder
	dbx := db.Create(&folder)
	if dbx.Error != nil {
		log.WithFields(logFields).WithError(dbx.Error).Error("Error Creating Segment Folder")
		return http.StatusConflict
	}
	
	return http.StatusCreated
}


func (store *MemSQL) GetAllSegmentFolders(projectID int64, folderType string) ([]model.SegmentFolder, int) {
	
	

	logFields := log.Fields{
		"project_id": projectID,
		"folderType": folderType,
	}
	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return nil, http.StatusBadRequest
	}
	var segmentFolders []model.SegmentFolder

	db := C.GetServices().Db
	err := db.Model(&model.SegmentFolder{}).Where("project_id = ? and folder_type = ?", projectID, folderType).Find(&segmentFolders).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed while getting all Segment Folders by ProjectId.")
		return nil, http.StatusInternalServerError
	}
	return segmentFolders, http.StatusFound
}

func (store *MemSQL) UpdateSegmentFolderByID(projectID int64, id string, name string, folderType string) int {
	if(name == ""){
		return http.StatusBadRequest
	}
	logFields := log.Fields{
		"project_id": projectID,
		"id":    id,
		"name": name,
		"folderType": folderType,
	}
	db := C.GetServices().Db

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return http.StatusBadRequest
	}

	if store.isFolderExists(projectID, name, folderType, "") {
		return http.StatusConflict
	}
	// Updating SegmentFolder
	folder := model.SegmentFolder{Name: name}
	err := db.Model(&model.SegmentFolder{}).Where("project_id = ? and id = ? and folder_type = ?", projectID, id, folderType).Update(folder).Error
	if err != nil {
		return http.StatusInternalServerError
	}
	
	return http.StatusAccepted
}

func (store *MemSQL) DeleteSegmentFolderByID(projectID int64, id string, folderType string) int {
	
	
	var err error
	logFields := log.Fields{
		"project_id": projectID,
		"id":    id,
		"folderType": folderType,
	}

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return http.StatusBadRequest
	}
	db := C.GetServices().Db
	// Update segment.folder_id = ''
	err =  db.Exec("UPDATE segments SET folder_id = '' WHERE project_id = ? and folder_id = ?", projectID, id).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to update segments")
		return http.StatusInternalServerError
	}
	
	// Delete Segment Folder
	err = db.Where("project_id = ? and id = ? and folder_type = ?",projectID, id, folderType).Delete(&model.SegmentFolder{}).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to Delete Segment Folders")
		return http.StatusInternalServerError
	}
	

	return http.StatusAccepted
}

func (store *MemSQL) MoveSegmentFolderItem(projectID int64, segmentID string, folderID string, folderType string) int {
	
	logFields := log.Fields{
		"project_id": projectID,
		"segmentID": segmentID,
		"folderID": folderID,
		"folderType": folderType,
	}
	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return http.StatusBadRequest
	}
	if folderID != "" {
		// Checking If Folder exists or not.
		if store.isFolderExists(projectID, "", folderType, folderID) == false {
			return http.StatusNotFound
		}
	}
	
	db := C.GetServices().Db
	// Updating segment with folderID
	err := db.Exec("UPDATE segments SET folder_id = ? WHERE project_id = ? and id = ?", folderID, projectID, segmentID).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to move segment")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
func (store *MemSQL) MoveSegmentToNewFolder(projectID int64, segmentID string, folderName string, folderType string) int {


	var folder model.SegmentFolder
	logFields := log.Fields{
		"project_id": projectID,
		"segmentID": segmentID,
		"folderName": folderName,
		"folderType": folderType,
	}
	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		log.WithFields(logFields).Error("Invalid Profile Type")
		return http.StatusBadRequest
	}
	errCode := store.CreateSegmentFolder(projectID, folderName, folderType)
	if errCode != http.StatusCreated {
		// Can't create Folder
		log.WithFields(logFields).Error("Failed to create Segment Folder")
		return http.StatusConflict
	}
	db := C.GetServices().Db
	err := db.Model(&model.SegmentFolder{}).Where("project_id = ? and name = ? and folder_type = ?", projectID, folderName, folderType).Find(&folder).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to get newly created segment Folder")
		return http.StatusInternalServerError
	}

	err = db.Exec("UPDATE segments SET folder_id = ? WHERE project_id = ? and id = ?", folder.Id, projectID, segmentID).Error 
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to Update Segment Folder")
		return http.StatusInternalServerError
	}
	
	return http.StatusAccepted
}