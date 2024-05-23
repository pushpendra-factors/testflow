package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const SECONDS_IN_A_WEEK int64 = 7 * 24 * 60 * 60

type UserCount struct {
	UniqueUserCount int64 `json:"unique_user_count"`
}

func (store *MemSQL) GetUniqueUsersCountWithFormSubmissionEvent(projectId int64, startTimeStamp int64, endTimeStamp int64) (UserCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var count UserCount
	formSubmittedEventName, errCode := store.GetEventName(U.EVENT_NAME_FORM_SUBMITTED, projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for form submitted event")
		return count, fmt.Errorf("failed to get event name id for form submitted event")
	}

	db := C.GetServices().Db
	err := db.Table("events").Select("COUNT(DISTINCT(user_id)) as unique_user_count").
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id =?", projectId, startTimeStamp, endTimeStamp, formSubmittedEventName.ID).
		Find(&count).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get unique user count for form submission event")
		return count, err
	}

	return count, nil

}

func (store *MemSQL) GetUniqueUsersCountWithWebsiteSessionEvent(projectId int64, startTimeStamp int64, endTimeStamp int64) (UserCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var count UserCount
	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for form submitted event")
		return count, fmt.Errorf("failed to get event name id for form submitted event")
	}

	db := C.GetServices().Db
	err := db.Table("events").Select("COUNT(DISTINCT(user_id)) as unique_user_count").
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id =?", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Find(&count).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get unique user count for form submission event")
		return count, err
	}

	return count, nil

}

type DomainCount struct {
	UniqueDomainCount int64 `json:"unique_domain_count"`
}

func (store *MemSQL) GetUniqueFactorsDeanonIdentifiedCompaniesCount(projectId, startTimeStamp, endTimeStamp int64) (DomainCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	var count DomainCount
	err := db.Table("events").Select("COUNT(DISTINCT(JSON_EXTRACT_STRING(user_properties,'$6Signal_domain'))) as unique_domain_count").
		Where("project_id= ? and timestamp>=? and timestamp<=?", projectId, startTimeStamp, endTimeStamp).Find(&count).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get unique domains count.")
		return count, err
	}

	return count, nil

}

func (store *MemSQL) GetTotalUniqueFactorsDeanonIdentifiedCompaniesCount(projectId int64) (DomainCount, error) {

	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	var count DomainCount
	err := db.Table("events").Select("COUNT(DISTINCT(JSON_EXTRACT_STRING(user_properties,'$6Signal_domain'))) as unique_domain_count").
		Where("project_id= ? ", projectId).Find(&count).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get total unique domains count.")
		return count, err
	}

	return count, nil

}

type propCount struct {
	PropValue   string `json:"prop_value"`
	CountDomain int64  `json:"count_domain"`
}

func (store *MemSQL) GetFactorsDeanonPropertyTopValueCount(propName string, projectId, startTimeStamp, endTimeStamp int64) (propCount, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
		"propName":       propName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var propertyCount propCount

	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for session event")
		return propertyCount, fmt.Errorf("failed to get event name id for session event")
	}

	db := C.GetServices().Db
	err := db.Limit(1).Table("events").
		Select("JSON_EXTRACT_STRING(user_properties, ?) as prop_value, COUNT(DISTINCT(JSON_EXTRACT_STRING(user_properties,'$6Signal_domain'))) as count_domain", propName).
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id = ? and prop_value IS NOT NULL", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Group("prop_value").Order("count_domain desc").Find(&propertyCount).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get the property top value and count.")
		return propertyCount, err
	}

	return propertyCount, nil

}

type SessionTopPropertyValue struct {
	PropValue string `json:"prop_value"`
}

func (store *MemSQL) GetPropertyTopValueForSessionEvent(propName string, projectId, startTimeStamp, endTimeStamp int64) (SessionTopPropertyValue, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
		"propName":       propName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var propTopValue SessionTopPropertyValue

	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for session event")
		return propTopValue, fmt.Errorf("failed to get event name id for session event")
	}

	db := C.GetServices().Db
	err := db.Limit(1).Table("events").Select("JSON_EXTRACT_STRING(properties, ?) as prop_value", propName).
		Where("project_id=? and timestamp>=? and timestamp<=? and event_name_id = ? and prop_value IS NOT NULL", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Group("prop_value").Order("COUNT(*) desc").Find(&propTopValue).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property top value for session event")
		return propTopValue, err
	}

	return propTopValue, nil
}

func (store *MemSQL) GetMostActiveCompanies(projectId, startTimeStamp, endTimeStamp int64) ([]model.ActiveCompanyProp, error) {

	logFields := log.Fields{
		"project_id":     projectId,
		"startTimestamp": startTimeStamp,
		"endTimestamp":   endTimeStamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var activeCompaniesProp []model.ActiveCompanyProp

	sessionEventName, errCode := store.GetEventName(U.EVENT_NAME_SESSION, projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get event name id for session event")
		return []model.ActiveCompanyProp{}, fmt.Errorf("failed to get event name id for session event")
	}

	db := C.GetServices().Db

	err := db.Limit(5).Table("events").
		Select("JSON_EXTRACT_STRING(user_properties,'$6Signal_domain') as domain,JSON_EXTRACT_STRING(user_properties,'$6Signal_industry') as industry,JSON_EXTRACT_STRING(user_properties,'$6Signal_country') as country,JSON_EXTRACT_STRING(user_properties,'$6Signal_revenue_range')  as revenue").
		Where("project_id= ? and timestamp>= ? and timestamp<= ? and event_name_id=? and domain IS NOT NULL and domain !='' ", projectId, startTimeStamp, endTimeStamp, sessionEventName.ID).
		Group("domain").Order("COUNT(*) desc").Find(&activeCompaniesProp).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get most active companies")
		return []model.ActiveCompanyProp{}, err
	}

	return activeCompaniesProp, nil

}

