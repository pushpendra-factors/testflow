package handler

import (
	C "factors/config"
	mid "factors/middleware"
	"factors/model/store"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"factors/model/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
// GetUserHandler godoc
// @Summary Get a user for the given project and user id.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} model.User
// @Router /{project_id}/users/{user_id} [get]
func GetUserHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("GetUser Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Params.ByName("user_id")
	if id == "" {
		log.WithFields(log.Fields{"project_id": projectId}).Error("GetUser Failed. Missing id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, errCode := store.GetStore().GetUser(projectId, id)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, user)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users
// curl -i -X GET http://localhost:8080/projects/1/users?offset=50&limit=10
// GetUsersHandler godoc
// @Summary Gets users for the given project id.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param offset query integer false "Offset"
// @Param limit query integer false "Limit"
// @Success 200 {array} model.User
// @Router /{project_id}/users [get]
func GetUsersHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	qParams := c.Request.URL.Query()

	var offset uint64 = 0
	offsets := qParams["offset"]
	if offsets != nil {
		offsetStr := offsets[0]
		if offsetParse, err := strconv.ParseUint(offsetStr, 10, 64); err != nil {
			log.WithError(err).Error("GetUsers Failed. Offset parse failed.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			offset = offsetParse
		}
	}

	var limit uint64 = 10
	limits := qParams["limit"]
	if limits != nil {
		limitStr := limits[0]
		if limitParse, err := strconv.ParseUint(limitStr, 10, 64); err != nil {
			log.WithError(err).Error("GetUsers Failed. Limit parse failed.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			limit = limitParse
		}
	}

	users, errCode := store.GetStore().GetUsers(projectId, offset, limit)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, users)
	}
}

// GetUserPropertiesHandler Test command.
// curl -i -X GET http://localhost:8080/projects/1/user_properties
// GetUserPropertiesHandler godoc
// @Summary Gets users properties for the given project id.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "map[string]string"
// @Router /{project_id}/user_properties [get]
func GetUserPropertiesHandler(c *gin.Context) {
	// NOTE: Change in GetRequiredUserPropertiesByProject when this changes.
	var err error
	var properties map[string][]string
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	isExplain := c.Query("is_explain")
	version := c.Query("version")
	isDisplayNameEnabled := c.Query("is_display_name_enabled")
	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	if isExplain != "true" {
		properties, err = store.GetStore().GetUserPropertiesByProject(projectId, 2500,
			C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.WithError(err).Error("Failed to get user properties by project.")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if len(properties) == 0 {
			logCtx.WithError(err).Warn("No user properties Returned.")
		}
	} else {
		var status int
		var errMsg string
		properties, status, errMsg = getUserPropertiesFromPatternServer(projectId, modelId)
		if status != 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  errMsg,
				"status": status,
			})
			return
		}
	}

	//removing all account properties from user_properties
	properties = removeAllAccountPropertiesFromUserProperties(&properties)

	properties = U.ClassifyDateTimePropertyKeys(&properties)
	U.FillMandatoryDefaultUserProperties(&properties)
	_, overrides := store.GetStore().GetPropertyOverridesByType(projectId, U.PROPERTY_OVERRIDE_BLACKLIST, model.GetEntity(true))
	U.FilterDisabledCoreUserProperties(overrides, &properties)

	//adding constant properties
	appendCustomColumnProperties(&properties)

	if isDisplayNameEnabled == "true" {
		_, displayNames := store.GetStore().GetDisplayNamesForAllUserProperties(projectId)
		standardProperties := U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES
		displayNamesOp := make(map[string]string)
		for property, displayName := range standardProperties {
			displayNamesOp[property] = strings.Title(displayName)
		}
		for property, displayName := range displayNames {
			displayNamesOp[property] = strings.Title(displayName)
		}

		_, displayNames = store.GetStore().GetDisplayNamesForObjectEntities(projectId)
		for property, displayName := range displayNames {
			displayNamesOp[property] = strings.Title(displayName)
		}
		for _, props := range properties {
			for _, prop := range props {
				displayName := U.CreateVirtualDisplayName(prop)
				_, exist := displayNamesOp[prop]
				if !exist {
					displayNamesOp[prop] = displayName
				}
			}
		}
		dupCheck := make(map[string]bool)
		for _, name := range displayNamesOp {
			_, exists := dupCheck[name]
			if exists {
				logCtx.Warningf(fmt.Sprintf("Duplicate display name %s", name))
			}
			dupCheck[name] = true
		}
		if version != "2" {
			c.JSON(http.StatusOK, gin.H{"properties": U.FilterEmptyKeysAndValues(projectId, properties), "display_names": U.FilterDisplayNameEmptyKeysAndValues(projectId, displayNamesOp), "disabled_event_user_properties": U.DISABLED_EVENT_USER_PROPERTIES})
		} else {
			c.JSON(http.StatusOK, gin.H{"properties": model.CategorizeProperties(U.FilterEmptyKeysAndValues(projectId, properties), "user"), "display_names": U.FilterDisplayNameEmptyKeysAndValues(projectId, displayNamesOp), "disabled_event_user_properties": U.DISABLED_EVENT_USER_PROPERTIES})
		}
		return
	}
}

func appendCustomColumnProperties(properties *map[string][]string) {

	if _, exists := (*properties)["categorical"]; exists {
		(*properties)["categorical"] = append((*properties)["categorical"], U.USER_PROPERTIES_WITH_COLUMN...)
	}

}

func removeAllAccountPropertiesFromUserProperties(properties *map[string][]string) map[string][]string {
	newProperties := make(map[string][]string)

	//make map of all account properties
	accountPropMap := make(map[string]bool)
	for _, accountProp := range U.ALL_ACCOUNTS_PROPERTIES {
		accountPropMap[accountProp] = true
	}
	for _, accountProp := range U.ALL_ACCOUNT_DEFAULT_PROPERTIES {
		accountPropMap[accountProp] = true
	}

	for key, props := range *properties {
		for _, p := range props {
			if !accountPropMap[p] {
				newProperties[key] = append(newProperties[key], p)
			}
		}
	}

	return newProperties
}

// GetUserPropertyValuesHandler curl -i -X GET http://localhost:8080/projects/1/user_properties/$country
// GetUserPropertiesHandler godoc
// @Summary Get property values for given property name.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param property_name path string true "Property Name"
// @Success 200 {string} json "[]string"
// @Router /{project_id}/user_properties/{property_name}/values [get]
func GetUserPropertyValuesHandler(c *gin.Context) {
	var err error
	var propertyValues []string
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	isExplain := c.Query("is_explain")
	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		logCtx.WithField("property_name", propertyName).Error("null propertyname")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	//updating property name if inproperties
	appendCustomColumnPropertyNames(&propertyName)

	if isExplain != "true" {
		propertyValues, err = store.GetStore().GetPropertyValuesByUserProperty(projectId, propertyName, 2500, C.GetLookbackWindowForEventUserCache())
		if err != nil {
			logCtx.WithError(err).Error("get property values by user property")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if len(propertyValues) == 0 {
			logCtx.WithError(err).Error(fmt.Sprintf("No user property values Returned - ProjectID - %v, propertyName - %s", projectId, propertyName))
		}
	} else {
		var status int
		var errMsg string
		propertyValues, status, errMsg = getUserPropertyValuesFromPatternServer(projectId, modelId, propertyName)
		if status != 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  errMsg,
				"status": status,
			})
			return
		}
	}

	label := c.Query("label")
	if label == "true" {
		propertyValueLabel, err, isSourceEmpty := store.GetStore().GetPropertyValueLabel(projectId, propertyName, propertyValues)
		if err != nil {
			logCtx.WithError(err).Error("get user properties labels and values by property name")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if isSourceEmpty {
			logCtx.WithField("property_name", propertyName).Warning("source is empty")
		}

		if len(propertyValueLabel) == 0 {
			logCtx.WithField("property_name", propertyName).Error("No user property value labels Returned")
		}

		c.JSON(http.StatusOK, U.FilterDisplayNameEmptyKeysAndValues(projectId, propertyValueLabel))
		return
	}

	propertyValues = U.SortRangePropertyValues(propertyName, propertyValues)
	c.JSON(http.StatusOK, U.FilterEmptyArrayValues(propertyValues))
}

