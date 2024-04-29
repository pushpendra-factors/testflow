package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreatePropertyMapping(propertyMapping model.PropertyMapping) (*model.PropertyMapping, string, int) {
	logFields := log.Fields{
		"property_mappings": propertyMapping,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if propertyMapping.ProjectID == 0 {
		logCtx.WithField("propertyMapping", propertyMapping).Warn("Invalid project ID for property mappin")
		return &model.PropertyMapping{}, "Invalid project ID for property mappin", http.StatusBadRequest
	}
	_, _, statusCode := store.GetPropertyMappingByProjectIDAndName(propertyMapping.ProjectID, propertyMapping.Name)
	if statusCode == http.StatusOK {
		logCtx.WithField("propertyMapping", propertyMapping).Warn("Failed to create property mapping. Duplicate record")
		return &model.PropertyMapping{}, "Failed to create property mapping. Duplicate record", http.StatusConflict
	}

	propertyMapping.ID = uuid.New().String()

	db := C.GetServices().Db
	if err := db.Create(&propertyMapping).Error; err != nil {
		if IsDuplicateRecordError(err) {
			log.WithError(err).WithField("propertyMapping", propertyMapping).Warn("Failed to create property mapping. Duplicate record")
			return &model.PropertyMapping{}, err.Error(), http.StatusConflict
		}
		logCtx.WithError(err).WithField("propertyMapping", propertyMapping).Error("Failed while creating Property Mapping")
		return &model.PropertyMapping{}, err.Error(), http.StatusInternalServerError
	}

	return &propertyMapping, "", http.StatusCreated
}

func (store *MemSQL) GetPropertyMappingByID(projectID int64, id string) (*model.PropertyMapping, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.Error("Invalid project ID for Property Mapping")
		return nil, "Invalid project ID for Property Mapping", http.StatusBadRequest
	}

	db := C.GetServices().Db
	var propertyMapping model.PropertyMapping
	if err := db.Where("project_id = ? AND id = ? AND is_deleted = ?", projectID, id, false).First(&propertyMapping).Error; err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error("Failed while retrieving Property Mapping")
		return nil, err.Error(), http.StatusInternalServerError
	}

	return &propertyMapping, "", http.StatusOK
}

// Creates a map of display category to property for a given project and property mapping name
// This map can be used to get the property for a given display category during query execution
func (store *MemSQL) GetDisplayCategoryToPropertiesByProjectIDAndPropertyMappingName(projectID int64, name string) (map[string]model.Property, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"name":       name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.Error("Invalid project ID for Property Mapping")
		return nil, "Invalid project ID for Property Mapping", http.StatusBadRequest
	}

	propertyMapping, errStr, errCode := store.GetPropertyMappingByProjectIDAndName(projectID, name)
	if errCode != http.StatusOK {
		logCtx.WithField("project_id", projectID).WithField("err_code", errCode).Error("Failed while retrieving Property Mapping")
		return nil, errStr, errCode
	}

	var properties []model.Property
	displayCategoryToPropertiesMap := make(map[string]model.Property)
	err := U.DecodePostgresJsonbToStructType(propertyMapping.Properties, &properties)
	if err != nil {
		log.WithError(err).Error("Failed while decoding property mapping properties")
		return displayCategoryToPropertiesMap, err.Error(), http.StatusInternalServerError
	}

	for _, property := range properties {
		displayCategoryToPropertiesMap[property.DisplayCategory] = property
	}
	return displayCategoryToPropertiesMap, "", http.StatusOK
}

func (store *MemSQL) GetPropertyMappingByProjectIDAndName(projectID int64, name string) (*model.PropertyMapping, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"name":       name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.Error("Invalid project ID for Property Mapping")
		return nil, "Invalid project ID for Property Mapping", http.StatusBadRequest
	}

	db := C.GetServices().Db
	var propertyMapping model.PropertyMapping
	if err := db.Where("project_id = ? AND name = ? AND is_deleted = ?", projectID, name, false).First(&propertyMapping).Error; err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error("Failed while retrieving Property Mapping")
		return nil, err.Error(), http.StatusInternalServerError
	}

	return &propertyMapping, "", http.StatusOK
}

