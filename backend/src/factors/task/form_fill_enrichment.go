package task

import (
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"

	"factors/model/model"
	"factors/model/store"

	U "factors/util"

	SDK "factors/sdk"
)

func FormFillProcessing() int {
	projectIds, err := store.GetStore().GetFormFillEnabledProjectIDs()
	if err != nil {
		log.Error("Failed to get projectids  form fill event by ID. Invalid parameters")
		return http.StatusNotFound
	}

	rowsUpadtedBeforeTenMinutes, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	if err != nil {
		log.Error("Failed to get projectids  form fill event by ID. Invalid parameters")
		return http.StatusNotFound
	}

	rowsByForm := make(map[string][]model.FormFill, 0)
	for _, makeMap := range rowsUpadtedBeforeTenMinutes {
		if _, ok := rowsByForm[makeMap.FormId]; !ok {
			rowsByForm[makeMap.FormId] = make([]model.FormFill, 0)
		}
		rowsByForm[makeMap.FormId] = append(rowsByForm[makeMap.FormId], makeMap)
	}
	sortedFieldUpdates := make(map[string][]model.FormFill)
	for key, rows := range rowsByForm {
		sort.Sort(structSorter(rows))
		sortedFieldUpdates[key] = rows
	}

	properties := make(U.PropertiesMap)
	for _, rows := range sortedFieldUpdates {
		if rows == nil {
			return http.StatusNotFound
		}
		properties[U.EP_TIME_SPENT_ON_FORM] = int64(rows[0].UpdatedAt.Sub(rows[len(rows)-1].UpdatedAt).Seconds())
		for _, fieldValue := range rows {
			if (U.IsEmail(fieldValue.Value)) && (properties[U.UP_EMAIL] == nil || U.IsBetterEmail(properties[U.UP_EMAIL].(string), fieldValue.Value)) {
				properties[U.UP_EMAIL] = fieldValue.Value
			}
			if (U.IsValidPhone(fieldValue.Value)) && (properties[U.UP_PHONE] == nil || U.IsBetterPhone((properties[U.UP_PHONE]).(string), fieldValue.Value)) {
				properties[U.UP_PHONE] = fieldValue.Value
			}
			trackPayload := SDK.TrackPayload{
				Name:            U.EVENT_NAME_FORM_FILL,
				Timestamp:       rows[0].UpdatedAt.UTC().Unix(),
				ProjectId:       fieldValue.ProjectID,
				Auto:            false,
				RequestSource:   model.UserSourceWeb,
				EventProperties: properties,
			}
			errCode, _ := SDK.Track(fieldValue.ProjectID, &trackPayload, false, SDK.SourceJSSDK, "")
			if errCode != http.StatusOK {
				return http.StatusBadRequest
			}
			_, err := store.GetStore().DeleteFormFillProcessedRecords(fieldValue.ProjectID, fieldValue.FormId, fieldValue.FieldId)
			if err != nil {
				log.Error("Failed to delete processed record ")
				return http.StatusBadRequest
			}
		}
	}
	return http.StatusOK
}

type structSorter []model.FormFill

func (a structSorter) Len() int {
	return len(a)
}
func (a structSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a structSorter) Less(i, j int) bool {
	return a[i].UpdatedAt.After(a[j].UpdatedAt)
}