func appendCustomColumnPropertyNames(propertyName *string) {

	if _, exists := model.USER_PROPERTY_TO_COLUMN_MAP[*propertyName]; exists {
		*propertyName = model.USER_PROPERTY_TO_COLUMN_MAP[*propertyName]
	}
}
func getUserPropertyValuesFromPatternServer(projectId int64, modelId uint64, propertyName string) ([]string, int, string) {
	propertyValues := make([]string, 0)
	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		return propertyValues, http.StatusBadRequest, err.Error()
	}
	userInfo := ps.GetUserAndEventsInfo()
	propertyValues = make([]string, 0)
	if userInfo.UserPropertiesInfo != nil {
		for property, values := range (*userInfo.UserPropertiesInfo).CategoricalPropertyKeyValues {
			if property == propertyName {
				for value, _ := range values {
					propertyValues = append(propertyValues, value)
				}
			}
		}
	}
	return propertyValues, 0, ""
}

func getUserPropertiesFromPatternServer(projectId int64, modelId uint64) (map[string][]string, int, string) {
	properties := make(map[string][]string)
	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		return properties, http.StatusBadRequest, err.Error()
	}
	userInfo := ps.GetUserAndEventsInfo()
	properties = make(map[string][]string)
	properties[U.PropertyTypeNumerical] = make([]string, 0)
	properties[U.PropertyTypeCategorical] = make([]string, 0)
	if userInfo.UserPropertiesInfo != nil {
		for property := range (*userInfo.UserPropertiesInfo).NumericPropertyKeys {
			properties[U.PropertyTypeNumerical] = append(properties[U.PropertyTypeNumerical], property)
		}

		for property := range (*userInfo.UserPropertiesInfo).CategoricalPropertyKeyValues {
			properties[U.PropertyTypeCategorical] = append(properties[U.PropertyTypeCategorical], property)
		}
	}
	return properties, 0, ""
}