func (store *MemSQL) GetPropertyMappingsByProjectId(projectID int64) ([]*model.PropertyMapping, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.Error("Invalid project ID for Property Mappings")
		return nil, "Invalid project ID for Property Mappings", http.StatusBadRequest
	}

	db := C.GetServices().Db
	var propertyMappings []*model.PropertyMapping
	if err := db.Where("project_id = ? AND is_deleted = ?", projectID, false).Find(&propertyMappings).Error; err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error("Failed while retrieving Property Mappings")
		return nil, err.Error(), http.StatusInternalServerError
	}

	return propertyMappings, "", http.StatusOK
}

func (store *MemSQL) GetPropertyMappingsByProjectIdAndSectionBitMap(projectID int64, sectionBitMap int64) ([]*model.PropertyMapping, string, int) {
	logFields := log.Fields{
		"project_id":     projectID,
		"section_bitMap": sectionBitMap,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.Error("Invalid project ID for Property Mappings")
		return nil, "Invalid project ID for Property Mappings", http.StatusBadRequest
	}

	db := C.GetServices().Db
	var propertyMappings []*model.PropertyMapping
	if err := db.Where("project_id = ? AND is_deleted = ? AND section_bit_map & ? = ?", projectID, false, sectionBitMap, sectionBitMap).Find(&propertyMappings).Error; err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error("Failed while retrieving Property Mappings")
		return nil, err.Error(), http.StatusInternalServerError
	}

	return propertyMappings, "", http.StatusOK
}

// Returns a array of string containing display_category from properties json.
func (store *MemSQL) GenerateSectionBitMapFromProperties(properties []model.Property, projectID int64) (int64, string) {

	displayCategories := make([]string, 0)
	for _, property := range properties {
		displayCategories = append(displayCategories, property.DisplayCategory)
	}

	return store.GenerateSectionBitMap(displayCategories, projectID)
}

// Takes list of display category
// Returns sectionBitMap
// Binary bits are marked based on display_category from properties from left to right.
func (store *MemSQL) GenerateSectionBitMap(displayCategories []string, projectID int64) (int64, string) {

	sources, statusCode := store.GetCustomAdsSourcesByProject(projectID)
	if statusCode != http.StatusOK {
		return 0, "Error during get of custom ads source"
	}

	sectionBitMap := int64(0)
	for _, displayCategory := range displayCategories {
		log.WithField("displayCategory", displayCategory).Warn("kark2")
		bitPosition, present := model.SectionBitMapping[displayCategory]
		if present {
			log.WithField("displayCategory", displayCategory).WithField("present", present).Warn("kark2-1")
			// Mark the bit as per position of display category
			sectionBitMap = sectionBitMap | (1 << (bitPosition - 1))
		} else {
			if U.ContainsStringInArray(sources, displayCategory) {
				baseForCustomAds := 30
				for i, source := range sources {
					if source == displayCategory {
						log.WithField("displayCategory", displayCategory).WithField("source", source).WithField("sources", sources).Warn("kark2-2")
						bitPosition = baseForCustomAds + i
						sectionBitMap = sectionBitMap | (1 << (bitPosition - 1))
					}

				}
			} else {
				return 0, "Invalid object type for property mapping"
			}

		}
	}

	return sectionBitMap, ""
}

// Soft deletes the property mapping by ID
func (store *MemSQL) DeletePropertyMappingByID(projectID int64, id string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		logCtx.Error("Invalid project ID for Property Mapping")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Model(&model.PropertyMapping{}).Where("project_id = ? AND id = ?", projectID, id).Update("is_deleted", true).Error; err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error("Failed while deleting Property Mapping")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
