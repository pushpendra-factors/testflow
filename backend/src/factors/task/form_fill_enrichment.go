package task

import (
	"fmt"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"

	"factors/config"
	"factors/model/model"
	"factors/model/store"

	SDK "factors/sdk"
	U "factors/util"
)

type UpdateTimestamp struct {
	First int64
	Last  int64
}

func FormFillProcessing() int {
	projectIdWithToken, errCode := store.GetStore().GetFormFillEnabledProjectIDWithToken()
	if errCode == http.StatusInternalServerError {
		log.Error("Failed to get projectids form fill event by ID. Invalid parameters")
		return http.StatusNotFound
	} else if errCode == http.StatusNotFound {
		log.Error("No projects have enabled forms fills.")
		return http.StatusNotFound
	}

	projectIDs := U.GetKeysOfInt64StringMap(projectIdWithToken)
	rowsUpadtedBeforeTenMinutes, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	if err != nil {
		log.Error("Failed to get projectids  form fill event by ID. Invalid parameters")
		return http.StatusNotFound
	}

	rowsByForm := make(map[string][]model.FormFill, 0)
	for _, r := range rowsUpadtedBeforeTenMinutes {
		key := fmt.Sprintf("%d.%s.%s", r.ProjectID, r.UserId, r.FormId)
		if _, ok := rowsByForm[key]; !ok {
			rowsByForm[key] = make([]model.FormFill, 0)
		}
		rowsByForm[key] = append(rowsByForm[key], r)
	}

	// Sorts all entries of form by timestamp
	rowsByFormSorted := make(map[string][]model.FormFill)
	for key, rows := range rowsByForm {
		sort.Sort(structSorter(rows))
		rowsByFormSorted[key] = rows
	}

	for _, formFills := range rowsByFormSorted {
		if formFills == nil {
			return http.StatusNotFound
		}

		rowsByField := map[string]*model.FormFill{}
		timestampUpdatesMap := map[string]*UpdateTimestamp{}

		properties := make(U.PropertiesMap)
		// Difference between first field entry - last field entry in seconds.
		properties[U.EP_TIME_SPENT_ON_FORM] = int64(formFills[0].CreatedAt.Sub(formFills[len(formFills)-1].CreatedAt).Seconds())

		// Selects one row with valid phone or email for each field on a form.
		for rowIndex := range formFills {
			row := formFills[rowIndex]
			key := fmt.Sprintf("%s.%s", row.FormId, row.FieldId)

			if _, exist := rowsByField[key]; !exist {
				rowsByField[key] = &model.FormFill{}
				timestampUpdatesMap[key] = &UpdateTimestamp{First: row.CreatedAt.UTC().Unix()}
			}
			prevValue := rowsByField[key].Value

			if (U.IsEmail(row.Value)) && U.IsBetterEmail(prevValue, row.Value) {
				rowsByField[key] = &row
			}
			if (U.IsValidPhone(row.Value)) && U.IsBetterPhone(prevValue, row.Value) {
				rowsByField[key] = &row
			}

			timestampUpdatesMap[key].Last = row.CreatedAt.UTC().Unix()

			store.GetStore().DeleteFormFillProcessedRecords(row.ProjectID, row.UserId, row.FormId, row.FieldId)
		}

		// Track form fills for selected rows.
		for field, row := range rowsByField {
			if row.Value == "" {
				continue
			}

			var hasValidValue bool
			var email string
			if U.IsEmail(row.Value) {
				email = row.Value
				properties[U.UP_EMAIL] = email
				hasValidValue = true
			}
			if U.IsValidPhone(row.Value) {
				properties[U.UP_PHONE] = row.Value
				hasValidValue = true
			}

			// Special property to check the captured value.
			properties[U.EP_FORM_FIELD_VALUE] = row.Value

			if !hasValidValue {
				continue
			}

			properties[U.EP_TIME_SPENT_ON_FORM_FIELD] = timestampUpdatesMap[field].Last - timestampUpdatesMap[field].First

			logCtx := log.WithFields(log.Fields{"project_id": row.ProjectID})

			// Add Page Event Properties to properties.
			if row.EventProperties != nil {
				pageProperties, err := U.DecodePostgresJsonbAsPropertiesMap(row.EventProperties)
				if err != nil {
					logCtx.WithError(err).Error("Failed decode event properties into properties_map.")
				} else {
					for k, v := range *pageProperties {
						properties[k] = v
					}
				}
			}

			projectToken := (*projectIdWithToken)[row.ProjectID]
			trackPayload := &SDK.TrackPayload{
				ProjectId:       row.ProjectID,
				UserId:          row.UserId,
				Name:            U.EVENT_NAME_FORM_FILL,
				Timestamp:       row.CreatedAt.UTC().Unix(),
				Auto:            false,
				RequestSource:   model.UserSourceWeb,
				EventProperties: properties,
			}
			logCtx = logCtx.WithFields(log.Fields{"track_payload": trackPayload})
			errCode, _ := SDK.TrackWithQueue(projectToken, trackPayload, config.GetSDKRequestQueueAllowedTokens())
			if errCode != http.StatusOK {
				logCtx.
					WithField("payload", trackPayload).WithField("err_code", errCode).
					Error("Failed to track form fill.")
				continue
			}

			logCtx.WithField("tag", "form_fill_debug").WithField("user_id", row.UserId).Info("Tracked form fills")
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
	return a[i].CreatedAt.After(a[j].CreatedAt)
}