func (store *MemSQL) GetWeeklyMailmodoEmailsMetrics(projectId, startTimeStamp, endTimeStamp int64) (model.WeeklyMailmodoEmailMetrics, error) {

	var metrics model.WeeklyMailmodoEmailMetrics

	lastWeekStartTimeStamp := startTimeStamp - SECONDS_IN_A_WEEK
	lastWeekEndTimeStamp := endTimeStamp - SECONDS_IN_A_WEEK

	sessionUniqueUserCount, err := store.GetUniqueUsersCountWithWebsiteSessionEvent(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	formSubmittedUniqueUserCount, err := store.GetUniqueUsersCountWithFormSubmissionEvent(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	factorsDeanonIdentifiedCompaniesCount, err := store.GetUniqueFactorsDeanonIdentifiedCompaniesCount(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	factorsDeanonIdentifiedCompaniesCountLastWeek, err := store.GetUniqueFactorsDeanonIdentifiedCompaniesCount(projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	totalFactorsDeanonIdentifiedCompaniesCount, err := store.GetTotalUniqueFactorsDeanonIdentifiedCompaniesCount(projectId)
	if err != nil {
		return metrics, err
	}

	industryTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_INDUSTRY, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	industryTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_INDUSTRY, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	employeeRangeTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_EMPLOYEE_RANGE, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	employeeRangeTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_EMPLOYEE_RANGE, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	countryTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_COUNTRY, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	countryTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_COUNTRY, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	revenueRangeTopValue, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_REVENUE_RANGE, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	revenueRangeTopValueLastWeek, err := store.GetFactorsDeanonPropertyTopValueCount(U.SIX_SIGNAL_REVENUE_RANGE, projectId, lastWeekStartTimeStamp, lastWeekEndTimeStamp)
	if err != nil {
		return metrics, err
	}

	sessionCount, err := store.GetSessionCount(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	channelTopValue, err := store.GetPropertyTopValueForSessionEvent(U.EP_CHANNEL, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	campaignTopValue, err := store.GetPropertyTopValueForSessionEvent(U.EP_CAMPAIGN, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	sourceTopValue, err := store.GetPropertyTopValueForSessionEvent(U.EP_SOURCE, projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	activeCompaniesProp, err := store.GetMostActiveCompanies(projectId, startTimeStamp, endTimeStamp)
	if err != nil {
		return metrics, err
	}

	metrics = model.WeeklyMailmodoEmailMetrics{
		SessionUniqueUserCount:        sessionUniqueUserCount.UniqueUserCount,
		FormSubmittedUniqueUserCount:  formSubmittedUniqueUserCount.UniqueUserCount,
		IdentifiedCompaniesCount:      factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount,
		TotalIdentifiedCompaniesCount: totalFactorsDeanonIdentifiedCompaniesCount.UniqueDomainCount,
		Industry: model.TopValueAndCountObject{
			TopValue:                industryTopValue.PropValue,
			TopValuePercent:         (industryTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount,
			TopValueLastWeek:        industryTopValueLastWeek.PropValue,
			TopValuePercentLastWeek: (industryTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount,
		},

		Country: model.TopValueAndCountObject{
			TopValue:                countryTopValue.PropValue,
			TopValuePercent:         (countryTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount,
			TopValueLastWeek:        countryTopValueLastWeek.PropValue,
			TopValuePercentLastWeek: (countryTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount,
		},
		EmployeeRange: model.TopValueAndCountObject{
			TopValue:                employeeRangeTopValue.PropValue,
			TopValuePercent:         (employeeRangeTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount,
			TopValueLastWeek:        employeeRangeTopValueLastWeek.PropValue,
			TopValuePercentLastWeek: (employeeRangeTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount,
		},
		RevenueRange: model.TopValueAndCountObject{
			TopValue:                revenueRangeTopValue.PropValue,
			TopValuePercent:         (revenueRangeTopValue.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCount.UniqueDomainCount,
			TopValueLastWeek:        revenueRangeTopValueLastWeek.PropValue,
			TopValuePercentLastWeek: (revenueRangeTopValueLastWeek.CountDomain * 100) / factorsDeanonIdentifiedCompaniesCountLastWeek.UniqueDomainCount,
		},
		SessionCount: sessionCount.SessionCount,
		TopChannel:   channelTopValue.PropValue,
		TopSource:    sourceTopValue.PropValue,
		TopCampaign:  campaignTopValue.PropValue,
		Company1:     activeCompaniesProp[0],
		Company2:     activeCompaniesProp[1],
		Company3:     activeCompaniesProp[2],
		Company4:     activeCompaniesProp[3],
		Company5:     activeCompaniesProp[4],
	}

	return metrics, nil

}
